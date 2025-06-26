package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/repository"
	"errors"
	"fmt"
)

type CartService interface {
	GetCartList(userID int64) ([]model.CartWithProduct, error)
	AddToCart(userID, productID int64) error
	UpdateCartItem(userID, cartID int64, quantity int) error
	DeleteCartItem(userID, cartID int64) error
	CheckoutCart(userID int64, cartIDs []int64) (*model.CheckoutResponse, error)
}

type cartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartService(cr repository.CartRepository, pr repository.ProductRepository) CartService {
	return &cartService{
		cartRepo:    cr,
		productRepo: pr,
	}
}

func (cs *cartService) GetCartList(userID int64) ([]model.CartWithProduct, error) {
	// 获取购物车列表并关联商品信息
	cartItems, err := cs.cartRepo.GetByUserIDWithProduct(userID)
	if err != nil {
		return nil, err
	}

	return cartItems, nil
}
func (cs *cartService) AddToCart(userID, productID int64) error {
	// 检查商品是否存在
	_, err := cs.productRepo.GetByID(productID)
	if err != nil {
		return err
	}

	// 检查是否已存在购物车
	existingItem, err := cs.cartRepo.GetByUserAndProduct(userID, productID)
	if err == nil && existingItem != nil {
		// 已存在则增加数量
		return cs.cartRepo.UpdateQuantity(existingItem.ID, existingItem.Quantity+1)
	}

	// 不存在则创建
	cart := &model.Cart{
		UserID:    userID,
		ProductID: productID,
		Quantity:  1,
		Selected:  true,
	}
	return cs.cartRepo.Create(cart)
}

func (cs *cartService) UpdateCartItem(userID, cartID int64, quantity int) error {
	// 验证购物车项属于该用户
	_, err := cs.cartRepo.GetByIDAndUser(cartID, userID)
	if err != nil {
		return err
	}

	return cs.cartRepo.UpdateQuantity(cartID, quantity)
}

func (cs *cartService) DeleteCartItem(userID, cartID int64) error {
	// 验证购物车项属于该用户
	_, err := cs.cartRepo.GetByIDAndUser(cartID, userID)
	if err != nil {
		return err
	}

	return cs.cartRepo.Delete(cartID)
}

func (cs *cartService) CheckoutCart(userID int64, cartIDs []int64) (*model.CheckoutResponse, error) {
	// 1. 获取购物车项
	cartItems, err := cs.cartRepo.GetByIDs(cartIDs)
	if err != nil {
		return nil, fmt.Errorf("用户 %v 获取购物车失败: %v", userID, err)
	}
	if len(cartItems) == 0 {
		return nil, errors.New("购物车中没有选中商品")
	}
	// 2. 获取商品信息
	productIDs := make([]int64, len(cartItems))
	for i, item := range cartItems {
		productIDs[i] = item.ProductID
	}
	products, err := cs.productRepo.GetByIDs(productIDs)
	if err != nil {
		return nil, fmt.Errorf("获取商品信息失败: %v", err)
	}
	// 3. 构建返回数据
	var orderItems []*model.OrderItem
	var totalAmount float64
	productMap := make(map[int64]*model.Product)
	for _, p := range products {
		productMap[p.ID] = p
	}
	for _, cartItem := range cartItems {
		product, exists := productMap[cartItem.ProductID]
		if !exists {
			return nil, fmt.Errorf("商品ID %d 不存在", cartItem.ProductID)
		}

		// 检查库存
		if product.Stock < cartItem.Quantity {
			return nil, fmt.Errorf("商品 %s 库存不足", product.Name)
		}

		itemTotal := product.SellerPrice * float64(cartItem.Quantity)
		totalAmount += itemTotal

		orderItems = append(orderItems, &model.OrderItem{
			ProductId:    product.ID,
			ProductName:  product.Name,
			ProductImage: product.Image,
			ProductPrice: product.SellerPrice,
			Quantity:     cartItem.Quantity,
			Unit:         product.Unit,
			TotalPrice:   itemTotal,
		})
	}
	// 4. 计算运费等 (这里简化处理，实际业务需要计算运费、优惠等)
	shippingFee := 0.0
	if totalAmount < 100 { // 假设满100免运费
		shippingFee = 10.0
	}

	return &model.CheckoutResponse{
		OrderItems:    orderItems,
		TotalAmount:   totalAmount,
		ShippingFee:   shippingFee,
		PaymentAmount: totalAmount + shippingFee,
	}, nil
}
