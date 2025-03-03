package gore

import (
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"os"
)

type IGService interface {
	Restart() error
	Upgrade(string) error
	Uninstall() error
}
type GService interface {
	OnInit() *service.Config
	OnVersion() string
	OnRun(IGService) error
}

type BaseService interface {
	GService
	GetAny() any
}

type coreService struct {
	svr GService
}

func (c coreService) Start(s service.Service) error {
	defer glog.Flush()
	status, err := s.Status()
	glog.Printf("启动服务【%s】\r\n", s.String())
	glog.Println("StatusUnknown=0；StatusRunning=1；StatusStopped=2；status", status, err)
	glog.Println("Platform", s.Platform())
	go c.svr.OnRun(NewGoreService(s))
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

func NewCoreService(svr GService) service.Shutdowner {
	return &coreService{
		svr: svr,
	}
}
