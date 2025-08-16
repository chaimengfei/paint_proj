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

// BatchInboundStock 批量入库操作
func (sc *StockController) BatchInboundStock(c *gin.Context) {
	var req model.BatchInboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	err := sc.stockService.BatchInboundStock(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "批量入库失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "批量入库成功"})
}

// BatchOutboundStock 批量出库操作
func (sc *StockController) BatchOutboundStock(c *gin.Context) {
	var req model.BatchOutboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	err := sc.stockService.BatchOutboundStock(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "批量出库失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "批量出库成功"})
}

// GetStockOperations 获取库存操作列表
func (sc *StockController) GetStockOperations(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	operations, total, err := sc.stockService.GetStockOperations(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取库存操作列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      operations,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetStockOperationDetail 获取库存操作详情
func (sc *StockController) GetStockOperationDetail(c *gin.Context) {
	operationIDStr := c.Param("id")
	operationID, err := strconv.ParseInt(operationIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "操作ID格式错误"})
		return
	}

	operation, items, err := sc.stockService.GetStockOperationDetail(operationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取库存操作详情失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"operation": operation,
			"items":     items,
		},
	})
}
