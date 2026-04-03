package controllers

import (
	"Focogram/global"
	"Focogram/models"
	"Focogram/utils"
	"context"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// 获取用户通知列表（优先从缓存查询）
func GetNotifications(c *gin.Context) {
	userid := c.GetString("userid")
	if userid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	ctx := context.Background()

	// 1. 先从Redis缓存获取消息ID列表
	ids, total, err := utils.GetNotificationIDsFromUserSet(ctx, userid, page, size)
	if err == nil && total > 0 && len(ids) > 0 {
		// 2. 批量查询数据库（根据ID列表）
		var notifications []models.Notification
		if err := global.Db.Where("id IN (?)", ids).Find(&notifications).Error; err != nil {
			log.Printf("从数据库查询消息失败: %v", err)
		} else {
			// 3. 按时间倒序排序（因为数据库查询结果可能无序）
			sort.Slice(notifications, func(i, j int) bool {
				return notifications[i].CreatedAt.After(notifications[j].CreatedAt)
			})
			// 返回结果
			returnNotifications(c, notifications, total, page, size)
			return
		}
	}

	// 缓存未命中或查询失败，从数据库查询并重建缓存
	var notifications []models.Notification
	var dbTotal int64

	// 查询总数
	if err := global.Db.Model(&models.Notification{}).
		Where("userid = ?", userid).
		Count(&dbTotal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询通知总数失败"})
		return
	}

	// 查询分页数据
	if err := global.Db.Where("userid = ?", userid).
		Order("created_at DESC").
		Limit(size).
		Offset((page - 1) * size).
		Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询通知失败"})
		return
	}

	// 重建缓存（批量添加到Redis）
	pipeline := global.Redis.Pipeline()
	for _, n := range notifications {
		pipeline.ZAdd(ctx, utils.UserNotificationSetPrefix+userid, &redis.Z{
			Score:  float64(n.CreatedAt.UnixNano()),
			Member: n.ID,
		})
	}
	// 设置过期时间
	pipeline.Expire(ctx, utils.UserNotificationSetPrefix+userid, 7*24*time.Hour)
	_, err = pipeline.Exec(ctx)
	if err != nil {
		log.Printf("重建消息缓存失败: %v", err)
	}

	// 返回结果
	returnNotifications(c, notifications, dbTotal, page, size)
}

// 格式化通知响应
func returnNotifications(c *gin.Context, notifications []models.Notification, total int64, page, size int) {
	result := make([]gin.H, 0, len(notifications))
	for _, n := range notifications {
		result = append(result, gin.H{
			"id":           n.ID,
			"senderid":     n.Senderid,
			"content_type": n.ContentType,
			"contentid":    n.Contentid,
			"content":      n.Content,
			"created_at":   n.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"size":  size,
		"items": result,
	})
}
