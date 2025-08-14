package controller

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StockController struct {
	stockService service.StockService
}

func NewStockController(s service.StockService) *StockController {
	return &StockController{stockService: s}
}

// InboundStock 入库操作
func (sc *StockController) InboundStock(c *gin.Context) {
	var req model.StockOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	// 从认证中获取操作人信息
	operator := c.GetString("admin_name") // 假设认证中间件设置了admin_name
	if operator == "" {
		operator = "管理员"
	}

	err := sc.stockService.InboundStock(req.ProductID, req.Quantity, operator, req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "入库失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "入库成功"})
}

// OutboundStock 出库操作
func (sc *StockController) OutboundStock(c *gin.Context) {
	var req model.StockOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	// 从认证中获取操作人信息
	operator := c.GetString("admin_name")
	if operator == "" {
		operator = "管理员"
	}

	err := sc.stockService.OutboundStock(req.ProductID, req.Quantity, operator, "", req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "出库失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "出库成功"})
}

// ReturnStock 退货操作
func (sc *StockController) ReturnStock(c *gin.Context) {
	var req model.StockOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	// 从认证中获取操作人信息
	operator := c.GetString("admin_name")
	if operator == "" {
		operator = "管理员"
	}

	err := sc.stockService.ReturnStock(req.ProductID, req.Quantity, operator, "", req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "退货失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "退货成功"})
}

// GetStockLogs 获取库存日志
func (sc *StockController) GetStockLogs(c *gin.Context) {
	productIDStr := c.Query("product_id")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	var productID int64
	if productIDStr != "" {
		var err error
		productID, err = strconv.ParseInt(productIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "商品ID格式错误"})
			return
		}
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	stockLogs, total, err := sc.stockService.GetStockLogs(productID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取库存日志失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      stockLogs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
