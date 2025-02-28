package gservice

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/inconshreveable/go-update"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type gservice struct {
	daemon  *gore.Daemon
	srv     gore.IService
	conf    *service.Config
	svr     service.Service
	workDir string
}

func Run(srv gore.IService) error {
	bpath, err := os.Executable()
	if err == nil {
		workDir := filepath.Dir(bpath)
		binName := filepath.Base(bpath)
		glog.SetLogFile(workDir, "install.log")
		glog.SetNoHeader(true)
		oldFileBinName := fmt.Sprintf(".%s.old", binName)
		if err = os.Remove(oldFileBinName); err != nil {
			//fmt.Printf("remove old file bin err: %s\n", oldFileBinName)
		}
	}
	bconfig := srv.OnConfig()
	if bconfig == nil {
		return fmt.Errorf("请实现OnConfig() *service.Config方法")
	}
	if bconfig.Name == "" {
		return fmt.Errorf("应用名不能为空")
	}
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
		workDir: filepath.Join(gore.DefaultInstallPath, bconfig.Name),
	}
	if gore.IsWindows() {
		bconfig.Name = bconfig.Name + ".exe"
	}

	bconfig.Executable = filepath.Join(this.workDir, bconfig.Name)
	binDir := filepath.Dir(os.Args[0])
	os.Chdir(binDir)
	d, err := gore.NewDaemon(gore.NewCoreService(srv, this), bconfig)
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
		case "upgrade":
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
func restartForWindows() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "restart")
	// 设置进程属性，创建新会话
	if !gore.IsWindows() {
	}
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("Error starting update process: %v\n", err)
	}
	return nil
}

func (this *gservice) Restart() error {
	if gore.IsWindows() {
		return restartForWindows()
	}
	if this.daemon == nil {
		return errors.New("daemon is nil")
	}
	if this.daemon.GetService() == nil {
		return errors.New("service is nil")
	}
	return this.daemon.GetService().Restart()
}

