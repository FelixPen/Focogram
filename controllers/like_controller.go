package controllers

// import (
// 	"Focogram/global"
// 	"Focogram/models"
// 	"Focogram/utils"
// 	"encoding/json"

// 	"context"
// 	"log"
// 	"net/http"

// 	"sync"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/go-redis/redis/v8"
// 	"gorm.io/gorm/clause"
// )

// var (
// 	cachekey     = "posts:"
// 	likeSyncPool = sync.Pool{
// 		New: func() interface{} {
// 			return &models.Like{}
// 		},
// 	}
// 	// 使用包级 sync.Once 确保只初始化一次
// 	batchWriterOnce sync.Once
// )

// // 利用Redis队列实现持久化
// var likeQueueKey = "like_queue"

// // 初始化批量写入器
// func initBatchWriter() {
// 	batchWriterOnce.Do(func() {
// 		go BatchWriteToDB()
// 	})
// }

// // 批量写入数据库（每5秒或累计50条时执行一次）
// func BatchWriteToDB() {
// 	ticker := time.NewTicker(5 * time.Second)
// 	defer ticker.Stop()

// 	batch := make([]*models.Like, 0, 50)

// 	for {
// 		select {
// 		case <-ticker.C:
// 			// 定时触发批量保存
// 			if len(batch) > 0 {
// 				saveBatch(batch)
// 				batch = batch[:0]
// 			}
// 		default:
// 			// 从redis队列读取数据，超时时间为5秒
// 			result, err := global.Redis.BLPop(context.Background(), 5*time.Second, likeQueueKey).Result()
// 			if err != nil {
// 				if err == redis.Nil {
// 					// 队列为空，继续等待
// 					continue
// 				}
// 				// 真正的Redis错误
// 				log.Printf("failed to get like from redis, got error: %v", err)
// 				continue
// 			}

// 			// 成功获取到数据
// 			var like models.Like
// 			if err := json.Unmarshal([]byte(result[1]), &like); err != nil {
// 				log.Printf("failed to unmarshal like from redis, got error: %v", err)
// 				continue
// 			}

// 			batch = append(batch, &like)

// 			// 达到批量大小立即保存
// 			if len(batch) >= 50 {
// 				saveBatch(batch)
// 				batch = batch[:0]
// 			}
// 		}
// 	}
// }

// // 执行批量保存
// func saveBatch(batch []*models.Like) {
// 	//安全检查：过滤掉nil指针
// 	validBatch := make([]*models.Like, 0, len(batch))
// 	for _, like := range batch {
// 		if like != nil {
// 			validBatch = append(validBatch, like)
// 		}
// 	}
// 	if len(validBatch) == 0 {
// 		return
// 	}
// 	//使用事务批量处理
// 	tx := global.Db.Begin()
// 	if tx.Error != nil {
// 		global.Db.Logger.Error(context.TODO(), "批量写入开启事务失败：%v", tx.Error)
// 		return
// 	}
// 	//使用批量插入
// 	//用postid和userid作为联合唯一索引，避免重复
// 	err := tx.Clauses(clause.OnConflict{
// 		Columns:   []clause.Column{{Name: "postid"}, {Name: "userid"}},
// 		DoUpdates: clause.Assignments(map[string]interface{}{"liked": clause.Expr{SQL: "VALUES(liked)"}}),
// 	}).CreateInBatches(validBatch, len(validBatch)).Error
// 	if err != nil {
// 		tx.Rollback()
// 		global.Db.Logger.Error(context.TODO(), "批量写入数据库失败：%v", err)
// 		return
// 	}
// 	if err := tx.Commit().Error; err != nil {
// 		global.Db.Logger.Error(context.TODO(), "批量提交事务失败：%v", err)
// 		return
// 	}
// 	//归还对象池
// 	for _, like := range batch {
// 		like.Postid = ""
// 		like.Userid = ""
// 		like.Liked = false
// 		likeSyncPool.Put(like)
// 	}
// }

// // 同步更新数据库（针对取消点赞操作使用同步更新）
// func SyscUpdateDB(like *models.Like) error {
// 	tx := global.Db.Begin()
// 	if tx.Error != nil {
// 		return tx.Error
// 	}
// 	//对于取消点赞仅更新状态
// 	if !like.Liked {
// 		if err := tx.Model(&models.Like{}).
// 			Where("postid=? AND userid=?", like.Postid, like.Userid).
// 			Update("liked", false).Error; err != nil {
// 			tx.Rollback()
// 			return err
// 		}
// 	} else {
// 		//点赞操作使用UPSERT
// 		if err := tx.Clauses(clause.OnConflict{
// 			Columns:   []clause.Column{{Name: "postid"}, {Name: "userid"}},
// 			DoUpdates: clause.Assignments(map[string]interface{}{"liked": true}),
// 		}).Create(like).Error; err != nil {
// 			tx.Rollback()
// 			return err
// 		}
// 	}
// 	return tx.Commit().Error
// }

