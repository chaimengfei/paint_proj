package repository

import (
	"cmf/paint_proj/model"
	"time"

	"gorm.io/gorm"
)

type StockRepository interface {
	// 库存操作
	UpdateProductStock(productID int64, quantity int) error
	GetProductStock(productID int64) (int, error)

	// 库存操作主表+子表
	CreateStockOperation(operation *model.StockOperation) error
	CreateStockOperationItems(items []model.StockOperationItem) error
	GetStockOperations(page, pageSize int, types *int8) ([]model.StockOperation, int64, error)
	GetStockOperationsByShop(page, pageSize int, types *int8, shopID int64) ([]model.StockOperation, int64, error)
	GetStockOperationByID(operationID int64) (*model.StockOperation, error)
	GetStockOperationItems(operationID int64) ([]model.StockOperationItem, error)
	GetStockOperationItemsByOrderID(orderID int64) ([]model.StockOperationItem, error)
	GetStockOperationItemsByShop(page, pageSize int, shopID int64, productID *int64) ([]model.StockOperationItem, int64, error)

	// 更新出库单支付完成状态
	UpdateOutboundPaymentStatus(operationID int64, paymentFinishStatus model.PaymentStatusCode, paymentFinishTime *time.Time) error

	// 供货商管理
	GetSupplierList() ([]*model.Supplier, error)

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
	// 直接Create，测试GORM是否会使用我们设置的CreatedAt值
	return sr.db.Create(operation).Error
}

// CreateStockOperationItems 创建库存操作子表记录
func (sr *stockRepository) CreateStockOperationItems(items []model.StockOperationItem) error {
	return sr.db.Create(&items).Error
}

// GetStockOperations 获取库存操作主表列表
func (sr *stockRepository) GetStockOperations(page, pageSize int, types *int8) ([]model.StockOperation, int64, error) {
	var operations []model.StockOperation
	var total int64

	// 构建查询条件
	query := sr.db.Model(&model.StockOperation{})
	if types != nil {
		query = query.Where("types = ?", *types)
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
		Find(&operations).Error; err != nil {
		return nil, 0, err
	}

	return operations, total, nil
}

// GetStockOperationsByShop 根据店铺获取库存操作主表列表
func (sr *stockRepository) GetStockOperationsByShop(page, pageSize int, types *int8, shopID int64) ([]model.StockOperation, int64, error) {
	var operations []model.StockOperation
	var total int64

	// 构建查询条件
	query := sr.db.Model(&model.StockOperation{}).Where("shop_id = ?", shopID)
	if types != nil {
		query = query.Where("types = ?", *types)
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

// GetStockOperationItemsByShop 根据店铺获取库存操作明细列表
func (sr *stockRepository) GetStockOperationItemsByShop(page, pageSize int, shopID int64, productID *int64) ([]model.StockOperationItem, int64, error) {
	var items []model.StockOperationItem
	var total int64

	// 构建查询条件
	query := sr.db.Model(&model.StockOperationItem{}).Where("shop_id = ?", shopID)
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
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
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
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
			operation.Items[i].CreatedAt = operation.CreatedAt
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
			// 3.1 更新库存
			if err := tx.Model(&model.Product{}).
				Where("id = ?", item.ProductID).
				Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}

			// 检查是否需要更新成本价
			var product model.Product
			if err := tx.Model(&model.Product{}).
				Select("cost, name, shipping_cost, product_cost").
				Where("id = ?", item.ProductID).
				First(&product).Error; err != nil {
				return err
			}

			// 如果新进价有变化，则更新进价和成本价
			if item.ProductCost != product.ProductCost {
				// 计算新的成本价 = 进价 + 运费成本
				newCost := item.ProductCost + product.ShippingCost

				// 更新进价和成本价
				if err := tx.Model(&model.Product{}).Where("id = ?", item.ProductID).
					Updates(map[string]interface{}{
						"product_cost": item.ProductCost,
						"cost":         newCost,
					}).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// UpdateOutboundPaymentStatus 更新出库单支付完成状态
func (sr *stockRepository) UpdateOutboundPaymentStatus(operationID int64, paymentFinishStatus model.PaymentStatusCode, paymentFinishTime *time.Time) error {
	updates := map[string]interface{}{
		"payment_finish_status": paymentFinishStatus,
	}

	// 如果支付完成状态为已支付，则设置支付完成时间
	if paymentFinishStatus == model.PaymentStatusPaid && paymentFinishTime != nil {
		updates["payment_finish_time"] = paymentFinishTime
	}

	return sr.db.Model(&model.StockOperation{}).
		Where("id = ?", operationID).
		Updates(updates).Error
}

// GetSupplierList 获取供货商列表
func (sr *stockRepository) GetSupplierList() ([]*model.Supplier, error) {
	var suppliers []*model.Supplier
	err := sr.db.Model(&model.Supplier{}).Find(&suppliers).Error
	return suppliers, err
}
