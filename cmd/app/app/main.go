package main

import (
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/cmd/app/app/service"
	"github.com/xxl6097/go-service/gservice"
	"github.com/xxl6097/go-service/pkg"
	"os"
)

//go:generate goversioninfo -icon=resource/icon.ico -manifest=resource/goversioninfo.exe.manifest
func main() {
	svr := service.Service{}
	if len(os.Args) > 1 {
		if os.Args[1] == "test" {
			service.Server(9090, &svr)
			return
		}
	}
	err := gservice.Run(&svr)
	glog.Println(pkg.AppName, err, glog.GetCrossPlatformDataDir())
}
