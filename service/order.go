package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/pkg"
	"cmf/paint_proj/repository"
	"context"
	"errors"
	"fmt"
)

type OrderService interface {
	CheckoutOrder(ctx context.Context, userID int64, req *model.CheckoutOrderRequest) (*model.CheckoutResponse, error)
	GetOrderList(ctx context.Context, req *model.OrderListRequest) ([]model.Order, int64, error) // 获取订单列表
	GetOrderDetail(ctx context.Context, userID int64, orderNo string) (*model.Order, error)      // 获取订单详情

	CancelOrder(ctx context.Context, userID int64, order *model.Order) error // 取消订单
	DeleteOrder(ctx context.Context, userID int64, order *model.Order) error // 删除订单
}

type orderService struct {
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
	addressRepo repository.AddressRepository
	stockRepo   repository.StockRepository
}

func NewOrderService(or repository.OrderRepository, cr repository.CartRepository, pr repository.ProductRepository, ar repository.AddressRepository, sr repository.StockRepository) OrderService {
	return &orderService{
		orderRepo:   or,
		cartRepo:    cr,
		productRepo: pr,
		addressRepo: ar,
		stockRepo:   sr,
	}
}

func (os *orderService) CheckoutOrder(ctx context.Context, userID int64, req *model.CheckoutOrderRequest) (*model.CheckoutResponse, error) {
	// 1. 数据校验和准备阶段
	// 1.1 获取用户收货地址
	var addressDbData *model.Address
	var err error
	if req.AddressID == 0 {
		addressDbData, err = os.addressRepo.GetDefaultOrFirstAddressID(userID)
	} else {
		addressDbData, err = os.addressRepo.GetByUserAppointId(req.UserID, req.AddressID)
	}
	var addressInfo *model.AddressInfo
	if addressDbData != nil && addressDbData.ID > 0 {
		isDefault := addressDbData.IsDefault == 1
		addressInfo = &model.AddressInfo{
			AddressID:      addressDbData.ID,
			RecipientName:  addressDbData.RecipientName,
			RecipientPhone: addressDbData.RecipientPhone,
			Province:       addressDbData.Province,
			City:           addressDbData.City,
			District:       addressDbData.District,
			Detail:         addressDbData.Detail,
			IsDefault:      &isDefault,
		}
	}

	// 1.2 获取订单商品并校验库存
	var items []model.StockOperationItem
	var totalAmount model.Amount
	if len(req.CartIDs) > 0 {
		// 从购物车创建订单
		items, totalAmount, err = os.getOrderItemsFromCart(ctx, req.UserID, req.CartIDs)
	} else {
		// 立即购买创建订单
		items, totalAmount, err = os.getOrderItemsFromBuyNow(ctx, req.UserID, req.BuyNowItems)
	}
	if err != nil {
		return nil, err
	}

	// 1.3 计算订单金额
	shippingFee := model.Amount(0)
	if totalAmount < 100 { // 假设满100免运费
		shippingFee = 1000 //单位:分(10块的运费)
	}
	paymentAmount := totalAmount + shippingFee

	// 1.4 准备订单数据
	orderNo := pkg.GenerateOrderNo(pkg.OrderPrefix, req.UserID)
	order := &model.Order{
		OrderNo:         orderNo,
		UserId:          req.UserID,
		OrderStatus:     model.OrderStatusPendingPayment,
		PaymentStatus:   model.PaymentStatusUnpaid,
		ReceiverName:    "柴梦妃",                // TODO 根据 user_id 获取name  ,还是根据address直接获取name addressDbData.Name
		ReceiverPhone:   "13671210659",        // TODO 根据 user_id 获取phone  ,还是根据address直接获取phone addressDbData.Phone
		ReceiverAddress: "河北省廊坊市三河市燕郊镇四季花都一期", // TODO 待实现address.FullAddress(),
		TotalAmount:     totalAmount,
		PaymentAmount:   paymentAmount,
		Items:           items,
	}

	// 1.5 准备订单日志数据
	log := model.OrderLog{
		OrderNo:      order.OrderNo,
		Action:       "create_order",
		Operator:     fmt.Sprintf("user:%d", req.UserID),
		OperatorID:   req.UserID,
		OperatorType: model.OperatorTypeUser,
		Content:      "用户创建订单",
	}

	// 1.6 准备库存操作数据
	operationNo := pkg.GenerateOrderNo(pkg.StockPrefix, req.UserID)
	operation := &model.StockOperation{
		OperationNo:  operationNo,
		Types:        model.StockTypeOutbound,
		OutboundType: model.OutboundTypeMiniProgram, // 小程序购买出库
		Operator:     "",
		OperatorID:   0,
		OperatorType: model.OperatorTypeUser,
		UserName:     "小程序用户", // 可以从用户表获取真实姓名
		UserID:       userID,
		UserAccount:  "", // 可以从用户表获取账号
		Remark:       "小程序用户购买",
		TotalAmount:  totalAmount,
	}

	// 1.7 校验库存并准备库存操作明细
	var operationItems []model.StockOperationItem
	for _, item := range items {
		// 获取商品信息
		product, err := os.productRepo.GetByID(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("获取商品信息失败: %v", err)
		}

		// 检查库存是否足够
		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("商品 %s 库存不足，当前库存: %d，需要数量: %d", product.Name, product.Stock, item.Quantity)
		}

		// 构建库存操作明细
		operationItem := model.StockOperationItem{
			ProductID:     item.ProductID,
			ProductName:   product.Name,
			Specification: product.Specification,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			TotalPrice:    item.TotalPrice,
			BeforeStock:   product.Stock,
			AfterStock:    product.Stock - item.Quantity,
			Cost:          0, // 出库时不记录成本价
			ShippingCost:  0, // 出库时不记录运费
			ProductCost:   0, // 出库时不记录货物成本
			Remark:        "小程序用户购买",
		}
		operationItems = append(operationItems, operationItem)
	}

	// 2. 事务处理阶段 - 所有数据库操作在一个事务中执行
	err = os.orderRepo.ProcessCheckoutTransaction(order, operation, operationItems, req.CartIDs, &log)
	if err != nil {
		return nil, err
	}

	// 3. 返回订单信息
	return &model.CheckoutResponse{
		Items:         items,
		OrderNo:       orderNo,
		TotalAmount:   totalAmount,
		ShippingFee:   shippingFee,
		PaymentAmount: paymentAmount,
		AddressData:   addressInfo,
	}, nil
}

