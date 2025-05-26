package main

import (
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/cmd/app/app/srv"
	"github.com/xxl6097/go-service/pkg/gs"
	"os"
)

//go:generate goversioninfo -icon=resource/icon.ico -manifest=resource/goversioninfo.exe.manifest
func main() {
	s := srv.Service{}
	if len(os.Args) > 1 {
		if os.Args[1] == "test" {
			srv.Server(9090, &s)
			return
		}
	}
	err := gs.Run(&s)
	glog.Debug("程序结束", err)

}
