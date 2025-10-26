package utils

import (
	"Focogram/global"
	"Focogram/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"math/rand"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/golang-jwt/jwt"
)

func GenerateRandomUserIDSecure() string {
	const digits = "0123456789"
	const length = 9

	rand.Seed(time.Now().UnixNano())
	result := make([]byte, length)
	for i := range result {
		result[i] = digits[rand.Intn(len(digits))]
	}
	return string(result)
}

func HashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	return string(hash), err
}
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateJWT(userid string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid": userid,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	})

	signedToken, err := token.SignedString([]byte("secret"))
	return "Bearer " + signedToken, err
}

func ParseJWT(tokenstring string) (string, error) {
	if len(tokenstring) > 7 && tokenstring[:7] == "Bearer " {
		tokenstring = tokenstring[7:]
	}
	token, err := jwt.Parse(tokenstring, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte("secret"), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userid, ok := claims["userid"].(string)
		if !ok {
			return "", errors.New("invalid token")
		}
		return userid, err
	}
	return "", nil
}

// GetUserPostedNum 获取用户当前已发帖的数目
func GetUserPostedNum(userid string) (int, error) {
	if userid == "" {
		return 0, fmt.Errorf("用户ID不能为空")
	}

	var user models.User
	err := global.Db.Model(&models.User{}).
		Where("userid = ?", userid).
		First(&user).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("用户不存在")
		}
		return 0, fmt.Errorf("查询用户信息失败: %w", err)
	}
	return user.PostNum, nil
}

func AddPostNum(userid string) error {
	return global.Db.Model(&models.User{}).Where("userid=?", userid).Update("post_num", gorm.Expr("post_num + ?", 1)).Error
}
func SubPostNum(userid string) error {
	return global.Db.Model(&models.User{}).Where("userid=?", userid).Update("post_num", gorm.Expr("post_num - ?", 1)).Error
}

// 通用接口限流和防抖检查
func CheckRateLimitAndDebounce(actionType, userid, targetID string) error {
	ctx := context.Background()

	// 接口限流，使用操作类型区分
	limitkey := fmt.Sprintf("%s_limit:%s", actionType, userid)
	count, err := global.Redis.Incr(ctx, limitkey).Result()
	if err != nil {
		return fmt.Errorf("系统繁忙，请稍后再试")
	}
	if count == 1 {
		global.Redis.Expire(ctx, limitkey, 1*time.Minute)
	}

	// 根据不同操作类型设置不同的限制
	var maxLimit int64
	switch actionType {
	case "like":
		maxLimit = 50 // 点赞每分钟50次
	case "comment":
		maxLimit = 20 // 评论每分钟20次
	default:
		maxLimit = 30 // 默认限制
	}

	if count > maxLimit {
		return fmt.Errorf("操作过于频繁，请稍后再试")
	}

	// 防抖操作，使用操作类型区分
	lastActionkey := fmt.Sprintf("%s_last_action:%s:%s", actionType, targetID, userid)
	lastTimeStr, err := global.Redis.Get(ctx, lastActionkey).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("系统错误")
	}

	if lastTimeStr != "" {
		lastTime, err := time.Parse(time.RFC3339Nano, lastTimeStr)
		if err == nil && time.Since(lastTime) < 300*time.Millisecond {
			return fmt.Errorf("操作过于频繁")
		}
	}

	global.Redis.Set(ctx, lastActionkey, time.Now().Format(time.RFC3339Nano), 1*time.Minute)
	return nil
}

// CheckPostAuthor 检查用户是否为帖子作者
func CheckPostAuthor(postid string, userid string) (bool, error) {
	if postid == "" || userid == "" {
		return false, nil // 空参数直接返回无权限
	}

	// 查询帖子作者ID
	var post models.Post
	err := global.Db.Model(&models.Post{}).
		Where("postid = ?", postid).
		First(&post).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 帖子不存在时，返回无权限
			return false, nil
		}
		// 数据库查询错误返回错误信息
		return false, fmt.Errorf("查询帖子失败: %w", err)
	}

	// 比较帖子作者ID与当前用户ID
	return post.Userid == userid, nil
}
