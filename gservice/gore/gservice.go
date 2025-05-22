package gore

import (
	"context"
	"errors"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore/util"
	"github.com/xxl6097/go-service/gservice/utils"
	"os"
	"os/exec"
	"path/filepath"
)

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
func (this *goreservice) Upgrade(ctx context.Context, destFilePath string, args ...string) error {
	var newFilePath string
	if utils.IsURL(destFilePath) {
		filePath, err := utils.DownloadFileWithCancel(ctx, destFilePath)
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

func (this *goreservice) RunCmd(args ...string) error {
	binpath, err := os.Executable()
	if err != nil {
		return err
	}
	err = this.runChildProcess(binpath, args...)
	if err != nil {
		glog.Errorf("RunChildProcess错误: %v\n", err)
		return fmt.Errorf("RunChildProcess错误: %v\n", err)
	}
	glog.Println("子进程启动成功", binpath)
	return err
}

func (this *goreservice) Restart() error {
	if utils.IsWindows() {
		return utils.RestartForWindows()
	}
	if this.s == nil {
		return errors.New("daemon is nil")
	}
	return this.s.Restart()
}

func (this *goreservice) Uninstall() error {
	if this.s == nil {
		return errors.New("daemon is nil")
	}
	e := this.s.Uninstall()
	if e != nil {
		glog.Errorf("原生函数卸载失败 %+v", e)
		binpath, err := os.Executable()
		if err != nil {
			return err
		}
		fileName := filepath.Base(binpath)
		destDir := glog.GetCrossPlatformDataDir("temp", utils.SecureRandomID())
		destFilePath := filepath.Join(destDir, fileName)
		err = utils.Copy(binpath, destFilePath)
		if err != nil {
			return err
		}
		defer utils.DeleteAll(destDir, "删除卸载临时文件")
		err = os.Chmod(destFilePath, 0755)
		if err != nil {
			glog.Errorf("赋权限错误: %v %s %v\n", utils.FileExists(destFilePath), destFilePath, err)
			return fmt.Errorf("赋权限错误: %v\n", err)
		}
		glog.Println("当前进程ID:", os.Getpid())
		err = this.runChildProcess(destFilePath, "uninstall")
		if err != nil {
			glog.Errorf("RunChildProcess错误: %v\n", err)
			return fmt.Errorf("Error starting update process: %v\n", err)
		}
		glog.Println("程序卸载成功", destFilePath)
		return err
	}
	return e
}

func (this *goreservice) Uninstall1() error {
	if this.s == nil {
		return errors.New("daemon is nil")
	}
	return this.s.Uninstall()
}
