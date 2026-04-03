package controllers

import (
	"Focogram/global"
	"Focogram/models"
	"Focogram/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成随机账号
	userID := utils.GenerateRandomUserIDSecure()
	// 加密密码
	hashPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 创建用户对象
	user := models.User{
		Userid:   userID,
		Username: req.Username,
		Email:    req.Email,
		Password: hashPwd,
		Gender:   req.Gender,
		Age:      req.Age,
		Describe: "",
		Address:  "",
	}

	// 生成JWT token
	token, err := utils.GenerateJWT(user.Userid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 创建用户
	if err := global.Db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
		"token":   token,
		"user_id": user.Userid,
	})
}

func Login(c *gin.Context) {
	var input struct {
		Userid   string `json:"userid"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user models.User

	if err := global.Db.Where("userid = ?", input.Userid).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号错误"})
		return
	}
	if !utils.CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}
	token, err := utils.GenerateJWT(user.Userid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
	})
}

func GetUserInfo(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索不能为空"})
		return
	}

	var users []models.User

	//同时搜索用户名和账号
	if err := global.Db.Where("userid = ? OR username = ?", keyword, keyword).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//检查是否找到用户
	if len(users) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到用户"})
		return
	}
	//构建返回的用户信息列表
	var userList []gin.H

	for _, user := range users {
		userList = append(userList, gin.H{
			"user_id":  user.Userid,
			"username": user.Username,
			"email":    user.Email,
			"gender":   user.Gender,
			"age":      user.Age,
			"describe": user.Describe,
			"address":  user.Address,
			"post_num": user.PostNum,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"counr":   len(users),
		"keyword": keyword,
		"users":   userList,
	})

}

func UpdateUserInfo(c *gin.Context) {
	var input struct {
		Username string `json:"username" validate:"omitempty,min=1,max=15"`
		Gender   string `json:"gender" validate:"omitempty,oneof=男 女 未知"`
		Age      int    `json:"age" validate:"omitempty,min=0,max=150"`
		Describe string `json:"describe" validate:"omitempty,max=100"`
		Address  string `json:"address" validate:"omitempty,max=50"`
	}
	//绑定并验证输入
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//获取当前用户ID
	userid := c.GetString("userid")
	if userid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}
	var user models.User
	if err := global.Db.Where("userid = ?", userid).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updateData := make(map[string]interface{})
	if input.Username != "" {
		updateData["username"] = input.Username
	}
	if input.Gender != "" {
		updateData["gender"] = input.Gender
	}
	if input.Age != 0 {
		updateData["age"] = input.Age
	}
	if input.Describe != "" {
		updateData["describe"] = input.Describe
	}
	if input.Address != "" {
		updateData["address"] = input.Address
	}
	if err := global.Db.Model(&user).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
	})
}

// UpdatePassword 处理用户密码更新请求
// 流程：验证输入 -> 确认用户身份 -> 校验旧密码 -> 加密新密码 -> 更新数据库
func UpdatePassword(c *gin.Context) {
	// 1. 绑定并验证输入参数
	var req models.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "输入参数错误：" + err.Error()})
		return
	}

	// 2. 获取当前登录用户ID（从AuthMiddleware中传递的上下文）
	userID := c.GetString("userid")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问，请先登录"})
		return
	}

	// 3. 查询数据库，获取用户信息（用于验证旧密码）
	var user models.User
	if err := global.Db.Where("userid = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "身份验证失败，用户不存在或已删除"})
		return
	}

	// 4. 验证旧密码是否正确（防止他人恶意修改）
	if !utils.CheckPasswordHash(req.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "旧密码错误，请重新输入"})
		return
	}

	// 5. 加密新密码（使用bcrypt算法，避免明文存储）
	hashedNewPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败：" + err.Error()})
		return
	}

	// 6. 更新数据库中的密码字段
	if err := global.Db.Model(&user).Update("password", hashedNewPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码更新失败：" + err.Error()})
		return
	}

	// 7. 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message": "密码更新成功，请使用新密码登录",
	})
}
