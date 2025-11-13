package controllers

import (
	"Focogram/global"
	"Focogram/models"
	"Focogram/utils"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	// 缓存键前缀定义
	convCacheKey     = "conv:"      // conv:{convID} -> Hash
	userConvCacheKey = "user:conv:" // user:conv:{userID} -> Sorted Set (score: lastMessageAt)
	msgCacheKey      = "msg:"       // msg:{msgID} -> Hash
	convMsgCacheKey  = "conv:msg:"  // conv:msg:{convID} -> Sorted Set (score: createdAt)
	msgQueueKey      = "msg_queue"  // 消息异步写入队列
	convQueueKey     = "conv_queue" // 对话异步写入队列
	msgBatchOnce     sync.Once      // 确保批量写入器只初始化一次
)

// 初始化消息批量写入器
func initMsgBatchWriter() {
	msgBatchOnce.Do(func() {
		go msgBatchWriteToDB()
		go convBatchWriteToDB()
	})
}

// 消息批量写入数据库
func msgBatchWriteToDB() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	batch := make([]*models.PrivateMessage, 0, 50)
	for {
		select {
		case <-ticker.C:
			if len(batch) > 0 {
				saveMsgBatch(batch)
				batch = batch[:0]
			}
		default:
			result, err := global.Redis.BLPop(context.Background(), 5*time.Second, msgQueueKey).Result()
			if err != nil {
				if err != redis.Nil {
					log.Printf("获取消息队列数据失败: %v", err)
				}
				continue
			}

			var msg models.PrivateMessage
			if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
				log.Printf("解析消息数据失败: %v", err)
				continue
			}
			batch = append(batch, &msg)

			if len(batch) >= 50 {
				saveMsgBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// 对话批量写入数据库
func convBatchWriteToDB() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	batch := make([]*models.Conversation, 0, 50)
	for {
		select {
		case <-ticker.C:
			if len(batch) > 0 {
				saveConvBatch(batch)
				batch = batch[:0]
			}
		default:
			result, err := global.Redis.BLPop(context.Background(), 5*time.Second, convQueueKey).Result()
			if err != nil {
				if err != redis.Nil {
					log.Printf("获取对话队列数据失败: %v", err)
				}
				continue
			}

			var conv models.Conversation
			if err := json.Unmarshal([]byte(result[1]), &conv); err != nil {
				log.Printf("解析对话数据失败: %v", err)
				continue
			}
			batch = append(batch, &conv)

			if len(batch) >= 50 {
				saveConvBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// 保存消息批次到数据库
func saveMsgBatch(batch []*models.PrivateMessage) {
	if len(batch) == 0 {
		return
	}

	tx := global.Db.Begin()
	if tx.Error != nil {
		log.Printf("开启消息事务失败: %v", tx.Error)
		return
	}

	err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "message_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"content":    clause.Expr{SQL: "VALUES(content)"},
			"updated_at": time.Now(),
		}),
	}).CreateInBatches(batch, 50).Error

	if err != nil {
		tx.Rollback()
		log.Printf("批量保存消息失败: %v", err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("提交消息事务失败: %v", err)
	}
}

// 保存对话批次到数据库
func saveConvBatch(batch []*models.Conversation) {
	if len(batch) == 0 {
		return
	}

	tx := global.Db.Begin()
	if tx.Error != nil {
		log.Printf("开启对话事务失败: %v", tx.Error)
		return
	}

	err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "conversation_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_message":    clause.Expr{SQL: "VALUES(last_message)"},
			"last_message_at": clause.Expr{SQL: "VALUES(last_message_at)"},
			"updated_at":      time.Now(),
		}),
	}).CreateInBatches(batch, 50).Error

	if err != nil {
		tx.Rollback()
		log.Printf("批量保存对话失败: %v", err)
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("提交对话事务失败: %v", err)
	}
}

