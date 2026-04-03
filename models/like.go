package models

import (
	"Focogram/global"

	"gorm.io/gorm"
)

type Like struct {
	gorm.Model
	Postid string `gorm:"type:varchar(36);not null;index:idx_post_user,unique" json:"postid"`
	Userid string `gorm:"type:varchar(36);not null;index:idx_post_user,unique" json:"userid"`
	Liked  bool   `gorm:"default:true" json:"liked"`
}

// 从数据库中获取帖子的点赞记录
func GetLikeFromDB(postid string) ([]Like, error) {
	var likes []Like
	err := global.Db.Where("postid=? AND liked=?", postid, true).Find(&likes).Error
	return likes, err
}

// 批量创建点赞记录
func CreateLikes(likes []Like) error {
	return global.Db.Create(&likes).Error
}
func (l *Like) Reset() {
	l.Postid = ""
	l.Userid = ""
	l.Liked = false
}
