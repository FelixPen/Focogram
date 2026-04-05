package controllers

import (
	"Focogram/global"
	"Focogram/models"
	"Focogram/utils"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

const (
	LikeQueueKey    = "like_queue"
	CachePrefix     = "like:posts:"
	BatchSize       = 50
	BatchInterval   = 5 * time.Second
	HotThreshold    = 100
	LongExpireTime  = 7 * 24 * time.Hour
	ShortExpireTime = 1 * time.Hour
)

// LikeService 点赞服务
type LikeService struct {
	queue          chan *models.Like
	batchProcessor *BatchProcessor
	once           sync.Once
}

// BatchProcessor 批量处理器
type BatchProcessor struct {
	batchSize int
	interval  time.Duration
	batch     []*models.Like
	mutex     sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

var likeService *LikeService
var serviceOnce sync.Once

// GetLikeService 获取点赞服务实例
func GetLikeService() *LikeService {
	serviceOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		likeService = &LikeService{
			queue: make(chan *models.Like, 1000), // 增加缓冲区
			batchProcessor: &BatchProcessor{
				batchSize: BatchSize,
				interval:  BatchInterval,
				batch:     make([]*models.Like, 0, BatchSize),
				ctx:       ctx,
				cancel:    cancel,
			},
		}
		// 启动批量处理器
		go likeService.batchProcessor.Start()
	})
	return likeService
}

// Start 启动批量处理器
func (bp *BatchProcessor) Start() {
	ticker := time.NewTicker(bp.interval)
	defer ticker.Stop()

	for {
		select {
		case <-bp.ctx.Done():
			return
		case like := <-likeService.queue:
			bp.addLike(like)
		case <-ticker.C:
			bp.flushBatch()
		}
	}
}

// addLike 添加点赞到批次
func (bp *BatchProcessor) addLike(like *models.Like) {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	bp.batch = append(bp.batch, like)

	if len(bp.batch) >= bp.batchSize {
		batch := make([]*models.Like, len(bp.batch))
		copy(batch, bp.batch)
		bp.batch = bp.batch[:0]
		go bp.saveBatch(batch)
	}
}

// flushBatch 刷新批次
func (bp *BatchProcessor) flushBatch() {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	if len(bp.batch) > 0 {
		batch := make([]*models.Like, len(bp.batch))
		copy(batch, bp.batch)
		bp.batch = bp.batch[:0]
		go bp.saveBatch(batch)
	}
}

// saveBatch 保存批次到数据库
func (bp *BatchProcessor) saveBatch(batch []*models.Like) {
	if len(batch) == 0 {
		return
	}

	tx := global.Db.Begin()
	if tx.Error != nil {
		global.Db.Logger.Error(context.TODO(), "批量写入开启事务失败："+tx.Error.Error())
		bp.returnToPool(batch)
		return
	}

	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "postid"}, {Name: "userid"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"liked": clause.Expr{SQL: "VALUES(liked)"}}),
	}).CreateInBatches(batch, len(batch)).Error

	if err != nil {
		tx.Rollback()
		global.Db.Logger.Error(context.TODO(), "批量写入数据库失败："+err.Error())
		bp.returnToPool(batch)
		return
	}

	if err := tx.Commit().Error; err != nil {
		global.Db.Logger.Error(context.TODO(), "批量提交事务失败："+err.Error())
		bp.returnToPool(batch)
		return
	}

	// 成功后归还对象到池
	bp.returnToPool(batch)
}

// returnToPool 归还对象到池
func (bp *BatchProcessor) returnToPool(batch []*models.Like) {
	for _, like := range batch {
		if like != nil {
			like.Reset() // 假设模型中有Reset方法清空字段
		}
	}
}

// LikePost 点赞/取消点赞
func LikePost2(c *gin.Context) {
	userid := c.GetString("userid")
	if userid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "登录后才能点赞哦!"})
		return
	}

	postid := c.Param("postid")
	if postid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求错误"})
		return
	}

	// 检查限流和防抖
	if err := utils.CheckRateLimitAndDebounce("like", userid, postid); err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	// 检查帖子是否存在，不存在直接返回成功
	exists, _ := models.CheckPostExists(postid)
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"message":   "帖子不存在",
			"likeCount": 0,
			"isLiked":   false,
			"postid":    postid,
		})
		return
	}

	service := GetLikeService()
	cacheKey := fmt.Sprintf("%s%s", CachePrefix, postid)

	// 检查是否已点赞
	isLiked, err := global.Redis.SIsMember(context.Background(), cacheKey, userid).Result()
	if err != nil {
		global.Db.Logger.Error(context.TODO(), "Redis检查点赞状态失败: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取点赞状态失败，请稍后再试"})
		return
	}

	like := &models.Like{
		Postid: postid,
		Userid: userid,
		Liked:  !isLiked,
	}

	var action string
	newLikedStatus := !isLiked

	if isLiked {
		// 取消点赞
		action = "取消点赞"
		_, err = global.Redis.SRem(context.Background(), cacheKey, userid).Result()
		like.Liked = false

		if err != nil {
			global.Db.Logger.Error(context.TODO(), "Redis取消点赞失败: "+err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": action + "失败，请稍后再试"})
			return
		}

		// 同步更新数据库
		if err := updateDatabaseSync(like); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误" + err.Error()})
			return
		}
	} else {
		// 点赞
		action = "点赞"
		_, err = global.Redis.SAdd(context.Background(), cacheKey, userid).Result()
		like.Liked = true

		if err != nil {
			global.Db.Logger.Error(context.TODO(), "Redis点赞失败: "+err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": action + "失败，请稍后再试"})
			return
		}

		// 异步保存到队列
		select {
		case service.queue <- like:
			// 发送通知
			sendNotification(userid, postid)
		default:
			// 队列满，直接同步保存
			if err := global.Db.Save(like).Error; err != nil {
				global.Db.Logger.Error(context.TODO(), "数据库写入失败："+err.Error())
			}
		}
	}

	// 设置缓存过期时间
	setCacheExpiration(cacheKey)

	// 获取点赞数
	likeCount, err := global.Redis.SCard(context.Background(), cacheKey).Result()
	if err != nil {
		global.Db.Logger.Error(context.TODO(), "获取点赞数失败: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取点赞数失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   action + "成功",
		"likeCount": likeCount,
		"isLiked":   newLikedStatus,
		"postid":    postid,
	})
}

