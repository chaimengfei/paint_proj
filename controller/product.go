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
	userService    service.UserService
}

func NewProductController(s service.ProductService, us service.UserService) *ProductController {
	return &ProductController{productService: s, userService: us}
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
	// 获取用户ID和店铺ID（从JWT token中解析）
	userID := c.GetInt64("user_id")
	shopID := c.GetInt64("shop_id")

	var categories []model.Category
	var productMap map[int64][]model.ProductSimple
	var err error

	if userID > 0 && shopID > 0 {
		// 根据用户店铺获取商品列表
		categories, productMap, err = pc.productService.GetProductListByShop(shopID)
	} else {
		// 如果没有用户信息，返回所有商品（兼容性）
		categories, productMap, err = pc.productService.GetProductList()
	}

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
	shopIDStr := c.DefaultQuery("shop_id", "0")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil {
		shopID = 0
	}

	// 验证店铺权限
	validShopID, isValid := pkg.ValidateShopPermission(c, shopID)
	if !isValid {
		return
	}
	shopID = validShopID

	products, total, err := pc.productService.GetAdminProductList(page, pageSize, shopID)
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

	// 验证店铺权限
	shopID, isValid := pkg.ValidateShopPermission(c, req.ShopID)
	if !isValid {
		return
	}

	if shopID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "缺少店铺信息"})
		return
	}

	// 检查商品名称是否已存在（在同一店铺内）
	exists, err := pc.productService.CheckProductNameExists(req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "检查商品名称失败: " + err.Error()})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": pkg.ErrProductNameExists})
		return
	}

	// 转换为完整的Product结构体
	product := &model.Product{
		Name:          req.Name,
		CategoryId:    req.CategoryId,
		Image:         req.Image,
		SellerPrice:   req.SellerPrice,
		Specification: req.Specification,
		Unit:          req.Unit,
		Remark:        req.Remark,
		IsOnShelf:     req.IsOnShelf,
		ShopID:        shopID,
		// 成本相关字段由入库单自动更新，初始化为0
		Cost:         0,
		ShippingCost: 0,
		ProductCost:  0,
		Stock:        0,
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

	var req model.EditProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证店铺权限
	shopID, isValid := pkg.ValidateShopPermission(c, req.ShopID)
	if !isValid {
		return
	}

	if shopID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "缺少店铺信息"})
		return
	}

	// 检查商品名称是否已存在（排除当前编辑的商品）
	exists, err := pc.productService.CheckProductNameExists(req.Name, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "检查商品名称失败: " + err.Error()})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": pkg.ErrProductNameExists})
		return
	}

	// 转换为完整的Product结构体
	product := &model.Product{
		ID:            id,
		Name:          req.Name,
		Image:         req.Image,
		SellerPrice:   req.SellerPrice,
		Specification: req.Specification,
		Remark:        req.Remark,
		IsOnShelf:     req.IsOnShelf,
		ShopID:        shopID,
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
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "商品ID格式错误"})
		return
	}

	// 1. 先查询商品信息
	product, err := pc.productService.GetProductByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取商品信息失败: " + err.Error()})
		return
	}

	// 2. 验证店铺权限
	operatorShopID := c.GetInt64("shop_id")
	isRoot := c.GetBool("is_root")

	if !isRoot && product.ShopID != operatorShopID {
		c.JSON(http.StatusForbidden, gin.H{"code": -1, "message": "无权限删除该商品"})
		return
	}

	// 3. 执行删除操作
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

	// 1. 先查询商品信息
	product, err := pc.productService.GetProductByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取商品信息失败: " + err.Error()})
		return
	}

	// 2. 验证店铺权限
	operatorShopID := c.GetInt64("shop_id")
	isRoot := c.GetBool("is_root")

	if !isRoot && product.ShopID != operatorShopID {
		c.JSON(http.StatusForbidden, gin.H{"code": -1, "message": "无权限查看该商品"})
		return
	}

	// 3. 返回商品信息
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": product,
	})
}

// GetCategories 获取所有分类（后台）
func (pc *ProductController) GetCategories(c *gin.Context) {
	categories, err := pc.productService.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取分类失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": categories,
	})
}

// AddCategory 新增分类（后台）
func (pc *ProductController) AddCategory(c *gin.Context) {
	var req model.AddCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	category := &model.Category{
		Name:      req.Name,
		SortOrder: req.SortOrder,
	}

	if err := pc.productService.AddCategory(category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "添加分类失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "添加分类成功"})
}

// EditCategory 编辑分类（后台）
func (pc *ProductController) EditCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "分类ID格式错误"})
		return
	}

	var req model.EditCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	category := &model.Category{
		ID:        id,
		Name:      req.Name,
		SortOrder: req.SortOrder,
	}

	if err = pc.productService.UpdateCategory(category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "编辑分类失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "编辑分类成功"})
}

// DeleteCategory 删除分类（后台）
func (pc *ProductController) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "分类ID格式错误"})
		return
	}

	if err := pc.productService.DeleteCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "删除分类失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除分类成功"})
}
