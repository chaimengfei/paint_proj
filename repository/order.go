package repository

import (
	"cmf/paint_proj/model"
	"gorm.io/gorm"
	"time"
)

type OrderRepository interface {
	CreateOrder(order *model.Order, cartIDs []int64, orderItems []model.OrderItem, orderLog *model.OrderLog) error
	GetOrderList(req *model.OrderListRequest) ([]*model.Order, int64, error)
	GetOrderItemList(orderID int64) ([]model.OrderItem, error)
	GetOrderByIDAndUserID(userID, orderID int64) (*model.Order, error)

	DeleteOrder(orderID int64, order *model.Order, orderLog *model.OrderLog) error
	CancelOrder(userID int64, order *model.Order, orderLog *model.OrderLog) error
	UpdateOrder(orderID int64, order *model.Order) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (or *orderRepository) CreateOrder(order *model.Order, cartIDs []int64, orderItems []model.OrderItem, orderLog *model.OrderLog) error {
	err := or.db.Transaction(func(tx *gorm.DB) error {
		// 1.创建订单
		err := tx.Model(&model.Order{}).Create(order).Error
		if err != nil {
			return err
		}
		// 2.创建订单商品
		for i := range orderItems {
			orderItems[i].OrderNo = order.OrderNo
			orderItems[i].OrderId = order.ID
		}
		err = tx.Model(&model.OrderItem{}).Create(orderItems).Error
		if err != nil {
			return err
		}
		// 3.记录订单日志
		orderLog.OrderId = order.ID
		err = tx.Model(&model.OrderLog{}).Create(&orderLog).Error
		if err != nil {
			return err
		}
		// 4. 如果是购物车下单，删除购物车
		if len(cartIDs) > 0 {
			err = tx.Model(&model.Cart{}).Delete("id in ?", cartIDs).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
func (or *orderRepository) GetOrderList(req *model.OrderListRequest) ([]*model.Order, int64, error) {
	orders := make([]*model.Order, 0)
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
func (or *orderRepository) GetOrderItemList(orderID int64) ([]model.OrderItem, error) {
	var orderItems []model.OrderItem
	err := or.db.Model(&model.OrderItem{}).Where("order_id = ?", orderID).Find(&orderItems).Error
	if err != nil {
		return nil, err
	}
	return orderItems, nil
}
func (or *orderRepository) GetOrderByIDAndUserID(userID, orderID int64) (*model.Order, error) {
	var order model.Order
	err := or.db.Model(&model.Order{}).Where("user_id = ? and id= ?", userID, orderID).First(&order).Error
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
