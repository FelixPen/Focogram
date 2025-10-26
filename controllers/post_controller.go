package controllers

import (
	"Focogram/global"
	"Focogram/models"
	"Focogram/utils"

	"net/http"

	"github.com/gin-gonic/gin"
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
		Postid:  postid,
		Userid:  userid,
		Content: input.Content,
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

func GetUserPosts(c *gin.Context) {
	var post []models.Post
	userid := c.Param("userid")
	if userid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	if err := global.Db.Where("userid=?", userid).Order("display_order").Find(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//构建只包含发布时间、内容、作者的响应数据
	var response []gin.H
	for _, post := range post {
		response = append(response, gin.H{
			"postid":   post.Postid,
			"content":  post.Content,
			"posttime": post.CreatedAt,
			"userid":   post.Userid,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"count": len(post),
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
	if err := global.Db.Where("postid=? AND userid=?", postid, userid).First(&post).Error; err != nil {
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
	if err := global.Db.Delete(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败：" + err.Error()})
		return
	}
	if err := utils.SubPostNum(userid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})

}
