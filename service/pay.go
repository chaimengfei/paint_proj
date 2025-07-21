package service

import (
	"cmf/paint_proj/configs"
	"cmf/paint_proj/model"
	"cmf/paint_proj/pkg"
	"cmf/paint_proj/repository"
	"context"
	"errors"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"time"
)

type PayService interface {
	PayOrder(ctx context.Context, openid, orderNo string, total model.Amount) (*jsapi.PrepayWithRequestPaymentResponse, error)
	PaidCallback(ctx context.Context, req *model.PaidCallbackData) error // 订单支付成功回调
}

type payService struct {
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewPayService(or repository.OrderRepository, cr repository.CartRepository, pr repository.ProductRepository) PayService {
	return &payService{
		orderRepo:   or,
		cartRepo:    cr,
		productRepo: pr,
	}
}

func (ps *payService) PayOrder(ctx context.Context, openid, orderNo string, total model.Amount) (*jsapi.PrepayWithRequestPaymentResponse, error) {
	// 1. 获取订单
	order, err := ps.orderRepo.GetOrderByOrderNo(orderNo)
	if err != nil {
		return nil, err
	}
	// 2. 检查订单状态是否可以支付
	if order.OrderStatus != model.OrderStatusPendingPayment {
		return nil, errors.New("订单状态异常 无法支付")
	}

	privateKey, _ := utils.LoadPrivateKeyWithPath("apiclient_key.pem")

	client, err := pkg.InitWechatPayClient(pkg.MchID, pkg.SerialNo, pkg.APIv3Key, privateKey)
	if err != nil {
		return nil, err
	}
	jsapiService := jsapi.JsapiApiService{Client: client}
	//float64Fen := total * float64(100)
	//totalFen := int32(float64Fen)
	resp, _, _ := jsapiService.PrepayWithRequestPayment(context.Background(), jsapi.PrepayRequest{
		Appid:       core.String(configs.Cfg.Wechat.AppID),
		Mchid:       core.String(pkg.MchID),
		Description: core.String("订单支付测试 " + time.Now().String()),
		OutTradeNo:  core.String(orderNo),
		NotifyUrl:   core.String("https://your-domain.com/api/pay/notify"), // TODO 待填充
		Amount: &jsapi.Amount{
			Total:    core.Int32(int32(total)),
			Currency: core.String("CNY"),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(openid),
		},
	})
	return resp, nil

}
func (ps *payService) PaidCallback(ctx context.Context, req *model.PaidCallbackData) error {
	panic("implement me")
}
