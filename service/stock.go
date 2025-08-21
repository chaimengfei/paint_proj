package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/pkg"
	"cmf/paint_proj/repository"
	"errors"
	"fmt"
)

type StockService interface {
	// 批量库存操作
	BatchInboundStock(req *model.BatchInboundRequest) error
	BatchOutboundStock(req *model.BatchOutboundRequest) error

	// 库存操作查询
	GetStockOperations(page, pageSize int) ([]model.StockOperation, int64, error)
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

	// 验证所有商品是否存在
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("商品ID %d 的入库数量必须大于0", item.ProductID)
		}

		_, err := ss.productRepo.GetByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("商品ID %d 不存在", item.ProductID)
		}
	}

	// 生成操作单号
	operationNo := pkg.GenerateOrderNo(pkg.StockPrefix, req.OperatorID)

	// 验证前端计算的总金额是否正确
	var calculatedTotalAmount model.Amount
	for _, item := range req.Items {
		if item.UnitPrice > 0 {
			calculatedTotalAmount += model.Amount(int64(item.UnitPrice) * int64(item.Quantity))
		}
	}

	// 如果前端提供了总金额，验证是否与计算值一致
	if req.TotalAmount > 0 && req.TotalAmount != calculatedTotalAmount {
		return fmt.Errorf("总金额计算错误，前端计算: %d，后端计算: %d", req.TotalAmount, calculatedTotalAmount)
	}

	// 使用前端提供的总金额，如果没有提供则使用计算值
	totalAmount := req.TotalAmount
	if totalAmount == 0 {
		totalAmount = calculatedTotalAmount
	}

	// 创建库存操作主表记录
	operation := &model.StockOperation{
		OperationNo:  operationNo,
		Type:         model.StockTypeInbound,
		Operator:     req.Operator,
		OperatorID:   req.OperatorID,
		OperatorType: model.OperatorTypeAdmin,
		Remark:       req.Remark,
		TotalAmount:  totalAmount,
	}

	err := ss.stockRepo.CreateStockOperation(operation)
	if err != nil {
		return fmt.Errorf("创建库存操作记录失败: %v", err)
	}

	// 批量处理入库并创建子表记录
	var operationItems []*model.StockOperationItem
	for _, item := range req.Items {
		err := ss.processInboundItemWithNewStructure(item, operation.ID, operationNo)
		if err != nil {
			return fmt.Errorf("商品ID %d 入库失败: %v", item.ProductID, err)
		}

		// 构建子表记录
		product, _ := ss.productRepo.GetByID(item.ProductID)
		beforeStock, _ := ss.stockRepo.GetProductStock(item.ProductID)
		afterStock := beforeStock + item.Quantity

		unitPrice := item.UnitPrice
		if unitPrice == 0 {
			unitPrice = product.SellerPrice // 如果没有提供单价，使用商品售价
		}

		operationItem := &model.StockOperationItem{
			OperationID:   operation.ID,
			ProductID:     item.ProductID,
			ProductName:   product.Name,
			Specification: product.Specification,
			Quantity:      item.Quantity,
			UnitPrice:     unitPrice,
			TotalPrice:    model.Amount(int64(unitPrice) * int64(item.Quantity)),
			BeforeStock:   beforeStock,
			AfterStock:    afterStock,
			Remark:        item.Remark,
		}
		operationItems = append(operationItems, operationItem)

		// 自动补齐商品信息到请求项中（用于前端显示）
		item.ProductName = product.Name
		item.Specification = product.Specification
		item.Unit = product.Unit
		item.TotalPrice = model.Amount(int64(unitPrice) * int64(item.Quantity))
	}

	// 批量创建子表记录
	err = ss.stockRepo.CreateStockOperationItems(operationItems)
	if err != nil {
		return fmt.Errorf("创建库存操作明细失败: %v", err)
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
		OperationNo:  operationNo,
		Type:         model.StockTypeOutbound,
		Operator:     req.Operator,
		OperatorID:   req.OperatorID,
		OperatorType: model.OperatorTypeAdmin,
		UserName:     req.UserName,
		UserID:       req.UserID,
		UserAccount:  req.UserAccount,
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
		unitPrice := item.UnitPrice
		if unitPrice == 0 {
			// 如果没有提供单价，获取商品售价
			product, err := ss.productRepo.GetByID(item.ProductID)
			if err != nil {
				return fmt.Errorf("获取商品ID %d 信息失败: %v", item.ProductID, err)
			}
			unitPrice = product.SellerPrice
		}

		afterStock := beforeStock - item.Quantity

		operationItem := &model.StockOperationItem{
			OperationID:   operation.ID,
			ProductID:     item.ProductID,
			ProductName:   item.ProductName,   // 使用前端传入的商品名称
			Specification: item.Specification, // 使用前端传入的规格
			Quantity:      item.Quantity,
			UnitPrice:     unitPrice,
			TotalPrice:    model.Amount(int64(unitPrice) * int64(item.Quantity)),
			BeforeStock:   beforeStock,
			AfterStock:    afterStock,
			Remark:        item.Remark,
		}
		operationItems = append(operationItems, operationItem)
	}

	// 将子表记录放入主表记录中
	for _, item := range operationItems {
		operation.Items = append(operation.Items, *item)
	}
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
func (ss *stockService) GetStockOperations(page, pageSize int) ([]model.StockOperation, int64, error) {
	operations, total, err := ss.stockRepo.GetStockOperations(page, pageSize)
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
