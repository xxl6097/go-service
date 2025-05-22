package service

import (
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore"
	"github.com/xxl6097/go-service/gservice/utils"
	"github.com/xxl6097/go-service/pkg"
	"time"
)

type Service struct {
	service   gore.IGService
	timestamp string
}

func (t *Service) GetAny(binDir string) any {
	return t.menu()
}

func (t *Service) OnInit() *service.Config {
	return &service.Config{
		Name:        pkg.AppName,
		DisplayName: fmt.Sprintf("A AAATest1 Service %s", pkg.AppVersion),
		Description: "A Golang AAATest1 Service..",
	}
}

func (t *Service) OnVersion() string {
	pkg.Version()
	return pkg.AppVersion
}

func (t *Service) OnRun(service gore.IGService) error {
	t.service = service
	//glog.SetLogFile("./logs", "app.log")
	go Serve(t)
	for {
		t.timestamp = time.Now().Format(time.RFC3339)
		glog.Println("run", t.timestamp)
		time.Sleep(time.Second * 1)
	}
}

func (t *Service) menu() any {
	appName := utils.InputString(fmt.Sprintf("请输入应用名称："))
	return []string{appName}
}
