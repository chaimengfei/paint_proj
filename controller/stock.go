package controller

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/service"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StockController struct {
	stockService   service.StockService
	productService service.ProductService
}

func NewStockController(s service.StockService, ps service.ProductService) *StockController {
	return &StockController{stockService: s, productService: ps}
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

	// 验证请求
	if err := sc.validateBatchOutboundRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "验证失败: " + err.Error()})
		return
	}
	// 业务处理
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

// validateBatchOutboundRequest 验证批量出库请求
func (sc *StockController) validateBatchOutboundRequest(req *model.BatchOutboundRequest) error {
	if len(req.Items) == 0 {
		return errors.New("出库商品列表不能为空")
	}
	// 验证所有商品是否存在且库存充足
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("商品ID %d 的出库数量必须大于0", item.ProductID)
		}

		product, err := sc.productService.GetProductByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("商品ID %d 不存在", item.ProductID)
		}
		// 检查库存是否足够
		if product.Stock < item.Quantity {
			return fmt.Errorf("商品 %s 库存不足，当前库存: %d，需要出库: %d", product.Name, product.Stock, item.Quantity)
		}
	}

	// 验证前端计算的总金额是否正确
	var calculatedTotalAmount model.Amount
	for _, item := range req.Items {
		if item.UnitPrice > 0 {
			calculatedTotalAmount += model.Amount(int64(item.UnitPrice) * int64(item.Quantity))
		}
	}
	if req.TotalAmount > 0 && req.TotalAmount != calculatedTotalAmount {
		return fmt.Errorf("总金额计算错误，前端计算: %d，后端计算: %d", req.TotalAmount, calculatedTotalAmount)
	}

	return nil
}