// processCheckoutTransaction 已废弃，功能已移至 repository 层
// 保留此方法用于向后兼容，但建议使用 repository 层的方法
func (os *orderService) processCheckoutTransaction(order *model.Order, operation *model.StockOperation, operationItems []model.StockOperationItem, cartIDs []int64, log *model.OrderLog) error {
	// 此方法已废弃，请使用 orderRepo.ProcessCheckoutTransaction 方法
	return fmt.Errorf("此方法已废弃，请使用 orderRepo.ProcessCheckoutTransaction 方法")
}

// getOrderItemsFromCart 从购物车获取订单商品和总金额
func (os *orderService) getOrderItemsFromCart(ctx context.Context, userID int64, cartIDs []int64) ([]model.StockOperationItem, model.Amount, error) {
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
	productMap := make(map[int64]model.Product)
	for _, product := range products {
		productMap[product.ID] = product
	}

	// 5. 构建订单商品项并计算总金额
	var orderItems []model.StockOperationItem
	var totalAmount model.Amount

	for _, cartItem := range cartItems {
		product, exists := productMap[cartItem.ProductID]
		if !exists {
			return nil, 0, fmt.Errorf("商品ID %d 不存在", cartItem.ProductID)
		}

		// 注意：库存检查在事务中进行，这里只做数据准备
		// 计算商品总价
		itemTotalPrice := int64(product.SellerPrice) * int64(cartItem.Quantity)

		// 构建订单商品项
		orderItem := model.StockOperationItem{
			ProductID:     product.ID,
			ProductName:   product.Name,
			Specification: product.Specification,
			Quantity:      cartItem.Quantity,
			UnitPrice:     product.SellerPrice,
			TotalPrice:    model.Amount(itemTotalPrice),
			Cost:          0,
			ShippingCost:  0,
			ProductCost:   0,
			Remark:        "从购物车创建订单",
		}
		orderItems = append(orderItems, orderItem)
		totalAmount += model.Amount(itemTotalPrice)
	}
	return orderItems, totalAmount, nil
}

