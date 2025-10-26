package middlewares

import (
	"Focogram/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"error": "未授权"})
			c.Abort()
			return
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
