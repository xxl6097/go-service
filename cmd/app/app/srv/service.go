package srv

import (
	"fmt"

	"github.com/kardianos/service"
	"github.com/xxl6097/glog/pkg/z"
	_ "github.com/xxl6097/go-service/assets/buffer"
	"github.com/xxl6097/go-service/pkg"
	"github.com/xxl6097/go-service/pkg/gs/igs"
	"github.com/xxl6097/go-service/pkg/ukey"
	"github.com/xxl6097/go-service/pkg/utils"
	"go.uber.org/zap"

	"os"
)

type Service struct {
	timestamp string
	gs        igs.Service
}

func (t *Service) OnStop() {
	z.L().Info("service stop")
}

func (t *Service) OnShutdown() {
	z.L().Info("OnShutdown")
}

func (this *Service) OnFinish() {
}

type Config struct {
	//ukey.KeyBuffer
	AppTesting string `json:"appTesting"`
	ServerPort int    `json:"serverPort"`
}

func load() (*Config, error) {
	byteArray, err := ukey.Load()
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = ukey.GobToStruct(byteArray, &cfg)
	//err = json.Unmarshal(byteArray, &cfg)
	if err != nil {
		z.L().Error("ClientConfig解析错误", zap.Error(err))
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
		z.L().Debug("cfg", zap.Any("cfg", cfg))
	}
	return pkg.AppVersion
}

func (this *Service) OnRun(service igs.Service) error {
	this.gs = service
	cfg, err := load()
	if err != nil {
		return err
	}
	z.L().Debug("程序运行", zap.Strings("args", os.Args))
	Server(cfg.ServerPort, this)
	//for {
	//	this.timestamp = time.Now().Format(time.RFC3339)
	//	glog.Println("run", pkg.AppVersion, pkg.BuildTime, this.timestamp)
	//	time.Sleep(time.Second * 10)
	//}
	return nil
}

func (this *Service) GetAny(binDir string) ([]byte, []string) {
	return this.menu(), []string{"--conf", "hello"}
}

func (this *Service) menu() []byte {
	appName := utils.InputStringEmpty(fmt.Sprintf("测试输入："), "测试数据")
	port := utils.InputIntDefault(fmt.Sprintf("测试输入端口(%d)：", 9090), 9090)
	cfg := &Config{AppTesting: appName, ServerPort: port}
	bb, e := ukey.StructToGob(cfg)
	if e != nil {
		return nil
	}
	return bb
}
