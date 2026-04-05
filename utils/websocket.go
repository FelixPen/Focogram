package utils

import (
	"Focogram/global"
	"Focogram/models"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

// Redis键前缀
const (
	UserNotificationSetPrefix = "notification:user:" // 每个用户的消息集合 key: notification:user:{userid}
)

// WebSocket连接管理器
type WsManager struct {
	clients    map[string]*websocket.Conn // userid -> WebSocket连接
	clientMux  sync.RWMutex               // 保护clients的互斥锁
	broadcast  chan *models.Notification  // 广播消息通道
	register   chan *Client               // 注册通道
	unregister chan *Client               // 注销通道
}

// 客户端连接
type Client struct {
	conn   *websocket.Conn
	userID string
}

// 全局WebSocket管理器
var WsMgr = &WsManager{
	clients:    make(map[string]*websocket.Conn),
	broadcast:  make(chan *models.Notification, 100),
	register:   make(chan *Client, 100),
	unregister: make(chan *Client, 100),
}

// 启动WebSocket管理器
func (m *WsManager) Start() {
	for {
		select {
		case client := <-m.register:
			m.clientMux.Lock()
			m.clients[client.userID] = client.conn
			log.Printf("用户 %s 已连接WebSocket，当前在线: %d", client.userID, len(m.clients))
			m.clientMux.Unlock()

		case client := <-m.unregister:
			m.clientMux.Lock()
			if _, ok := m.clients[client.userID]; ok {
				delete(m.clients, client.userID)
				client.conn.Close()
				log.Printf("用户 %s 已断开WebSocket，当前在线: %d", client.userID, len(m.clients))
			}
			m.clientMux.Unlock()

		case notification := <-m.broadcast:
			// 向指定用户推送通知
			m.clientMux.RLock()
			conn, exists := m.clients[notification.Userid]
			m.clientMux.RUnlock()

			if exists {
				// 格式化消息
				msg, err := json.Marshal(map[string]interface{}{
					"content_type": notification.ContentType,
					"senderid":     notification.Senderid,
					"content":      notification.Content,
					"time":         notification.CreatedAt.Format("2006-01-02 15:04:05"),
					"id":           notification.ID,
				})
				if err != nil {
					log.Printf("❌ WebSocket消息序列化失败: %v", err)
					continue
				}

				// 发送消息
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					log.Printf("❌ WebSocket消息发送失败: %v", err)
					// 发送失败时移除连接
					m.unregister <- &Client{userID: notification.Userid, conn: conn}
				}
			}
		}
	}
}

// WebSocket升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有跨域请求，生产环境需根据实际情况修改
		return true
	},
}

// WebSocket连接处理函数
func WsHandler(c *gin.Context) {
	// 从JWT获取用户ID
	userID := c.GetString("userid")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	client := &Client{
		conn:   conn,
		userID: userID,
	}

	// 注册客户端
	WsMgr.register <- client

	// 补发离线未读私信 + 未读统计
	go SendOfflineMessages(userID, conn)

	// 监听客户端消息（主要用于心跳检测）
	go func() {
		defer func() {
			WsMgr.unregister <- client
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket读取错误: %v", err)
				}
				break
			}
		}
	}()
}

// 补发离线未读私信及未读统计
func SendOfflineMessages(userID string, conn *websocket.Conn) {
	// 1. 获取每个对话的未读消息数
	type UnreadCount struct {
		ConversationID uint   `json:"conversation_id"`
		Count          int64  `json:"count"`
		OtherUserID    string `json:"other_user_id"`
	}
	var unreadCounts []UnreadCount

	rows, err := global.Db.Table("private_messages").
		Select("conversation_id, COUNT(*) as count, sender_id as other_user_id").
		Where("receiver_id = ? AND is_read = ?", userID, false).
		Group("conversation_id, sender_id").
		Rows()

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var uc UnreadCount
			rows.Scan(&uc.ConversationID, &uc.Count, &uc.OtherUserID)
			unreadCounts = append(unreadCounts, uc)
		}
	}

	// 2. 计算总未读数
	totalUnread := int64(0)
	for _, uc := range unreadCounts {
		totalUnread += uc.Count
	}

	// 3. 先发送未读汇总消息（像微信一样）
	summaryMsg, _ := json.Marshal(map[string]interface{}{
		"content_type":   "message_unread_summary",
		"total_unread":   totalUnread,
		"unread_details": unreadCounts,
		"time":           time.Now().Format("2006-01-02 15:04:05"),
	})
	conn.WriteMessage(websocket.TextMessage, summaryMsg)

	// 4. 补发所有未读私信
	var unreadMessages []models.PrivateMessage
	global.Db.Where("receiver_id = ? AND is_read = ?", userID, false).
		Order("created_at ASC").
		Find(&unreadMessages)

	for _, msg := range unreadMessages {
		msgData, _ := json.Marshal(map[string]interface{}{
			"content_type":    "private_message",
			"message_id":      msg.MessageID,
			"conversation_id": msg.ConversationID,
			"sender_id":       msg.SenderID,
			"content":         msg.Content,
			"created_at":      msg.CreatedAt.Format("2006-01-02 15:04:05"),
			"is_read":         msg.IsRead,
		})
		conn.WriteMessage(websocket.TextMessage, msgData)
	}
}

