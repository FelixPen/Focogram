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
			"time":         n.CreatedAt.Format("2006-01-02 15:04:05"),
			"is_read":      n.IsRead,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":         total,
		"page":          page,
		"size":          size,
		"notifications": result,
	})
}

// 标记全部通知为已读
func MarkNotificationsAsRead(c *gin.Context) {
	userid := c.GetString("userid")
	if userid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	// 批量更新未读通知
	if err := global.Db.Model(&models.Notification{}).
		Where("userid = ? AND is_read = ?", userid, false).
		Update("is_read", true).Error; err != nil {
		log.Printf("标记通知已读失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "标记已读失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "标记已读成功"})
}

func DeleteNotification(c *gin.Context) {
	userid := c.GetString("userid")
	notificationID := c.Param("id")

	if userid == "" || notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	id, err := strconv.ParseUint(notificationID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID格式错误"})
		return
	}

	result := global.Db.Unscoped().
		Where("id = ? AND userid = ?", id, userid).
		Delete(&models.Notification{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知不存在"})
		return
	}

	ctx := context.Background()
	global.Redis.ZRem(ctx, utils.UserNotificationSetPrefix+userid, id)

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func BatchDeleteNotifications(c *gin.Context) {
	userid := c.GetString("userid")
	if userid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	var req struct {
		Ids []uint64 `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if len(req.Ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要删除的通知"})
		return
	}

	result := global.Db.Unscoped().
		Where("id IN (?) AND userid = ?", req.Ids, userid).
		Delete(&models.Notification{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	ctx := context.Background()
	ids := make([]interface{}, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = id
	}
	global.Redis.ZRem(ctx, utils.UserNotificationSetPrefix+userid, ids...)

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
		"deleted": result.RowsAffected,
	})
}
