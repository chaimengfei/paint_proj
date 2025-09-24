package auth

import (
	"cmf/paint_proj/pkg"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 小程序认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析 Header 里的 Token
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "缺少 Authorization "})
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
		c.Set("auth_type", "mini_program")
		c.Next()
	}
}

// 后台管理认证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析 Header 里的 Token
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "web后台缺少 Authorization header"})
			return
		}

		// 移除 "Bearer " 前缀
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		operatorID, operatorName, shopID, isRoot, err := pkg.ParseAdminJWTToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效 token"})
			return
		}

		c.Set("operator_id", operatorID)
		c.Set("operator_name", operatorName)
		c.Set("shop_id", shopID)
		c.Set("is_root", isRoot)
		c.Set("auth_type", "admin")
		c.Next()
	}
}