// updateDatabaseSync 同步更新数据库
func updateDatabaseSync(like *models.Like) error {
	tx := global.Db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if !like.Liked {
		// 取消点赞
		if err := tx.Model(&models.Like{}).
			Where("postid=? AND userid=?", like.Postid, like.Userid).
			Update("liked", false).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// 点赞 - 使用UPSERT
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "postid"}, {Name: "userid"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"liked": true}),
		}).Create(like).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// sendNotification 发送通知
func sendNotification(userid, postid string) {
	go func() {
		postAuthorID, err := models.GetPostAuthor(postid)
		if err != nil || postAuthorID == userid {
			return
		}

		var liker models.User
		if err := global.Db.Select("username").Where("userid = ?", userid).First(&liker).Error; err != nil {
			return
		}

		notification := &models.Notification{
			Userid:      postAuthorID,
			Senderid:    userid,
			ContentType: models.NotificationTypeLike,
			Contentid:   postid,
			Content:     liker.Username + " 点赞了你的帖子",
		}
		utils.PushToWebSocket(notification)
	}()
}

// setCacheExpiration 设置缓存过期时间
func setCacheExpiration(key string) {
	count, _ := global.Redis.SCard(context.Background(), key).Result()
	expiration := ShortExpireTime
	if count > HotThreshold {
		expiration = LongExpireTime
	}
	global.Redis.Expire(context.Background(), key, expiration)
}

// GetPostLikeCount 获取点赞数
func GetPostLikeCount2(c *gin.Context) {
	postid := c.Param("postid")
	if postid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求错误"})
		return
	}

	cacheKey := fmt.Sprintf("%s%s", CachePrefix, postid)

	// 尝试从缓存获取
	likeCount, err := global.Redis.SCard(context.Background(), cacheKey).Result()
	if err == nil && likeCount > 0 {
		setCacheExpiration(cacheKey)
		c.JSON(http.StatusOK, gin.H{
			"message":   "获取成功",
			"postid":    postid,
			"likeCount": likeCount,
		})
		return
	}

	// 从数据库加载
	likes, err := models.GetLikeFromDB(postid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "从数据库中加载失败" + err.Error()})
		return
	}

	// 重建缓存
	if len(likes) > 0 {
		userids := make([]interface{}, len(likes))
		for i, like := range likes {
			userids[i] = like.Userid
		}
		global.Redis.SAdd(context.Background(), cacheKey, userids...)
	}

	likeCount = int64(len(likes))
	setCacheExpiration(cacheKey)

	c.JSON(http.StatusOK, gin.H{
		"message":   "获取成功",
		"postid":    postid,
		"likeCount": likeCount,
	})
}

// GetPostLikeUsers 获取点赞用户列表
func GetPostLikeUsers2(c *gin.Context) {
	postid := c.Param("postid")
	if postid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求错误"})
		return
	}

	cacheKey := fmt.Sprintf("%s%s", CachePrefix, postid)

	// 尝试从缓存获取
	users, err := global.Redis.SMembers(context.Background(), cacheKey).Result()
	if err == nil && len(users) > 0 {
		setCacheExpiration(cacheKey)
		c.JSON(http.StatusOK, gin.H{
			"message": "获取成功",
			"postid":  postid,
			"users":   users,
		})
		return
	}

	// 从数据库加载
	likes, err := models.GetLikeFromDB(postid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "从数据库中加载失败" + err.Error()})
		return
	}

	// 重建缓存
	if len(likes) > 0 {
		userids := make([]interface{}, len(likes))
		for i, like := range likes {
			userids[i] = like.Userid
		}
		global.Redis.SAdd(context.Background(), cacheKey, userids...)
	}

	// 转换为用户ID列表
	users = make([]string, len(likes))
	for i, like := range likes {
		users[i] = like.Userid
	}

	setCacheExpiration(cacheKey)

	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"postid":  postid,
		"users":   users,
	})
}

// Close 关闭服务
func (ls *LikeService) Close() {
	if ls.batchProcessor != nil && ls.batchProcessor.cancel != nil {
		ls.batchProcessor.cancel()
	}
}
