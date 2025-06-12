package auth

import "github.com/gin-gonic/gin"

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 假设通过 token 验证用户并获取 userID
		userID := int64(123)
		c.Set("user_id", userID)
		c.Next()
	}
}
