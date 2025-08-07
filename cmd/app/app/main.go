package main

import (
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/cmd/app/app/srv"
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
		fmt.Println("Hello World")
	}
}

//go:generate goversioninfo -icon=resource/icon.ico -manifest=resource/goversioninfo.exe.manifest
func main() {
	s := srv.Service{}
	if IsMacOs() {
		srv.Server(9091, &s)
		return
	}
	err := gs.Run(&s)
	glog.Debug("程序结束", err)

}
