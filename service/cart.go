package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/repository"
)

type CartService interface {
	GetCartList(userID int64, shopID int64) ([]model.CartWithProduct, error)
	AddToCart(userID, productID int64, shopID int64) error
	UpdateCartItem(userID, shopID, cartID int64, quantity int) error
	DeleteCartItem(userID, shopID, cartID int64) error
}

type cartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
	userRepo    repository.UserRepository
}

func NewCartService(cr repository.CartRepository, pr repository.ProductRepository, ur repository.UserRepository) CartService {
	return &cartService{
		cartRepo:    cr,
		productRepo: pr,
		userRepo:    ur,
	}
}

func (cs *cartService) GetCartList(userID int64, shopID int64) ([]model.CartWithProduct, error) {
	// 根据用户店铺获取购物车商品
	cartItems, err := cs.cartRepo.GetByUserIDAndShopWithProduct(userID, shopID)
	if err != nil {
		return nil, err
	}

	return cartItems, nil
}
func (cs *cartService) AddToCart(userID, productID int64, shopID int64) error {
	// 检查商品是否属于该店铺
	_, err := cs.productRepo.GetByIDAndShop(productID, shopID)
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
		ShopID:    shopID,
		ProductID: productID,
		Quantity:  1,
		Selected:  true,
	}
	return cs.cartRepo.Create(cart)
}

func (cs *cartService) UpdateCartItem(userID, shopID, cartID int64, quantity int) error {
	// 验证购物车项属于该用户和店铺
	_, err := cs.cartRepo.GetByIDAndUserAndShop(cartID, userID, shopID)
	if err != nil {
		return err
	}

	return cs.cartRepo.UpdateQuantity(cartID, quantity)
}

func (cs *cartService) DeleteCartItem(userID, shopID, cartID int64) error {
	// 验证购物车项属于该用户和店铺
	_, err := cs.cartRepo.GetByIDAndUserAndShop(cartID, userID, shopID)
	if err != nil {
		return err
	}

	return cs.cartRepo.Delete(cartID)
}
