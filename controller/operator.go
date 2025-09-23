package controller

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OperatorController struct {
	operatorService service.OperatorService
}

func NewOperatorController(operatorService service.OperatorService) *OperatorController {
	return &OperatorController{
		operatorService: operatorService,
	}
}

// AdminLogin 后台管理员登录
// @Summary 后台管理员登录
// @Description 后台管理员登录接口
// @Tags 后台管理
// @Accept json
// @Produce json
// @Param request body model.AdminLoginRequest true "登录请求"
// @Success 200 {object} model.AdminLoginResponse
// @Router /admin/operator/login [post]
func (oc *OperatorController) AdminLogin(c *gin.Context) {
	var req model.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	response, err := oc.operatorService.AdminLogin(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "登录失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "登录成功",
		"data":    response,
	})
}

// GetOperatorList 获取管理员列表
// @Summary 获取管理员列表
// @Description 获取管理员列表（需要超级管理员权限）
// @Tags 后台管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} gin.H
// @Router /admin/operator/list [get]
func (oc *OperatorController) GetOperatorList(c *gin.Context) {
	// 检查是否为超级管理员
	isRoot := c.GetBool("is_root")
	if !isRoot {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    -1,
			"message": "权限不足，需要超级管理员权限",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")

	operators, total, err := oc.operatorService.GetOperatorList(page, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "获取管理员列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      operators,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
		"message": "获取管理员列表成功",
	})
}

// GetOperatorByID 根据ID获取管理员信息
// @Summary 根据ID获取管理员信息
// @Description 根据ID获取管理员信息
// @Tags 后台管理
// @Accept json
// @Produce json
// @Param id path int true "管理员ID"
// @Success 200 {object} gin.H
// @Router /admin/operator/{id} [get]
func (oc *OperatorController) GetOperatorByID(c *gin.Context) {
	operatorIDStr := c.Param("id")
	operatorID, err := strconv.ParseInt(operatorIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "管理员ID格式错误",
		})
		return
	}

	operator, err := oc.operatorService.GetOperatorByID(operatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "获取管理员信息失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    operator,
		"message": "获取管理员信息成功",
	})
}
