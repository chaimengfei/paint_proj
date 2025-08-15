package repository

import (
	"cmf/paint_proj/model"

	"gorm.io/gorm"
)

type ProductRepository interface {
	GetProductCategory() ([]model.Category, map[int64]string, error) //  从product表查分类
	GetAllProduct() ([]model.Product, error)                         //  获取所有商品

	GetByID(productID int64) (*model.Product, error)
	GetByIDs(productIDs []int64) ([]*model.Product, error)
	GetList(offset, limit int) ([]model.Product, int64, error)

	Create(product *model.Product) error
	Update(product *model.Product) error
	Delete(id int64) error
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
	if err := p.db.Table("product p").
		Select("distinct p.category_id as id,c.name").
		Joins("INNER JOIN category c ON p.category_id = c.id").
		Where("p.is_on_shelf = ?", 1).
		Scan(&categories).Error; err != nil {
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
	if err := p.db.Model(&model.Product{}).Where("is_on_shelf = ?", 1).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
func (p *productRepository) GetByID(productID int64) (*model.Product, error) {
	var product model.Product
	err := p.db.Model(&model.Product{}).First(&product, productID).Error
	return &product, err
}
func (p *productRepository) GetByIDs(productIDs []int64) ([]*model.Product, error) {
	var products []*model.Product
	err := p.db.Model(&model.Product{}).Where("id in ?", productIDs).Find(&products).Error
	return products, err
}
func (p *productRepository) GetList(offset, limit int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	if err := p.db.Model(&model.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := p.db.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (p *productRepository) Create(product *model.Product) error {
	return p.db.Create(product).Error
}

func (p *productRepository) Update(product *model.Product) error {
	return p.db.Model(&model.Product{}).Where("id = ?", product.ID).Updates(product).Error
}

func (p *productRepository) Delete(id int64) error {
	return p.db.Delete(&model.Product{}, id).Error
}
