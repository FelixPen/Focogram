package controllers

import (
	"Focogram/global"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UploadAvatar(c *gin.Context) {
	userid := c.GetString("userid")
	if userid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的图片"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支持 JPG、PNG、GIF、WEBP 格式"})
		return
	}

	maxSize := int64(5 * 1024 * 1024)
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片大小不能超过 5MB"})
		return
	}

	filename := fmt.Sprintf("%s_%s%s", userid, uuid.New().String()[:8], ext)
	savePath := filepath.Join("uploads", "avatars", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败: " + err.Error()})
		return
	}

	avatarURL := fmt.Sprintf("http://localhost:8080/uploads/avatars/%s", filename)

	c.JSON(http.StatusOK, gin.H{
		"message":   "上传成功",
		"avatarUrl": avatarURL,
	})
}

func UploadPostImage(c *gin.Context) {
	userid := c.GetString("userid")
	if userid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的图片"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支持 JPG、PNG、GIF、WEBP 格式"})
		return
	}

	maxSize := int64(10 * 1024 * 1024)
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片大小不能超过 10MB"})
		return
	}

	filename := fmt.Sprintf("post_%s_%s%s", userid, uuid.New().String()[:8], ext)
	savePath := filepath.Join("uploads", "posts", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败: " + err.Error()})
		return
	}

	imageURL := fmt.Sprintf("http://localhost:8080/uploads/posts/%s", filename)

	c.JSON(http.StatusOK, gin.H{
		"message":  "上传成功",
		"imageUrl": imageURL,
	})
}

func init() {
	go func() {
		time.Sleep(1 * time.Second)
		global.Db.Exec("ALTER TABLE users ADD COLUMN avatar_url VARCHAR(500) DEFAULT ''")
		global.Db.Exec("ALTER TABLE users ADD COLUMN banner_color VARCHAR(200) DEFAULT ''")
		global.Db.Exec("ALTER TABLE users ADD COLUMN birth_date VARCHAR(20) DEFAULT ''")
		global.Db.Exec("ALTER TABLE posts ADD COLUMN image_url VARCHAR(500) DEFAULT ''")
	}()
}
