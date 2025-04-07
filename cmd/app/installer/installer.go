package main

import (
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore"
	"github.com/xxl6097/go-service/gservice/gore/util"
	"github.com/xxl6097/go-service/gservice/utils"
	"github.com/xxl6097/go-service/pkg"
	"os"
	"os/exec"
	"path/filepath"
)

type BinInfo struct {
	BinPath string
}

type Installer struct {
	service gore.IGService
}

func (t Installer) OnInit() *service.Config {
	arr := t.menu()
	return &service.Config{
		Name:        arr[0],
		DisplayName: fmt.Sprintf("%s %s", arr[2], pkg.AppVersion),
		Description: arr[3],
	}
}

func (t Installer) OnVersion() string {
	pkg.Version()
	return pkg.AppVersion
}

func (t Installer) GetAny(binDir string) any {
	appName := glog.GetNameByPath(os.Args[1])
	ext := filepath.Ext(os.Args[1])
	dstBinPath := filepath.Join(binDir, appName, ext)
	err := os.Rename(os.Args[1], dstBinPath)
	if err != nil {
		fmt.Println("移动失败:", err)
		return err
	}
	fmt.Println(binDir)
	return &BinInfo{binDir}
}

func (t Installer) OnRun(service gore.IGService) error {
	t.service = service
	executable := ""
	arg := make([]string, 0)
	arg = append(arg, "upgrade")
	cmd := exec.Command(executable, arg...)
	util.SetPlatformSpecificAttrs(cmd)
	glog.Printf("运行进程 %s %v\n", executable, arg)
	err := cmd.Start()
	cmd.Wait()
	fmt.Println(err)
	return err
}

func (t Installer) menu() []string {
	if len(os.Args) > 1 {
		panic("请输入可执行文件路径")
	}
	if !utils.IsPathExist(os.Args[1]) {
		panic(fmt.Sprintf("无效文件路径:%s", os.Args[1]))
	}
	appName := glog.GetNameByPath(os.Args[1])
	appName = utils.InputStringEmpty(fmt.Sprintf("请输入应用名称(%s)：", appName), appName)
	displayName := utils.InputStringEmpty(fmt.Sprintf("请输入应用显示名(%s)：", appName), appName)
	describe := utils.InputStringEmpty(fmt.Sprintf("请输入应用描述(%s)：", appName), appName)
	return []string{appName, displayName, describe}
}