// 创建对话
func CreateConversation(c *gin.Context) {
	initMsgBatchWriter()
	userID := c.GetString("userid")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	// 检查当前登录用户是否存在
	exists, err := CheckUserExists(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误：查询用户信息失败"})
		return
	}
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "当前用户不存在，请重新登录"})
		return
	}

	var req models.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误：" + err.Error()})
		return
	}

	// 不能和自己创建对话
	if userID == req.TargetUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能和自己创建对话"})
		return
	}

	// 检查目标用户是否存在
	exists, err = CheckUserExists(req.TargetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误：查询目标用户信息失败"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标用户不存在"})
		return
	}

	// 检查是否已存在对话
	var existingConv models.Conversation
	err = global.Db.Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
		userID, req.TargetUserID, req.TargetUserID, userID).First(&existingConv).Error

	if err == nil {
		// 对话已存在，返回现有对话ID
		c.JSON(http.StatusOK, gin.H{
			"conversation_id": existingConv.ConversationID,
			"message":         "对话已存在",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("查询对话失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询对话失败"})
		return
	}

	// 创建新对话
	conv := &models.Conversation{
		User1ID:       userID,
		User2ID:       req.TargetUserID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		LastMessageAt: time.Now(),
	}

	// 保存到数据库获取自增ID（关键错误点优化）
	if err := global.Db.Create(conv).Error; err != nil {
		// 打印详细错误日志
		log.Printf("创建对话失败: user1_id=%s, user2_id=%s, 错误详情: %v", userID, req.TargetUserID, err)

		// 解析MySQL特定错误
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			// 外键约束错误（1452为MySQL外键约束失败代码）
			if mysqlErr.Number == 1452 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":  "创建对话失败：用户信息无效",
					"detail": "请确认双方用户是否存在于系统中",
				})
				return
			}
			// 其他MySQL错误
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "数据库错误",
				"detail": mysqlErr.Message,
			})
			return
		}

		// 通用错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建对话失败"})
		return
	}

	// 更新缓存
	ctx := context.Background()
	convKey := convCacheKey + strconv.FormatUint(uint64(conv.ConversationID), 10)
	pipe := global.Redis.Pipeline()

	// 存储对话信息
	pipe.HSet(ctx, convKey, map[string]interface{}{
		"conversation_id": conv.ConversationID,
		"user1_id":        conv.User1ID,
		"user2_id":        conv.User2ID,
		"last_message":    "",
		"last_message_at": conv.LastMessageAt.Unix(),
		"created_at":      conv.CreatedAt.Unix(),
	})
	pipe.Expire(ctx, convKey, 24*time.Hour)

	// 添加到双方的对话列表
	user1ConvKey := userConvCacheKey + conv.User1ID
	user2ConvKey := userConvCacheKey + conv.User2ID
	pipe.ZAdd(ctx, user1ConvKey, &redis.Z{
		Score:  float64(conv.LastMessageAt.Unix()),
		Member: conv.ConversationID,
	})
	pipe.ZAdd(ctx, user2ConvKey, &redis.Z{
		Score:  float64(conv.LastMessageAt.Unix()),
		Member: conv.ConversationID,
	})
	pipe.Expire(ctx, user1ConvKey, 24*time.Hour)
	pipe.Expire(ctx, user2ConvKey, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("更新对话缓存失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation_id": conv.ConversationID,
		"message":         "对话创建成功",
	})
}

