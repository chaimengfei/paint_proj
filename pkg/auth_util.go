package pkg

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetOperatorShopID 获取操作员的有效店铺ID
// 优先使用前端传递的shopID，但需要验证权限
// 如果前端没有传递shopID，则使用JWT中的店铺ID
func GetOperatorShopID(c *gin.Context, frontendShopID int64) (int64, error) {
	operatorShopID := c.GetInt64("shop_id")
	isRoot := c.GetBool("is_root")

	// 如果前端没有传递shop_id，使用JWT中的店铺ID
	if frontendShopID == 0 {
		return operatorShopID, nil
	}

	// 如果前端传递了shop_id，需要验证权限
	if !isRoot && frontendShopID != operatorShopID {
		c.JSON(http.StatusForbidden, gin.H{"code": -1, "message": "无权限操作该店铺的数据"})
		return 0, gin.Error{Err: nil, Type: gin.ErrorTypePublic}
	}

	return frontendShopID, nil
}

// ValidateShopPermission 验证店铺权限
// 如果验证失败，会直接返回错误响应
func ValidateShopPermission(c *gin.Context, targetShopID int64) (int64, bool) {
	operatorShopID := c.GetInt64("shop_id")
	isRoot := c.GetBool("is_root")

	// 如果目标店铺ID为0，使用操作员的店铺ID
	if targetShopID == 0 {
		return operatorShopID, true
	}

	// 验证权限：超级管理员可以操作所有店铺，普通管理员只能操作自己的店铺
	if !isRoot && targetShopID != operatorShopID {
		c.JSON(http.StatusForbidden, gin.H{"code": -1, "message": "无权限操作该店铺的数据"})
		return 0, false
	}

	return targetShopID, true
}
