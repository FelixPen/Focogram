package controllers

import (
	"Focogram/global"
	"Focogram/models"
	"Focogram/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm/clause"
)

var (
	// 缓存键前缀
	followingCacheKey = "following:" // following:{userid} -> Set 存储关注的用户ID
	followerCacheKey  = "follower:"  // follower:{userid} -> Set 存储粉丝用户ID
	followQueueKey    = "follow_queue"
	followOnce        sync.Once
)

// 初始化关注批量写入器
func initFollowBatchWriter() {
	followOnce.Do(func() {
		go followBatchWriteToDB()
	})
}

// 批量将关注记录写入数据库
func followBatchWriteToDB() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	batch := make([]*models.Follow, 0, 50)
	for {
		select {
		case <-ticker.C:
			if len(batch) > 0 {
				saveFollowBatch(batch)
				batch = batch[:0]
			}
		default:
			// 从Redis队列读取数据
			result, err := global.Redis.BLPop(context.Background(), 5*time.Second, followQueueKey).Result()
			if err != nil {
				if err != redis.Nil {
					log.Printf("failed to get follow from redis: %v", err)
				}
				continue
			}

			var follow models.Follow
			if err := json.Unmarshal([]byte(result[1]), &follow); err != nil {
				log.Printf("failed to unmarshal follow: %v", err)
				continue
			}
			batch = append(batch, &follow)

			if len(batch) >= 50 {
				saveFollowBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// 保存批量关注记录到数据库
func saveFollowBatch(batch []*models.Follow) {
	if len(batch) == 0 {
		return
	}

	tx := global.Db.Begin()
	if tx.Error != nil {
		log.Printf("开启事务失败: %v", tx.Error)
		return
	}

	// 批量插入，冲突时忽略（已存在的关注记录）
	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "followerid"}, {Name: "followedid"}},
		DoNothing: true,
	}).CreateInBatches(batch, 50).Error

	if err != nil {
		tx.Rollback()
		log.Printf("批量写入关注记录失败: %v", err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("提交关注事务失败: %v", err)
	}
}

// 推送关注记录到Redis队列
func pushFollowToQueue(follow *models.Follow) error {
	data, err := json.Marshal(follow)
	if err != nil {
		return err
	}
	return global.Redis.RPush(context.Background(), followQueueKey, data).Err()
}

// 关注用户
func FollowUser(c *gin.Context) {
	initFollowBatchWriter()
	followerID := c.GetString("userid")
	if followerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	followedID := c.Param("userid")
	if followedID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 不能关注自己
	if followerID == followedID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能关注自己"})
		return
	}

	// 检查限流
	if err := utils.CheckRateLimitAndDebounce("follow", followerID, followedID); err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否存在
	exists, err := CheckUserExists(followedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 检查是否已关注
	followingKey := followingCacheKey + followerID
	isFollowing, err := global.Redis.SIsMember(context.Background(), followingKey, followedID).Result()
	if err != nil {
		log.Printf("检查关注状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取关注状态失败"})
		return
	}

	if isFollowing {
		c.JSON(http.StatusBadRequest, gin.H{"error": "已关注该用户"})
		return
	}

	// 更新缓存
	ctx := context.Background()
	pipe := global.Redis.Pipeline()
	pipe.SAdd(ctx, followingKey, followedID)
	pipe.Expire(ctx, followingKey, 24*time.Hour)

	followerKey := followerCacheKey + followedID
	pipe.SAdd(ctx, followerKey, followerID)
	pipe.Expire(ctx, followerKey, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("更新关注缓存失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "关注失败"})
		return
	}

	// 推送至队列异步写入数据库
	follow := &models.Follow{
		Followerid: followerID,
		Followedid: followedID,
	}
	if err := pushFollowToQueue(follow); err != nil {
		// 队列失败时降级为直接写入
		go func() {
			if err := global.Db.Create(follow).Error; err != nil {
				log.Printf("直接写入关注记录失败: %v", err)
			}
		}()
	}

	// 关注成功后发送通知
	var follower models.User
	if err := global.Db.Select("username").Where("userid = ?", followerID).First(&follower).Error; err == nil {
		notification := &models.Notification{
			Userid:      followedID,
			Senderid:    followerID,
			ContentType: models.NotificationTypeFollow,
			Contentid:   followerID,
			Content:     follower.Username + " 关注了你",
		}
		go utils.PushToWebSocket(notification)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "关注成功",
		"followedid": followedID,
	})
}