// processStockOutboundWithNewStructure 已废弃，功能已整合到 processCheckoutTransaction 中
// 保留此方法用于向后兼容，但建议使用新的事务方法
func (os *orderService) processStockOutboundWithNewStructure(orderItems []model.StockOperationItem, orderNo string, orderID int64, userID int64) error {
	// 此方法已废弃，请使用 processCheckoutTransaction 方法
	return fmt.Errorf("此方法已废弃，请使用 processCheckoutTransaction 方法")
}

// getOrderItemsFromBuyNow 从立即购买获取订单商品和总金额
func (os *orderService) getOrderItemsFromBuyNow(ctx context.Context, userID int64, buyNowItems []*model.BuyNowItem) ([]model.StockOperationItem, model.Amount, error) {
	// 1. 参数校验
	if len(buyNowItems) == 0 {
		return nil, 0, errors.New("立即购买商品不能为空")
	}

	// 2. 获取商品ID列表
	productIDs := make([]int64, 0, len(buyNowItems))
	for _, item := range buyNowItems {
		productIDs = append(productIDs, item.ProductID)
	}

	// 3. 批量查询商品信息
	products, err := os.productRepo.GetByIDs(productIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("获取商品信息失败: %v", err)
	}

	// 4. 构建商品ID到商品信息的映射
	productMap := make(map[int64]model.Product)
	for _, product := range products {
		productMap[product.ID] = product
	}

	// 5. 构建订单商品项并计算总金额
	var orderItems []model.StockOperationItem
	var totalAmount model.Amount

	for _, buyNowItem := range buyNowItems {
		product, exists := productMap[buyNowItem.ProductID]
		if !exists {
			return nil, 0, fmt.Errorf("商品ID %d 不存在", buyNowItem.ProductID)
		}

		// 注意：库存检查在事务中进行，这里只做数据准备

		// 检查购买数量是否合法
		if buyNowItem.Quantity <= 0 {
			return nil, 0, fmt.Errorf("商品 %s 购买数量必须大于0", product.Name)
		}

		// 计算商品总价
		itemTotalPrice := int64(product.SellerPrice) * int64(buyNowItem.Quantity)

		// 构建订单商品项
		orderItem := model.StockOperationItem{
			ProductID:     product.ID,
			ProductName:   product.Name,
			Specification: product.Specification,
			Quantity:      buyNowItem.Quantity,
			UnitPrice:     product.SellerPrice,
			TotalPrice:    model.Amount(itemTotalPrice),
			Cost:          0,
			ShippingCost:  0,
			ProductCost:   0,
			Remark:        "立即购买创建订单",
		}
		orderItems = append(orderItems, orderItem)
		totalAmount += model.Amount(itemTotalPrice)
	}

	return orderItems, totalAmount, nil
}

func (os *orderService) GetOrderList(ctx context.Context, req *model.OrderListRequest) ([]model.Order, int64, error) {
	orders, total, err := os.orderRepo.GetOrderList(req)
	if err != nil {
		return nil, 0, err
	}
	// 查询每个订单的商品（从stock_operation_item表获取）
	for i := range orders {
		items, err := os.stockRepo.GetStockOperationItemsByOrderID(orders[i].ID)
		if err != nil {
			return nil, 0, err
		}
		orders[i].Items = items
	}
	return orders, total, nil
}
func (os *orderService) GetOrderDetail(ctx context.Context, userID int64, orderNo string) (*model.Order, error) {
	order, err := os.orderRepo.GetOrderByNo(userID, orderNo)
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
		OperatorID:   userID,
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
		OperatorID:   userID,
		OperatorType: model.OperatorTypeUser,
		Content:      "用户删除订单",
	}
	err := os.orderRepo.DeleteOrder(userID, order, log)
	return err
}
