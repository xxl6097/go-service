package gservice

import (
	"context"
	"errors"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore"
	"github.com/xxl6097/go-service/gservice/gore/util"
	"github.com/xxl6097/go-service/gservice/utils"
	"github.com/xxl6097/go-service/pkg/ukey"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type gservice struct {
	daemon  *gore.Daemon
	srv     gore.GService
	conf    *service.Config
	workDir string
	dirs    []string
}

func Run(srv gore.GService) error {
	defer glog.Flush()
	bconfig := srv.OnInit()
	if bconfig == nil {
		return fmt.Errorf("请实现OnConfig() *service.Config方法")
	}
	//bconfig.Option = map[string]interface{}{
	//	"RunAtLoad": true,
	//}
	//u, err := user.Current()
	//if err == nil {
	//	bconfig.UserName = u.Username
	//}

	if bconfig.Name == "" {
		return fmt.Errorf("应用名不能为空")
	}
	if len(os.Args) > 1 {
		glog.LogDefaultLogSetting(fmt.Sprintf("%s.log", os.Args[1]))
	} else {
		glog.LogDefaultLogSetting("app.log")
	}
	glog.Debugf("运行参数：%+v", os.Args)
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
	}
	if utils.IsWindows() {
		bconfig.Name = bconfig.Name + ".exe"
	}

	bconfig.Executable = filepath.Join(this.workDir, bconfig.Name)
	binDir := filepath.Dir(os.Args[0])
	_ = os.Chdir(binDir)
	this.dirs = []string{this.workDir, glog.GetCrossPlatformDataDir()}
	core := gore.NewCoreService(srv, this.dirs)
	d, err := gore.NewDaemon(core, bconfig)
	if err != nil {
		return err
	}
	this.daemon = d
	//if this.daemon.IsRunning() {
	//	return this.run()
	//} else {
	//	return srv.OnRun(this)
	//}
	return this.run(srv)
}

func (this *gservice) run(srv gore.GService) error {
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
			err := this.uninstall()
			glog.Debugf("卸载进程: %v,err : %v", os.Getpid(), err)
			for _, dir := range this.dirs {
				e := os.RemoveAll(dir)
				if e != nil {
					glog.Error("删除失败", dir, e)
				}
			}
			return err
		case "upgrade", "update":
			return this.upgrade()
		case "start":
			return this.startService()
		case "stop":
			return this.stopService()
		case "restart":
			return this.restart()
		//case "run":
		//	glog.Printf("运行服务【%s】%v\n", this.conf.DisplayName, this.daemon.IsRunning())
		//	return this.daemon.Run()
		case "r", "run":
			glog.Printf("运行服务【%s】%v\n", this.conf.DisplayName, this.daemon.IsRunning())
			return srv.OnRun(nil)
		}
	}
	if ukey.CanShowMenu() {
		glog.Debug("运行菜单")
		return this.runMenu()
	}
	glog.Printf("运行服务【%s】%v\n", this.conf.DisplayName, this.daemon.IsRunning())
	return this.daemon.Run()
}

func (this *gservice) runMenu() error {
	defer func() {
		//fmt.Print("按回车键退出程序...")
		//endKey := make([]byte, 1)
		//_, _ = os.Stdin.Read(endKey) // 等待用户输入任意内容后按回车
		//fmt.Println("程序已退出")
		utils.Exit()
	}()
	fmt.Println("1. 安装程序")
	fmt.Println("2. 卸载程序")
	fmt.Println("3. 升级程序")
	fmt.Println("4. 重启程序")
	fmt.Println("5. 停止程序")
	fmt.Println("6. 查看版本")
	index := utils.InputInt("请根据菜单选择：")
	switch index {
	case 1:
		return this.install()
	case 2:
		return this.uninstall()
	case 3:
		return this.upgrade()
	case 4:
		return this.restart()
	case 5:
		return this.stopService()
	case 6:
		this.srv.OnVersion()
		break
	default:
		fmt.Println("未知选项", index)
		break
	}
	return nil
}

//func (this *gservice) canMenu() bool {
//	_, err := ukey.Load()
//	if err != nil {
//		return true
//	}
//	return false
//}

//
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

//func (this *gservice) upgrade() error {
//	defer glog.Flush()
//	glog.Debugf(">>>>>>>进入升级流程[%d] %v\n", os.Getpid(), os.Args)
//	if len(os.Args) <= 2 {
//		glog.Error("参数错误，请重新配置参数")
//		return errors.New("参数错误，请重新配置参数")
//	}
//	fileUrlOrLocalPath := os.Args[2]
//	return this.update(fileUrlOrLocalPath)
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
	defer utils.DeleteAll(fileUrlOrLocalPath, "升级文件")

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
	err = gore.Update(this.srv, context.Background(), this.conf.Executable, fileUrlOrLocalPath)
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
			this.daemon.Uninstall()
			//err = this.uninstall()
			//if err != nil {
			//	//return err
			//	glog.Error(err)
			//}
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
		this.daemon.Uninstall()
		//e := this.daemon.Uninstall()
		//if e != nil {
		//	glog.Printf("服务【%s】卸载失败 %v\n", this.conf.DisplayName, e)
		//} else {
		//	glog.Println("服务成功卸载！")
		//}
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

	err = gore.Install(this.srv, currentBinPath, this.conf.Executable)
	if err != nil {
		glog.Error(err)
		return err
	}

	err = os.Chdir(this.workDir)
	if err != nil {
		glog.Println("os.Chdir error:", err)
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
	defer glog.Flush()
	err := this.daemon.Stop() //.Control("stop", "", nil)
	if err != nil {           // service maybe not install
		glog.Printf("服务【%s】未运行 %v\n", this.conf.DisplayName, err)
		_ = glog.Flush()
		//return err
	}
	err = this.daemon.Uninstall() //Control("uninstall", "", nil)
	if err != nil {
		glog.Printf("服务【%s】卸载失败 %v\n", this.conf.DisplayName, err)
		_ = glog.Flush()
	} else {
		glog.Printf("服务【%s】成功卸载\n", this.conf.DisplayName)
		_ = glog.Flush()
	}
	time.Sleep(time.Second * 2)
	// 尝试删除自身
	_ = utils.DeleteAll(glog.GetCrossPlatformDataDir(), "app文件夹")
	glog.Println("尝试删除自身:", this.workDir)
	_ = glog.Flush()
	if err := os.RemoveAll(this.workDir); err != nil {
		fmt.Printf("Error removing executable: %v\n", err)
		_ = glog.Flush()
		time.Sleep(time.Second * 3)
		os.Exit(1)
	} else {
		glog.Println("尝试删除成功")
		_ = glog.Flush()
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
