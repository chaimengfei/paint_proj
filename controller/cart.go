package controller

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CartController struct {
	cartService service.CartService
}

func NewCartController(s service.CartService) *CartController {
	return &CartController{cartService: s}
}

// GetCartList 获取购物车列表
func (cc *CartController) GetCartList(c *gin.Context) {
	userID := c.GetInt64("user_id")
	shopID := c.GetInt64("shop_id")

	cartItems, err := cc.cartService.GetCartList(userID, shopID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取购物车失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": cartItems})
}

// AddToCart 添加商品到购物车
func (cc *CartController) AddToCart(c *gin.Context) {
	userID := c.GetInt64("user_id") // 从认证中获取用户ID
	shopID := c.GetInt64("shop_id") // 从认证中获取店铺ID

	var req model.ProductIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	err := cc.cartService.AddToCart(userID, req.ProductID, shopID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "添加购物车失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "添加成功"})
}

// UpdateCartItem 更新购物车商品数量
func (cc *CartController) UpdateCartItem(c *gin.Context) {
	userID := c.GetInt64("user_id")
	shopID := c.GetInt64("shop_id") // 从认证中获取店铺ID
	var req model.UpdateCartItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	err := cc.cartService.UpdateCartItem(userID, shopID, req.CartID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}

// DeleteCartItem 删除购物车商品
func (cc *CartController) DeleteCartItem(c *gin.Context) {
	userID := c.GetInt64("user_id")
	shopID := c.GetInt64("shop_id") // 从认证中获取店铺ID

	idStr := c.Param("id")
	cartId, _ := strconv.ParseInt(idStr, 10, 64)

	err := cc.cartService.DeleteCartItem(userID, shopID, cartId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}
