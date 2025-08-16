package repository

import (
	"cmf/paint_proj/model"

	"gorm.io/gorm"
)

type StockRepository interface {
	// 库存操作
	UpdateProductStock(productID int64, quantity int) error
	GetProductStock(productID int64) (int, error)

	// 库存操作主表+子表
	CreateStockOperation(operation *model.StockOperation) error
	CreateStockOperationItems(items []*model.StockOperationItem) error
	GetStockOperations(page, pageSize int) ([]model.StockOperation, int64, error)
	GetStockOperationByID(operationID int64) (*model.StockOperation, error)
	GetStockOperationItems(operationID int64) ([]model.StockOperationItem, error)
}

type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) StockRepository {
	return &stockRepository{db: db}
}

// UpdateProductStock 更新商品库存
func (sr *stockRepository) UpdateProductStock(productID int64, quantity int) error {
	return sr.db.Model(&model.Product{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock + ?", quantity)).
		Error
}

// GetProductStock 获取商品库存
func (sr *stockRepository) GetProductStock(productID int64) (int, error) {
	var product model.Product
	err := sr.db.Model(&model.Product{}).
		Select("stock").
		Where("id = ?", productID).
		First(&product).Error
	return product.Stock, err
}

// CreateStockOperation 创建库存操作主表记录
func (sr *stockRepository) CreateStockOperation(operation *model.StockOperation) error {
	return sr.db.Create(operation).Error
}

// CreateStockOperationItems 创建库存操作子表记录
func (sr *stockRepository) CreateStockOperationItems(items []*model.StockOperationItem) error {
	return sr.db.Create(items).Error
}

// GetStockOperations 获取库存操作主表列表
func (sr *stockRepository) GetStockOperations(page, pageSize int) ([]model.StockOperation, int64, error) {
	var operations []model.StockOperation
	var total int64

	// 获取总数
	if err := sr.db.Model(&model.StockOperation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := sr.db.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&operations).Error; err != nil {
		return nil, 0, err
	}

	return operations, total, nil
}

// GetStockOperationByID 根据ID获取库存操作主表记录
func (sr *stockRepository) GetStockOperationByID(operationID int64) (*model.StockOperation, error) {
	var operation model.StockOperation
	err := sr.db.Model(&model.StockOperation{}).
		Where("id = ?", operationID).
		First(&operation).Error
	return &operation, err
}

// GetStockOperationItems 获取库存操作子表记录
func (sr *stockRepository) GetStockOperationItems(operationID int64) ([]model.StockOperationItem, error) {
	var items []model.StockOperationItem
	err := sr.db.Model(&model.StockOperationItem{}).
		Where("operation_id = ?", operationID).
		Find(&items).Error
	return items, err
}