// // 推送点赞记录到Redis队列
// func PushLikeToQueue(like *models.Like) error {
// 	data, err := json.Marshal(like)
// 	if err != nil {
// 		return err
// 	}
// 	return global.Redis.RPush(context.Background(), likeQueueKey, data).Err()
// }

// // 点赞/取消点赞
// func LikePost(c *gin.Context) {
// 	//启动批量写入器
// 	initBatchWriter()
// 	//获取当前用户id
// 	userid := c.GetString("userid")
// 	if userid == "" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "登录后才能点赞哦!"})
// 		return
// 	}
// 	postid := c.Param("postid")
// 	if postid == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "请求错误"})
// 		return
// 	}

// 	// 检查限流和防抖
// 	if err := utils.CheckRateLimitAndDebounce("like", userid, postid); err != nil {
// 		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
// 		return
// 	}
// 	// 判断帖文是否存在 - 加强验证
// 	exists, err := models.CheckPostExists(postid)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误"})
// 		return
// 	}
// 	if !exists {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "帖文不存在"})
// 		return
// 	}

// 	//构建redis键名（格式：like：posts：{postid}）
// 	key := "like:" + cachekey + postid

// 	//判断用户是否点赞过该帖文
// 	isLiked, err := global.Redis.SIsMember(context.Background(), key, userid).Result()
// 	if err != nil {
// 		// 只有真正的Redis错误（如连接失败）才返回错误
// 		global.Db.Logger.Error(context.TODO(), "Redis检查点赞状态失败: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取点赞状态失败，请稍后再试"})
// 		return
// 	}

// 	//从对象池获取like对象
// 	like := likeSyncPool.Get().(*models.Like)
// 	like.Postid = postid
// 	like.Userid = userid

// 	var action string
// 	var newLikedStatus bool

// 	if isLiked {
// 		//已点赞，取消点赞
// 		_, err = global.Redis.SRem(context.Background(), key, userid).Result()
// 		action = "取消点赞"
// 		newLikedStatus = false
// 		like.Liked = newLikedStatus
// 		if err != nil {
// 			global.Db.Logger.Error(context.TODO(), "Redis取消点赞失败: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": action + "失败，请稍后再试"})
// 			return
// 		}

// 		//同步更新数据库，确保立即生效
// 		if err := SyscUpdateDB(like); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误" + err.Error()})
// 			return
// 		}
// 	} else {
// 		//未点赞，执行点赞
// 		_, err = global.Redis.SAdd(context.Background(), key, userid).Result()
// 		action = "点赞"
// 		newLikedStatus = true
// 		like.Liked = newLikedStatus
// 		if err != nil {
// 			global.Db.Logger.Error(context.TODO(), "Redis点赞失败: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": action + "失败，请稍后再试"})
// 			return
// 		}
// 		//推送至Redis队列
// 		if err := PushLikeToQueue(like); err != nil {
// 			//队列失败时降级为直接写入
// 			go func(l *models.Like) {
// 				if err := global.Db.Save(l).Error; err != nil {
// 					global.Db.Logger.Error(context.TODO(), "数据库写入失败：%v", err)
// 				}
// 				l.Postid = ""
// 				l.Userid = ""
// 				l.Liked = false
// 				likeSyncPool.Put(l)
// 			}(like)
// 		} else {
// 			//归还对象池
// 			like.Postid = ""
// 			like.Userid = ""
// 			like.Liked = false
// 			likeSyncPool.Put(like)
// 		}
// 		// 点赞成功后发送通知
// 		if newLikedStatus {
// 			// 获取帖子作者ID
// 			postAuthorID, err := models.GetPostAuthor(postid)
// 			if err == nil && postAuthorID != userid { // 不是自己点赞自己的帖子
// 				notification := &models.Notification{
// 					Userid:      postAuthorID,
// 					Senderid:    userid,
// 					ContentType: models.NotificationTypeLike,
// 					Contentid:   postid,
// 					Content:     userid + "点赞了你的帖文",
// 				}
// 				// 推送通知
// 				go utils.PushToWebSocket(notification)
// 			}
// 		}
// 	}
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": action + "失败" + err.Error()})
// 		return
// 	}
// 	// 设置缓存过期时间
// 	global.Redis.Expire(context.Background(), key, 24*time.Hour)

