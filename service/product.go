package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/repository"
)

type ProductService interface {
	GetProductList() ([]model.Category, map[int64][]model.ProductSimple, error)

	GetAdminProductList(page, pageSize int) ([]model.Product, int64, error)
	AddProduct(p *model.Product) error
	UpdateProduct(p *model.Product) error
	DeleteProduct(id int64) error
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

func (ps *productService) GetAdminProductList(page, pageSize int) ([]model.Product, int64, error) {
	offset := (page - 1) * pageSize
	return ps.productRepo.GetList(offset, pageSize)
}

func (ps *productService) AddProduct(p *model.Product) error {
	return ps.productRepo.Create(p)
}

func (ps *productService) UpdateProduct(p *model.Product) error {
	return ps.productRepo.Update(p)
}

func (ps *productService) DeleteProduct(id int64) error {
	return ps.productRepo.Delete(id)
}
