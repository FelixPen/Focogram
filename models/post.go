package models

import (
	"Focogram/global"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	Postid  string `gorm:"primarykey;size:36;not null;unique" json:"postid"`
	Userid  string `gorm:"size:36;not null;index" json:"userid"` // 外键，需与 users.userid 类型一致
	Content string `gorm:"type:text;size:500;not null"`

	//关联：一个帖子属于一个用户（多对一）
	User User `gorm:"foreignKey:Userid;references:Userid"`
}

type PostRequest struct {
	Content string `gorm:"type:text;size:500;not null"`
}

// 假设原函数如下（请根据实际路径调整）：
func GetPostFromDB(postid string) (*Post, error) {
	var post Post
	err := global.Db.Where("postid = ?", postid).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// 判断帖子是否存在
func CheckPostExists(postid string) (bool, error) {
	if postid == "" {
		log.Printf("CheckPostExists: 传入的postid为空")
		return false, fmt.Errorf("帖文ID不能为空")
	}

	var post Post
	result := global.Db.Where("postid = ?", postid).First(&post)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("CheckPostExists: 帖子不存在 (postid=%s)，数据库无此记录", postid)
			return false, fmt.Errorf("帖文不存在或已被删除")
		}
		log.Printf("CheckPostExists: 查询数据库失败 (postid=%s)，错误: %v", postid, result.Error)
		return false, fmt.Errorf("系统错误，请稍后再试")
	}

	// 修复点：正确判断是否逻辑删除
	if post.DeletedAt.Valid && !post.DeletedAt.Time.IsZero() {
		log.Printf("CheckPostExists: 帖子已被逻辑删除 (postid=%s)", postid)
		return false, fmt.Errorf("帖文不存在或已被删除")
	}

	log.Printf("CheckPostExists: 帖子存在 (postid=%s, 作者=%s)", postid, post.Userid)
	return true, nil
}

func CheckPostAuthor(postid, userid string) (bool, error) {
	post, err := GetPostFromDB(postid)
	if err != nil {
		return false, err
	}
	return post.Userid == userid, nil
}

func GetPostAuthor(postid string) (string, error) {
	post, err := GetPostFromDB(postid)
	if err != nil {
		return "", err
	}
	return post.Userid, nil
}

func CreatePostid() string {
	return uuid.New().String()
}
