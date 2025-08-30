package repository

import (
	"cmf/paint_proj/model"
	"errors"

	"gorm.io/gorm"
)

type ProductRepository interface {
	GetProductCategory() ([]model.Category, map[int64]string, error) //  从product表查分类
	GetAllCategories() ([]model.Category, error)                     //  获取所有分类
	GetAllProduct() ([]model.Product, error)                         //  获取所有商品

	GetByID(productID int64) (*model.Product, error)
	GetByIDs(productIDs []int64) ([]model.Product, error)
	GetList(offset, limit int) ([]model.Product, int64, error)

	Create(product *model.Product) error
	Update(product *model.Product) error
	Delete(id int64) error

	// 分类管理方法
	CreateCategory(category *model.Category) error
	UpdateCategory(category *model.Category) error
	DeleteCategory(id int64) error
	GetCategoryByID(id int64) (*model.Category, error)
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

// GetAllCategories 获取所有分类
func (p *productRepository) GetAllCategories() ([]model.Category, error) {
	var categories []model.Category
	if err := p.db.Model(&model.Category{}).Order("sort_order desc, id asc").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
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
func (p *productRepository) GetByIDs(productIDs []int64) ([]model.Product, error) {
	var products []model.Product
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

// 分类管理方法实现
func (p *productRepository) CreateCategory(category *model.Category) error {
	return p.db.Create(category).Error
}

func (p *productRepository) UpdateCategory(category *model.Category) error {
	return p.db.Model(&model.Category{}).Where("id = ?", category.ID).Updates(category).Error
}

func (p *productRepository) DeleteCategory(id int64) error {
	// 检查是否有商品使用此分类
	var count int64
	if err := p.db.Model(&model.Product{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该分类下还有商品，无法删除")
	}
	return p.db.Delete(&model.Category{}, id).Error
}

func (p *productRepository) GetCategoryByID(id int64) (*model.Category, error) {
	var category model.Category
	err := p.db.Model(&model.Category{}).First(&category, id).Error
	return &category, err
}
