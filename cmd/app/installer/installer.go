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

type SvrInstall struct {
	service gore.IGService
	cfg     *service.Config
}

func (t *SvrInstall) OnInit() *service.Config {
	arr := t.menu()
	t.cfg = &service.Config{
		Name:        arr[0],
		DisplayName: arr[1],
		Description: arr[2],
	}
	return t.cfg
}

func (t *SvrInstall) OnVersion() string {
	pkg.Version()
	return pkg.AppVersion
}

func (t *SvrInstall) copyCfg(binDir string) {
	if len(os.Args) > 3 {
		args := os.Args[2:]
		var newArgs []string
		for _, arg := range args {
			if utils.IsPathExist(arg) {
				_, name := filepath.Split(arg)
				newCfgPath := filepath.Join(binDir, name)
				_ = utils.Copy(arg, newCfgPath)
				newArgs = append(newArgs, newCfgPath)
			} else {
				newArgs = append(newArgs, arg)
			}
		}
		fmt.Println("newArgs", newArgs)
		fmt.Println("Arguments", t.cfg.Arguments)
		t.cfg.Arguments = newArgs
	}
}

func (t *SvrInstall) copyBin(binDir string) error {
	curBinPath := os.Args[1]
	fmt.Println("curBinPath", curBinPath, utils.IsPathExist(curBinPath))
	appName := glog.GetNameByPath(curBinPath)
	if utils.IsWindows() {
		appName = appName + ".exe"
	}
	dstBinPath := filepath.Join(binDir, appName)
	fmt.Println("dstBinPath", dstBinPath)
	err := utils.Copy(curBinPath, dstBinPath)
	if err != nil {
		fmt.Println("移动失败:", err)
		return err
	}
	return nil
}

func (t *SvrInstall) OnInstall(string, binPath string) error {
	binDir := filepath.Dir(binPath)
	t.copyCfg(binDir)
	return t.copyBin(binDir)
}

//func (t *SvrInstall) GetAny(binDir string) any {
//	appName := glog.GetNameByPath(os.Args[1])
//	ext := filepath.Ext(os.Args[1])
//	dstBinPath := filepath.Join(binDir, appName, ext)
//	err := os.Rename(os.Args[1], dstBinPath)
//	if err != nil {
//		fmt.Println("移动失败:", err)
//		return err
//	}
//	fmt.Println(binDir)
//	return &BinInfo{binDir}
//}

func (t *SvrInstall) OnRun(service gore.IGService) error {
	t.service = service
	executable := ""
	arg := make([]string, 0)
	arg = append(arg, "upgrade")
	cmd := exec.Command(executable, arg...)
	util.SetPlatformSpecificAttrs(cmd)
	glog.Printf("运行进程 %s %v\n", executable, arg)
	err := cmd.Start()
	err = cmd.Wait()
	fmt.Println(err)
	return err
}

func (t *SvrInstall) menu() []string {
	if len(os.Args) < 1 {
		panic("请输入可执行文件路径")
	}
	b := utils.IsPathExist(os.Args[1])
	fmt.Println(os.Args[1], b)
	if !b {
		panic(fmt.Sprintf("无效文件路径:%s", os.Args[1]))
	}

	appName := glog.GetNameByPath(os.Args[1])
	appName = utils.InputStringEmpty(fmt.Sprintf("请输入应用名称(%s)：", appName), appName)
	displayName := utils.InputStringEmpty(fmt.Sprintf("请输入应用显示名(%s)：", appName), appName)
	describe := utils.InputStringEmpty(fmt.Sprintf("请输入应用描述(%s)：", appName), appName)
	return []string{appName, displayName, describe}
}
