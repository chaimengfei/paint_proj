package controller

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/pkg"
	"cmf/paint_proj/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ProductController struct {
	productService service.ProductService
}

func NewProductController(s service.ProductService) *ProductController {
	return &ProductController{productService: s}
}

// GetProductList 获取商品列表
// @Summary 获取商品分类及列表
// @Description 获取所有分类及对应的商品列表
// @Tags 商品
// @Accept json
// @Produce json
// @Success 200 {object} model.ProductListResponse
// @Router /api/products [get]
func (pc *ProductController) GetProductList(c *gin.Context) {
	categories, productMap, err := pc.productService.GetProductList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取商品列表失败"})
		return
	}
	// 返回结果
	response := model.ProductListResponse{
		Categories: categories,
		Products:   productMap,
	}

	c.JSON(http.StatusOK, response) //
}

func (pc *ProductController) UploadImageForAdmin(c *gin.Context) {
	fileURL, err := pkg.UploadImage(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": fileURL})
}
