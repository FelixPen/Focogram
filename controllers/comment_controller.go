package controllers

import (
	"Focogram/global"
	"Focogram/models"
	"Focogram/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	cachekeyComment  = "comment:"       //comment:{commentid} ->Hash
	cachekeyPostCmts = "post:comments:" //post:comments:{postid}->Sorted Set
	commentQueuekey  = "comment_queue"  //Redis队列，用于异步落库
)

// 初始化评论批量写入器（启动时执行）
func InitCommentBatchWriter() {
	go CommentBatchWriteToDB()
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
	initBatchWriter()
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
	c.JSON(http.StatusOK, gin.H{
		"message":   "评论成功",
		"content":   input.Content,
		"commentid": commentid,
	})
	postAuthorID, err := models.GetPostAuthor(postid)
	if err == nil && postAuthorID != userid { // 不是自己评论自己的帖子
		notification := &models.Notification{
			Userid:      postAuthorID,
			Senderid:    userid,
			ContentType: models.NotificationTypeComment,
			Contentid:   commentid,
			Content:     userid + "评论了你的帖文: " + input.Content,
		}
		go utils.PushToWebSocket(notification)
	}

}

func GetPostComments(c *gin.Context) {
	postid := c.Param("postid")
	if postid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数"})
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

	//2.解析分页参数（默认第1页，每页20条）
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	start := (page - 1) * size
	end := start + size - 1

	//3.构建缓存键
	key := cachekeyPostCmts + postid

	//4.先从redis缓存查询
	//4.1获取评论总数（用于分页）
	total, err := global.Redis.ZCard(context.Background(), key).Result()
	if err != nil {
		log.Printf("获取评论总数失败：%v", err)
	}
	//4.2获取当前页的评论id(按时间顺序，最新的在前面)
	commentids, err := global.Redis.ZRange(context.Background(), key, int64(start), int64(end)).Result()
	if err == nil && len(commentids) > 0 {
		//4.3批量获取评论详情
		comments := make([]map[string]interface{}, 0, len(commentids))
		for _, commentid := range commentids {
			cmtData, err := global.Redis.HGetAll(context.Background(), cachekeyComment+commentid).Result()
			if err != nil {
				log.Printf("获取评论详情失败：%v", err)
				continue
			}
			if len(cmtData) == 0 {
				continue
			}
			// 构建包含时间的评论数据（从缓存获取createdAt）
			comments = append(comments, map[string]interface{}{
				"commentid": cmtData["commentid"],
				"content":   cmtData["content"],
				"postid":    cmtData["postid"],
				"userid":    cmtData["userid"],
				"createdAt": cmtData["createdAt"], // 添加时间字段
			})
		}
		//4.4返回缓存结果
		c.JSON(http.StatusOK, gin.H{
			"total":   total,
			"page":    page,
			"size":    size,
			"content": comments,
		})
		return
	}
	//5.缓存未命中，从数据库查询
	var dbComments []models.Comment
	//5.1先查询总数
	var count int64
	if err := global.Db.Model(&models.Comment{}).Where("postid = ?", postid).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询评论总数失败"})
		return
	}

	//5.2查询您当前页数据
	if err := global.Db.Where("postid = ?", postid).
		Order("created_at DESC").
		Limit(size).Offset(start).
		Find(&dbComments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询评论失败"})
		return
	}

	//6.重建缓存
	pipeline := global.Redis.Pipeline()
	for _, cmt := range dbComments {
		//6.1存储评论详情（Hash）
		cmtKey := cachekeyComment + cmt.Commentid
		pipeline.HSet(context.Background(), cmtKey, map[string]interface{}{
			"commentid": cmt.Commentid,
			"content":   cmt.Content,
			"postid":    cmt.Postid,
			"userid":    cmt.Userid,
			"createdAt": cmt.CreatedAt.Unix(),
		})
		pipeline.Expire(context.Background(), cmtKey, 24*time.Hour) // 缓存1天
		//6.2添加到帖子评论有序集合
		pipeline.ZAdd(context.Background(), key, &redis.Z{
			Score:  float64(cmt.CreatedAt.Unix()),
			Member: cmt.Commentid,
		})
	}
	pipeline.Expire(context.Background(), key, 24*time.Hour) //集合缓存一天
	if _, err := pipeline.Exec(context.Background()); err != nil {
		log.Printf("重建缓存失败：%v", err)
	}

	// 7. 格式化数据库结果并返回
	result := make([]map[string]interface{}, 0, len(dbComments))
	for _, cmt := range dbComments {
		result = append(result, map[string]interface{}{
			"commentid": cmt.Commentid,
			"content":   cmt.Content,
			"userid":    cmt.Userid,
			"createdAt": cmt.CreatedAt.Unix(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"total":   count,
		"page":    page,
		"size":    size,
		"content": result,
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
	if err := global.Db.Where("commentid=?", commentid).First(&comment).Error; err != nil {
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
	if !isPostAuthor {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限删除此评论"})
		return
	}

	//3.开启事务删除数据库记录
	tx := global.Db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}
	if err := tx.Delete(&comment).Error; err != nil {
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
