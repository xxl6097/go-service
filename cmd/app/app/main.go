package main

import (
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/cmd/app/app/service"
	"github.com/xxl6097/go-service/gservice"
	"github.com/xxl6097/go-service/pkg"
)

func main() {
	svr := service.Service{}
	//service.Serve(&svr)
	err := gservice.Run(&svr)
	glog.Println(pkg.AppName, err, glog.GetCrossPlatformDataDir())
}