// 推送私信到WebSocket
func PushPrivateMessage(msg *models.PrivateMessage) {
	WsMgr.clientMux.RLock()
	conn, exists := WsMgr.clients[msg.ReceiverID]
	WsMgr.clientMux.RUnlock()

	if exists {
		msgData, _ := json.Marshal(map[string]interface{}{
			"content_type":    "private_message",
			"message_id":      msg.MessageID,
			"conversation_id": msg.ConversationID,
			"sender_id":       msg.SenderID,
			"content":         msg.Content,
			"created_at":      msg.CreatedAt.Format("2006-01-02 15:04:05"),
			"is_read":         msg.IsRead,
		})
		conn.WriteMessage(websocket.TextMessage, msgData)
	}
}

// 推送通知到WebSocket并写入缓存
func PushToWebSocket(notification *models.Notification) {
	// 1. 保存到数据库
	if err := global.Db.Create(notification).Error; err != nil {
		log.Printf("保存通知到数据库失败: %v", err)
		return
	}

	// 2. 写入Redis缓存（用户的消息集合）
	ctx := context.Background()
	// 使用消息的创建时间戳作为分数（保证排序）
	timestamp := notification.CreatedAt.UnixNano() // 纳秒级时间戳，确保唯一性
	if err := AddNotificationToUserSet(ctx, notification.Userid, uint64(notification.ID), timestamp); err != nil {
		log.Printf("添加消息到Redis缓存失败: %v", err)
	} else {
		// 设置过期时间（例如7天，可根据业务调整）
		SetUserNotificationSetExpire(ctx, notification.Userid, 7*24*time.Hour)
	}

	// 3. 推送到WebSocket
	WsMgr.broadcast <- notification
}

// 添加消息ID到用户的Redis集合
// userID: 接收消息的用户ID
// notificationID: 消息ID
// timestamp: 消息创建时间戳（用于排序）
func AddNotificationToUserSet(ctx context.Context, userID string, notificationID uint64, timestamp int64) error {
	key := UserNotificationSetPrefix + userID
	// ZAdd: 将消息ID添加到有序集合，分数为时间戳（保证按时间排序）
	return global.Redis.ZAdd(ctx, key, &redis.Z{
		Score:  float64(timestamp),
		Member: notificationID,
	}).Err()
}

// 从用户的Redis集合中获取消息ID列表（分页）
// userID: 用户ID
// page: 页码（从1开始）
// size: 每页数量
// return: 消息ID列表、总数量
func GetNotificationIDsFromUserSet(ctx context.Context, userID string, page, size int) ([]uint64, int64, error) {
	key := UserNotificationSetPrefix + userID
	start := int64((page - 1) * size)
	end := int64(page*size - 1)

	// ZCard: 获取集合总数量
	total, err := global.Redis.ZCard(ctx, key).Result()
	if err != nil {
		return nil, 0, err
	}

	// ZRevRange: 按分数倒序（最新的在前）获取指定范围的消息ID
	// 注意：返回的是string类型，需要转换为uint64
	members, err := global.Redis.ZRevRange(ctx, key, start, end).Result()
	if err != nil {
		return nil, 0, err
	}

	// 转换为uint64
	ids := make([]uint64, 0, len(members))
	for _, m := range members {
		id, err := strconv.ParseUint(m, 10, 64)
		if err != nil {
			continue // 忽略无效ID
		}
		ids = append(ids, id)
	}

	return ids, total, nil
}

// 设置用户消息集合的过期时间（可选，避免缓存膨胀）
func SetUserNotificationSetExpire(ctx context.Context, userID string, expire time.Duration) error {
	key := UserNotificationSetPrefix + userID
	return global.Redis.Expire(ctx, key, expire).Err()
}
