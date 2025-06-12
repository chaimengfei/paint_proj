package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/repository"
)

type ProductService interface {
	GetProductList() ([]model.Category, map[int64][]model.ProductSimple, error)
}

type productService struct {
	productRepo repository.ProductRepository
}

func NewProductService(pr repository.ProductRepository) ProductService {
	return &productService{
		productRepo: pr,
	}
}
func (p productService) GetProductList() ([]model.Category, map[int64][]model.ProductSimple, error) {
	// 1 从product表查分类
	categories, categoryMap, err := p.productRepo.GetProductCategory()
	if err != nil {
		return nil, nil, err
	}
	// 2.获取所有商品
	products, err := p.productRepo.GetAllProduct()
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
