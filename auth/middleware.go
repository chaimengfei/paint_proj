package auth

import (
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		/*// 解析 Header 里的 Token
		tokenStr := c.GetHeader("Authorization")
		userID, err := pkg.ParseJWTToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效 token"})
			return
		}
		c.Set("user_id", userID)
		c.Next()*/

		//TODO 柴梦妃 临时测试
		userID := int64(123)
		c.Set("user_id", userID)
		c.Next()
	}
}
