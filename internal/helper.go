package internal

import (
	"context"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore/util"
	utils2 "github.com/xxl6097/go-service/gservice/utils"
	"github.com/xxl6097/go-service/internal/core"
	utils3 "github.com/xxl6097/go-service/pkg/utils"
	"os"
	"time"
)

func (this *CoreService) install() error {
	s, err := this.srv.Status()
	var isRemoved = true
	if err == nil {
		no := utils2.InputString(fmt.Sprintf("%s%s%s", "检测到", this.config.Name, "程序已经安装，卸载/更新/取消?(y/u/n):"))
		switch no {
		case "y", "Y", "Yes", "YES":
			isRemoved = true
			e := this.srv.Uninstall()
			if e != nil {
				glog.Error("卸载失败", this.config.Name, e)
			} else {
				glog.Error("卸载成功", this.config.Name)
			}
			break
		case "u", "U", "Update", "UPDATE":
			isRemoved = false
			e := this.srv.Stop()
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
		e := this.srv.Uninstall()
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

	err = this.srv.Install() //.Control("install", this.binPath, []string{"-d"})
	if err == nil {
		glog.Printf("服务【%s】安装成功!\n", this.config.DisplayName)
	} else {
		glog.Printf("服务【%s】安装失败，错误信息:%v\n", this.config.DisplayName, err)
	}
	time.Sleep(time.Second * 2)
	err = this.srv.Start()
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
		if !utils2.IsWindows() {
			err := this.srv.Stop()
			glog.Debug("尝试停止服务", err)
		}
		_ = glog.Flush()
	}()
	if utils2.IsWindows() {
		err := this.srv.Stop()
		glog.Debug("尝试停止服务", err)
	}
	err := this.srv.Uninstall() //Control("uninstall", "", nil)
	if err != nil {
		glog.Printf("服务卸载失败 %v\n", err)
		_ = glog.Flush()
	} else {
		glog.Printf("服务成功卸载\n")
		_ = glog.Flush()
	}
	time.Sleep(time.Second * 2)
	// 尝试删除自身
	return err
}

// 1. 检测升级文件是本地还是网络文件（下载）；
// 2. 升级文件最终需要被删除，所以使用defer删除；
// 3. 给升级文件赋予0755权限
func (this *CoreService) upgrade(ctx context.Context, binUrlOrLocal string) error {
	newFilePath, err := utils3.CheckFileOrDownload(ctx, binUrlOrLocal)
	if err != nil {
		return err
	}
	defer func() {
		e := os.Remove(newFilePath)
		if e != nil {
			glog.Error("升级文件删除失败", newFilePath, e)
		} else {
			glog.Debug("升级文件删除成功", newFilePath)
		}
	}()
	err = os.Chmod(newFilePath, 0755)
	if err != nil {
		eMsg := fmt.Sprintf("赋权限错误: %s %v\n", newFilePath, err)
		return fmt.Errorf(eMsg)
	}
	glog.Println("当前进程ID:", os.Getpid())
	err = utils3.PerformUpdate(newFilePath)
	if err != nil {
		return err
	}
	utils3.RestartWindowsApplication()
	return nil
}

func (this *CoreService) clear() {
	glog.CleanGLog(glog.StdGLog)
	glog.Debugf("删除安装文件，pid: %v", os.Getpid())
	e := os.RemoveAll(this.workDir)
	if e != nil {
		glog.Errorf("删除失败[%s] err:%v", this.workDir, e)
	} else {
		glog.Debugf("删除成功[%s]", this.workDir)
	}
	appDir := glog.GetCrossPlatformDataDir()
	e = os.RemoveAll(appDir)
	if e != nil {
		glog.Errorf("删除失败[%s] err:%v", appDir, e)
	} else {
		glog.Debugf("删除成功[%s]", appDir)
	}
}