// 取消关注
func UnfollowUser(c *gin.Context) {
	followerID := c.GetString("userid")
	if followerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	followedID := c.Param("userid")
	if followedID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 同步删除数据库记录
	tx := global.Db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误"})
		return
	}

	if err := tx.Unscoped().Where("followerid = ? AND followedid = ?", followerID, followedID).
		Delete(&models.Follow{}).Error; err != nil {
		tx.Rollback()
		log.Printf("删除关注记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消关注失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消关注失败"})
		return
	}

	// 清除缓存
	ctx := context.Background()
	pipe := global.Redis.Pipeline()
	pipe.SRem(ctx, followingCacheKey+followerID, followedID)
	pipe.SRem(ctx, followerCacheKey+followedID, followerID)
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("清除关注缓存失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "取消关注成功",
		"followedid": followedID,
	})
}

// 查看我的关注列表
func GetMyFollowing(c *gin.Context) {
	userID := c.GetString("userid")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	key := followingCacheKey + userID
	ctx := context.Background()

	// 从缓存获取
	followingList, err := global.Redis.SMembers(ctx, key).Result()
	if err == nil && len(followingList) > 0 {
		// 热点数据续期
		if len(followingList) > 100 {
			global.Redis.Expire(ctx, key, 7*24*time.Hour)
		} else {
			global.Redis.Expire(ctx, key, 24*time.Hour)
		}

		c.JSON(http.StatusOK, gin.H{
			"count":     len(followingList),
			"following": followingList,
		})
		return
	}

	// 缓存未命中，从数据库获取
	var follows []models.Follow
	if err := global.Db.Where("followerid = ?", userID).Find(&follows).Error; err != nil {
		log.Printf("查询关注列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取关注列表失败"})
		return
	}

	// 构建关注列表
	followingList = make([]string, 0, len(follows))
	for _, follow := range follows {
		followingList = append(followingList, follow.Followedid)
	}

	// 重建缓存
	if len(followingList) > 0 {
		members := make([]interface{}, len(followingList))
		for i, id := range followingList {
			members[i] = id
		}
		global.Redis.SAdd(ctx, key, members...)
		global.Redis.Expire(ctx, key, 24*time.Hour)
	}

	c.JSON(http.StatusOK, gin.H{
		"count":     len(followingList),
		"following": followingList,
	})
}

// 查看我的粉丝列表
func GetMyFollowers(c *gin.Context) {
	userID := c.GetString("userid")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	key := followerCacheKey + userID
	ctx := context.Background()

	// 从缓存获取
	followerList, err := global.Redis.SMembers(ctx, key).Result()
	if err == nil && len(followerList) > 0 {
		// 热点数据续期
		if len(followerList) > 100 {
			global.Redis.Expire(ctx, key, 7*24*time.Hour)
		} else {
			global.Redis.Expire(ctx, key, 24*time.Hour)
		}

		c.JSON(http.StatusOK, gin.H{
			"count":     len(followerList),
			"followers": followerList,
		})
		return
	}

	// 缓存未命中，从数据库获取
	var follows []models.Follow
	if err := global.Db.Where("followedid = ?", userID).Find(&follows).Error; err != nil {
		log.Printf("查询粉丝列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取粉丝列表失败"})
		return
	}

	// 构建粉丝列表
	followerList = make([]string, 0, len(follows))
	for _, follow := range follows {
		followerList = append(followerList, follow.Followerid)
	}

	// 重建缓存
	if len(followerList) > 0 {
		members := make([]interface{}, len(followerList))
		for i, id := range followerList {
			members[i] = id
		}
		global.Redis.SAdd(ctx, key, members...)
		global.Redis.Expire(ctx, key, 24*time.Hour)
	}

	c.JSON(http.StatusOK, gin.H{
		"count":     len(followerList),
		"followers": followerList,
	})
}

