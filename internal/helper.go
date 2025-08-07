package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/internal/core"
	"github.com/xxl6097/go-service/pkg/ukey"
	"github.com/xxl6097/go-service/pkg/utils"
	"github.com/xxl6097/go-service/pkg/utils/util"
	"os"
	"path/filepath"
	"time"
)

func (this *CoreService) install() error {
	//this.reqeustWindowsUser()
	s, err := this.statusService()
	var isRemoved = true
	if err == nil {
		no := utils.InputString(fmt.Sprintf("%s%s%s", "检测到", this.config.Name, "程序已经安装，卸载/更新/取消?(y/u/n):"))
		switch no {
		case "y", "Y", "Yes", "YES":
			isRemoved = true
			e := this.uninstall()
			if e != nil {
				glog.Error("卸载失败", this.config.Name, e)
			} else {
				glog.Error("卸载成功", this.config.Name)
			}
			break
		case "u", "U", "Update", "UPDATE":
			isRemoved = false
			e := this.stopService()
			if err != nil {
				//return err
				glog.Error(e)
			}
			break
		default:
			glog.Debug("结束安装.")
			time.Sleep(time.Second * 3)
			os.Exit(0)
			return err
		}
	} else if s != service.StatusUnknown {
		e := this.uninstallService()
		if e != nil {
			glog.Error("卸载失败", e)
		}
	}
	util.SetFirewall(this.config.Name, this.config.Executable)
	e1 := util.SetRLimit()
	if e1 != nil {
		glog.Error("SetRLimit", e1)
	}

	if isRemoved {
		_ = utils.ResetDirector(this.workDir)
	} else {
		_ = utils.CheckDirector(this.workDir)
	}

	//这个地方是取的当前运行的执行文件
	currentBinPath, e := os.Executable()
	if e != nil {
		glog.Fatal("os.Executable() error", e)
		return e
	}

	ee := core.Install(this.isrv, currentBinPath, this.config.Executable)
	if ee != nil {
		glog.Error(ee)
		return ee
	}

	e = os.Chdir(this.workDir)
	if e != nil {
		glog.Println("os.Chdir error:", e)
		return e
	}

	err = this.installService()
	if err == nil {
		glog.Printf("[%s]安装成功!\n", this.config.DisplayName)
	} else {
		glog.Printf("[%s]安装失败，错误信息:%v\n", this.config.DisplayName, err)
	}
	time.Sleep(time.Second * 1)
	err = this.startService()
	if err != nil {
		glog.Printf("[%s]启动失败，错误信息:%v\n", this.config.DisplayName, err)
	} else {
		glog.Printf("[%s]启动成功!\n", this.config.DisplayName)
	}
	time.Sleep(time.Second * 1)
	glog.Info(this.Status())
	return nil
}
func (this *CoreService) uninstall() error {
	defer func() {
		this.clearForUninstall()
		glog.Debug("1尝试停止服务")
		err := this.stopService()
		glog.Debug("2尝试停止服务", err)
		this.clearForUninstall()
		glog.Flush()
	}()

	if utils.IsOpenWRT() {
		err := this.stopService()
		if err != nil {
			glog.Errorf("[%s]停止失败 %v", this.config.Name, err)
		} else {
			glog.Errorf("[%s]停止成功", this.config.Name)
		}
	}
	err := this.uninstallService()
	if err != nil {
		glog.Printf("[%s]卸载失败 %v\n", this.config.Name, err)
	} else {
		glog.Printf("[%s]成功卸载\n", this.config.Name)
	}
	return err
}

func (this *CoreService) patchUpgrade(ctx context.Context, binUrlOrLocal string) error {
	downFilePath, err := utils.CheckFileOrDownload(ctx, binUrlOrLocal)
	if err != nil {
		glog.Debug("升级失败", err)
		return err
	}

	if !utils.FileExists(downFilePath) {
		return fmt.Errorf("差分升级文件不存在 %s", downFilePath)
	}
	return this.update(downFilePath, true)
}

