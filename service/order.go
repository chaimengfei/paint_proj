package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/pkg"
	"cmf/paint_proj/repository"
	"context"
	"errors"
	"fmt"
	"time"
)

type OrderService interface {
	CheckoutOrder(ctx context.Context, userID int64, req *model.CheckoutOrderRequest) (*model.CheckoutResponse, error)
	GetOrderList(ctx context.Context, req *model.OrderListRequest) ([]*model.Order, int64, error) // 获取订单列表
	GetOrderDetail(ctx context.Context, userID int64, orderNo string) (*model.Order, error)       // 获取订单详情

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
	// 1. 参数校验
	if len(req.CartIDs) == 0 && len(req.BuyNowItems) == 0 {
		return nil, errors.New("购物车或立即购买商品不能为空")
	}

	//2. 获取用户收货地址
	var addressDbData *model.Address
	var err error
	if req.AddressID == 0 {
		addressDbData, err = os.addressRepo.GetDefaultOrFirstAddressID(userID)
	} else {
		addressDbData, err = os.addressRepo.GetByUserAppointId(req.UserID, req.AddressID)
	}
	var addressInfo *model.AddressInfo = nil
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

	// 3. 创建订单
	orderNo := pkg.GenerateOrderNo(req.UserID)
	order := &model.Order{
		OrderNo:         orderNo,
		UserId:          req.UserID,
		OrderStatus:     model.OrderStatusPendingPayment,
		PaymentStatus:   model.PaymentStatusUnpaid,
		ReceiverName:    "柴梦妃",                // TODO 根据 user_id 获取name  ,还是根据address直接获取name addressDbData.Name
		ReceiverPhone:   "13671210659",        // TODO 根据 user_id 获取phone  ,还是根据address直接获取phone addressDbData.Phone
		ReceiverAddress: "河北省廊坊市三河市燕郊镇四季花都一期", // TODO 待实现address.FullAddress(),
	}
	// 4. 获取订单商品
	var orderItems []*model.OrderItem
	var totalAmount model.Amount
	if len(req.CartIDs) > 0 {
		// 4.1. 从购物车创建订单
		orderItems, totalAmount, err = os.getOrderItemsFromCart(ctx, req.UserID, req.CartIDs)
	} else {
		// 4.2. 立即购买创建订单
		orderItems, totalAmount, err = os.getOrderItemsFromBuyNow(ctx, req.UserID, req.BuyNowItems)
	}
	if err != nil {
		return nil, err
	}
	// 4.4. 计算运费等 (这里简化处理，实际业务需要计算运费、优惠等)
	shippingFee := model.Amount(0)
	if totalAmount < 100 { // 假设满100免运费
		shippingFee = 1000 //单位:分(10块的运费)
	}
	paymentAmount := totalAmount + shippingFee
	// 4.3. 计算优惠金额等
	order.TotalAmount = totalAmount
	order.PaymentAmount = paymentAmount // 实付金额(这里简化处理，实际可能有优惠券、运费等)

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

	// 6.1. 处理库存出库（使用新的主表+子表结构）
	err = os.processStockOutboundWithNewStructure(orderItems, order.OrderNo, req.UserID)
	if err != nil {
		// 注意：这里如果库存出库失败，订单已经创建成功，实际业务中可能需要回滚订单
		// 或者记录错误日志，后续手动处理
		return nil, fmt.Errorf("订单创建成功但库存出库失败: %v", err)
	}

	// 7.返回订单信息
	return &model.CheckoutResponse{
		OrderItems:    orderItems,
		OrderNo:       orderNo,
		TotalAmount:   totalAmount,
		ShippingFee:   shippingFee,
		PaymentAmount: paymentAmount,
		AddressData:   addressInfo,
	}, nil
}

// getOrderItemsFromCart 从购物车获取订单商品和总金额
func (os *orderService) getOrderItemsFromCart(ctx context.Context, userID int64, cartIDs []int64) ([]*model.OrderItem, model.Amount, error) {
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
	var orderItems []*model.OrderItem
	var totalAmount model.Amount

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
		itemTotalPrice := int64(product.SellerPrice) * int64(cartItem.Quantity)

		// 构建订单商品项
		orderItem := model.OrderItem{
			ProductId:    product.ID,
			ProductName:  product.Name,
			ProductImage: product.Image,
			ProductPrice: product.SellerPrice,
			Quantity:     cartItem.Quantity,
			Unit:         product.Unit,
			TotalPrice:   model.Amount(itemTotalPrice),
		}
		orderItems = append(orderItems, &orderItem)
		totalAmount += model.Amount(itemTotalPrice)
	}
	return orderItems, totalAmount, nil
}

