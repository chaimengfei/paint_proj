package repository

import (
	"cmf/paint_proj/model"

	"gorm.io/gorm"
)

type CartRepository interface {
	Create(cart *model.Cart) error
	UpdateQuantity(id int64, quantity int) error
	Delete(id int64) error

	GetByID(id int64) (*model.Cart, error)
	GetByIDs(ids []int64) ([]model.Cart, error)
	GetByIDAndUser(id, userID int64) (*model.Cart, error)
	GetByUserAndProduct(userID, productID int64) (*model.Cart, error)
	GetByUserID(userID int64) ([]model.Cart, error)
	GetByUserIDWithProduct(userID int64) ([]model.CartWithProduct, error)
	GetByUserIDAndShop(userID int64, shopID int64) ([]model.Cart, error)
	GetByUserIDAndShopWithProduct(userID int64, shopID int64) ([]model.CartWithProduct, error)
}

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

func (cr *cartRepository) Create(cart *model.Cart) error {
	return cr.db.Model(&model.Cart{}).Create(cart).Error
}
func (cr *cartRepository) UpdateQuantity(id int64, quantity int) error {
	return cr.db.Model(&model.Cart{}).Model(&model.Cart{}).Where("id = ?", id).Update("quantity", quantity).Error
}
func (cr *cartRepository) Delete(id int64) error {
	return cr.db.Model(&model.Cart{}).Delete(&model.Cart{}, id).Error
}

func (cr *cartRepository) GetByID(id int64) (*model.Cart, error) {
	var cart model.Cart
	err := cr.db.Model(&model.Cart{}).First(&cart, id).Error
	return &cart, err
}
func (cr *cartRepository) GetByIDs(ids []int64) ([]model.Cart, error) {
	var carts []model.Cart
	err := cr.db.Model(&model.Cart{}).Where("id in ?", ids).Find(&carts).Error
	return carts, err
}
func (cr *cartRepository) GetByIDAndUser(id, userID int64) (*model.Cart, error) {
	var cart model.Cart
	err := cr.db.Model(&model.Cart{}).Where("id = ? AND user_id = ?", id, userID).First(&cart).Error
	return &cart, err
}

func (cr *cartRepository) GetByUserAndProduct(userID, productID int64) (*model.Cart, error) {
	var cart model.Cart
	err := cr.db.Model(&model.Cart{}).Where("user_id = ? AND product_id = ?", userID, productID).First(&cart).Error
	return &cart, err
}

func (cr *cartRepository) GetByUserID(userID int64) ([]model.Cart, error) {
	var carts []model.Cart
	err := cr.db.Model(&model.Cart{}).Where("user_id = ?", userID).Find(&carts).Error
	return carts, err
}

func (cr *cartRepository) GetByUserIDWithProduct(userID int64) ([]model.CartWithProduct, error) {
	var carts []model.CartWithProduct
	err := cr.db.Table("cart c").
		Select("c.*, p.name as product_name, p.image as product_image, p.seller_price as product_seller_price, p.unit as product_unit").
		Joins("LEFT JOIN product p ON c.product_id = p.id").
		Where("c.user_id = ?", userID).
		Scan(&carts).Error
	return carts, err
}

func (cr *cartRepository) GetByUserIDAndShop(userID int64, shopID int64) ([]model.Cart, error) {
	var carts []model.Cart
	err := cr.db.Model(&model.Cart{}).Where("user_id = ? AND shop_id = ?", userID, shopID).Find(&carts).Error
	return carts, err
}

func (cr *cartRepository) GetByUserIDAndShopWithProduct(userID int64, shopID int64) ([]model.CartWithProduct, error) {
	var carts []model.CartWithProduct
	err := cr.db.Table("cart c").
		Select("c.*, p.name as product_name, p.image as product_image, p.seller_price as product_seller_price, p.unit as product_unit").
		Joins("LEFT JOIN product p ON c.product_id = p.id").
		Where("c.user_id = ? AND c.shop_id = ?", userID, shopID).
		Scan(&carts).Error
	return carts, err
}
