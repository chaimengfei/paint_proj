package repository

import (
	"cmf/paint_proj/model"

	"gorm.io/gorm"
)

type StockRepository interface {
	// 库存操作
	UpdateProductStock(productID int64, quantity int) error
	GetProductStock(productID int64) (int, error)

	// 库存日志
	CreateStockLog(stockLog *model.StockLog) error
	GetStockLogs(productID int64, page, pageSize int) ([]model.StockLog, int64, error)
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

// CreateStockLog 创建库存日志
func (sr *stockRepository) CreateStockLog(stockLog *model.StockLog) error {
	return sr.db.Create(stockLog).Error
}

// GetStockLogs 获取库存日志列表
func (sr *stockRepository) GetStockLogs(productID int64, page, pageSize int) ([]model.StockLog, int64, error) {
	var stockLogs []model.StockLog
	var total int64

	query := sr.db.Model(&model.StockLog{})
	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&stockLogs).Error; err != nil {
		return nil, 0, err
	}

	return stockLogs, total, nil
}
