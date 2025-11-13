package models

import (
	"gorm.io/gorm"
)

const (
	NotificationTypeLike    = "like"
	NotificationTypeComment = "comment"
	NotificationTypeFollow  = "follow"
	NotificationTypeMessage = "message"
)

type Notification struct {
	gorm.Model
	Userid      string `gorm:"type:varchar(36);not null;index" json:"userid"`    // 接收通知的用户ID
	Senderid    string `gorm:"type:varchar(36);not null;index" json:"senderid"`  // 发送通知的用户ID
	ContentType string `gorm:"type:varchar(20);not null" json:"content_type"`    // 通知类型
	Contentid   string `gorm:"type:varchar(36);not null;index" json:"contentid"` // 关联内容ID
	Content     string `gorm:"type:text;size:500;not null" json:"content"`       // 通知内容
}
