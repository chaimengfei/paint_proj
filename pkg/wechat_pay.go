package pkg

import (
	"context"
	"crypto/rsa"
	"fmt"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
)

// InitWechatPayClient 初始化微信支付客户端（v0.2.1 新版）
func InitWechatPayClient(mchID, mchSerialNo, apiV3Key string, privateKey *rsa.PrivateKey) (*core.Client, error) {
	client, err := core.NewClient(
		context.Background(),
		option.WithMerchantCredential(mchID, mchSerialNo, privateKey),
		option.WithWechatPayAutoAuthCipher(mchID, mchSerialNo, privateKey, apiV3Key), // 自动下载平台证书 + 自动验签
	)
	if err != nil {
		return nil, fmt.Errorf("初始化微信支付客户端失败: %w", err)
	}
	return client, nil
}
