package main

import (
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/cmd/app/test"
	"github.com/xxl6097/go-service/gservice"
)

func main() {
	err := gservice.Run(test.Test{})
	glog.Println("main", err)
}
