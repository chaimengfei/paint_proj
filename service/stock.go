package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/pkg"
	"cmf/paint_proj/repository"
	"errors"
	"fmt"
	"time"
)

type StockService interface {
	// 批量库存操作
	BatchInboundStock(req *model.BatchInboundRequest) error
	BatchOutboundStock(req *model.BatchOutboundRequest) error

	// 更新出库单支付状态
	UpdateOutboundPaymentStatus(req *model.UpdateOutboundPaymentStatusRequest) error

	// 库存操作查询
	GetStockOperations(page, pageSize int, types *int8) ([]model.StockOperation, int64, error)
	GetStockOperationDetail(operationID int64) (*model.StockOperation, []model.StockOperationItem, error)
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

// BatchInboundStock 批量入库操作（新结构）
func (ss *stockService) BatchInboundStock(req *model.BatchInboundRequest) error {
	if len(req.Items) == 0 {
		return errors.New("入库商品列表不能为空")
	}

	// 使用前端提供的总金额，如果没有提供则使用计算值
	totalAmount := req.TotalAmount
	if totalAmount == 0 {
		var calculatedTotalAmount model.Amount
		for _, item := range req.Items {
			calculatedTotalAmount += model.Amount(int64(item.Cost) * int64(item.Quantity))
		}
		totalAmount = calculatedTotalAmount
	}

	// 生成操作单号
	operationNo := pkg.GenerateOrderNo(pkg.StockPrefix, req.OperatorID)

	// 创建库存操作主表记录
	operation := &model.StockOperation{
		OperationNo:  operationNo,
		Types:        model.StockTypeInbound,
		Operator:     req.Operator,
		OperatorID:   req.OperatorID,
		OperatorType: model.OperatorTypeAdmin,
		Remark:       req.Remark,
		TotalAmount:  totalAmount,
	}

	// 构建子表记录
	var operationItems []*model.StockOperationItem
	for _, item := range req.Items {
		// 获取当前库存
		beforeStock, err := ss.stockRepo.GetProductStock(item.ProductID)
		if err != nil {
			return fmt.Errorf("获取商品ID %d 库存失败: %v", item.ProductID, err)
		}

		afterStock := beforeStock + item.Quantity

		operationItem := &model.StockOperationItem{
			OperationID:   operation.ID,
			ProductID:     item.ProductID,
			ProductName:   item.ProductName,   // 使用前端传入的商品名称
			Specification: item.Specification, // 使用前端传入的规格
			Unit:          item.Unit,          // 使用前端传入的单位
			Quantity:      item.Quantity,
			UnitPrice:     0, // 入库时不记录单价
			TotalPrice:    model.Amount(int64(item.Cost) * int64(item.Quantity)),
			BeforeStock:   beforeStock,
			AfterStock:    afterStock,
			Cost:          item.Cost,         // 成本价
			ShippingCost:  item.ShippingCost, // 运费成本
			ProductCost:   item.ProductCost,  // 货物成本
			Remark:        item.Remark,
		}
		operationItems = append(operationItems, operationItem)
	}

	// 将子表记录放入主表记录中
	for _, item := range operationItems {
		operation.Items = append(operation.Items, *item)
	}

	// 执行事务：创建主表记录、子表记录、更新库存
	err := ss.stockRepo.ProcessInboundTransaction(operation)
	if err != nil {
		return fmt.Errorf("批量入库事务失败: %v", err)
	}

	return nil
}

// BatchOutboundStock 批量出库操作（新结构）
func (ss *stockService) BatchOutboundStock(req *model.BatchOutboundRequest) error {
	// 使用前端提供的总金额，如果没有提供则使用计算值
	totalAmount := req.TotalAmount
	if totalAmount == 0 {
		var calculatedTotalAmount model.Amount
		for _, item := range req.Items {
			if item.UnitPrice > 0 {
				calculatedTotalAmount += model.Amount(int64(item.UnitPrice) * int64(item.Quantity))
			}
		}
		totalAmount = calculatedTotalAmount
	}

	// 生成操作单号
	operationNo := pkg.GenerateOrderNo(pkg.StockPrefix, req.UserID)

	// 创建库存操作主表记录
	operation := &model.StockOperation{
		OperationNo:         operationNo,
		Types:               model.StockTypeOutbound,
		OutboundType:        model.OutboundTypeAdmin, // admin后台操作出库
		Operator:            req.Operator,
		OperatorID:          req.OperatorID,
		OperatorType:        model.OperatorTypeAdmin,
		UserName:            req.UserName,
		UserID:              req.UserID,
		Remark:              req.Remark,
		TotalAmount:         totalAmount,
		TotalProfit:         0,                         // 初始化为0，后面会计算
		PaymentFinishStatus: model.PaymentStatusUnpaid, // 初始化为未支付
	}

	// 如果前端传了操作时间，则设置到CreatedAt字段
	if req.OperateTime != nil {
		operation.CreatedAt = req.OperateTime
	}

	// 构建子表记录并计算利润
	var operationItems []model.StockOperationItem
	var totalProfit model.Amount

	for _, item := range req.Items {
		// 获取商品信息（包含库存、成本价、售价等）
		product, err := ss.productRepo.GetByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("获取商品ID %d 信息失败: %v", item.ProductID, err)
		}

		// 从商品信息中获取当前库存
		beforeStock := product.Stock

		// 确定单价：优先使用前端传入的单价，如果没有则使用商品售价
		unitPrice := item.UnitPrice
		if unitPrice == 0 {
			unitPrice = product.SellerPrice
		}

		// 计算利润：(卖价 - 成本价) * 数量
		profit := model.Amount((int64(unitPrice) - int64(product.Cost)) * int64(item.Quantity))
		totalProfit += profit

		afterStock := beforeStock - item.Quantity

		operationItem := model.StockOperationItem{
			OperationID:   operation.ID,
			ProductID:     item.ProductID,
			ProductName:   item.ProductName,   // 使用前端传入的商品名称
			Specification: item.Specification, // 使用前端传入的规格
			Unit:          item.Unit,          // 使用前端传入的单位
			Quantity:      item.Quantity,
			UnitPrice:     unitPrice,
			TotalPrice:    model.Amount(int64(unitPrice) * int64(item.Quantity)),
			BeforeStock:   beforeStock,
			AfterStock:    afterStock,
			Cost:          product.Cost, // 记录成本价
			ShippingCost:  product.ShippingCost,
			ProductCost:   product.ProductCost,
			Profit:        profit, // 记录利润
			Remark:        item.Remark,
		}
		operationItems = append(operationItems, operationItem)
	}

	// 设置总利润
	operation.TotalProfit = totalProfit
	operation.Items = operationItems

	// 执行事务：创建主表记录、子表记录、更新库存
	err := ss.stockRepo.ProcessOutboundTransaction(operation)
	if err != nil {
		return fmt.Errorf("批量出库事务失败: %v", err)
	}

	return nil
}