// 检查是否已关注某个用户
func CheckFollow(c *gin.Context) {
	followerID := c.GetString("userid")
	if followerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	followedID := c.Param("userid")
	if followedID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	ctx := context.Background()
	key := followingCacheKey + followerID

	isMember, err := global.Redis.SIsMember(ctx, key, followedID).Result()
	if err != nil {
		var count int64
		if err := global.Db.Model(&models.Follow{}).Where("followerid = ? AND followedid = ?", followerID, followedID).Count(&count).Error; err != nil {
			log.Printf("检查关注状态失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误"})
			return
		}
		isMember = count > 0
	}

	c.JSON(http.StatusOK, gin.H{
		"isFollowing": isMember,
		"followerid":  followerID,
		"followedid":  followedID,
	})
}

// 检查用户是否存在（完善实现）
func CheckUserExists(userID string) (bool, error) {
	var count int64
	// 明确指定用户表，确保查询正确
	err := global.Db.Table("users").Where("userid = ?", userID).Count(&count).Error
	if err != nil {
		log.Printf("检查用户存在性失败 (userID=%s): %v", userID, err)
		return false, err
	}
	return count > 0, nil
}

// 获取指定用户的关注列表
func GetUserFollowing(c *gin.Context) {
	targetUserID := c.Param("userid")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	key := followingCacheKey + targetUserID
	ctx := context.Background()

	// 从缓存获取
	followingList, err := global.Redis.SMembers(ctx, key).Result()
	if err == nil && len(followingList) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"count":     len(followingList),
			"following": followingList,
		})
		return
	}

	// 缓存未命中，从数据库获取
	var follows []models.Follow
	if err := global.Db.Where("followerid = ?", targetUserID).Find(&follows).Error; err != nil {
		log.Printf("查询用户关注列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取关注列表失败"})
		return
	}

	// 构建关注列表
	followingList = make([]string, 0, len(follows))
	for _, follow := range follows {
		followingList = append(followingList, follow.Followedid)
	}

	// 重建缓存
	if len(followingList) > 0 {
		members := make([]interface{}, len(followingList))
		for i, id := range followingList {
			members[i] = id
		}
		global.Redis.SAdd(ctx, key, members...)
		global.Redis.Expire(ctx, key, 24*time.Hour)
	}

	c.JSON(http.StatusOK, gin.H{
		"count":     len(followingList),
		"following": followingList,
	})
}

// 获取指定用户的粉丝列表
func GetUserFollowers(c *gin.Context) {
	targetUserID := c.Param("userid")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	key := followerCacheKey + targetUserID
	ctx := context.Background()

	// 从缓存获取
	followerList, err := global.Redis.SMembers(ctx, key).Result()
	if err == nil && len(followerList) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"count":     len(followerList),
			"followers": followerList,
		})
		return
	}

	// 缓存未命中，从数据库获取
	var follows []models.Follow
	if err := global.Db.Where("followedid = ?", targetUserID).Find(&follows).Error; err != nil {
		log.Printf("查询用户粉丝列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取粉丝列表失败"})
		return
	}

	// 构建粉丝列表
	followerList = make([]string, 0, len(follows))
	for _, follow := range follows {
		followerList = append(followerList, follow.Followerid)
	}

	// 重建缓存
	if len(followerList) > 0 {
		members := make([]interface{}, len(followerList))
		for i, id := range followerList {
			members[i] = id
		}
		global.Redis.SAdd(ctx, key, members...)
		global.Redis.Expire(ctx, key, 24*time.Hour)
	}

	c.JSON(http.StatusOK, gin.H{
		"count":     len(followerList),
		"followers": followerList,
	})
}
