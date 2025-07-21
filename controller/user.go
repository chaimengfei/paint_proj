package controller

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/service"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(s service.UserService) *UserController {
	return &UserController{userService: s}
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
