package gore

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/inconshreveable/go-update"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/utils"
	"os"
)

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

	file, err := os.Open(newFilePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()
	// 使用 bufio.NewReader 创建带缓冲的读取器
	err = update.Apply(bufio.NewReader(file), update.Options{})
	if err != nil {
		glog.Error(err)
		return err
	}
	return nil

}

func (this *goreservice) Restart() error {
	if this.s == nil {
		return errors.New("daemon is nil")
	}
	if utils.IsMacOs() {
		//cmd := exec.Command("sudo", "launchctl", "kickstart", "-k", "aatest")
		//util.SetPlatformSpecificAttrs(cmd)
		//glog.Printf("运行子进程 \n")
		//return cmd.Start()
		//c, err := utils.RunCmdWithSudo("launchctl", "kickstart", "-k", "aatest")
		//c, err := utils.RunCmdWithSudo("launchctl", "load", "/Library/LaunchDaemons/aatest.plist")
		//if c != nil {
		//	glog.Debugf("result: %v", string(c))
		//}
		err := this.RunCmd("restart")
		return err
	}
	return this.s.Restart()
}
