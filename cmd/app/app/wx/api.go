package wx

import (
	"fmt"
	"sync"
	"time"

	"github.com/xxl6097/glog/pkg/z"
	"go.uber.org/zap"
)

var (
	instance *WxApi
	once     sync.Once
)

type WxApi struct {
	appID           string
	appSecret       string
	cachedToken     string
	tokenExpireTime time.Time
}

func Api() *WxApi {
	once.Do(func() {
		instance = &WxApi{
			tokenExpireTime: time.Now().Add(time.Minute),
		}
	})
	return instance
}

func (this *WxApi) Load(appID, appSecret string) {
	this.appID = appID
	this.appSecret = appSecret
	token := this.GetToken()
	if token == "" {
		z.L().Fatal("load app access token err")
	} else {
		z.L().Info("token", zap.String("token", this.cachedToken))
	}
}

func (this *WxApi) GetToken() string {
	if this.cachedToken == "" || time.Now().After(this.tokenExpireTime) {
		// 缓存无效，重新获取
		tokenInfo, err := GetStableAccessToken(this.appID, this.appSecret, false)
		if err != nil {
			fmt.Printf("get stable access token err:%v\n", err)
			return ""
		}
		this.cachedToken = tokenInfo.AccessToken
		this.tokenExpireTime = time.Now().Add(time.Duration(tokenInfo.ExpiresIn-120) * time.Second) // 提前2分钟过期
	}
	return this.cachedToken
}
