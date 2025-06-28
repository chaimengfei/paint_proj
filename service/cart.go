package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/repository"
)

type CartService interface {
	GetCartList(userID int64) ([]model.CartWithProduct, error)
	AddToCart(userID, productID int64) error
	UpdateCartItem(userID, cartID int64, quantity int) error
	DeleteCartItem(userID, cartID int64) error
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