func (this *gservice) update(upgradeBinPath string) error {
	if gore.IsURL(upgradeBinPath) {
		resp, err := http.Get(upgradeBinPath)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		err = update.Apply(resp.Body, update.Options{})
		if err != nil {
			glog.Error(err)
			return err
		}
		return nil
	} else if gore.FileExists(upgradeBinPath) {
		// 打开文件
		file, err := os.Open(upgradeBinPath)
		if err != nil {
			return fmt.Errorf("Error opening file: %v", err)
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
	return fmt.Errorf("位置文件路径:%s", upgradeBinPath)
}

func (this *gservice) Upgrade(upgradeBinPath string) error {
	var err error
	defer func() {
		if err == nil {
			go func() {
				time.Sleep(time.Second)
				err = this.Restart()
				if err != nil {
					glog.Errorf("Error restarting: %v\n", err)
				}
			}()
		}
	}()
	err = this.update(upgradeBinPath)
	return err
}

//	func Upgrade(upgradeBinPath string, args ...string) error {
//		//删除可执行文件
//		if _, err := os.Stat(upgradeBinPath); !os.IsNotExist(err) {
//			if err != nil {
//				glog.Error("文件不存在", upgradeBinPath)
//				return err
//			}
//		}
//		err := os.Chmod(upgradeBinPath, 0755)
//		if err != nil {
//			return fmt.Errorf("赋权限错误: %v\n", err)
//		}
//		a := []string{"upgrade", upgradeBinPath}
//		a = append(a, args...)
//		glog.Println("当前进程ID:", os.Getpid())
//
//		// 启动子进程
//		cmd := exec.Command(upgradeBinPath, a...)
//		// 设置进程属性，创建新会话
//		if !gore.IsWindows() {
//			//cmd.SysProcAttr = &syscall.SysProcAttr{
//			//	Setsid: true,
//			//}
//		}
//		err = cmd.Start()
//		if err != nil {
//			return fmt.Errorf("Error starting update process: %v\n", err)
//		}
//		glog.Println("升级进程启动成功", upgradeBinPath)
//		// 主进程退出
//		defer os.Exit(0)
//		return nil
//	}
func (this *gservice) upgrade() error {
	//fileUrlOrLocalPath := this.srv.OnUpgrade()
	defer glog.Flush()
	glog.Debugf("进入升级流程[%d] %v\n", os.Getpid(), os.Args)
	if len(os.Args) <= 2 {
		glog.Error("参数错误，请重新配置参数")
		return errors.New("参数错误，请重新配置参数")
	}
	fileUrlOrLocalPath := os.Args[2]
	glog.Debug("升级文件地址", fileUrlOrLocalPath)

	//return nil

	if this.daemon.IsRunning() {
		glog.Debug("停止主进程", os.Args)
		glog.Flush()
		err := this.daemon.Stop() //.Control("stop", "", nil)
		if err != nil {           // service maybe not install
			glog.Println("服务停止失败，错误信息：", err)
			return err
		}
	}
	_, err := this.daemon.Status()
	if err == nil {
		err := this.daemon.Uninstall()
		if err != nil {
			glog.Println("服务卸载失败，错误信息：", err)
			return err
		} else {
			glog.Println("服务成功卸载！")
		}
	}

	if len(os.Args) <= 1 {
		glog.Error("参数错误，请重新配置参数")
		return errors.New("参数错误，请重新配置参数")
	}
	if strings.Compare(os.Args[1], "upgrade") != 0 {
		glog.Error("参数错误，请重新配置参数")
		return errors.New("参数错误，请重新配置参数")
	}

	if gore.FileExists(fileUrlOrLocalPath) {
		newPath := fileUrlOrLocalPath

		if _, err := os.Stat(this.conf.Executable); !os.IsNotExist(err) {
			err := os.Remove(this.conf.Executable)
			if err != nil {
				glog.Error("删除失败", this.conf.Executable)
				return err
			}
		}

		err = gore.Copy(newPath, this.conf.Executable)
		if err != nil {
			glog.Error("拷贝失败", err)
			return err
		}

	} else if gore.IsURL(fileUrlOrLocalPath) {
		if _, err := os.Stat(this.conf.Executable); !os.IsNotExist(err) {
			err := os.Remove(this.conf.Executable)
			if err != nil {
				glog.Error("删除失败", this.conf.Executable)
				return err
			}
		}
		glog.Debug("下载文件", fileUrlOrLocalPath)
		err = gore.Download(fileUrlOrLocalPath, this.conf.Executable)
		if err != nil {
			glog.Error("下载失败", err)
			return err
		}
		glog.Debug("下载成功.", this.conf.Executable)
	} else {
		glog.Error("参数错误，请输入正确的URL", fileUrlOrLocalPath)
		return errors.New("参数错误，请输入正确的URL")
	}

	err = os.Chmod(this.conf.Executable, 0755)
	if err == nil {
		glog.Debug(this.conf.Executable, "赋予0755权限成功")
	} else {
		glog.Error(this.conf.Executable, "赋予0755权限失败", err)
	}

	_, args := this.srv.OnInstall(this.workDir)
	err = this.daemon.Install(args) //.Control("install", this.binPath, []string{"-d"})
	if err == nil {
		glog.Println("服务升级成功!")
	} else {
		glog.Println("服务升级失败，错误信息:", err)
	}
	time.Sleep(time.Second * 1)
	err = this.daemon.Start()
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
		no := gore.InputString(fmt.Sprintf("%s%s%s", "检测到", this.conf.Name, "程序已经安装，卸载/更新/取消?(y/u/n):"))
		switch no {
		case "y", "Y", "Yes", "YES":
			isRemoved = true
			err = this.uninstall()
			if err != nil {
				return err
			}
			break
		case "u", "U", "Update", "UPDATE":
			isRemoved = false
			err = this.stopService()
			if err != nil {
				return err
			}
			break
		default:
			glog.Debug("结束安装.")
			time.Sleep(time.Second * 3)
			os.Exit(0)
			return err
		}
	}
	gore.SetFirewall(this.conf.Name, this.conf.Executable)
	err = gore.SetRLimit()
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

	isCopy, args := this.srv.OnInstall(this.conf.Executable)
	//
	if isCopy {
		err = gore.Copy(currentBinPath, this.conf.Executable)
		//err = utils.GenerateBin(currentBinPath, this.binPath, cfg.B, cfg.Size, cfg.CfgBytes)
		if err != nil {
			glog.Printf("文件拷贝失败，错误信息：%s", err)
			return err
		}
	}
	err = this.daemon.Install(args) //.Control("install", this.binPath, []string{"-d"})
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
			glog.Println("卸载失败，错误信息：", this.conf.Name, err)
			return err
		}
	}
	_, err := this.daemon.Status()
	if err != nil {
		glog.Printf("服务【%s】未安装!\n", this.conf.DisplayName)
		return err
	}
	err = this.daemon.Uninstall() //Control("uninstall", "", nil)
	if err != nil {
		glog.Printf("服务【%s】卸载失败，错误信息：%v\n", this.conf.DisplayName, err)
	} else {
		glog.Printf("服务【%s】成功卸载\n", this.conf.DisplayName)
	}
	//os.Remove(this.binPath + "0")
	//os.Remove(this.binPath)
	// 尝试删除自身
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
