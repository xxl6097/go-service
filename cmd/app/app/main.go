package main

import (
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/cmd/app/app/service"
	"github.com/xxl6097/go-service/gservice"
	"github.com/xxl6097/go-service/pkg"
	"os"
)

func main() {
	svr := service.Service{}
	if len(os.Args) > 1 {
		if os.Args[1] == "test" {
			service.Server(&svr)
			return
		}
	}
	err := gservice.Run(&svr)
	glog.Println(pkg.AppName, err, glog.GetCrossPlatformDataDir())
}