// 1. 检测升级文件是本地还是网络文件（下载）；
// 2. 升级文件最终需要被删除，所以使用defer删除；
// 3. 给升级文件赋予0755权限
func (this *CoreService) upgrade(ctx context.Context, binUrlOrLocal string) error {
	defer glog.Flush()
	downFilePath, err := utils.CheckFileOrDownload(ctx, binUrlOrLocal)
	if err != nil {
		glog.Debug("升级失败", err)
		return err
	}
	var signFilePath string
	patch := false
	if filepath.Ext(downFilePath) == ".patch" {
		signFilePath = downFilePath
		patch = true
	} else {
		tempFilePath, e := ukey.SignFileByOldFileKey(this.config.Executable, downFilePath)
		//签名完后会生产出新的签名文件，那么下载的文件需要被删除
		_ = utils.DeleteAllDirector(downFilePath)
		if e != nil {
			glog.Debug("升级签名错误", e)
			return err
		}
		signFilePath = tempFilePath
	}

	if !utils.FileExists(signFilePath) {
		return fmt.Errorf("升级文件不存在 %s", signFilePath)
	}

	return this.update(signFilePath, patch)
	//err = os.Chmod(signFilePath, 0755)
	//if err != nil {
	//	eMsg := fmt.Sprintf("赋权限错误: %s %v\n", signFilePath, err)
	//	glog.Error(eMsg)
	//	return fmt.Errorf(eMsg)
	//}
	//glog.Println("当前进程ID:", os.Getpid(), this.config.Executable)
	//err = utils.PerformUpdate(signFilePath, this.config.Executable, patch)
	////同样，更新完后，需要删除签名文件
	//_ = utils.DeleteAllDirector(filepath.Dir(filepath.Dir(signFilePath)))
	//if err != nil {
	//	glog.Error("升级失败", err)
	//	return err
	//}
	//glog.Error("升级成功")
	//if utils.IsWindows() {
	//	glog.Debug(utils.RunCmd("dir"))
	//} else {
	//	glog.Debug(utils.RunCmd("ls", "-l"))
	//}
	//return this.RunCMD("restart")
}

func (this *CoreService) changeSelf(buffer []byte) error {
	if buffer == nil || len(buffer) == 0 {
		return errors.New("配置buffer空")
	}
	binFilePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取当前可执行文件路径出错: %v\n", err)
	}

	tempFilePath, e := ukey.SignFileByBuffer(buffer, binFilePath)
	if e != nil {
		glog.Debug("升级签名错误", e)
		return err
	}
	signFilePath := tempFilePath
	if !utils.FileExists(signFilePath) {
		return fmt.Errorf("自升级文件不存在 %s", signFilePath)
	}
	return this.update(signFilePath, false)
}

func (this *CoreService) update(signFilePath string, patch bool) error {
	err := os.Chmod(signFilePath, 0755)
	if err != nil {
		eMsg := fmt.Sprintf("赋权限错误: %s %v\n", signFilePath, err)
		glog.Error(eMsg)
		return fmt.Errorf(eMsg)
	}
	if !patch {
		err = utils.IsMissMatchOsApp(signFilePath)
		if err != nil {
			return err
		}
	}
	glog.Println("当前进程ID:", os.Getpid(), this.config.Executable)
	err = utils.PerformUpdate(signFilePath, this.config.Executable, patch)
	//同样，更新完后，需要删除签名文件
	_ = utils.DeleteAllDirector(signFilePath)
	if err != nil {
		glog.Error("升级失败", err)
		return err
	}
	glog.Error("升级成功")
	if utils.IsWindows() {
		glog.Debug(utils.RunCmd("dir"))
	} else {
		glog.Debug(utils.RunCmd("ls", "-l"))
	}
	return this.RunCMD("restart")
}

func (this *CoreService) clearForUninstall() {
	glog.CloseLog()
	_ = utils.DeleteAllDirector(this.workDir)
	appDir := glog.AppHome()
	_ = utils.DeleteAllDirector(appDir)
}
func (this *CoreService) clearAppData() error {
	appDir := glog.AppHome()
	return utils.DeleteAllDirector(appDir)
}
