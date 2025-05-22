package main

import (
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/cmd/app/test1/test1"
	"github.com/xxl6097/go-service/gservice"
)

func main() {
	err := gservice.Run(&test1.Test1{})
	glog.Println("main", err)
	//test1.ServeTesting()
}
