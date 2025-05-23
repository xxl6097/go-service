package main

import (
	_ "github.com/xxl6097/go-service/assets/we"
	"github.com/xxl6097/go-service/cmd/app/app/service"
)

func main() {
	svr := service.Service{}
	//service.Serve(&svr)
	//err := gservice.Run(&svr)
	//glog.Println(pkg.AppName, err, glog.GetCrossPlatformDataDir())
	service.Server(&svr)
}
