package internal

import (
	"errors"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore/util"
	"github.com/xxl6097/go-service/gservice/utils"
	"github.com/xxl6097/go-service/pkg/gs/igs"
	utils2 "github.com/xxl6097/go-service/pkg/utils"
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

func (this *CoreService) Run() error {
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
	tempFilePath := filepath.Join(this.workDir, fmt.Sprintf(".%s.old", this.config.Name))
	_ = utils2.DeleteAllDirector(tempFilePath)
	binDir := filepath.Dir(os.Args[0])
	_ = os.Chdir(binDir)
	e := this.createService()
	if e != nil {
		return e
	}
	glog.Debug("运行参数", os.Args, os.Getpid())
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

func NewCore(srv igs.IService) igs.DefaultService {
	return &CoreService{iService: srv}
}
