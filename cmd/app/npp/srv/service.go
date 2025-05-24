package srv

import (
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/go-service/gservice/utils"
	"github.com/xxl6097/go-service/pkg"
	"github.com/xxl6097/go-service/pkg/gs/igs"
	"time"
)

type Service struct {
	timestamp string
	port      int
	gs        igs.Service
}

type Config struct {
	//ukey.KeyBuffer
	AppTesting string `json:"appTesting"`
}

func (this *Service) OnConfig() *service.Config {
	cfg := service.Config{
		Name: pkg.AppName,
		//UserName:    "root",
		DisplayName: fmt.Sprintf("A AAATest1 Service %s", pkg.AppVersion),
		Description: "A Golang AAATest1 Service..",
	}
	return &cfg
}

func (this *Service) OnVersion() string {
	pkg.Version()
	return pkg.AppVersion
}

func (this *Service) OnRun(service igs.Service) error {
	this.gs = service
	//glog.SetLogFile("./logs", "app.log")
	go Server(this.port, this)
	for {
		this.timestamp = time.Now().Format(time.RFC3339)
		//glog.Println("run", t.timestamp)
		time.Sleep(time.Second * 1)
	}
}

func (this *Service) GetAny(s2 string) any {
	return this.menu()
}

func (this *Service) menu() any {
	appName := utils.InputStringEmpty(fmt.Sprintf("测试输入："), "册书数据")
	port := utils.InputIntDefault(fmt.Sprintf("测试输入端口(%d)：", 9090), 9090)
	this.port = port
	return &Config{AppTesting: appName}
}
