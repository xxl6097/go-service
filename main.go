package main

import (
	"github.com/kardianos/service"
	"github.com/xxl6097/go-glog/glog"
	"github.com/xxl6097go-service/svr"
	"os"
	"time"
)

type Main struct{}

// Shutdown 服务结束回调
func (i *Main) Shutdown(s service.Service) error {
	defer glog.Flush()
	status, err := s.Status()
	glog.Println("Shutdown")
	glog.Println("Status", status, err)
	glog.Println("Platform", s.Platform())
	glog.Println("String", s.String())
	return nil
}

// Start 服务启动回调
func (i *Main) Start(s service.Service) error {
	defer glog.Flush()
	status, err := s.Status()
	glog.Println("启动服务")
	glog.Println("Status", status, err)
	glog.Println("Platform", s.Platform())
	glog.Println("String", s.String())
	go run()
	return nil
}

// Stop 服务停止回调
func (i *Main) Stop(s service.Service) error {
	defer glog.Flush()
	glog.Println("停止服务")

	if service.Interactive() {
		glog.Println("停止deamon")
		os.Exit(0)
	}
	return nil
}

func (i *Main) Config() *service.Config {
	return &service.Config{
		Name:        "AAATest1",
		DisplayName: "A AAATest1 Service",
		Description: "A Golang AAATest1 Service..",
	}
}

func (i *Main) Version() string {
	return "v0.0.1"
}

func run() {
	for {
		glog.Println("run", time.Now().Format("2006-01-02 15:04:05"))
		time.Sleep(time.Second * 5)
	}
}
func main() {
	glog.Println("hello")
	svr.Run(&Main{})
}
