package gservice

import (
	"github.com/xxl6097/go-glog/glog"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func initLog(installPath string) {
	glog.SetLogFile(filepath.Join(installPath, "logs"), "app.log")
	glog.SetMaxSize(1 * 1024 * 1024)
	glog.SetMaxAge(15)
	glog.SetNoHeader(false)
}

// Run func Run(config *service.Config, version string, runner service.Interface) {
func Run(iService IService) {
	defer glog.Flush()
	glog.SetNoHeader(true)
	if iService == nil {
		glog.Error("config is nil")
		return
	}
	if len(os.Args) > 1 {
		k := os.Args[1]
		switch k {
		case "version", "-v", "--version":
			glog.Println(iService.Version())
			return
		}
	}
	binDir := iService.Config().Name
	if !IsWindows() {
		binDir = strings.ToLower(binDir)
	}
	installPath := filepath.Join(defaultInstallPath, binDir)
	rand.Seed(time.Now().UnixNano())
	baseDir := filepath.Dir(os.Args[0])
	os.Chdir(baseDir) // for system service
	//glog.Info("Run...", len(os.Args), os.Args)
	//glog.Infof("config...%+v", iService.Config())
	installer = NewInstaller(iService, installPath)
	if installer == nil {
		glog.Error("installer is nil")
		return
	}
	if len(os.Args) > 1 {
		k := os.Args[1]
		switch k {
		case "install":
			installer.Install()
			return
		case "uninstall":
			installer.Uninstall()
			return
		case "upgrade":
			installer.Upgrade()
			return
		case "start":
			installer.StartService()
		case "stop":
			installer.StopService()
		case "restart":
			installer.Restart()
			return
		case "-d":
			glog.Flush()
			initLog(installPath)
			glog.Println("创建进程..")
			installer.Run()
			glog.Println("进程结束..")
			return
		default:
			iService.Unkown(k, installPath)
		}
	} else {
		//if installer.IsInstalled() {
		//	glog.Flush()
		//	initLog(installPath)
		//	glog.Println("创建进程..")
		//	installer.Run()
		//	glog.Println("进程结束..")
		//} else {
		//	installer.InstallByFilename()
		//}
		installer.InstallByFilename()
	}
}

var installer *Installer

func Uninstall() {
	if installer != nil {
		installer.Uninstall()
	}
}
