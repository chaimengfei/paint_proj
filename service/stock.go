package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/repository"
	"errors"
	"time"
)

type StockService interface {
	// 库存操作
	InboundStock(productID int64, quantity int, operator string, remark string) error
	OutboundStock(productID int64, quantity int, operator string, orderNo string, remark string) error
	ReturnStock(productID int64, quantity int, operator string, orderNo string, remark string) error

	// 库存日志
	GetStockLogs(productID int64, page, pageSize int) ([]model.StockLog, int64, error)
}

type stockService struct {
	stockRepo   repository.StockRepository
	productRepo repository.ProductRepository
}

func NewStockService(sr repository.StockRepository, pr repository.ProductRepository) StockService {
	return &stockService{
		stockRepo:   sr,
		productRepo: pr,
	}
}

// InboundStock 入库操作
func (ss *stockService) InboundStock(productID int64, quantity int, operator string, remark string) error {
	if quantity <= 0 {
		return errors.New("入库数量必须大于0")
	}

	// 获取商品信息
	product, err := ss.productRepo.GetByID(productID)
	if err != nil {
		return err
	}

	// 获取操作前库存
	beforeStock, err := ss.stockRepo.GetProductStock(productID)
	if err != nil {
		return err
	}

	// 更新库存
	err = ss.stockRepo.UpdateProductStock(productID, quantity)
	if err != nil {
		return err
	}

	// 获取操作后库存
	afterStock := beforeStock + quantity

	// 创建库存日志
	now := time.Now()
	stockLog := &model.StockLog{
		ProductID:    productID,
		ProductName:  product.Name,
		Types:        model.StockTypeInbound,
		Quantity:     quantity,
		BeforeStock:  beforeStock,
		AfterStock:   afterStock,
		Remark:       remark,
		Operator:     operator,
		OperatorType: model.OperatorTypeAdmin,
		CreatedAt:    &now,
	}

	return ss.stockRepo.CreateStockLog(stockLog)
}

// OutboundStock 出库操作
func (ss *stockService) OutboundStock(productID int64, quantity int, operator string, orderNo string, remark string) error {
	if quantity <= 0 {
		return errors.New("出库数量必须大于0")
	}

	// 获取商品信息
	product, err := ss.productRepo.GetByID(productID)
	if err != nil {
		return err
	}

	// 获取操作前库存
	beforeStock, err := ss.stockRepo.GetProductStock(productID)
	if err != nil {
		return err
	}

	// 检查库存是否足够
	if beforeStock < quantity {
		return errors.New("库存不足")
	}

	// 更新库存（出库为负数）
	err = ss.stockRepo.UpdateProductStock(productID, -quantity)
	if err != nil {
		return err
	}

	// 获取操作后库存
	afterStock := beforeStock - quantity

	// 创建库存日志
	now := time.Now()
	stockLog := &model.StockLog{
		ProductID:    productID,
		ProductName:  product.Name,
		Types:        model.StockTypeOutbound,
		Quantity:     quantity,
		BeforeStock:  beforeStock,
		AfterStock:   afterStock,
		OrderNo:      orderNo,
		Remark:       remark,
		Operator:     operator,
		OperatorType: model.OperatorTypeAdmin,
		CreatedAt:    &now,
	}

	return ss.stockRepo.CreateStockLog(stockLog)
}

// ReturnStock 退货操作
func (ss *stockService) ReturnStock(productID int64, quantity int, operator string, orderNo string, remark string) error {
	if quantity <= 0 {
		return errors.New("退货数量必须大于0")
	}

	// 获取商品信息
	product, err := ss.productRepo.GetByID(productID)
	if err != nil {
		return err
	}

	// 获取操作前库存
	beforeStock, err := ss.stockRepo.GetProductStock(productID)
	if err != nil {
		return err
	}

	// 更新库存（退货为正数，增加库存）
	err = ss.stockRepo.UpdateProductStock(productID, quantity)
	if err != nil {
		return err
	}

	// 获取操作后库存
	afterStock := beforeStock + quantity

	// 创建库存日志
	now := time.Now()
	stockLog := &model.StockLog{
		ProductID:    productID,
		ProductName:  product.Name,
		Types:        model.StockTypeReturn,
		Quantity:     quantity,
		BeforeStock:  beforeStock,
		AfterStock:   afterStock,
		OrderNo:      orderNo,
		Remark:       remark,
		Operator:     operator,
		OperatorType: model.OperatorTypeAdmin,
		CreatedAt:    &now,
	}

	return ss.stockRepo.CreateStockLog(stockLog)
}

// GetStockLogs 获取库存日志
func (ss *stockService) GetStockLogs(productID int64, page, pageSize int) ([]model.StockLog, int64, error) {
	return ss.stockRepo.GetStockLogs(productID, page, pageSize)
}
