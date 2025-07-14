package controller

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type OrderController struct {
	orderService service.OrderService
}

func NewOrderController(s service.OrderService) *OrderController {
	return &OrderController{orderService: s}
}

// CheckoutOrder 订单结算
func (oc *OrderController) CheckoutOrder(c *gin.Context) {
	userID := c.GetInt64("user_id") // 从认证中获取用户ID
	var req struct {
		CartIDs   []int64 `json:"cart_ids"`
		ProductID int64   `json:"product_id"`
		Quantity  int     `json:"quantity"`
		AddressID int64   `json:"address_id"`
		CouponID  int64   `json:"coupon_id"`
		Note      string  `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}
	// 构建服务请求
	svcReq := &model.CheckoutOrderRequest{
		UserID:    userID,
		AddressID: req.AddressID,
		CouponID:  req.CouponID,
		Note:      req.Note,
	}

	// 判断是购物车下单还是立即购买
	if len(req.CartIDs) > 0 {
		svcReq.CartIDs = req.CartIDs
	} else if req.ProductID > 0 && req.Quantity > 0 {
		svcReq.BuyNowItems = []*model.BuyNowItem{
			{
				ProductID: req.ProductID,
				Quantity:  req.Quantity,
			},
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "请选择商品或购物车"})
		return
	}
	// 真实的业务处理
	checkoutData, err := oc.orderService.CheckoutOrder(c.Request.Context(), userID, svcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "购物车结算失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": checkoutData,
	})
}

// GetOrderList 获取订单列表
func (oc *OrderController) GetOrderList(c *gin.Context) {
	userID := c.GetInt64("user_id")
	statusStr := c.Query("status")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	status, _ := strconv.Atoi(statusStr)
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	req := &model.OrderListRequest{
		UserID:   userID,
		Status:   int32(status),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	orders, total, err := oc.orderService.GetOrderList(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取订单列表失败:" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"list":  orders,
		"total": total,
	}})
}

// GetOrderDetail 获取订单详情
func (oc *OrderController) GetOrderDetail(c *gin.Context) {
	userID := c.GetInt64("user_id")
	orderIDStr := c.Query("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "订单ID格式错误"})
		return
	}
	order, err := oc.orderService.GetOrderDetail(c.Request.Context(), userID, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取订单详情失败:" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": order})
}

// CancelOrder 取消订单
func (oc *OrderController) CancelOrder(c *gin.Context) {
	userID := c.GetInt64("user_id")
	orderIDStr := c.Query("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "订单ID格式错误"})
		return
	}
	order, err := oc.orderService.GetOrderDetail(c.Request.Context(), userID, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "查询订单异常:" + err.Error()})
		return
	}
	if err = oc.orderService.CancelOrder(c.Request.Context(), userID, order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "取消订单失败:" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "取消订单成功"})
}

// DeleteOrder 删除订单
func (oc *OrderController) DeleteOrder(c *gin.Context) {
	userID := c.GetInt64("user_id")
	orderIDStr := c.Query("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "订单ID格式错误"})
		return
	}
	order, err := oc.orderService.GetOrderDetail(c.Request.Context(), userID, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "查询订单异常:" + err.Error()})
		return
	}
	if err := oc.orderService.DeleteOrder(c.Request.Context(), userID, order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "删除订单失败:" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除订单成功"})
}
