package internal

import (
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"os"
)

func (c *CoreService) Start(s service.Service) error {
	status, err := s.Status()
	defer glog.Flush()
	glog.Printf("启动服务【%s】\r\n", s.String())
	glog.Println("StatusUnknown=0；StatusRunning=1；StatusStopped=2；status", status, err)
	glog.Println("Platform", s.Platform())
	go func() {
		e := c.iService.OnRun(c)
		if e != nil {
			glog.Error("运行失败", e)
		} else {
			glog.Debug("运行成功")
		}
	}()
	return nil
}

func (c *CoreService) Stop(s service.Service) error {
	defer glog.Flush()
	status, err := s.Status()
	ok := service.Interactive()
	//注意，这个地方在非windows下不行！
	//c.clear()
	glog.Println("停止服务", ok, s.String(), s.Platform(), status, err)
	if ok {
		glog.Println("停止deamon")
		os.Exit(0)
	}
	return nil
}

func (c *CoreService) Shutdown(s service.Service) error {
	defer glog.Flush()
	status, err := s.Status()
	glog.Println("Shutdown")
	glog.Println("Status", status, err)
	glog.Println("Platform", s.Platform())
	glog.Println("String", s.String())
	return nil
}
