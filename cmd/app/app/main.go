package main

import (
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/cmd/app/app/srv"
	"github.com/xxl6097/go-service/cmd/app/app/wx"
	"github.com/xxl6097/go-service/pkg"
	"github.com/xxl6097/go-service/pkg/gs"
	"runtime"
	"strings"
)

func IsMacOs() bool {
	if strings.Compare(runtime.GOOS, "darwin") == 0 {
		return true
	}
	return false
}
func init() {
	if IsMacOs() {
		pkg.AppVersion = "v0.0.3"
		pkg.BinName = "aatest_v0.0.20_darwin_arm64"
		fmt.Println("Hello World...1")
	}
}

//go:generate goversioninfo -icon=resource/icon.ico -manifest=resource/goversioninfo.exe.manifest
func main() {
	appID := "wxbe2c2961b236427f"                   // 替换为你的微信公众号或小程序的 AppID
	appSecret := "667fc391b1ca8f4c58d1b5f224356ad5" // 替换为你的 AppSecret
	wx.Api().Load(appID, appSecret)
	s := srv.Service{}
	if IsMacOs() {
		srv.Server(9091, &s)
		return
	}
	err := gs.Run(&s)
	glog.Debug("程序结束", err)

}
