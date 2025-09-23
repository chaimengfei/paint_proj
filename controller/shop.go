package controller

import (
	"cmf/paint_proj/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShopController struct {
	shopService service.ShopService
}

func NewShopController(shopService service.ShopService) *ShopController {
	return &ShopController{
		shopService: shopService,
	}
}

// GetShopList 获取店铺列表
// @Summary 获取店铺列表
// @Description 获取所有启用的店铺信息，无需token验证
// @Tags 店铺
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/shops [get]
// @Router /admin/shops [get]
func (sc *ShopController) GetShopList(c *gin.Context) {
	shopList, err := sc.shopService.GetAllActiveShopsSimple()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "获取店铺列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    shopList,
		"message": "获取店铺列表成功",
	})
}
