package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	Commentid string `gorm:"primarykey;type:varchar(36);not null;unique" json:"commentid"`
	Content   string `gorm:"type:text;size:500;not null"`
	Postid    string `gorm:"type:varchar(36);not null;" json:"postid"`
	Userid    string `gorm:"type:varchar(36);not null" json:"userid"`
}

type CommentRequest struct {
	Content string `gorm:"type:text;size:500;not null"`
}

func CreateCommentid() string {
	return uuid.New().String()
}
