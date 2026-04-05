package controllers

import (
	"Focogram/global"
	"Focogram/models"
	"Focogram/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	cachekeyComment        = "comment:"       //comment:{commentid} ->Hash
	cachekeyPostCmts       = "post:comments:" //post:comments:{postid}->Sorted Set
	commentQueuekey        = "comment_queue"  //Redis队列，用于异步落库
	CommentbatchWriterOnce sync.Once
)

// 初始化评论批量写入器（启动时执行）
func InitCommentBatchWriter() {
	CommentbatchWriterOnce.Do(func() {
		go CommentBatchWriteToDB()
	})
}

// 批量将评论批量写入数据库
func CommentBatchWriteToDB() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	batch := make([]*models.Comment, 0, 50) //批量写入缓存
	for {
		select {
		case <-ticker.C:
			if len(batch) > 0 {
				saveCommentBatch(batch)
				batch = batch[:0]
			}
		default:
			//从redis队列读取数据（超时为5秒）
			result, err := global.Redis.BLPop(context.Background(), 5*time.Second, commentQueuekey).Result()
			if err != nil {
				if err != redis.Nil {
					log.Printf("failed to get comment from redis, got error: %v", err)
				}
				continue
			}
			// 成功获取到数据
			var comment models.Comment
			if err := json.Unmarshal([]byte(result[1]), &comment); err != nil {
				log.Printf("failed to unmarshal comment from redis, got error: %v", err)
				continue
			}
			batch = append(batch, &comment)

			// 达到批量大小立即保存
			if len(batch) >= 50 {
				log.Printf("达到批量大小，立即保存，数量: %d", len(batch))
				saveCommentBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// 保存批量评论到数据库
func saveCommentBatch(batch []*models.Comment) {
	if len(batch) == 0 {
		return
	}
	tx := global.Db.Begin()
	if tx.Error != nil {
		log.Printf("开启事务失败: %v", tx.Error)
		return
	}
	//批量插入，冲突时更新
	err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commentid"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"content": clause.Expr{SQL: "VALUES(content)"}, // 修复点：补充结构体大括号
		}),
	}).CreateInBatches(batch, 50).Error
	if err != nil {
		tx.Rollback()
		log.Printf("批量写入数据库失败: %v", err)
		return
	}
	if err := tx.Commit().Error; err != nil {
		log.Printf("提交事务失败: %v", err)
		return
	}
}

