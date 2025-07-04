package controller

import (
	"cmf/paint_proj/service"
	"github.com/gin-gonic/gin"
	"net/http"
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

	cartItems, err := cc.cartService.GetCartList(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取购物车失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": cartItems})
}

// AddToCart 添加商品到购物车
func (cc *CartController) AddToCart(c *gin.Context) {
	userID := c.GetInt64("user_id") // 从认证中获取用户ID

	var req struct {
		ProductID int64 `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	err := cc.cartService.AddToCart(userID, req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "添加购物车失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "添加成功"})
}

// UpdateCartItem 更新购物车商品数量
func (cc *CartController) UpdateCartItem(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		CartID   int64 `json:"cart_id" binding:"required"`
		Quantity int   `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	err := cc.cartService.UpdateCartItem(userID, req.CartID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
}

// DeleteCartItem 删除购物车商品
func (cc *CartController) DeleteCartItem(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		CartID int64 `json:"cart_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	err := cc.cartService.DeleteCartItem(userID, req.CartID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}
