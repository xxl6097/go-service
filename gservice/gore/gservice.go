package gore

import (
	"errors"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore/util"
	"github.com/xxl6097/go-service/gservice/utils"
	"os"
	"os/exec"
)

type FF interface {
	Restart() error
}

var DD FF

type goreservice struct {
	s service.Service
}

func NewGoreService(s service.Service) IGService {
	return &goreservice{
		s: s,
	}
}
func (this *goreservice) runChildProcess(executable string, args ...string) error {
	cmd := exec.Command(executable, args...)
	util.SetPlatformSpecificAttrs(cmd)
	glog.Printf("运行子进程 %s %v\n", executable, args)
	return cmd.Start()
}
func (this *goreservice) Upgrade(destFilePath string, args ...string) error {
	var newFilePath string
	if utils.IsURL(destFilePath) {
		filePath, err := utils.DownLoad(destFilePath)
		if err != nil {
			glog.Error("下载失败", err)
			return err
		}
		newFilePath = filePath
		glog.Debug("下载成功.", newFilePath)
	} else if utils.FileExists(destFilePath) {
		newFilePath = destFilePath
	} else {
		glog.Error("无法识别的文件", newFilePath)
		return errors.New("无法识别的文件" + newFilePath)
	}

	err := os.Chmod(newFilePath, 0755)
	if err != nil {
		glog.Errorf("赋权限错误: %v %s %v\n", utils.FileExists(newFilePath), newFilePath, err)
		return fmt.Errorf("赋权限错误: %v\n", err)
	}
	glog.Println("当前进程ID:", os.Getpid())
	arg := make([]string, 0)
	arg = append(arg, "upgrade")
	arg = append(arg, newFilePath)
	arg = append(arg, args...)
	err = this.runChildProcess(newFilePath, arg...)
	if err != nil {
		glog.Errorf("RunChildProcess错误: %v\n", err)
		return fmt.Errorf("Error starting update process: %v\n", err)
	}
	glog.Println("升级进程启动成功", newFilePath)
	return err
}

func (this *goreservice) Restart() error {
	if utils.IsWindows() {
		return utils.RestartForWindows()
	}
	return DD.Restart()
	//if this.s == nil {
	//	return errors.New("daemon is nil")
	//}
	//return this.s.Restart()
}

func (this *goreservice) Uninstall() error {
	if this.s == nil {
		return errors.New("daemon is nil")
	}
	return this.s.Uninstall()
}
