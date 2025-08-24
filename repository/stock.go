package repository

import (
	"cmf/paint_proj/model"

	"gorm.io/gorm"
)

type StockRepository interface {
	// 库存操作
	UpdateProductStock(productID int64, quantity int) error
	GetProductStock(productID int64) (int, error)
	UpdateProductCost(productID int64, newCost model.Amount) error
	GetProductCost(productID int64) (model.Amount, error)

	// 库存操作主表+子表
	CreateStockOperation(operation *model.StockOperation) error
	CreateStockOperationItems(items []model.StockOperationItem) error
	GetStockOperations(page, pageSize int) ([]model.StockOperation, int64, error)
	GetStockOperationByID(operationID int64) (*model.StockOperation, error)
	GetStockOperationItems(operationID int64) ([]model.StockOperationItem, error)
	GetStockOperationItemsByOrderID(orderID int64) ([]model.StockOperationItem, error)

	// 入库成本变更记录
	CreateInboundCostChange(change *model.InboundCostChange) error

	// 事务处理
	ProcessOutboundTransaction(operation *model.StockOperation) error
	ProcessInboundTransaction(operation *model.StockOperation) error
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
func (sr *stockRepository) CreateStockOperationItems(items []model.StockOperationItem) error {
	return sr.db.Create(&items).Error
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

// GetStockOperationItemsByOrderID 根据订单ID获取库存操作子表记录
func (sr *stockRepository) GetStockOperationItemsByOrderID(orderID int64) ([]model.StockOperationItem, error) {
	var items []model.StockOperationItem
	err := sr.db.Model(&model.StockOperationItem{}).
		Where("order_id = ?", orderID).
		Find(&items).Error
	return items, err
}

// ProcessOutboundTransaction 处理出库事务：创建主表记录、子表记录、更新库存
func (sr *stockRepository) ProcessOutboundTransaction(operation *model.StockOperation) error {
	return sr.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建主表记录
		if err := tx.Create(operation).Error; err != nil {
			return err
		}

		// 2. 创建子表记录
		for i := range operation.Items {
			operation.Items[i].OperationID = operation.ID
		}
		if err := tx.Create(&operation.Items).Error; err != nil {
			return err
		}

		// 3. 更新库存
		for _, item := range operation.Items {
			if err := tx.Model(&model.Product{}).
				Where("id = ?", item.ProductID).
				Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// ProcessInboundTransaction 处理入库事务：创建主表记录、子表记录、更新库存和成本价
func (sr *stockRepository) ProcessInboundTransaction(operation *model.StockOperation) error {
	return sr.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建主表记录
		if err := tx.Create(operation).Error; err != nil {
			return err
		}

		// 2. 创建子表记录
		for i := range operation.Items {
			operation.Items[i].OperationID = operation.ID
		}
		if err := tx.Create(&operation.Items).Error; err != nil {
			return err
		}

		// 3. 更新库存和成本价
		for _, item := range operation.Items {
			// 更新库存
			if err := tx.Model(&model.Product{}).
				Where("id = ?", item.ProductID).
				Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}

			// 检查是否需要更新成本价
			var product model.Product
			if err := tx.Model(&model.Product{}).
				Select("cost, name").
				Where("id = ?", item.ProductID).
				First(&product).Error; err != nil {
				return err
			}

			// 如果新成本价更低，则更新成本价并记录变更
			if item.Cost < product.Cost {
				// 更新成本价
				if err := tx.Model(&model.Product{}).
					Where("id = ?", item.ProductID).
					Update("cost", item.Cost).Error; err != nil {
					return err
				}

				// 创建成本变更记录
				costChange := &model.InboundCostChange{
					OperationID:  operation.ID,
					ProductID:    item.ProductID,
					ProductName:  product.Name,
					OldCost:      product.Cost,
					NewCost:      item.Cost,
					ChangeReason: "入库成本价降低",
					Operator:     operation.Operator,
					OperatorID:   operation.OperatorID,
				}
				if err := tx.Create(costChange).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// UpdateProductCost 更新商品成本价
func (sr *stockRepository) UpdateProductCost(productID int64, newCost model.Amount) error {
	return sr.db.Model(&model.Product{}).
		Where("id = ?", productID).
		Update("cost", newCost).Error
}

// GetProductCost 获取商品成本价
func (sr *stockRepository) GetProductCost(productID int64) (model.Amount, error) {
	var product model.Product
	err := sr.db.Model(&model.Product{}).
		Select("cost").
		Where("id = ?", productID).
		First(&product).Error
	return product.Cost, err
}

// CreateInboundCostChange 创建入库成本变更记录
func (sr *stockRepository) CreateInboundCostChange(change *model.InboundCostChange) error {
	return sr.db.Create(change).Error
}
