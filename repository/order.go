package repository

import (
	"cmf/paint_proj/model"
	"time"

	"gorm.io/gorm"
)

type OrderRepository interface {
	GetOrderList(req *model.OrderListRequest) ([]model.Order, int64, error)
	GetOrderByNo(userID int64, orderNo string) (*model.Order, error)
	GetOrderByOrderNo(orderNo string) (*model.Order, error)

	DeleteOrder(orderID int64, order *model.Order, orderLog *model.OrderLog) error
	CancelOrder(userID int64, order *model.Order, orderLog *model.OrderLog) error
	UpdateOrder(orderID int64, order *model.Order) error

	// ProcessCheckoutTransaction 处理结算事务：创建订单、记录日志、处理库存、删除购物车
	ProcessCheckoutTransaction(order *model.Order, operation *model.StockOperation, operationItems []model.StockOperationItem, cartIDs []int64, log *model.OrderLog) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}
func (or *orderRepository) GetOrderList(req *model.OrderListRequest) ([]model.Order, int64, error) {
	orders := make([]model.Order, 0)
	count := int64(0)

	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	queryDb := or.db.Model(&model.Order{}).Where("user_id = ? and deleted_at is NULL", req.UserID)
	if req.Status > 0 {
		queryDb = queryDb.Where("order_status = ?", req.Status)
	}
	err := queryDb.Limit(pageSize).Offset(offset).Order("id asc").Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}
	err = queryDb.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	return orders, count, nil
}

// GetOrderItemList 已废弃，订单商品信息现在通过stock_operation_item表获取
// 此方法已移除，请使用stock repository的GetStockOperationItemsByOrderID方法
func (or *orderRepository) GetOrderByNo(userID int64, orderNo string) (*model.Order, error) {
	var order model.Order
	err := or.db.Model(&model.Order{}).Where("user_id = ? and order_no= ?", userID, orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}
func (or *orderRepository) GetOrderByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := or.db.Model(&model.Order{}).Where("order_no = ? ", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}
func (or *orderRepository) DeleteOrder(userID int64, order *model.Order, orderLog *model.OrderLog) error {
	err := or.db.Transaction(func(tx *gorm.DB) error {
		// 1.更新订单状态
		now := time.Now()
		err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(&model.Order{DeletedAt: &now}).Error
		if err != nil {
			return err
		}
		// 2. 记录订单日志
		err = tx.Model(&model.OrderLog{}).Create(orderLog).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
func (or *orderRepository) CancelOrder(userID int64, order *model.Order, orderLog *model.OrderLog) error {
	err := or.db.Transaction(func(tx *gorm.DB) error {
		// 1.更新订单状态
		err := tx.Model(&model.Order{}).Where("id = ?", order.ID).Updates(&model.Order{OrderStatus: model.OrderStatusCancelled}).Error
		if err != nil {
			return err
		}
		// 2. 记录订单日志
		err = tx.Model(&model.OrderLog{}).Create(orderLog).Error
		if err != nil {
			return err
		}
		// 3. 如果已支付，需要退款
		if order.PaymentStatus == model.PaymentStatusPaid {
			// TODO 这里调用退款逻辑 ......
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
func (or *orderRepository) UpdateOrder(orderID int64, order *model.Order) error {
	// 1.更新订单状态
	err := or.db.Model(&model.Order{}).Where("id = ?", orderID).Updates(order).Error
	if err != nil {
		return err
	}
	return nil
}

// ProcessCheckoutTransaction 处理结算事务：创建订单、记录日志、处理库存、删除购物车
func (or *orderRepository) ProcessCheckoutTransaction(order *model.Order, operation *model.StockOperation, operationItems []model.StockOperationItem, cartIDs []int64, log *model.OrderLog) error {
	return or.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建订单
		if err := tx.Model(&model.Order{}).Create(order).Error; err != nil {
			return err
		}

		// 2. 记录订单日志
		log.OrderId = order.ID
		if err := tx.Model(&model.OrderLog{}).Create(log).Error; err != nil {
			return err
		}

		// 3. 创建库存操作主表记录
		if err := tx.Create(operation).Error; err != nil {
			return err
		}

		// 4. 处理库存出库和创建子表记录
		for _, item := range operationItems {
			// 更新库存（出库为负数）
			if err := tx.Model(&model.Product{}).
				Where("id = ?", item.ProductID).
				Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
				return err
			}

			// 设置关联ID并创建子表记录
			item.OperationID = operation.ID
			item.OrderID = order.ID
			item.OrderNo = order.OrderNo

			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		// 5. 如果是购物车下单，删除购物车
		if len(cartIDs) > 0 {
			if err := tx.Model(&model.Cart{}).Delete("id in ?", cartIDs).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
