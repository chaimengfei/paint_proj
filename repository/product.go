package repository

import (
	"cmf/paint_proj/model"
	"gorm.io/gorm"
)

type ProductRepository interface {
	GetProductCategory() ([]model.Category, map[int64]string, error) //  从product表查分类
	GetAllProduct() ([]model.Product, error)                         //  获取所有商品

	GetByID(productID int64) (*model.Product, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

// GetProductCategory  从product表查分类
func (p *productRepository) GetProductCategory() ([]model.Category, map[int64]string, error) {
	var categories []model.Category
	if err := p.db.Table("product p").Select("distinct p.category_id as id,c.name").Joins("INNER JOIN category c ON p.category_id = c.id").Scan(&categories).Error; err != nil {
		return nil, nil, err
	}
	var categoryMap = make(map[int64]string)
	for _, category := range categories {
		categoryMap[category.ID] = category.Name
	}
	return categories, categoryMap, nil
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
