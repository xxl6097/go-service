package internal

import (
	"errors"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/pkg"
	"github.com/xxl6097/go-service/pkg/gs/igs"
	"github.com/xxl6097/go-service/pkg/utils"
	"github.com/xxl6097/go-service/pkg/utils/util"
	"os"
	"path/filepath"
	"strings"
)

type CoreService struct {
	iService igs.IService
	srv      service.Service
	config   *service.Config
	workDir  string
}

func (this *CoreService) initLog() {
	if len(os.Args) > 1 {
		glog.LogDefaultLogSetting(fmt.Sprintf("%s.log", os.Args[1]))
	} else {
		bindir, err := os.Executable()
		var isSrvApp bool
		if err != nil {
			glog.LogDefaultLogSetting("app.log")
		} else {
			isSrvApp = strings.HasPrefix(strings.ToLower(bindir), strings.ToLower(util.DefaultInstallPath))
			if isSrvApp {
				glog.LogDefaultLogSetting("app.log")
			} else {
				glog.SetLogFile(filepath.Dir(bindir), fmt.Sprintf("install-%s.log", filepath.Base(bindir)))
			}
		}
	}
}

func (this *CoreService) Run() error {
	if len(os.Args) > 1 && os.Args[1] == utils.SYSTEM_CPU_INFO {
		fmt.Printf("%s/%s", pkg.OsType, pkg.Arch)
		return nil
	}
	this.initLog()
	this.config = this.iService.OnConfig()
	if this.config == nil {
		return errors.New("请设置服务配置信息～")
	}
	if this.config.Name == "" {
		return errors.New("【Name】应用服务名不能为空字符串")
	}
	if this.config.DisplayName == "" {
		return errors.New("【DisplayName】应用服务显示名不能为空字符串")
	}
	if this.config.Description == "" {
		return errors.New("【Description】应用服务显示名不能为空字符串")
	}
	this.config.Name = strings.ToLower(this.config.Name)
	this.workDir = filepath.Join(util.DefaultInstallPath, this.config.Name)
	if utils.IsWindows() {
		this.config.Name = fmt.Sprintf("%s.exe", this.config.Name)
	}
	this.config.Executable = filepath.Join(this.workDir, this.config.Name)

	this.deleteOld()
	binDir := filepath.Dir(os.Args[0])
	_ = os.Chdir(binDir)
	e := this.createService()
	if e != nil {
		return e
	}
	//glog.Debug("运行参数", os.Args, os.Getpid())
	return this.menu()
}

func (this *CoreService) reqeustWindowsUser() {
	if utils.IsWindows() {
		if this.config.UserName == "" {
			username := utils.InputStringEmpty("请输入windows登录用户名：", "")
			if username != "" {
				this.config.UserName = username
			}
			password := utils.InputStringEmpty("请输入windows登录用户密码：", "")
			if password != "" {
				if this.config.Option == nil {
					this.config.Option = make(map[string]interface{})
				}
				this.config.Option["Password"] = password
			}

		}
	}
}

// 在安装目录且服务处于运行状态
func (this *CoreService) isServiceApp() bool {
	binPath, err := os.Executable()
	if err != nil {
		return false
	}
	if strings.Compare(strings.ToLower(this.config.Executable), strings.ToLower(binPath)) == 0 {
		s, e := this.statusService()
		if e != nil {
			return false
		}
		if s == service.StatusRunning {
			return true
		}
	}
	return false
}

func (this *CoreService) deleteOld() {
	tempFilePath := filepath.Join(this.workDir, fmt.Sprintf(".%s.old", this.config.Name))
	if utils.FileExists(tempFilePath) {
		_ = utils.DeleteAllDirector(tempFilePath)
	}
}

func NewCore(srv igs.IService) igs.DefaultService {
	return &CoreService{iService: srv}
}
