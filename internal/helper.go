package internal

import (
	"context"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore/util"
	utils2 "github.com/xxl6097/go-service/gservice/utils"
	"github.com/xxl6097/go-service/internal/core"
	"github.com/xxl6097/go-service/pkg/ukey"
	utils3 "github.com/xxl6097/go-service/pkg/utils"
	"os"
	"path/filepath"
	"time"
)

func (this *CoreService) install() error {
	//if len(os.Args) > 1 {
	//	if os.Args[1] == "install" {
	//		this.reqeustWindowsUser()
	//	} else if os.Args[1] == "uninstall" {
	//		glog.CloseLog()
	//	}
	//}
	//this.reqeustWindowsUser()
	s, err := this.statusService()
	var isRemoved = true
	if err == nil {
		no := utils2.InputString(fmt.Sprintf("%s%s%s", "检测到", this.config.Name, "程序已经安装，卸载/更新/取消?(y/u/n):"))
		switch no {
		case "y", "Y", "Yes", "YES":
			isRemoved = true
			e := this.uninstallService()
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

	if _, e2 := os.Stat(this.workDir); !os.IsNotExist(e2) {
		if isRemoved {
			e := os.RemoveAll(this.workDir)
			if e != nil {
				glog.Error("删除失败", this.workDir)
			} else {
				e = os.MkdirAll(this.workDir, 0775)
				if e != nil {
					glog.Printf("MkdirAll %s error:%s", this.workDir, err)
					return e
				}
			}
		}

	} else {
		e := os.MkdirAll(this.workDir, 0775)
		if e != nil {
			glog.Printf("MkdirAll %s error:%s", this.workDir, e)
			return e
		}
	}

	//这个地方是取的当前运行的执行文件
	currentBinPath, e := os.Executable()
	if e != nil {
		glog.Fatal("os.Executable() error", e)
		return e
	}

	ee := core.Install(this.iService, currentBinPath, this.config.Executable)
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
		glog.Printf("服务【%s】安装成功!\n", this.config.DisplayName)
	} else {
		glog.Printf("服务【%s】安装失败，错误信息:%v\n", this.config.DisplayName, err)
	}
	time.Sleep(time.Second * 2)
	err = this.startService()
	if err != nil {
		glog.Printf("服务【%s】启动失败，错误信息:%v\n", this.config.DisplayName, err)
	} else {
		glog.Printf("服务【%s】启动成功!\n", this.config.DisplayName)
	}
	return nil
}
func (this *CoreService) uninstall() error {
	defer func() {
		this.clear()
		glog.Debug("1尝试停止服务")
		err := this.stopService()
		glog.Debug("2尝试停止服务", err)
		this.clear()
		glog.Flush()
	}()
	//glog.CloseLog()
	//if utils2.IsWindows() {
	//	err := this.srv.Stop()
	//	glog.Debug("尝试停止服务", err)
	//}
	err := this.uninstallService()
	if err != nil {
		glog.Printf("服务卸载失败 %v\n", err)
	} else {
		glog.Printf("服务成功卸载\n")
	}
	//time.Sleep(time.Second * 2)
	// 尝试删除自身
	return err
}

// 1. 检测升级文件是本地还是网络文件（下载）；
// 2. 升级文件最终需要被删除，所以使用defer删除；
// 3. 给升级文件赋予0755权限
func (this *CoreService) upgrade(ctx context.Context, binUrlOrLocal string) error {
	defer glog.Flush()
	downFilePath, err := utils3.CheckFileOrDownload(ctx, binUrlOrLocal)
	if err != nil {
		glog.Debug("升级失败", err)
		return err
	}
	signFilePath, e := ukey.SignFileByOldFileKey(this.config.Executable, downFilePath)
	_ = utils3.DeleteAllDirector(filepath.Dir(filepath.Dir(downFilePath)))
	if e != nil {
		glog.Debug("升级签名错误", e)
		return err
	}
	err = os.Chmod(signFilePath, 0755)
	if err != nil {
		eMsg := fmt.Sprintf("赋权限错误: %s %v\n", signFilePath, err)
		glog.Error(eMsg)
		return fmt.Errorf(eMsg)
	}
	glog.Println("当前进程ID:", os.Getpid())
	err = utils3.PerformUpdate(signFilePath, this.config.Executable)
	//_ = utils3.DeleteAllDirector(filepath.Dir(filepath.Dir(signFilePath)), "签名文件")
	if err != nil {
		glog.Error("升级失败", err)
		return err
	}
	glog.Error("升级成功")
	if utils2.IsWindows() {
		return this.RunCMD("restart")
	}
	return this.restartService()
}

func (this *CoreService) clear() {
	glog.CloseLog()
	glog.Debugf("删除安装文件，pid: %v", os.Getpid())
	_ = utils3.DeleteAllDirector(this.workDir)
	appDir := glog.GetCrossPlatformDataDir()
	_ = utils3.DeleteAllDirector(appDir)
}
