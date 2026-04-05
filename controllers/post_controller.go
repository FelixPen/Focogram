package controllers

import (
	"Focogram/global"
	"Focogram/models"
	"Focogram/utils"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func CreatePost(c *gin.Context) {
	var input models.PostRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userid := c.GetString("userid")
	if userid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	postid := models.CreatePostid()

	var post = models.Post{
		Postid:   postid,
		Userid:   userid,
		Content:  input.Content,
		ImageUrl: input.ImageUrl,
	}

	if err := global.Db.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := utils.AddPostNum(userid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "发布成功"})

}

type PostWithUsername struct {
	Postid      string
	Content     string
	ImageUrl    string
	CreatedAt   string
	Userid      string
	Username    string
	AvatarUrl   string
	AvatarColor string
}

func GetPostDetail(c *gin.Context) {
	postid := c.Param("postid")
	if postid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少帖子ID"})
		return
	}
	type PostWithUsername struct {
		Postid      string
		Content     string
		ImageUrl    string
		CreatedAt   string
		Userid      string
		Username    string
		AvatarUrl   string
		AvatarColor string
	}

	var post PostWithUsername
	err := global.Db.Table("posts").
		Select("posts.postid, posts.content, posts.image_url, posts.created_at, posts.userid, users.username, users.avatar_url, users.avatar_color").
		Joins("LEFT JOIN users ON posts.userid = users.userid").
		Where("posts.postid = ?", postid).
		Scan(&post).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if post.Postid == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		return
	}

	avatarURL := post.AvatarUrl
	if avatarURL != "" && !strings.HasPrefix(avatarURL, "http") {
		avatarURL = "http://localhost:8080" + avatarURL
	}

	c.JSON(http.StatusOK, gin.H{
		"postid":      post.Postid,
		"content":     post.Content,
		"image_url":   post.ImageUrl,
		"posttime":    post.CreatedAt,
		"userid":      post.Userid,
		"username":    post.Username,
		"avatarUrl":   avatarURL,
		"avatarColor": post.AvatarColor,
	})
}

func GetUserPosts(c *gin.Context) {
	var posts []PostWithUsername
	userid := c.Param("userid")
	if userid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	err := global.Db.Table("posts").
		Select("posts.postid, posts.content, posts.image_url, posts.created_at, posts.userid, users.username, users.avatar_url, users.avatar_color").
		Joins("LEFT JOIN users ON posts.userid = users.userid").
		Where("posts.userid = ?", userid).
		Order("posts.created_at DESC").
		Scan(&posts).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []gin.H
	for _, post := range posts {
		avatarURL := post.AvatarUrl
		if avatarURL != "" && !strings.HasPrefix(avatarURL, "http") {
			avatarURL = "http://localhost:8080" + avatarURL
		}
		response = append(response, gin.H{
			"postid":      post.Postid,
			"content":     post.Content,
			"image_url":   post.ImageUrl,
			"posttime":    post.CreatedAt,
			"userid":      post.Userid,
			"username":    post.Username,
			"avatarUrl":   avatarURL,
			"avatarColor": post.AvatarColor,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"count": len(posts),
		"posts": response,
	})
}

func GetUserLikedPosts(c *gin.Context) {
	var posts []PostWithUsername
	userid := c.Param("userid")
	if userid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户ID"})
		return
	}

	err := global.Db.Table("likes").
		Select("posts.postid, posts.content, posts.created_at, posts.userid, users.username, users.avatar_url, users.avatar_color").
		Joins("LEFT JOIN posts ON likes.postid = posts.postid").
		Joins("LEFT JOIN users ON posts.userid = users.userid").
		Where("likes.userid = ? AND likes.liked = ?", userid, true).
		Order("posts.created_at DESC").
		Scan(&posts).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []gin.H
	for _, post := range posts {
		avatarURL := post.AvatarUrl
		if avatarURL != "" && !strings.HasPrefix(avatarURL, "http") {
			avatarURL = "http://localhost:8080" + avatarURL
		}
		response = append(response, gin.H{
			"postid":      post.Postid,
			"content":     post.Content,
			"image_url":   post.ImageUrl,
			"posttime":    post.CreatedAt,
			"userid":      post.Userid,
			"username":    post.Username,
			"avatarUrl":   avatarURL,
			"avatarColor": post.AvatarColor,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"count": len(posts),
		"posts": response,
	})
}

func DeletePost(c *gin.Context) {
	userid := c.GetString("userid")
	postid := c.Param("postid")
	if userid == "" || postid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var post models.Post
	if err := global.Db.Unscoped().Where("postid=? AND userid=?", postid, userid).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "帖子不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tx := global.Db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}

	if err := tx.Where("postid=?", postid).Delete(&models.Like{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除点赞失败：" + err.Error()})
		return
	}

	if err := tx.Where("postid=?", postid).Delete(&models.Comment{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除评论失败：" + err.Error()})
		return
	}

	if err := tx.Unscoped().Where("postid=?", postid).Delete(&models.Post{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败：" + err.Error()})
		return
	}

	ctx := context.Background()
	pipe := global.Redis.Pipeline()

	likeCacheKey := fmt.Sprintf("like:posts:%s", postid)
	pipe.Del(ctx, likeCacheKey)

	postCommentsKey := fmt.Sprintf("post:comments:%s", postid)
	commentIds, err := global.Redis.ZRange(ctx, postCommentsKey, 0, -1).Result()
	if err == nil && len(commentIds) > 0 {
		for _, cid := range commentIds {
			commentKey := fmt.Sprintf("comment:%s", cid)
			pipe.Del(ctx, commentKey)
		}
	}
	pipe.Del(ctx, postCommentsKey)

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		global.Db.Logger.Error(ctx, "Redis清理帖子缓存失败: "+err.Error())
	}

	if err := utils.SubPostNum(userid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func GetFollowingPosts(c *gin.Context) {
	userid := c.GetString("userid")
	if userid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	type PostWithUsername struct {
		Postid      string
		Content     string
		ImageUrl    string
		CreatedAt   string
		Userid      string
		Username    string
		AvatarUrl   string
		AvatarColor string
	}

	var posts []PostWithUsername

	err := global.Db.Table("posts").
		Select("posts.postid, posts.content, posts.image_url, posts.created_at, posts.userid, users.username, users.avatar_url, users.avatar_color").
		Joins("LEFT JOIN users ON posts.userid = users.userid").
		Where(`posts.userid = ? OR posts.userid IN (
			SELECT followedid FROM follows WHERE followerid = ?
		)`, userid, userid).
		Order("posts.created_at DESC").
		Scan(&posts).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []gin.H
	for _, post := range posts {
		avatarURL := post.AvatarUrl
		if avatarURL != "" && !strings.HasPrefix(avatarURL, "http") {
			avatarURL = "http://localhost:8080" + avatarURL
		}
		response = append(response, gin.H{
			"postid":      post.Postid,
			"content":     post.Content,
			"image_url":   post.ImageUrl,
			"posttime":    post.CreatedAt,
			"userid":      post.Userid,
			"username":    post.Username,
			"avatarUrl":   avatarURL,
			"avatarColor": post.AvatarColor,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"count": len(posts),
		"posts": response,
	})
}
