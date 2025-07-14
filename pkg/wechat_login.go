package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type WxLoginResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// 是用来首次登录小程序时，通过 code 换取 openid 的一步 <这一步只需要在用户第一次进入时调用即可>
//之后应该缓存用户的 openid（或你生成的 userID），在用户每次发请求时带上，后端只需要解析 token 并还原 userID
func GetOpenIDByCode(code string) (string, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		AppID, AppSecret, code,
	)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("请求微信接口失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result WxLoginResp
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析微信返回数据失败: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("微信接口错误: %d - %s", result.ErrCode, result.ErrMsg)
	}
	return result.OpenID, nil
}