// 	//获取当前点赞数
// 	likeCount, err := global.Redis.SCard(context.Background(), key).Result()

// 	if err != nil {
// 		global.Db.Logger.Error(context.TODO(), "获取点赞数失败: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取点赞数失败"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{
// 		"message":   action + "成功",
// 		"likeCount": likeCount,
// 		"isLiked":   newLikedStatus,
// 		"postid":    postid,
// 	})

// }

// // 获取当前帖子的点赞数
// func GetPostLikeCount(c *gin.Context) {
// 	postid := c.Param("postid")
// 	if postid == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "请求错误"})
// 		return
// 	}
// 	key := "like:" + cachekey + postid
// 	likecount, err := global.Redis.SCard(context.Background(), key).Result()

// 	// 缓存命中处理
// 	if err == nil && likecount > 0 {
// 		// 热点帖子续期
// 		if likecount > 100 {
// 			global.Redis.Expire(context.Background(), key, 7*24*time.Hour)
// 		} else {
// 			global.Redis.Expire(context.Background(), key, 1*time.Hour)
// 		}
// 		c.JSON(http.StatusOK, gin.H{
// 			"message":   "获取成功",
// 			"postid":    postid,
// 			"likeCount": likecount,
// 		})
// 		return
// 	}
// 	//Redis未命中或错误
// 	likes, err := models.GetLikeFromDB(postid)
// 	//从数据库中加载
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "从数据库中加载失败" + err.Error()})
// 		return
// 	}

// 	//重建缓存
// 	if len(likes) > 0 {
// 		userids := make([]interface{}, len(likes))
// 		for i, like := range likes {
// 			userids[i] = like.Userid
// 		}
// 		_, err := global.Redis.SAdd(context.Background(), key, userids...).Result()
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "重建缓存失败" + err.Error()})
// 			return
// 		}
// 		// 设置初始过期时间
// 		if len(likes) > 100 {
// 			global.Redis.Expire(context.Background(), key, 7*24*time.Hour)
// 		} else {
// 			global.Redis.Expire(context.Background(), key, 1*time.Hour)
// 		}
// 	}
// 	likecount = int64(len(likes))
// 	c.JSON(http.StatusOK, gin.H{
// 		"message":   "获取成功",
// 		"postid":    postid,
// 		"likeCount": likecount,
// 	})
// }

// // 获取当前帖子点赞的用户
// func GetPostLikeUsers(c *gin.Context) {
// 	postid := c.Param("postid")
// 	if postid == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "请求错误"})
// 		return
// 	}
// 	key := "like:" + cachekey + postid
// 	users, err := global.Redis.SMembers(context.Background(), key).Result()

// 	// 缓存命中处理
// 	if err == nil && len(users) > 0 {
// 		// 热点帖子续期
// 		if len(users) > 100 {
// 			global.Redis.Expire(context.Background(), key, 7*24*time.Hour)
// 		} else {
// 			global.Redis.Expire(context.Background(), key, 1*time.Hour)
// 		}
// 		c.JSON(http.StatusOK, gin.H{
// 			"message": "获取成功",
// 			"postid":  postid,
// 			"users":   users,
// 		})
// 		return
// 	}

// 	//Redis未命中或错误
// 	likes, err := models.GetLikeFromDB(postid)
// 	//从数据库中记载帖子点赞用户
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "从数据库中加载失败" + err.Error()})
// 		return
// 	}
// 	//重建缓存
// 	if len(likes) > 0 {
// 		userids := make([]interface{}, len(likes))
// 		for i, like := range likes {
// 			userids[i] = like.Userid
// 		}
// 		_, err := global.Redis.SAdd(context.Background(), key, userids...).Result()
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "重建缓存失败" + err.Error()})
// 			return
// 		}
// 		if len(likes) > 100 {
// 			global.Redis.Expire(context.Background(), key, 7*24*time.Hour)
// 		} else {
// 			global.Redis.Expire(context.Background(), key, 1*time.Hour)
// 		}
// 	}
// 	//转换为用户ID列表
// 	users = make([]string, len(likes))
// 	for i, like := range likes {
// 		users[i] = like.Userid

// 	}
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "获取成功",
// 		"postid":  postid,
// 		"users":   users,
// 	})
// }
