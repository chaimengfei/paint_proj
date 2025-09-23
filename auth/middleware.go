package auth

import (
	"cmf/paint_proj/pkg"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析 Header 里的 Token
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "缺少 Authorization header"})
			return
		}

		// 移除 "Bearer " 前缀
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		userID, shopID, err := pkg.ParseJWTToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效 token"})
			return
		}

		c.Set("user_id", userID)
		c.Set("shop_id", shopID)
		c.Next()
	}
}
