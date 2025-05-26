package utils

import (
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore/util"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func RestartWindowsApplication() {
	if IsWindows() {
		// Windows特有重启逻辑
		exe, _ := os.Executable()
		cmd := exec.Command("cmd", "/C", "timeout 2 && "+exe)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			glog.Error(err)
		}
		os.Exit(0)
	}
}

func IsWindows() bool {
	if strings.Compare(runtime.GOOS, "windows") == 0 {
		return true
	}
	return false
}
func IsOpenWRT() bool {
	_, err := os.Stat("/etc/openwrt_release")
	if err == nil {
		return true
	}
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return false
	}
	src := strings.ToLower(string(data))
	dst := strings.ToLower("OpenWRT")
	return strings.Contains(src, dst)
}

func RunChildProcess(executable string, args ...string) error {
	//args = append([]string{executable}, args...)
	//cmd := exec.Command("sudo", args...)
	var cmd *exec.Cmd
	if !IsWindows() {
		arg := []string{executable}
		arg = append(arg, args...)
		cmd = exec.Command("sudo", arg...)
	} else {
		cmd = exec.Command(executable, args...)
	}
	cmd = exec.Command(executable, args...)
	util.SetPlatformSpecificAttrs(cmd)
	glog.Printf("运行子进程 %s %v\n", executable, args)
	return cmd.Start()
}

func RunCmdBySelf(args ...string) error {
	defer glog.Flush()
	binpath, err := os.Executable()
	if err != nil {
		return err
	}
	err = RunChildProcess(binpath, args...)
	if err != nil {
		glog.Errorf("RunChildProcess错误: %v\n", err)
		return fmt.Errorf("RunChildProcess错误: %v\n", err)
	}
	glog.Println("子进程启动成功", binpath)
	return err
}

func ExitAnyKey() {
	fmt.Print("按回车键退出程序...")
	endKey := make([]byte, 1)
	_, _ = os.Stdin.Read(endKey) // 等待用户输入任意内容后按回车
	os.Exit(0)
}

func ExitCountDown() {
	for i := 5; i >= 0; i-- {
		fmt.Printf("\r%d秒后退出程序..", i)
		time.Sleep(1 * time.Second)
	}
	os.Exit(0)
}
