package wx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const WxHost = "https://api.weixin.qq.com"

// StableTokenRequest 请求参数结构体
type StableTokenRequest struct {
	GrantType    string `json:"grant_type"`
	AppID        string `json:"appid"`
	Secret       string `json:"secret"`
	ForceRefresh bool   `json:"force_refresh,omitempty"` // 是否强制刷新，普通模式可省略或设为 false
}

// StableTokenResponse 响应参数结构体 (根据官方文档[2](@ref))
type StableTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"` // 有效期，单位秒
	ErrCode     int    `json:"errcode,omitempty"`
	ErrMsg      string `json:"errmsg,omitempty"`
}

// GetStableAccessToken 获取稳定版接口调用凭据
func GetStableAccessToken(appID, appSecret string, forceRefresh bool) (*StableTokenResponse, error) {
	apiUrl := fmt.Sprintf("%s/cgi-bin/stable_token", WxHost)
	// 构造请求体
	reqData := StableTokenRequest{
		GrantType:    "client_credential",
		AppID:        appID,
		Secret:       appSecret,
		ForceRefresh: forceRefresh,
	}
	// 将结构体序列化为 JSON 字节切片
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("序列化请求参数失败: %v", err)
	}

	// 创建 HTTP POST 请求
	req, err := http.NewRequest("POST", apiUrl, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	// 解析 JSON 响应
	var tokenResp StableTokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, fmt.Errorf("解析JSON响应失败: %v", err)
	}

	// 检查微信接口返回的错误码
	if tokenResp.ErrCode != 0 {
		return nil, fmt.Errorf("微信接口错误[%d]: %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	return &tokenResp, nil
}
