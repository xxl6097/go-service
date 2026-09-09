package main

import (
	"github.com/xxl6097/glog/pkg/z"
	"github.com/xxl6097/go-service/cmd/app/app/srv"
	"github.com/xxl6097/go-service/cmd/app/test5/svv"
	"github.com/xxl6097/go-service/pkg/gs"
	"github.com/xxl6097/go-service/pkg/utils"
	"go.uber.org/zap"
)

func main() {
	s := svv.Service{}
	if utils.IsMacOs() {
		srv.Server(9091, &s)
		return
	}
	err := gs.Run(&s)
	z.L().Debug("程序结束", zap.Error(err))

}
