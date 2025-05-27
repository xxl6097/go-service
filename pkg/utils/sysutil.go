package utils

import (
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/pkg/utils/util"
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

func RunCmdBySelf(name string, args ...string) error {
	defer glog.Flush()
	//binpath, err := os.Executable()
	//if err != nil {
	//	return err
	//}
	glog.Debug("RunCmdBySelf", name, args)
	err := RunChildProcess(name, args...)
	if err != nil {
		glog.Errorf("RunChildProcess错误: %v\n", err)
		return fmt.Errorf("RunChildProcess错误: %v\n", err)
	}
	glog.Println("子进程启动成功", name)
	return err
}

func RunCmdWithSudo(args ...string) ([]byte, error) {
	glog.Debug("run", args)
	cmd := exec.Command("sudo", args...)
	output, err := cmd.CombinedOutput() // 捕获标准输出和错误
	if err != nil {
		return nil, err
	}
	fmt.Println(string(output)) // 输出：hello world
	return output, err

	//cmd := exec.Command("sudo", args...)
	//cmd.Stdin = os.Stdin
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	//// 执行命令
	//return cmd.Run()
}

func RunCmd(name string, args ...string) string {
	glog.Debug("run", name, args)
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput() // 捕获标准输出和错误
	if err != nil {
		return err.Error()
	}
	return string(output)
}

func RunWithSudo() error {
	if os.Geteuid() == 0 {
		return nil // 已经拥有 root 权限
	}

	// 获取当前可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 构建 sudo 命令
	cmd := exec.Command("sudo", exePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 执行命令
	return cmd.Run()
}

// PrintByteArrayAsConstant 把字节数组以常量字节数组的形式打印出来
func PrintByteArrayAsConstant(bytes []byte) string {
	sb := strings.Builder{}
	sb.WriteString("[]byte{")
	for i, b := range bytes {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("0x%02X", b))
	}
	sb.WriteString("}")
	return sb.String()
}

func ExitAnyKey() {
	fmt.Print("按回车键退出程序...")
	endKey := make([]byte, 1)
	_, _ = os.Stdin.Read(endKey) // 等待用户输入任意内容后按回车
	os.Exit(0)
}

func ExitCountDown(count int) {
	for i := count; i >= 0; i-- {
		fmt.Printf("\r%d秒后退出程序..", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Println()
	os.Exit(0)
}
