package repository

import (
	"cmf/paint_proj/model"
	"gorm.io/gorm"
)

type ProductRepository interface {
	GetProductCategory() ([]model.Category, error) //  获取所有分类
	GetAllProduct() ([]model.Product, error)       //  获取所有分类

	GetByID(productID int64) (*model.Product, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

// GetProductCategory 从product表查有效的分类
func (p *productRepository) GetProductCategory() ([]model.Category, error) {
	var categories []model.Category
	if err := p.db.Model(&model.Product{}).Select("distinct category_id as id,category_name as name").Order("id").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// GetAllProduct 获取所有商品
func (p *productRepository) GetAllProduct() ([]model.Product, error) {
	var products []model.Product
	if err := p.db.Model(&model.Product{}).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
func (p *productRepository) GetByID(productID int64) (*model.Product, error) {
	var product model.Product
	err := p.db.Model(&model.Product{}).First(&product, productID).Error
	return &product, err
}
