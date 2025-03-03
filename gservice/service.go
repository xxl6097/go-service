package gservice

import (
	"errors"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore"
	"github.com/xxl6097/go-service/gservice/gore/util"
	"github.com/xxl6097/go-service/gservice/utils"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type gservice struct {
	daemon           *gore.Daemon
	srv              gore.GService
	conf             *service.Config
	workDir, tempDir string
}

func Run(srv gore.GService) error {
	bconfig := srv.OnInit()
	if bconfig == nil {
		return fmt.Errorf("请实现OnConfig() *service.Config方法")
	}
	if bconfig.Name == "" {
		return fmt.Errorf("应用名不能为空")
	}
	glog.SetLogFile(filepath.Join(os.TempDir(), bconfig.Name), "install.log")
	glog.SetNoHeader(true)
	if bconfig.DisplayName == "" {
		return fmt.Errorf("服务显示名不能为空")
	}
	if bconfig.Description == "" {
		return fmt.Errorf("服务描述不能为空")
	}
	bconfig.Name = strings.ToLower(bconfig.Name)
	this := &gservice{
		srv:     srv,
		conf:    bconfig,
		workDir: filepath.Join(util.DefaultInstallPath, bconfig.Name),
		tempDir: filepath.Join(os.TempDir(), bconfig.Name),
	}
	if utils.IsWindows() {
		bconfig.Name = bconfig.Name + ".exe"
	}

	bconfig.Executable = filepath.Join(this.workDir, bconfig.Name)
	binDir := filepath.Dir(os.Args[0])
	os.Chdir(binDir)
	d, err := gore.NewDaemon(gore.NewCoreService(srv), bconfig)
	if err != nil {
		return err
	}
	this.daemon = d
	//if this.daemon.IsRunning() {
	//	return this.run()
	//} else {
	//	return srv.OnRun(this)
	//}
	return this.run()
}

func (this *gservice) run() error {
	defer glog.Flush()
	if this.srv == nil {
		return errors.New("请继承gservice.IService接口！")
	}
	if len(os.Args) > 1 {
		k := os.Args[1]
		switch k {
		case "-v", "v", "version":
			this.srv.OnVersion()
			return nil
		case "install":
			return this.install()
		case "uninstall":
			return this.uninstall()
		case "upgrade", "update":
			return this.upgrade()
		case "start":
			return this.startService()
		case "stop":
			return this.stopService()
		case "restart":
			return this.restart()
		}
	}
	glog.Printf("运行服务【%s】%v\n", this.conf.DisplayName, this.daemon.IsRunning())
	return this.daemon.Run()
}

//func (this *gservice) update(upgradeBinPath string) error {
//	if utils.IsURL(upgradeBinPath) {
//		resp, err := http.Get(upgradeBinPath)
//		if err != nil {
//			return err
//		}
//		defer resp.Body.Close()
//		err = update.Apply(resp.Body, update.Options{})
//		if err != nil {
//			glog.Error(err)
//			return err
//		}
//		return nil
//	} else if utils.FileExists(upgradeBinPath) {
//		// 打开文件
//		file, err := os.Open(upgradeBinPath)
//		if err != nil {
//			return fmt.Errorf("Error opening file: %v", err)
//		}
//		defer file.Close()
//		// 使用 bufio.NewReader 创建带缓冲的读取器
//		err = update.Apply(bufio.NewReader(file), update.Options{})
//		if err != nil {
//			glog.Error(err)
//			return err
//		}
//		return nil
//	}
//	return fmt.Errorf("位置文件路径:%s", upgradeBinPath)
//}

//func (this *gservice) Upgrade1(upgradeBinPath string) error {
//	var err error
//	defer func() {
//		if err == nil {
//			go func() {
//				time.Sleep(100 * time.Millisecond)
//				err = this.Restart()
//				if err != nil {
//					glog.Errorf("Error restarting: %v\n", err)
//				}
//			}()
//		}
//	}()
//	err = this.update(upgradeBinPath)
//	return err
//}

func (this *gservice) upgrade() error {
	defer glog.Flush()
	glog.Debugf(">>>>>>>进入升级流程[%d] %v\n", os.Getpid(), os.Args)
	if len(os.Args) <= 2 {
		glog.Error("参数错误，请重新配置参数")
		return errors.New("参数错误，请重新配置参数")
	}
	fileUrlOrLocalPath := os.Args[2]
	glog.Debug("升级文件地址", fileUrlOrLocalPath)
	defer utils.Delete(fileUrlOrLocalPath, "升级文件")

	if utils.IsWindows() {
		glog.Printf("停止服务【%s】\n", this.conf.DisplayName)
		err := this.daemon.Stop() //.Control("stop", "", nil)
		if err != nil {           // service maybe not install
			glog.Printf("服务【%s】未运行 %v\n", this.conf.DisplayName, err)
		}
	}
	glog.Debug("开始卸载")
	err := this.daemon.Uninstall()
	if err != nil {
		glog.Printf("服务【%s】卸载失败 %v\n", this.conf.DisplayName, err)
	} else {
		glog.Println("服务成功卸载！")
	}

	if len(os.Args) <= 1 {
		glog.Error("参数错误，请重新配置参数")
		return errors.New("参数错误，请重新配置参数")
	}
	if strings.Compare(os.Args[1], "upgrade") != 0 {
		glog.Error("参数错误，请重新配置参数")
		return errors.New("参数错误，请重新配置参数")
	}

	util.SetFirewall(this.conf.Name, this.conf.Executable)
	err = util.SetRLimit()
	if err != nil {
		glog.Error(err)
	}

	glog.Debug("准备升级...")
	err = gore.Update(this.srv, this.conf.Executable, fileUrlOrLocalPath)
	if err != nil {
		glog.Error(err)
		return err
	}
	err = os.Chmod(this.conf.Executable, 0755)
	if err == nil {
		glog.Debug(this.conf.Executable, "赋予0755权限成功")
	} else {
		glog.Error(this.conf.Executable, "赋予0755权限失败", err)
	}

	err = this.daemon.Install() //.Control("install", this.binPath, []string{"-d"})
	if err == nil {
		glog.Println("服务升级成功!")
	} else {
		glog.Println("服务升级失败，错误信息:", err)
	}
	time.Sleep(time.Second * 1)
	if utils.IsWindows() {
		err = this.daemon.Start()
	} else {
		err = this.daemon.Restart()
	}
	if err != nil {
		glog.Println("服务启动失败，错误信息:", err)
	} else {
		glog.Println("服务启动成功！")
	}
	return err
}

func (this *gservice) install() error {
	_, err := this.daemon.Status()
	var isRemoved = true
	if err == nil {
		no := utils.InputString(fmt.Sprintf("%s%s%s", "检测到", this.conf.Name, "程序已经安装，卸载/更新/取消?(y/u/n):"))
		switch no {
		case "y", "Y", "Yes", "YES":
			isRemoved = true
			err = this.uninstall()
			if err != nil {
				//return err
				glog.Error(err)
			}
			break
		case "u", "U", "Update", "UPDATE":
			isRemoved = false
			err = this.stopService()
			if err != nil {
				//return err
			}
			break
		default:
			glog.Debug("结束安装.")
			time.Sleep(time.Second * 3)
			os.Exit(0)
			return err
		}
	} else {
		err := this.daemon.Uninstall()
		if err != nil {
			glog.Printf("服务【%s】卸载失败 %v\n", this.conf.DisplayName, err)
		} else {
			glog.Println("服务成功卸载！")
		}
	}
	util.SetFirewall(this.conf.Name, this.conf.Executable)
	err = util.SetRLimit()
	if err != nil {
		glog.Error(err)
	}

	if _, err := os.Stat(this.workDir); !os.IsNotExist(err) {
		if isRemoved {
			err5 := os.RemoveAll(this.workDir)
			if err5 != nil {
				glog.Error("删除失败", this.workDir)
			} else {
				err = os.MkdirAll(this.workDir, 0775)
				if err != nil {
					glog.Printf("MkdirAll %s error:%s", this.workDir, err)
					return err
				}
			}
		}

	} else {
		err = os.MkdirAll(this.workDir, 0775)
		if err != nil {
			glog.Printf("MkdirAll %s error:%s", this.workDir, err)
			return err
		}
	}

	//这个地方是取的当前运行的执行文件
	currentBinPath, err := os.Executable()
	if err != nil {
		glog.Fatal("os.Executable() error", err)
		return err
	}

	err = os.Chdir(this.workDir)
	if err != nil {
		glog.Println("os.Chdir error:", err)
		return err
	}

	err = gore.Install(this.srv, currentBinPath, this.conf.Executable)
	if err != nil {
		glog.Error(err)
		return err
	}
	err = this.daemon.Install() //.Control("install", this.binPath, []string{"-d"})
	if err == nil {
		glog.Printf("服务【%s】安装成功!\n", this.conf.DisplayName)
	} else {
		glog.Printf("服务【%s】安装失败，错误信息:%v\n", this.conf.DisplayName, err)
	}
	time.Sleep(time.Second * 2)
	err = this.daemon.Start() //Control("start", this.binPath, []string{"-d"})
	if err != nil {
		glog.Printf("服务【%s】启动失败，错误信息:%v\n", this.conf.DisplayName, err)
	} else {
		glog.Printf("服务【%s】启动成功!\n", this.conf.DisplayName)
	}
	return err

}
func (this *gservice) uninstall() error {
	if this.daemon.IsRunning() {
		err := this.daemon.Stop() //.Control("stop", "", nil)
		if err != nil {           // service maybe not install
			glog.Printf("服务【%s】未运行 %v\n", this.conf.DisplayName, err)
			//return err
		}
	}
	_, err := this.daemon.Status()
	if err != nil {
		glog.Printf("服务【%s】未运行 %v\n", this.conf.DisplayName, err)
		//return err
	}
	err = this.daemon.Uninstall() //Control("uninstall", "", nil)
	if err != nil {
		glog.Printf("服务【%s】卸载失败 %v\n", this.conf.DisplayName, err)
	} else {
		glog.Printf("服务【%s】成功卸载\n", this.conf.DisplayName)
	}
	//os.Remove(this.binPath + "0")
	//os.Remove(this.binPath)
	// 尝试删除自身
	utils.DeleteAll(this.tempDir, "临时文件夹")
	utils.DeleteAll(utils.GetUpgradeDir(), "升级文件夹")
	glog.Println("尝试删除自身:", this.workDir)
	if err := os.RemoveAll(this.workDir); err != nil {
		fmt.Printf("Error removing executable: %v\n", err)
		time.Sleep(time.Second * 3)
		os.Exit(1)
	} else {
		glog.Println("尝试删除成功")
	}

	return err
}
func (this *gservice) startService() error {
	glog.Println("startService")
	defer glog.Println("startService end")
	err := this.daemon.Start() //Control("start", "", nil)
	if err != nil {
		glog.Println("start system service error:", err)
	} else {
		glog.Println("start system service ok.")
	}
	return err
}
func (this *gservice) stopService() error {
	glog.Println("stopService")
	defer glog.Println("stopService end")
	err := this.daemon.Stop() //.Control("stop", "", nil)
	if err != nil {
		glog.Println("stop system service error:", err)
	} else {
		glog.Println("stop system service ok.")
	}
	return err
}
func (this *gservice) restart() error {
	defer glog.Println("restart end")
	glog.Println("重启...")
	err := this.daemon.Restart() //Control("restart", "", nil)
	if err != nil {
		glog.Printf("服务【%s】重启失败，错误信息：%v\n", this.conf.DisplayName, err)
	} else {
		glog.Printf("服务【%s】重启成功\n", this.conf.DisplayName)
	}

	return err
}