// processInboundItemWithNewStructure 处理单个入库商品（新结构）
func (ss *stockService) processInboundItemWithNewStructure(item model.BatchInboundItem, operationID int64, operationNo string) error {
	// 更新库存
	err := ss.stockRepo.UpdateProductStock(item.ProductID, item.Quantity)
	if err != nil {
		return err
	}

	return nil
}

// processOutboundItemWithNewStructure 处理单个出库商品（新结构）
func (ss *stockService) processOutboundItemWithNewStructure(item model.BatchOutboundItem, operationID int64, operationNo string) error {
	// 更新库存（出库为负数）
	err := ss.stockRepo.UpdateProductStock(item.ProductID, -item.Quantity)
	if err != nil {
		return err
	}

	return nil
}

// GetStockOperations 获取库存操作列表
func (ss *stockService) GetStockOperations(page, pageSize int, types *int8) ([]model.StockOperation, int64, error) {
	operations, total, err := ss.stockRepo.GetStockOperations(page, pageSize, types)
	if err != nil {
		return nil, 0, err
	}

	// 为每个操作填充Items字段
	for i := range operations {
		items, err := ss.stockRepo.GetStockOperationItems(operations[i].ID)
		if err != nil {
			return nil, 0, fmt.Errorf("获取操作ID %d 的明细失败: %v", operations[i].ID, err)
		}
		operations[i].Items = items
	}

	return operations, total, nil
}

// GetStockOperationDetail 获取库存操作详情
func (ss *stockService) GetStockOperationDetail(operationID int64) (*model.StockOperation, []model.StockOperationItem, error) {
	operation, err := ss.stockRepo.GetStockOperationByID(operationID)
	if err != nil {
		return nil, nil, err
	}

	items, err := ss.stockRepo.GetStockOperationItems(operationID)
	if err != nil {
		return nil, nil, err
	}

	return operation, items, nil
}

// UpdateOutboundPaymentStatus 更新出库单支付状态
func (ss *stockService) UpdateOutboundPaymentStatus(req *model.UpdateOutboundPaymentStatusRequest) error {
	// 验证出库单是否存在
	operation, err := ss.stockRepo.GetStockOperationByID(req.OperationID)
	if err != nil {
		return fmt.Errorf("出库单不存在: %v", err)
	}

	// 验证是否为出库单
	if operation.Types != model.StockTypeOutbound {
		return fmt.Errorf("只能更新出库单的支付状态")
	}

	// 验证支付完成状态是否有效（只允许未支付和已支付）
	if req.PaymentFinishStatus != model.PaymentStatusUnpaid && req.PaymentFinishStatus != model.PaymentStatusPaid {
		return fmt.Errorf("无效的支付完成状态，只允许设置为未支付(1)或已支付(3)")
	}

	// 设置支付完成时间
	var paymentFinishTime *time.Time
	if req.PaymentFinishStatus == model.PaymentStatusPaid {
		now := time.Now()
		paymentFinishTime = &now
	}

	// 更新支付完成状态
	err = ss.stockRepo.UpdateOutboundPaymentStatus(req.OperationID, req.PaymentFinishStatus, paymentFinishTime)
	if err != nil {
		return fmt.Errorf("更新支付状态失败: %v", err)
	}

	return nil
}