// processStockOutboundWithNewStructure 处理库存出库（使用新的主表+子表结构）
func (os *orderService) processStockOutboundWithNewStructure(orderItems []*model.OrderItem, orderNo string, userID int64) error {
	if len(orderItems) == 0 {
		return nil
	}

	// 生成操作单号
	operationNo := pkg.GenerateOrderNo(userID)

	// 计算总金额
	var totalAmount model.Amount
	for _, item := range orderItems {
		totalAmount += item.TotalPrice
	}

	// 创建库存操作主表记录
	now := time.Now()
	operation := &model.StockOperation{
		OperationNo:  operationNo,
		Type:         model.StockTypeOutbound,
		Operator:     fmt.Sprintf("user:%d", userID),
		OperatorID:   userID,
		OperatorType: model.OperatorTypeUser,
		UserName:     "小程序用户", // 可以从用户表获取真实姓名
		UserID:       userID,
		UserAccount:  "", // 可以从用户表获取账号
		PurchaseTime: &now,
		Remark:       "小程序用户购买",
		TotalAmount:  totalAmount,
		CreatedAt:    &now,
	}

	err := os.stockRepo.CreateStockOperation(operation)
	if err != nil {
		return fmt.Errorf("创建库存操作记录失败: %v", err)
	}

	// 批量处理出库并创建子表记录
	var operationItems []*model.StockOperationItem
	for _, item := range orderItems {
		// 获取商品信息
		product, err := os.productRepo.GetByID(item.ProductId)
		if err != nil {
			return fmt.Errorf("获取商品信息失败: %v", err)
		}

		// 获取操作前库存
		beforeStock, err := os.stockRepo.GetProductStock(item.ProductId)
		if err != nil {
			return fmt.Errorf("获取商品库存失败: %v", err)
		}

		// 检查库存是否足够
		if beforeStock < item.Quantity {
			return fmt.Errorf("商品 %s 库存不足", product.Name)
		}

		// 更新库存（出库为负数）
		err = os.stockRepo.UpdateProductStock(item.ProductId, -item.Quantity)
		if err != nil {
			return fmt.Errorf("更新商品库存失败: %v", err)
		}

		// 获取操作后库存
		afterStock := beforeStock - item.Quantity

		// 构建子表记录
		operationItem := &model.StockOperationItem{
			OperationID:   operation.ID,
			ProductID:     item.ProductId,
			ProductName:   product.Name,
			Specification: product.Specification,
			Quantity:      item.Quantity,
			UnitPrice:     item.ProductPrice,
			TotalPrice:    item.TotalPrice,
			BeforeStock:   beforeStock,
			AfterStock:    afterStock,
			CreatedAt:     &now,
		}
		operationItems = append(operationItems, operationItem)
	}

	// 批量创建子表记录
	err = os.stockRepo.CreateStockOperationItems(operationItems)
	if err != nil {
		return fmt.Errorf("创建库存操作明细失败: %v", err)
	}

	return nil
}

// getOrderItemsFromBuyNow 从立即购买获取订单商品和总金额
func (os *orderService) getOrderItemsFromBuyNow(ctx context.Context, userID int64, buyNowItems []*model.BuyNowItem) ([]*model.OrderItem, model.Amount, error) {
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
	productMap := make(map[int64]*model.Product)
	for _, product := range products {
		productMap[product.ID] = product
	}

	// 5. 构建订单商品项并计算总金额
	var orderItems []*model.OrderItem
	var totalAmount model.Amount

	for _, buyNowItem := range buyNowItems {
		product, exists := productMap[buyNowItem.ProductID]
		if !exists {
			return nil, 0, fmt.Errorf("商品ID %d 不存在", buyNowItem.ProductID)
		}

		// 检查商品库存
		if product.Stock < buyNowItem.Quantity {
			return nil, 0, fmt.Errorf("商品 %s 库存不足", product.Name)
		}

		// 检查购买数量是否合法
		if buyNowItem.Quantity <= 0 {
			return nil, 0, fmt.Errorf("商品 %s 购买数量必须大于0", product.Name)
		}

		// 计算商品总价
		itemTotalPrice := int64(product.SellerPrice) * int64(buyNowItem.Quantity)

		// 构建订单商品项
		orderItem := model.OrderItem{
			ProductId:    product.ID,
			ProductName:  product.Name,
			ProductImage: product.Image,
			ProductPrice: product.SellerPrice,
			Quantity:     buyNowItem.Quantity,
			Unit:         product.Unit,
			TotalPrice:   model.Amount(itemTotalPrice),
		}
		orderItems = append(orderItems, &orderItem)
		totalAmount += model.Amount(itemTotalPrice)
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
