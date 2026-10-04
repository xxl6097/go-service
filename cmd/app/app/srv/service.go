package srv

import (
	"fmt"

	"github.com/kardianos/service"
	"github.com/xxl6097/glog/pkg/z"
	_ "github.com/xxl6097/go-service/assets/buffer"
	"github.com/xxl6097/go-service/pkg/gs/igs"
	"github.com/xxl6097/go-service/pkg/ukey"
	"github.com/xxl6097/go-service/pkg/utils"
	"github.com/xxl6097/go-service/pkg/version"
	"go.uber.org/zap"

	"os"
)

type Service struct {
	timestamp string
	gs        igs.Service
}

func (t *Service) UnInstall() {
	//TODO implement me
	z.L().Info("service UnInstall")
}

func (t *Service) OnStop() {
	z.L().Info("service stop")
}

func (t *Service) OnShutdown() {
	z.L().Info("OnShutdown")
}

func (this *Service) OnFinish() {
}

func (this *Service) OnConfig() *service.Config {
	cfg := service.Config{
		Name: "aatest", //version.AppName
		//UserName:    "root",
		DisplayName: fmt.Sprintf("AAATest_%s", version.AppVersion),
		Description: "A Golang AAATest Service..",
	}
	// macOS 下安装目录在用户目录（见 util_darwin.go），服务也注册为用户级
	// LaunchAgent（~/Library/LaunchAgents/<name>.plist）；
	// 否则 kardianos 会写到 /Library/LaunchDaemons，仍然需要 root。
	if utils.IsMacOs() {
		cfg.Option = service.KeyValue{
			"UserService": true,
			// RunAtLoad 默认 false，不设的话登录后不会自动拉起
			"RunAtLoad": true,
		}
	}
	return &cfg
}

func (this *Service) OnVersion() string {
	version.Version()
	cfg, err := load()
	if err == nil {
		z.L().Debug("cfg", zap.Any("cfg", cfg))
	}
	return version.AppVersion
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
