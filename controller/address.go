package controller

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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
