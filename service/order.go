package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/repository"
	"cmf/paint_proj/util"
	"context"
	"errors"
	"fmt"
)

type OrderService interface {
	CreateOrder(ctx context.Context, req *model.CreateOrderRequest) (*model.Order, error)         // 创建订单
	GetOrderList(ctx context.Context, req *model.OrderListRequest) ([]*model.Order, int64, error) // 获取订单列表
	GetOrderDetail(ctx context.Context, userID, orderID int64) (*model.Order, error)              // 获取订单详情

	CancelOrder(ctx context.Context, userID int64, order *model.Order) error // 取消订单
	DeleteOrder(ctx context.Context, userID int64, order *model.Order) error // 删除订单
	PayOrder(ctx context.Context, userID, orderID int64, PaymentType int) (map[string]interface{}, error)
	OrderPaidCallback(ctx context.Context, req *model.OrderPaidCallbackRequest) error // 订单支付成功回调
	// ConfirmReceipt(ctx context.Context, userID, orderID int64) error    // 确认收货  TODO 用户几乎不会点'收货'(可主动触发'收货')。留着后期看
	// ShipOrder(ctx context.Context, req *model.ShipOrderRequest) error  // 发货 TODO 待不做,因为目前都是一来活就发
}

type orderService struct {
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewOrderService(or repository.OrderRepository, cr repository.CartRepository, pr repository.ProductRepository) OrderService {
	return &orderService{
		orderRepo:   or,
		cartRepo:    cr,
		productRepo: pr,
	}
}
func (os *orderService) CreateOrder(ctx context.Context, req *model.CreateOrderRequest) (*model.Order, error) {
	// 1. 参数校验
	if len(req.CartIDs) == 0 && len(req.BuyNowItems) == 0 {
		return nil, errors.New("购物车或立即购买商品不能为空")
	}

	// 2. 获取用户收货地址
	//address, err := os.getUserAddress(ctx, req.UserID, req.AddressID) // TODO 待实现获取收件人信息
	//if err != nil {
	//	return nil, err
	//}

	// 3. 创建订单
	order := &model.Order{
		OrderNo:         util.GenerateOrderNo(req.UserID),
		UserId:          req.UserID,
		OrderStatus:     model.OrderStatusPendingPayment,
		PaymentStatus:   model.PaymentStatusUnpaid,
		ReceiverName:    "柴梦妃",                               // TODO 根据 user_id 获取name  ,还是根据address直接获取name address.Name
		ReceiverPhone:   "13671210659",                          // TODO 根据 user_id 获取phone  ,还是根据address直接获取phone address.Phone
		ReceiverAddress: "河北省廊坊市三河市燕郊镇四季花都一期", // TODO 待实现address.FullAddress(),
	}

	// 4. 获取订单商品
	var orderItems []model.OrderItem
	var totalAmount float64
	var err error
	if len(req.CartIDs) > 0 {
		// 4.1. 从购物车创建订单
		orderItems, totalAmount, err = os.getOrderItemsFromCart(ctx, req.UserID, req.CartIDs)
	} else {
		// 4.2. 立即购买创建订单
		//orderItems, totalAmount, err = os.getOrderItemsFromBuyNow(ctx, req.UserID, req.BuyNowItems)
	}
	if err != nil {
		return nil, err
	}
	//order.OrderItems = orderItems
	// 4.3. 计算优惠金额等
	order.TotalAmount = totalAmount
	order.PaymentAmount = totalAmount // 实付金额(这里简化处理，实际可能有优惠券、运费等)

	// 5.记录订单log
	log := model.OrderLog{
		OrderNo:      order.OrderNo,
		Action:       "create_order",
		Operator:     fmt.Sprintf("user:%d", req.UserID),
		OperatorType: model.OperatorTypeUser,
		Content:      "用户创建订单",
	}
	// 6.真正的业务处理
	err = os.orderRepo.CreateOrder(order, req.CartIDs, orderItems, &log)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// getOrderItemsFromCart 从购物车获取订单商品和总金额
func (os *orderService) getOrderItemsFromCart(ctx context.Context, userID int64, cartIDs []int64) ([]model.OrderItem, float64, error) {
	// 1. 获取购物车项
	cartItems, err := os.cartRepo.GetByIDs(cartIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("获取购物车商品失败: %v", err)
	}
	if len(cartItems) == 0 {
		return nil, 0, errors.New("购物车中没有选中商品")
	}

	// 2. 获取商品ID列表
	productIDs := make([]int64, 0, len(cartItems))
	for _, item := range cartItems {
		productIDs = append(productIDs, item.ProductID)
	}

	// 3. 批量查询商品信息
	products, err := os.productRepo.GetByIDs(productIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("获取商品信息失败: %v", err)
	}

	// 4. 构建商品ID到商品信息的映射
	productMap := make(map[int64]*model.Product)
	for _, product := range products {
		productMap[product.ID] = product
	}

	// 5. 构建订单商品项并计算总金额
	var orderItems []model.OrderItem
	var totalAmount float64

	for _, cartItem := range cartItems {
		product, exists := productMap[cartItem.ProductID]
		if !exists {
			return nil, 0, fmt.Errorf("商品ID %d 不存在", cartItem.ProductID)
		}

		// 检查商品库存
		if product.Stock < cartItem.Quantity {
			return nil, 0, fmt.Errorf("商品 %s 库存不足", product.Name)
		}
		// 计算商品总价
		itemTotalPrice := product.SellerPrice * float64(cartItem.Quantity)

		// 构建订单商品项
		orderItem := model.OrderItem{
			ProductId:    product.ID,
			ProductName:  product.Name,
			ProductImage: product.Image,
			ProductPrice: product.SellerPrice,
			Quantity:     cartItem.Quantity,
			Unit:         product.Unit,
			TotalPrice:   itemTotalPrice,
		}
		orderItems = append(orderItems, orderItem)
		totalAmount += itemTotalPrice
	}

	return orderItems, totalAmount, nil
}
func (os *orderService) GetOrderList(ctx context.Context, req *model.OrderListRequest) ([]*model.Order, int64, error) {
	orders, total, err := os.orderRepo.GetOrderList(req)
	if err != nil {
		return nil, 0, err
	}
	// 查询每个订单的商品
	for _, order := range orders {
		items, err := os.orderRepo.GetOrderItemList(order.ID)
		if err != nil {
			return nil, 0, err
		}
		order.OrderItems = items
	}
	return orders, total, nil
}
func (os *orderService) GetOrderDetail(ctx context.Context, userID, orderID int64) (*model.Order, error) {
	order, err := os.orderRepo.GetOrderByIDAndUserID(userID, orderID)
	if err != nil {
		return nil, err
	}
	return order, nil
}
func (os *orderService) CancelOrder(ctx context.Context, userID int64, order *model.Order) error {
	log := &model.OrderLog{
		OrderId:      order.ID,
		OrderNo:      order.OrderNo,
		Action:       "cancel_order",
		Operator:     fmt.Sprintf("user:%d", userID),
		OperatorType: model.OperatorTypeUser,
		Content:      "用户取消订单",
	}
	err := os.orderRepo.CancelOrder(userID, order, log)
	return err
}

func (os *orderService) DeleteOrder(ctx context.Context, userID int64, order *model.Order) error {
	log := &model.OrderLog{
		OrderId:      order.ID,
		OrderNo:      order.OrderNo,
		Action:       "delete_order",
		Operator:     fmt.Sprintf("user:%d", userID),
		OperatorType: model.OperatorTypeUser,
		Content:      "用户删除订单",
	}
	err := os.orderRepo.DeleteOrder(userID, order, log)
	return err
}
func (os *orderService) PayOrder(ctx context.Context, userID, orderID int64, paymentType int) (map[string]interface{}, error) {
	// 1. 获取订单
	order, err := os.orderRepo.GetOrderByIDAndUserID(userID, orderID)
	if err != nil {
		return nil, err
	}
	// 2. 检查订单状态是否可以支付
	if order.OrderStatus != model.OrderStatusPendingPayment {
		return nil, errors.New("订单状态异常 无法支付")
	}

	// 3. 调用支付服务生成支付参数
	paymentParams, err := os.generatePaymentParams(order, paymentType)
	if err != nil {
		return nil, errors.New("generatePaymentParams Error:" + err.Error())
	}
	return paymentParams, nil
}

func (os *orderService) generatePaymentParams(order *model.Order, paymentType int) (map[string]interface{}, error) {
	// 这里实现具体的支付参数生成逻辑
	// 根据不同的支付方式(微信、支付宝等)生成不同的支付参数
	// ...
	return map[string]interface{}{
		"order_id":       order.ID,
		"order_no":       order.OrderNo,
		"payment_type":   paymentType,
		"payment_amount": order.PaymentAmount,
		// 其他支付参数...
	}, nil
}
func (os *orderService) OrderPaidCallback(ctx context.Context, req *model.OrderPaidCallbackRequest) error {
	panic("implement me")
}
