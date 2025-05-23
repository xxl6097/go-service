package service

import (
	"fmt"
	"github.com/kardianos/service"
	_ "github.com/xxl6097/go-service/assets/we"
	"github.com/xxl6097/go-service/gservice/gore"
	"github.com/xxl6097/go-service/gservice/utils"
	"github.com/xxl6097/go-service/pkg"
	"time"
)

type Service struct {
	gs        gore.IGService
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
	t.gs = service
	//glog.SetLogFile("./logs", "app.log")
	go Server(t)
	for {
		t.timestamp = time.Now().Format(time.RFC3339)
		//glog.Println("run", t.timestamp)
		time.Sleep(time.Second * 1)
	}
}

func (t *Service) menu() any {
	appName := utils.InputStringEmpty(fmt.Sprintf("测试输入："), "册书数据")
	return []string{appName}
}
