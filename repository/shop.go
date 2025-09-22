package repository

import (
	"cmf/paint_proj/model"

	"gorm.io/gorm"
)

type ShopRepository interface {
	GetAllActiveShops() ([]*model.Shop, error)
	GetShopByID(shopID int64) (*model.Shop, error)
	CreateShop(shop *model.Shop) error
	UpdateShop(shop *model.Shop) error
}

type shopRepository struct {
	db *gorm.DB
}

func NewShopRepository(db *gorm.DB) ShopRepository {
	return &shopRepository{db: db}
}

func (r *shopRepository) GetAllActiveShops() ([]*model.Shop, error) {
	var shops []*model.Shop
	err := r.db.Where("is_active = ?", 1).Find(&shops).Error
	return shops, err
}

func (r *shopRepository) GetShopByID(shopID int64) (*model.Shop, error) {
	var shop model.Shop
	err := r.db.Where("id = ?", shopID).First(&shop).Error
	return &shop, err
}

func (r *shopRepository) CreateShop(shop *model.Shop) error {
	return r.db.Create(shop).Error
}

func (r *shopRepository) UpdateShop(shop *model.Shop) error {
	return r.db.Save(shop).Error
}
