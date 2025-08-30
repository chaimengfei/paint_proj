package controller

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AddressController struct {
	addressService service.AddressService
}

func NewAddressController(s service.AddressService) *AddressController {
	return &AddressController{addressService: s}
}

func (ac *AddressController) GetAddressList(c *gin.Context) {
	userID := c.GetInt64("user_id")

	list, err := ac.addressService.GetAddressList(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取地址失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

// CreateAddress 创建地址
func (ac *AddressController) CreateAddress(c *gin.Context) {
	userID := c.GetInt64("user_id") // 从认证中获取用户ID
	var req model.CreateAddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}
	err := ac.addressService.CreateAddress(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "创建地址失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "创建地址成功"})
}

// SetDefultAddress 设置默认地址
func (ac *AddressController) SetDefultAddress(c *gin.Context) {
	userID := c.GetInt64("user_id") // 从认证中获取用户ID
	addressID, _ := strconv.ParseInt(c.PostForm("id"), 10, 64)

	err := ac.addressService.SetDefaultAddress(userID, addressID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "更新默认地址失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新默认地址成功"})
}

// UpdateAddress 更新购物车商品数量
func (ac *AddressController) UpdateAddress(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req model.UpdateAddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}
	err := ac.addressService.UpdateAddress(userID, req.Data.AddressID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "更新地址失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新地址成功"})
}

// DeleteAddress 删除地址
func (ac *AddressController) DeleteAddress(c *gin.Context) {
	userID := c.GetInt64("user_id")
	idStr := c.Param("id")
	addressID, _ := strconv.ParseInt(idStr, 10, 64)

	err := ac.addressService.DeleteAddress(userID, addressID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// GetAdminAddressList 获取地址列表（admin）
func (ac *AddressController) GetAdminAddressList(c *gin.Context) {
	userIdStr := c.Query("user_id")
	userName := c.Query("user_name")

	var userId int64
	if userIdStr != "" {
		userId, _ = strconv.ParseInt(userIdStr, 10, 64)
	}

	list, err := ac.addressService.GetAdminAddressList(userId, userName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取地址列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

// CreateAdminAddress 创建地址（admin）
func (ac *AddressController) CreateAdminAddress(c *gin.Context) {
	userIdStr := c.Query("user_id")

	var userId int64
	if userIdStr != "" {
		userId, _ = strconv.ParseInt(userIdStr, 10, 64)
	}

	var req model.CreateAddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	// 转换为普通创建地址请求，复用现有service方法
	createReq := model.CreateAddressReq{
		Data: model.AddressInfo{
			RecipientName:  req.Data.RecipientName,
			RecipientPhone: req.Data.RecipientPhone,
			Province:       req.Data.Province,
			City:           req.Data.City,
			District:       req.Data.District,
			Detail:         req.Data.Detail,
			IsDefault:      req.Data.IsDefault,
		},
	}

	err := ac.addressService.CreateAdminAddress(userId, &createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "创建地址失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "创建地址成功"})
}

// UpdateAdminAddress 更新地址（admin）
func (ac *AddressController) UpdateAdminAddress(c *gin.Context) {
	idStr := c.Param("id")
	addressID, _ := strconv.ParseInt(idStr, 10, 64)

	var req model.UpdateAddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误"})
		return
	}

	// 转换为普通更新地址请求，复用现有service方法
	updateReq := model.UpdateAddressReq{
		Data: model.AddressInfo{
			AddressID:      addressID,
			RecipientName:  req.Data.RecipientName,
			RecipientPhone: req.Data.RecipientPhone,
			Province:       req.Data.Province,
			City:           req.Data.City,
			District:       req.Data.District,
			Detail:         req.Data.Detail,
			IsDefault:      req.Data.IsDefault,
		},
	}

	err := ac.addressService.UpdateAdminAddress(addressID, &updateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "更新地址失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新地址成功"})
}

// DeleteAdminAddress 删除地址（admin）
func (ac *AddressController) DeleteAdminAddress(c *gin.Context) {
	idStr := c.Param("id")
	addressID, _ := strconv.ParseInt(idStr, 10, 64)

	// 从请求体中获取用户ID，或者从查询参数获取
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "缺少用户ID参数"})
		return
	}
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	err := ac.addressService.DeleteAddress(userID, addressID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "删除地址失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除地址成功"})
}
