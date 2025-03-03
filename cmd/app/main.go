package main

import (
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/cmd/app/test1"
	"github.com/xxl6097/go-service/gservice"
	"github.com/xxl6097/go-service/pkg"
)

func main() {
	if pkg.AppName == "" {
		pkg.AppName = "acsvr"
	}
	err := gservice.Run(&test1.Test1{})
	glog.Println("main", err)
}
