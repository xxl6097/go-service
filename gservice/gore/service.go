package gore

import (
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"os"
)

type IService interface {
	OnVersion() string
	OnConfig() *service.Config
	OnInstall(string) (bool, []string)
	OnRun(Install) error
}

type Install interface {
	Restart() error
	Upgrade(upgradeBinPath string) error
}

type coreService struct {
	svr IService
	ins Install
}

func (c coreService) Start(s service.Service) error {
	defer glog.Flush()
	status, err := s.Status()
	glog.Printf("启动服务【%s】\r\n", s.String())
	glog.Println("StatusUnknown=0；StatusRunning=1；StatusStopped=2；status", status, err)
	glog.Println("Platform", s.Platform())
	go c.svr.OnRun(c.ins)
	return err
}

func (c coreService) Stop(s service.Service) error {
	defer glog.Flush()
	glog.Println("停止服务")

	if service.Interactive() {
		glog.Println("停止deamon")
		os.Exit(0)
	}
	return nil
}

func (c coreService) Shutdown(s service.Service) error {
	defer glog.Flush()
	status, err := s.Status()
	glog.Println("Shutdown")
	glog.Println("Status", status, err)
	glog.Println("Platform", s.Platform())
	glog.Println("String", s.String())
	return nil
}

func NewCoreService(svr IService, ins Install) service.Shutdowner {
	return &coreService{
		svr: svr,
		ins: ins,
	}
}
