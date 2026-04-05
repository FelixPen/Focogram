package middlewares

import (
	"Focogram/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		// 如果Header中没有token，尝试从URL参数中获取（用于WebSocket连接）
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			c.JSON(401, gin.H{"error": "未授权：缺少 Token"})
			c.Abort()
			return
		}

		// 处理 Bearer 前缀
		const prefix = "Bearer "
		if len(token) > len(prefix) && token[:len(prefix)] == prefix {
			token = token[len(prefix):]
		}

		userid, err := utils.ParseJWT(token)

		if err != nil {
			c.JSON(401, gin.H{"error": "未授权"})
			c.Abort()
			return
		}
		c.Set("userid", userid)
		c.Next()

	}
}
