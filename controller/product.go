package controller

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/pkg"
	"cmf/paint_proj/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
	c.JSON(http.StatusOK, response)
}

func (pc *ProductController) UploadImageForAdmin(c *gin.Context) {
	fileURL, err := pkg.UploadImage(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": fileURL})
}

// 管理员分页获取商品列表
func (pc *ProductController) GetAdminProductList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	products, total, err := pc.productService.GetAdminProductList(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      products,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// AddProduct 新增商品（后台）
func (pc *ProductController) AddProduct(c *gin.Context) {
	var req model.AddOrEditSimpleProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	// 转换为完整的Product结构体
	product := &model.Product{
		Name:          req.Name,
		CategoryId:    req.CategoryId,
		Image:         req.Image,
		SellerPrice:   req.SellerPrice,
		Cost:          req.Cost,
		ShippingCost:  req.ShippingCost,
		ProductCost:   req.ProductCost,
		Specification: req.Specification,
		Unit:          req.Unit,
		Remark:        req.Remark,
		IsOnShelf:     req.IsOnShelf,
		// TODO 设置默认值
		Stock: 0,
	}

	if err := pc.productService.AddProduct(product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "添加商品失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "添加成功"})
}

// EditProduct 编辑商品（后台）
func (pc *ProductController) EditProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "商品ID格式错误"})
		return
	}

	var req model.AddOrEditSimpleProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	// 转换为完整的Product结构体
	product := &model.Product{
		ID:            id,
		Name:          req.Name,
		CategoryId:    req.CategoryId,
		Image:         req.Image,
		SellerPrice:   req.SellerPrice,
		Cost:          req.Cost,
		ShippingCost:  req.ShippingCost,
		ProductCost:   req.ProductCost,
		Specification: req.Specification,
		Unit:          req.Unit,
		Remark:        req.Remark,
		IsOnShelf:     req.IsOnShelf,
	}

	if err = pc.productService.UpdateProduct(product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "编辑商品失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "编辑成功"})
}

// DeleteProduct 删除商品（后台）
func (pc *ProductController) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	if err := pc.productService.DeleteProduct(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// GetProductByID 根据ID获取商品信息（后台）
func (pc *ProductController) GetProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "商品ID格式错误"})
		return
	}

	product, err := pc.productService.GetProductByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取商品信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": product,
	})
}