// 发送私信
func SendPrivateMessage(c *gin.Context) {
	initMsgBatchWriter()
	senderID := c.GetString("userid")
	receiverID := c.Param("receiver_id")
	if senderID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	var req models.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取对话ID
	convID, err := strconv.ParseUint(c.Param("conv_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的对话ID"})
		return
	}

	// 验证对话是否存在且当前用户为参与者
	var conv models.Conversation
	if err := global.Db.Where("conversation_id = ? AND (user1_id = ? OR user2_id = ?)",
		convID, senderID, senderID).First(&conv).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "对话不存在或无权访问"})
			return
		}
		log.Printf("查询对话失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误"})
		return
	}

	// 验证接收者是否为对话参与者
	if conv.User1ID != receiverID && conv.User2ID != receiverID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "接收者不在当前对话中"})
		return
	}

	// 创建消息
	msg := &models.PrivateMessage{
		ConversationID: uint(convID),
		SenderID:       senderID,
		ReceiverID:     receiverID,
		Content:        req.Content,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 更新缓存
	ctx := context.Background()
	msgKey := msgCacheKey + strconv.FormatUint(uint64(msg.MessageID), 10) // 临时ID，实际会被数据库自增ID替换
	convMsgKey := convMsgCacheKey + strconv.FormatUint(convID, 10)
	pipe := global.Redis.Pipeline()

	// 存储消息
	pipe.HSet(ctx, msgKey, map[string]interface{}{
		"message_id":      msg.MessageID,
		"conversation_id": msg.ConversationID,
		"sender_id":       msg.SenderID,
		"receiver_id":     msg.ReceiverID,
		"content":         msg.Content,
		"created_at":      msg.CreatedAt.Unix(),
	})
	pipe.Expire(ctx, msgKey, 24*time.Hour)

	// 添加到对话消息列表
	pipe.ZAdd(ctx, convMsgKey, &redis.Z{
		Score:  float64(msg.CreatedAt.Unix()),
		Member: msg.MessageID,
	})
	pipe.Expire(ctx, convMsgKey, 24*time.Hour)

	// 更新对话最后一条消息
	convKey := convCacheKey + strconv.FormatUint(convID, 10)
	pipe.HSet(ctx, convKey, map[string]interface{}{
		"last_message":    msg.Content,
		"last_message_at": msg.CreatedAt.Unix(),
		"updated_at":      msg.UpdatedAt.Unix(),
	})

	// 更新双方的对话列表排序
	user1ConvKey := userConvCacheKey + conv.User1ID
	user2ConvKey := userConvCacheKey + conv.User2ID
	pipe.ZAdd(ctx, user1ConvKey, &redis.Z{
		Score:  float64(msg.CreatedAt.Unix()),
		Member: convID,
	})
	pipe.ZAdd(ctx, user2ConvKey, &redis.Z{
		Score:  float64(msg.CreatedAt.Unix()),
		Member: convID,
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("更新消息缓存失败: %v", err)
	}

	// 推送消息到队列异步写入数据库
	data, _ := json.Marshal(msg)
	if err := global.Redis.RPush(ctx, msgQueueKey, data).Err(); err != nil {
		// 队列失败时直接写入数据库
		go func() {
			if err := global.Db.Create(msg).Error; err != nil {
				log.Printf("直接写入消息失败: %v", err)
			}
		}()
	}

	// 推送对话更新到队列
	conv.LastMessage = msg.Content
	conv.LastMessageAt = msg.CreatedAt
	conv.UpdatedAt = msg.UpdatedAt
	convData, _ := json.Marshal(conv)
	global.Redis.RPush(ctx, convQueueKey, convData)

	// 发送WebSocket通知
	notification := &models.Notification{
		Userid:      receiverID,
		Senderid:    senderID,
		ContentType: models.NotificationTypeMessage,
		Contentid:   strconv.FormatUint(uint64(convID), 10),
		Content:     "收到新消息: " + req.Content,
	}
	go utils.PushToWebSocket(notification)

	c.JSON(http.StatusOK, gin.H{
		"message_id":      msg.MessageID,
		"conversation_id": convID,
		"message":         "消息发送成功",
	})
}

// 获取对话列表
func GetConversations(c *gin.Context) {
	userID := c.GetString("userid")
	if userID == "" {
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

	start := (page - 1) * size
	end := start + size - 1

	ctx := context.Background()
	key := userConvCacheKey + userID

	// 从缓存获取对话ID列表（按最后消息时间倒序）
	convIDs, err := global.Redis.ZRevRange(ctx, key, int64(start), int64(end)).Result()
	if err == nil && len(convIDs) > 0 {
		// 获取总数
		total, _ := global.Redis.ZCard(ctx, key).Result()

		// 批量获取对话详情
		conversations := make([]map[string]interface{}, 0, len(convIDs))
		for _, idStr := range convIDs {
			convID, _ := strconv.ParseUint(idStr, 10, 64)
			convData, err := global.Redis.HGetAll(ctx, convCacheKey+idStr).Result()
			if err != nil || len(convData) == 0 {
				continue
			}

			// 确定对方用户ID
			otherUserID := convData["user1_id"]
			if otherUserID == userID {
				otherUserID = convData["user2_id"]
			}

			conversations = append(conversations, map[string]interface{}{
				"conversation_id": convID,
				"other_user_id":   otherUserID,
				"last_message":    convData["last_message"],
				"last_message_at": convData["last_message_at"],
				"created_at":      convData["created_at"],
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"total":         total,
			"page":          page,
			"size":          size,
			"conversations": conversations,
		})
		return
	}

	// 缓存未命中，从数据库查询
	var dbConvs []models.Conversation
	if err := global.Db.Where("user1_id = ? OR user2_id = ?", userID, userID).
		Order("last_message_at DESC").
		Limit(size).
		Offset((page - 1) * size).
		Find(&dbConvs).Error; err != nil {
		log.Printf("查询对话列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取对话列表失败"})
		return
	}

	// 获取总数
	var total int64
	global.Db.Where("user1_id = ? OR user2_id = ?", userID, userID).Count(&total)

	// 重建缓存
	pipe := global.Redis.Pipeline()
	conversations := make([]map[string]interface{}, 0, len(dbConvs))
	for _, conv := range dbConvs {
		convIDStr := strconv.FormatUint(uint64(conv.ConversationID), 10)
		convKey := convCacheKey + convIDStr

		// 存储对话详情
		pipe.HSet(ctx, convKey, map[string]interface{}{
			"conversation_id": conv.ConversationID,
			"user1_id":        conv.User1ID,
			"user2_id":        conv.User2ID,
			"last_message":    conv.LastMessage,
			"last_message_at": conv.LastMessageAt.Unix(),
			"created_at":      conv.CreatedAt.Unix(),
		})
		pipe.Expire(ctx, convKey, 24*time.Hour)

		// 添加到用户对话列表
		pipe.ZAdd(ctx, key, &redis.Z{
			Score:  float64(conv.LastMessageAt.Unix()),
			Member: conv.ConversationID,
		})

		// 构建返回数据
		otherUserID := conv.User1ID
		if otherUserID == userID {
			otherUserID = conv.User2ID
		}

		conversations = append(conversations, map[string]interface{}{
			"conversation_id": conv.ConversationID,
			"other_user_id":   otherUserID,
			"last_message":    conv.LastMessage,
			"last_message_at": conv.LastMessageAt.Unix(),
			"created_at":      conv.CreatedAt.Unix(),
		})
	}
	pipe.Expire(ctx, key, 24*time.Hour)
	_, _ = pipe.Exec(ctx)

	c.JSON(http.StatusOK, gin.H{
		"total":         total,
		"page":          page,
		"size":          size,
		"conversations": conversations,
	})
}

// 获取对话消息列表
func GetConversationMessages(c *gin.Context) {
	userID := c.GetString("userid")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	// 获取对话ID
	convID, err := strconv.ParseUint(c.Param("conv_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的对话ID"})
		return
	}

	// 验证权限
	var conv models.Conversation
	if err := global.Db.Where("conversation_id = ? AND (user1_id = ? OR user2_id = ?)",
		convID, userID, userID).First(&conv).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "对话不存在或无权访问"})
			return
		}
		log.Printf("查询对话失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误"})
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

	// 计算Redis分页范围（倒序，最新的在前）
	start := int64((page - 1) * size)
	end := start + int64(size) - 1

	ctx := context.Background()
	key := convMsgCacheKey + strconv.FormatUint(convID, 10)

	// 从缓存获取消息ID
	msgIDs, err := global.Redis.ZRevRange(ctx, key, start, end).Result()
	if err == nil && len(msgIDs) > 0 {
		// 获取总数
		total, _ := global.Redis.ZCard(ctx, key).Result()

		// 批量获取消息详情
		messages := make([]map[string]interface{}, 0, len(msgIDs))
		for _, idStr := range msgIDs {
			msgData, err := global.Redis.HGetAll(ctx, msgCacheKey+idStr).Result()
			if err != nil || len(msgData) == 0 {
				continue
			}

			messages = append(messages, map[string]interface{}{
				"message_id":      msgData["message_id"],
				"conversation_id": msgData["conversation_id"],
				"sender_id":       msgData["sender_id"],
				"receiver_id":     msgData["receiver_id"],
				"content":         msgData["content"],
				"created_at":      msgData["created_at"],
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"total":    total,
			"page":     page,
			"size":     size,
			"messages": messages,
		})
		return
	}

	// 缓存未命中，从数据库查询
	var dbMessages []models.PrivateMessage
	if err := global.Db.Where("conversation_id = ?", convID).
		Order("created_at DESC").
		Limit(size).
		Offset((page - 1) * size).
		Find(&dbMessages).Error; err != nil {
		log.Printf("查询消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取消息失败"})
		return
	}

	// 获取总数
	var total int64
	global.Db.Where("conversation_id = ?", convID).Count(&total)

	// 重建缓存
	pipe := global.Redis.Pipeline()
	messages := make([]map[string]interface{}, 0, len(dbMessages))
	for _, msg := range dbMessages {
		msgIDStr := strconv.FormatUint(uint64(msg.MessageID), 10)
		msgKey := msgCacheKey + msgIDStr

		// 存储消息详情
		pipe.HSet(ctx, msgKey, map[string]interface{}{
			"message_id":      msg.MessageID,
			"conversation_id": msg.ConversationID,
			"sender_id":       msg.SenderID,
			"receiver_id":     msg.ReceiverID,
			"content":         msg.Content,
			"created_at":      msg.CreatedAt.Unix(),
		})
		pipe.Expire(ctx, msgKey, 24*time.Hour)

		// 添加到对话消息列表
		pipe.ZAdd(ctx, key, &redis.Z{
			Score:  float64(msg.CreatedAt.Unix()),
			Member: msg.MessageID,
		})

		// 构建返回数据
		messages = append(messages, map[string]interface{}{
			"message_id":      msg.MessageID,
			"conversation_id": msg.ConversationID,
			"sender_id":       msg.SenderID,
			"receiver_id":     msg.ReceiverID,
			"content":         msg.Content,
			"created_at":      msg.CreatedAt.Unix(),
		})
	}
	pipe.Expire(ctx, key, 24*time.Hour)
	_, _ = pipe.Exec(ctx)

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"page":     page,
		"size":     size,
		"messages": messages,
	})
}
