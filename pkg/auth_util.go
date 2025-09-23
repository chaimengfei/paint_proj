package pkg

import "github.com/gin-gonic/gin"

// GetOperatorShopID 获取操作者的店铺ID（考虑超级管理员权限）
func GetOperatorShopID(c *gin.Context, targetShopID int64) int64 {
	authType := c.GetString("auth_type")

	if authType == "admin" {
		isRoot := c.GetBool("is_root")
		operatorShopID := c.GetInt64("shop_id")

		if isRoot {
			// 超级管理员可以操作指定店铺，如果不指定则操作所有店铺
			if targetShopID > 0 {
				return targetShopID
			}
			return 0 // 0 表示所有店铺
		}

		// 普通管理员只能操作自己店铺
		return operatorShopID
	}

	// 小程序用户
	return c.GetInt64("shop_id")
}

// IsRootOperator 检查是否为超级管理员
func IsRootOperator(c *gin.Context) bool {
	authType := c.GetString("auth_type")
	if authType == "admin" {
		return c.GetBool("is_root")
	}
	return false
}

// GetOperatorID 获取操作者ID
func GetOperatorID(c *gin.Context) int64 {
	authType := c.GetString("auth_type")

	if authType == "admin" {
		return c.GetInt64("operator_id")
	}

	// 小程序用户
	return c.GetInt64("user_id")
}
