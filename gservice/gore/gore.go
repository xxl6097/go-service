package gore

import (
	"context"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/utils"
	"os"
)

type IGService interface {
	Restart() error
	RunCmd(...string) error
	Upgrade(context.Context, string, ...string) error
	Uninstall() error
}
type GService interface {
	OnInit() *service.Config
	OnVersion() string
	OnRun(IGService) error
}

type BaseService interface {
	GService
	GetAny(string) any
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
	utils.DeleteAll(utils.GetUpgradeDir(), "升级文件夹")
	go c.svr.OnRun(NewGoreService(s))
	return err
}

func (c coreService) Stop(s service.Service) error {
	defer glog.Flush()
	status, err := s.Status()
	Stop(c.svr, s)
	ok := service.Interactive()
	glog.Println("停止服务", ok, s.String(), s.Platform(), status, err)
	if ok {
		glog.Println("停止deamon")
		os.Exit(0)
	}
	return nil
}

func (c coreService) Shutdown(s service.Service) error {
	defer glog.Flush()
	status, err := s.Status()
	ShutDown(c.svr, s)
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
