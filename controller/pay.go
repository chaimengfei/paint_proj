package controller

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/service"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PayController struct {
	payService service.PayService
}

func NewPayController(s service.PayService) *PayController {
	return &PayController{payService: s}
}

// PaymentData 获取支付数据
func (pc *PayController) PaymentData(c *gin.Context) {
	var req struct {
		Code    string  `json:"code"`     // ，前端通过 wx.login() 获取临时 code，后端就可以使用这个 code 请求微信服务器获取 openid 和 session_key
		Total   float64 `json:"total"`    // 单位：元
		OrderNo string  `json:"order_no"` // 订单号
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "PaymentData 参数错误"})
		return
	}
	//userID := c.GetInt64("user_id")     // 从认证中获取用户ID
	openid := "" // TODO getOpenIDByCode(req.Code) // 伪函数，请替换为真实获取 openid
	orderNo := req.OrderNo

	resp, err := pc.payService.PayOrder(context.Background(), orderNo, openid, req.Total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": "<UNK>"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// PaymentCallback 支付回调
func (pc *PayController) PaymentCallback(c *gin.Context) {
	// 解析回调参数
	var callbackReq struct {
		OrderNo       string  `json:"order_no"`
		PaymentNo     string  `json:"payment_no"`
		PaymentType   int     `json:"payment_type"`
		PaymentTime   int64   `json:"payment_time"`
		PaymentAmount float64 `json:"payment_amount"`
	}

	if err := c.ShouldBindJSON(&callbackReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 调用服务处理回调
	req := &model.OrderPaidCallbackRequest{
		OrderNo:       callbackReq.OrderNo,
		PaymentNo:     callbackReq.PaymentNo,
		PaymentType:   int32(callbackReq.PaymentType),
		PaymentTime:   callbackReq.PaymentTime,
		PaymentAmount: callbackReq.PaymentAmount,
	}

	if err := pc.payService.PaidCallback(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "处理回调失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}
