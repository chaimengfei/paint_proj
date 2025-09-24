package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/repository"
)

type ProductService interface {
	GetProductList() ([]model.Category, map[int64][]model.ProductSimple, error)
	GetProductListByShop(shopID int64) ([]model.Category, map[int64][]model.ProductSimple, error)

	GetAdminProductList(page, pageSize int, shopID int64) ([]model.Product, int64, error)
	GetProductByID(id int64) (*model.Product, error)
	GetProductByIDAndShop(id int64, shopID int64) (*model.Product, error)
	AddProduct(p *model.Product) error
	UpdateProduct(p *model.Product) error
	DeleteProduct(id int64) error
	CheckProductNameExists(name string, excludeID ...int64) (bool, error)

	// 分类管理方法
	GetAllCategories() ([]model.Category, error)
	GetCategoriesByShop(shopID int64) ([]model.Category, error)
	AddCategory(category *model.Category) error
	UpdateCategory(category *model.Category) error
	DeleteCategory(id int64) error
	GetCategoryByID(id int64) (*model.Category, error)
}

type productService struct {
	productRepo repository.ProductRepository
}

func NewProductService(pr repository.ProductRepository) ProductService {
	return &productService{
		productRepo: pr,
	}
}
func (ps *productService) GetProductList() ([]model.Category, map[int64][]model.ProductSimple, error) {
	// 1 从product表查分类
	categories, categoryMap, err := ps.productRepo.GetProductCategory()
	if err != nil {
		return nil, nil, err
	}
	// 2.获取所有商品
	products, err := ps.productRepo.GetAllProduct()
	if err != nil {
		return nil, nil, err
	}
	// 3.按分类分组
	productMap := make(map[int64][]model.ProductSimple)
	for _, p := range products {
		sp := model.ProductSimple{
			ID:           p.ID,
			Name:         p.Name,
			SellerPrice:  p.SellerPrice,
			CategoryID:   p.CategoryId,
			CategoryName: categoryMap[p.CategoryId],
			Image:        p.Image,
			Unit:         p.Unit,
			Remark:       p.Remark,
		}
		productMap[p.CategoryId] = append(productMap[p.CategoryId], sp)
	}
	return categories, productMap, nil
}

func (ps *productService) GetProductListByShop(shopID int64) ([]model.Category, map[int64][]model.ProductSimple, error) {
	// 1 根据店铺从product表查分类
	categories, categoryMap, err := ps.productRepo.GetProductCategoryByShop(shopID)
	if err != nil {
		return nil, nil, err
	}
	// 2.根据店铺获取所有商品
	products, err := ps.productRepo.GetAllProductByShop(shopID)
	if err != nil {
		return nil, nil, err
	}
	// 3.按分类分组
	productMap := make(map[int64][]model.ProductSimple)
	for _, p := range products {
		sp := model.ProductSimple{
			ID:           p.ID,
			Name:         p.Name,
			SellerPrice:  p.SellerPrice,
			CategoryID:   p.CategoryId,
			CategoryName: categoryMap[p.CategoryId],
			Image:        p.Image,
			Unit:         p.Unit,
			Remark:       p.Remark,
		}
		productMap[p.CategoryId] = append(productMap[p.CategoryId], sp)
	}
	return categories, productMap, nil
}

func (ps *productService) GetAllCategories() ([]model.Category, error) {
	return ps.productRepo.GetAllCategories()
}

func (ps *productService) GetCategoriesByShop(shopID int64) ([]model.Category, error) {
	return ps.productRepo.GetCategoriesByShop(shopID)
}

func (ps *productService) GetAdminProductList(page, pageSize int, shopID int64) ([]model.Product, int64, error) {
	offset := (page - 1) * pageSize
	if shopID > 0 {
		return ps.productRepo.GetListByShop(offset, pageSize, shopID)
	}
	return ps.productRepo.GetList(offset, pageSize)
}

func (ps *productService) AddProduct(p *model.Product) error {
	return ps.productRepo.Create(p)
}

func (ps *productService) UpdateProduct(p *model.Product) error {
	return ps.productRepo.Update(p)
}

func (ps *productService) GetProductByID(id int64) (*model.Product, error) {
	return ps.productRepo.GetByID(id)
}

func (ps *productService) GetProductByIDAndShop(id int64, shopID int64) (*model.Product, error) {
	return ps.productRepo.GetByIDAndShop(id, shopID)
}

func (ps *productService) DeleteProduct(id int64) error {
	return ps.productRepo.Delete(id)
}

// 分类管理方法实现
func (ps *productService) AddCategory(category *model.Category) error {
	return ps.productRepo.CreateCategory(category)
}

func (ps *productService) UpdateCategory(category *model.Category) error {
	return ps.productRepo.UpdateCategory(category)
}

func (ps *productService) DeleteCategory(id int64) error {
	return ps.productRepo.DeleteCategory(id)
}

func (ps *productService) GetCategoryByID(id int64) (*model.Category, error) {
	return ps.productRepo.GetCategoryByID(id)
}

// CheckProductNameExists 检查商品名称是否已存在
func (ps *productService) CheckProductNameExists(name string, excludeID ...int64) (bool, error) {
	return ps.productRepo.CheckNameExists(name, excludeID...)
}
