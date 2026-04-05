package models

import (
	"time"

	"gorm.io/gorm"
)

type Conversation struct {
	ConversationID uint             `gorm:"primaryKey;autoIncrement" json:"conversation_id"` // 自增主键（累加形式）
	User1ID        string           `gorm:"size:36;not null" json:"user1_id"`
	User2ID        string           `gorm:"size:36;not null" json:"user2_id"`
	LastMessage    string           `gorm:"type:text" json:"last_message"`
	LastMessageAt  time.Time        `json:"last_message_at"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	DeletedAt      gorm.DeletedAt   `gorm:"index" json:"-"`
	User1          User             `gorm:"foreignKey:User1ID;references:Userid" json:"user1,omitempty"`
	User2          User             `gorm:"foreignKey:User2ID;references:Userid" json:"user2,omitempty"`
	Messages       []PrivateMessage `gorm:"foreignKey:ConversationID" json:"messages,omitempty"` // 关联该对话的所有消息
}

type PrivateMessage struct {
	MessageID      uint           `gorm:"primaryKey;autoIncrement" json:"message_id"` // 自增主键（累加形式）
	ConversationID uint           `gorm:"not null" json:"conversation_id"`            // 关联对话ID（同步改为uint）
	SenderID       string         `gorm:"size:36;not null" json:"sender_id"`
	ReceiverID     string         `gorm:"size:36;not null" json:"receiver_id"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	IsRead         bool           `gorm:"default:false" json:"is_read"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	Sender         User           `gorm:"foreignKey:SenderID;references:Userid" json:"sender,omitempty"`
	Receiver       User           `gorm:"foreignKey:ReceiverID;references:Userid" json:"receiver,omitempty"`
}

// 发送私信请求
type SendMessageRequest struct {
	//ReceiverID string `json:"receiver_id" validate:"required"`
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// 创建对话请求
type CreateConversationRequest struct {
	TargetUserID string `json:"target_user_id" validate:"required"`
}
