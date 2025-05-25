package internal

import (
	"context"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/utils"
	"os"
)

func (this *CoreService) Upgrade(ctx context.Context, binUrl string) error {
	return this.upgrade(ctx, binUrl)
}

func (this *CoreService) UnInstall() error {
	return this.uninstall()
}

func (this *CoreService) RunCMD(args ...string) error {
	return RunCmdBySelf(args...)
}

func (this *CoreService) Restart() error {
	if utils.IsWindows() {
		return RunCmdBySelf("restart")
	}
	return this.restartService()
	//return this.srv.Restart()
	//return RunCmdBySelf("restart")
}

func RunCmdBySelf(args ...string) error {
	defer glog.Flush()
	binpath, err := os.Executable()
	if err != nil {
		return err
	}
	err = utils.RunChildProcess(binpath, args...)
	if err != nil {
		glog.Errorf("RunChildProcess错误: %v\n", err)
		return fmt.Errorf("RunChildProcess错误: %v\n", err)
	}
	glog.Println("子进程启动成功", binpath)
	return err
}