// 发布评论
func CreateComment(c *gin.Context) {
	InitCommentBatchWriter()
	var input models.Comment
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userid := c.GetString("userid")
	postid := c.Param("postid")
	if postid == "" || userid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数"})
		return
	}
	// 检查限流和防抖
	if err := utils.CheckRateLimitAndDebounce("comment", userid, postid); err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}
	// 判断帖文是否存在 - 加强验证
	exists, err := models.CheckPostExists(postid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖文不存在"})
		return
	}

	commentid := models.CreateCommentid()

	//1.写入Redis缓存
	//存储评论详情（Hash）
	commentkey := cachekeyComment + commentid
	if err := global.Redis.HSet(context.Background(), commentkey, map[string]interface{}{
		"commentid":  commentid,
		"content":    input.Content,
		"postid":     postid,
		"userid":     userid,
		"created_at": time.Now().Unix(),
	}).Err(); err != nil {
		log.Printf("写入Redis缓存失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	//2.加入帖子评论列表（Sorted Set,按时间排序）
	postCommentskey := cachekeyPostCmts + postid
	if err := global.Redis.ZAdd(context.Background(), postCommentskey, &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: commentid,
	}).Err(); err != nil {
		log.Printf("加入帖子评论列表失败：%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}
	global.Redis.Expire(context.Background(), postCommentskey, time.Hour*24)

	//3.推送至Redis队列，异步写入数据库
	comment := models.Comment{
		Commentid: commentid,
		Content:   input.Content,
		Postid:    postid,
		Userid:    userid,
	}
	data, _ := json.Marshal(comment)
	if err := global.Redis.RPush(context.Background(), commentQueuekey, data); err != nil {
		//推送至队列失败直接写入数据库
		go func() {
			if err := global.Db.Create(&comment).Error; err != nil {
				log.Printf("写入数据库失败：%v", err)
			}
		}()
	}
	// 发送评论通知
	go func() {
		postAuthorID, err := models.GetPostAuthor(postid)
		if err != nil || postAuthorID == userid {
			return
		}

		var commenter models.User
		if err := global.Db.Select("username").Where("userid = ?", userid).First(&commenter).Error; err != nil {
			return
		}

		shortContent := input.Content
		if len([]rune(shortContent)) > 20 {
			shortContent = string([]rune(shortContent)[:20]) + "..."
		}

		notification := &models.Notification{
			Userid:      postAuthorID,
			Senderid:    userid,
			ContentType: models.NotificationTypeComment,
			Contentid:   commentid,
			Content:     commenter.Username + " 评论了你的帖子: " + shortContent,
		}
		utils.PushToWebSocket(notification)
	}()

	c.JSON(http.StatusOK, gin.H{
		"message":   "评论成功",
		"content":   input.Content,
		"commentid": commentid,
	})
}

func GetPostComments(c *gin.Context) {
	postid := c.Param("postid")
	if postid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数"})
		return
	}
	exists, err := models.CheckPostExists(postid)
	if err != nil || !exists {
		c.JSON(http.StatusOK, gin.H{
			"total":    0,
			"comments": []interface{}{},
		})
		return
	}

	type CommentWithUsername struct {
		Commentid   string
		Content     string
		Userid      string
		Username    string
		AvatarUrl   string
		AvatarColor string
		CreatedAt   time.Time
	}

	var comments []CommentWithUsername
	var count int64

	// 先查总数
	if err := global.Db.Model(&models.Comment{}).Where("postid = ?", postid).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询评论总数失败"})
		return
	}

	// 直接从数据库查询，关联用户名，按时间正序（最早的在上面）
	err = global.Db.Table("comments").
		Select("comments.commentid, comments.content, comments.userid, comments.created_at, users.username, users.avatar_url, users.avatar_color").
		Joins("LEFT JOIN users ON comments.userid = users.userid").
		Where("comments.postid = ? AND comments.deleted_at IS NULL", postid).
		Order("comments.created_at ASC").
		Scan(&comments).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询评论失败"})
		return
	}

	// 格式化结果
	result := make([]map[string]interface{}, 0, len(comments))
	for _, cmt := range comments {
		// 格式化时间显示
		timeStr := cmt.CreatedAt.Format("2006-01-02 15:04")
		avatarURL := cmt.AvatarUrl
		if avatarURL != "" && !strings.HasPrefix(avatarURL, "http") {
			avatarURL = "http://localhost:8080" + avatarURL
		}
		result = append(result, map[string]interface{}{
			"commentid":   cmt.Commentid,
			"content":     cmt.Content,
			"userid":      cmt.Userid,
			"username":    cmt.Username,
			"avatarUrl":   avatarURL,
			"avatarColor": cmt.AvatarColor,
			"createdAt":   timeStr,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"total":    count,
		"comments": result,
	})
}

func DeleteComment(c *gin.Context) {
	commentid := c.Param("commentid")
	userid := c.GetString("userid")
	if userid == "" || commentid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问，请先登录"})
		return
	}

	//1.查询评论信息，验证验证存在性并获取相关ID
	var comment models.Comment
	if err := global.Db.Unscoped().Where("commentid=?", commentid).First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "评论不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误"})
		return
	}

	//2.验证权限
	isPostAuthor, err := utils.CheckPostAuthor(comment.Postid, userid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误"})
		return
	}
	if comment.Userid != userid && !isPostAuthor {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限删除此评论"})
		return
	}

	//3.开启事务删除数据库记录
	tx := global.Db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}
	if err := tx.Unscoped().Delete(&comment).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败：" + err.Error()})
		return
	}

	//4.删除缓存
	commentKey := cachekeyComment + commentid
	if err := global.Redis.Del(context.Background(), commentKey).Err(); err != nil {
		log.Printf("删除缓存失败：%v", err)
	}

	//4.2删除帖子评论有序集合中的记录
	postKey := cachekeyPostCmts + comment.Postid
	if err := global.Redis.ZRem(context.Background(), postKey, commentid).Err(); err != nil {
		log.Printf("删除缓存失败：%v", err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})

}
