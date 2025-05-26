package srv

import (
	"encoding/json"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/pkg"
	"github.com/xxl6097/go-service/pkg/gs/igs"
	"github.com/xxl6097/go-service/pkg/ukey"
	"github.com/xxl6097/go-service/pkg/utils"
	"os"
	"time"
)

type Service struct {
	timestamp string
	gs        igs.Service
}

type Config struct {
	//ukey.KeyBuffer
	AppTesting string `json:"appTesting"`
	ServerPort int    `json:"serverPort"`
}

func load() (*Config, error) {
	defer glog.Flush()
	byteArray, err := ukey.Load()
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(byteArray, &cfg)
	if err != nil {
		glog.Println("ClientConfig解析错误", err)
		return nil, err
	}
	pkg.Version()
	return &cfg, nil
}

func (this *Service) OnConfig() *service.Config {
	cfg := service.Config{
		Name: pkg.AppName,
		//UserName:    "root",
		DisplayName: fmt.Sprintf("AAATest_%s", pkg.AppVersion),
		Description: "A Golang AAATest Service..",
	}
	return &cfg
}

func (this *Service) OnVersion() string {
	pkg.Version()
	cfg, err := load()
	if err == nil {
		glog.Debugf("cfg:%+v", cfg)
	}
	return pkg.AppVersion
}

func (this *Service) OnRun(service igs.Service) error {
	this.gs = service
	cfg, err := load()
	if err != nil {
		return err
	}
	glog.Debug("程序运行", os.Args)
	go Server(cfg.ServerPort, this)
	for {
		this.timestamp = time.Now().Format(time.RFC3339)
		glog.Println("run", pkg.AppVersion, pkg.BuildTime, this.timestamp)
		time.Sleep(time.Second * 10)
	}
}

func (this *Service) GetAny(s2 string) any {
	return this.menu()
}

func (this *Service) menu() any {
	appName := utils.InputStringEmpty(fmt.Sprintf("测试输入："), "测试数据")
	port := utils.InputIntDefault(fmt.Sprintf("测试输入端口(%d)：", 9090), 9090)
	return &Config{AppTesting: appName, ServerPort: port}
}
