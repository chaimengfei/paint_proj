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
	dbData := model.Address{
		UserId:         userID,
		RecipientName:  req.RecipientName,
		RecipientPhone: req.RecipientPhone,
		Province:       req.Province,
		City:           req.City,
		District:       req.District,
		Detail:         req.Detail,
	}
	// 如果设置为默认，则取消用户其他地址的默认状态
	if req.IsDefault != nil {
		if *req.IsDefault {
			dbData.IsDefault = 1 // 设置默认
		} else {
			dbData.IsDefault = 0 // 取消默认
		}
	} else {
		dbData.IsDefault = 0 // 忽略默认地址设置
	}

	err := ac.addressService.CreateAddress(&dbData)
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
	dbData := map[string]interface{}{}
	if req.RecipientName != "" {
		dbData["recipient_name"] = req.RecipientName
	}
	if req.RecipientPhone != "" {
		dbData["recipient_phone"] = req.RecipientPhone
	}
	if req.Province != "" {
		dbData["province"] = req.Province
	}
	if req.City != "" {
		dbData["city"] = req.City
	}
	if req.District != "" {
		dbData["district"] = req.District
	}
	if req.Detail != "" {
		dbData["detail"] = req.Detail
	}
	// 如果设置为默认，则取消用户其他地址的默认状态
	if req.IsDefault != nil {
		if *req.IsDefault {
			dbData["is_default"] = 1 // 设置默认
		} else {
			dbData["is_default"] = 0 // 取消默认
		}
	} else {
		dbData["is_default"] = 0 // 忽略默认地址设置
	}
	err := ac.addressService.UpdateAddress(userID, req.AddressID, dbData)
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
