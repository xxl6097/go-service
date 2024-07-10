package gservice

import (
	"github.com/xxl6097/go-glog/glog"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func initLog(installPath string) {
	glog.SetLogFile(filepath.Join(installPath, "logs"), "app.log")
	glog.SetMaxSize(1 * 1024 * 1024)
	glog.SetMaxAge(15)
}

// Run func Run(config *service.Config, version string, runner service.Interface) {
func Run(iService IService) {
	if iService == nil {
		glog.Debug("config is nil")
		return
	}
	installPath := filepath.Join(defaultInstallPath, iService.Config().Name)
	rand.Seed(time.Now().UnixNano())
	baseDir := filepath.Dir(os.Args[0])
	os.Chdir(baseDir) // for system service
	glog.Info("Run...", len(os.Args), os.Args)
	glog.Infof("config...%+v", iService.Config())
	installer := NewInstaller(iService, installPath)
	if installer == nil {
		return
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "-v", "--version":
			glog.Println(iService.Version())
			return
		case "install":
			installer.Install()
			return
		case "uninstall":
			installer.Uninstall()
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
		}
	} else {
		installer.InstallByFilename()
	}
}
