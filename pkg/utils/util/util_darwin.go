package util

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

const (
	DefaultInstallPath = "/usr/local/" + MarketName
	// defaultBinName     = "AAServiceApp"
)

// GetDiskUsage 获取 Unix 系统磁盘使用情况
func GetDiskUsage(path string) (total, used, free uint64, err error) {
	var stat syscall.Statfs_t
	err = syscall.Statfs(path, &stat)
	if err != nil {
		return
	}
	total = stat.Blocks * uint64(stat.Bsize)
	free = stat.Bfree * uint64(stat.Bsize)
	used = total - free
	return
}

func SetPlatformSpecificAttrs(cmd *exec.Cmd) {
	if runtime.GOOS == "darwin" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid:  true, // 创建新会话，脱离终端
			Setpgid: true, // 创建新的进程组
			Pgid:    0,    // 子进程成为进程组领导者
			// 或者使用 Setsid: true 创建新会话（类似 nohup）
		}
		// 重定向输入输出（避免挂起）
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
	}
}

func execOutput(name string, args ...string) string {
	cmdGetOsName := exec.Command(name, args...)
	var cmdOut bytes.Buffer
	cmdGetOsName.Stdout = &cmdOut
	cmdGetOsName.Run()
	return cmdOut.String()
}
func getOsName() (osName string) {
	//fmt.Println(AppConfig.ProductName)
	output := execOutput("sw_vers", "-productVersion")
	osName = "Mac OS X " + strings.TrimSpace(output)
	return
}

func SetRLimit() error {
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}
	limit.Cur = 65536
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}
	return nil
}

func SetFirewall(ProductName, fullPath string) {
}
