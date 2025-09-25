package controller

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/pkg"
	"cmf/paint_proj/service"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
	shopService service.ShopService
}

func NewUserController(s service.UserService, shopService service.ShopService) *UserController {
	return &UserController{userService: s, shopService: shopService}
}

// Login 登录接口：code换openid，自动注册，返回token
func (uc *UserController) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.BindJSON(&req); err != nil || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误,缺少code"})
		return
	}

	// 获取或创建用户
	userId, token, err := uc.userService.LoginHandler(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败:" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id": userId,
		"token":   token,
	})

	// TODO 柴梦妃 临时测试，token也用入参
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"user_id": 123,
			"token":   req.Code,
		},
	})
}

func (uc *UserController) UpdateUserInfo(c *gin.Context) {
	var req model.UpdateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "手机号或昵称不能为空"})
		return
	}

	userID := c.GetInt64("user_id")
	err := uc.userService.UpdateUserInfo(context.Background(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新信息失败:" + err.Error()})
		return
	}
	c.JSON(200, gin.H{"msg": "更新信息成功"})
}

// AdminAddUser 后台添加用户
func (uc *UserController) AdminAddUser(c *gin.Context) {
	var req model.AdminUserAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证店铺权限
	validShopID, isValid := pkg.ValidateShopPermission(c, req.ShopID)
	if !isValid {
		return
	}
	req.ShopID = validShopID

	user, err := uc.userService.CreateUserByAdmin(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "添加用户失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "添加用户成功",
		"data":    user,
	})
}

// AdminGetUserList 后台获取用户列表
func (uc *UserController) AdminGetUserList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	keyword := c.Query("keyword")
	shopIDStr := c.Query("shop_id")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// 解析shop_id参数
	var shopID int64
	if shopIDStr != "" {
		shopID, err = strconv.ParseInt(shopIDStr, 10, 64)
		if err != nil {
			shopID = 0
		}
	}

	// 验证店铺权限
	validShopID, isValid := pkg.ValidateShopPermission(c, shopID)
	if !isValid {
		return
	}
	shopID = validShopID

	// 根据shopID决定调用哪个方法
	var users []*model.User
	var total int64
	if shopID > 0 {
		users, total, err = uc.userService.GetUserListByShop(page, pageSize, keyword, shopID)
	} else {
		users, total, err = uc.userService.GetUserList(page, pageSize, keyword)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取用户列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      users,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// AdminGetUserByID 后台根据ID获取用户
func (uc *UserController) AdminGetUserByID(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "用户ID格式错误"})
		return
	}

	// 1. 先查询用户信息
	user, err := uc.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取用户信息失败: " + err.Error()})
		return
	}

	// 2. 验证店铺权限
	operatorShopID := c.GetInt64("shop_id")
	isRoot := c.GetBool("is_root")

	if !isRoot && user.ShopID != operatorShopID {
		c.JSON(http.StatusForbidden, gin.H{"code": -1, "message": "无权限查看该用户信息"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": user,
	})
}

// AdminEditUser 后台编辑用户
func (uc *UserController) AdminEditUser(c *gin.Context) {
	var req model.AdminUserEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证店铺权限
	validShopID, isValid := pkg.ValidateShopPermission(c, req.ShopID)
	if !isValid {
		return
	}
	req.ShopID = validShopID

	err := uc.userService.UpdateUserByAdmin(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "更新用户失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新用户成功",
	})
}

// AdminDeleteUser 后台删除用户
func (uc *UserController) AdminDeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "用户ID格式错误"})
		return
	}

	// 1. 先查询用户信息
	user, err := uc.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "获取用户信息失败: " + err.Error()})
		return
	}

	// 2. 验证店铺权限
	operatorShopID := c.GetInt64("shop_id")
	isRoot := c.GetBool("is_root")

	if !isRoot && user.ShopID != operatorShopID {
		c.JSON(http.StatusForbidden, gin.H{"code": -1, "message": "无权限删除该用户"})
		return
	}

	// 3. 删除用户
	err = uc.userService.DeleteUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "删除用户失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除用户成功",
	})
}

// WechatBindMobile 小程序绑定手机号
func (uc *UserController) WechatBindMobile(c *gin.Context) {
	var req model.WechatBindMobileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "参数错误: " + err.Error()})
		return
	}

	userID := c.GetInt64("user_id")

	// 调用 service 层处理业务逻辑
	user, err := uc.userService.WechatBindMobile(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "绑定失败: " + err.Error()})
		return
	}

	// 根据返回的用户信息判断绑定结果
	if user.ID == userID {
		// 更新当前用户手机号
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "绑定成功",
			"data":    user,
		})
	} else {
		// 绑定到现有用户
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "绑定成功，已关联到现有用户",
			"data":    user,
		})
	}
}
