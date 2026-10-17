package util

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const (
	DefaultInstallPath = "/usr/local/" + MarketName
	//defaultBinName     = "AAServiceApp"
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
	// ⚠️ 只能设 Setsid，不能再设 Setpgid：Go 在 FreeBSD 上同样是
	// 「先 setsid、后 setpgid」，而 setsid 之后子进程已是会话首进程，
	// 会话首进程不能再改自己的进程组（POSIX），setpgid 会返回 EPERM，
	// exec 起不来并报 "fork/exec <path>: operation not permitted"。
	// 详见 util_darwin.go 里的同一处说明。
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // 创建新会话，脱离终端
	}
	// 重定向输入输出（避免挂起）
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
}

func getOsName() (osName string) {
	var sysnamePath string
	sysnamePath = "/etc/redhat-release"
	_, err := os.Stat(sysnamePath)
	if err != nil && os.IsNotExist(err) {
		str := "PRETTY_NAME="
		f, err := os.Open("/etc/os-release")
		if err == nil {
			buf := bufio.NewReader(f)
			for {
				line, err := buf.ReadString('\n')
				if err == nil {
					line = strings.TrimSpace(line)
					pos := strings.Count(line, str)
					if pos > 0 {
						len1 := len([]rune(str)) + 1
						rs := []rune(line)
						osName = string(rs[len1 : (len(rs))-1])
						break
					}
				} else {
					break
				}
			}
		}
	} else {
		buff, err := ioutil.ReadFile(sysnamePath)
		if err == nil {
			osName = string(bytes.TrimSpace(buff))
		}
	}
	if osName == "" {
		osName = "FreeBSD"
	}
	return
}

func SetRLimit() error {
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}
	limit.Max = 65536
	limit.Cur = limit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}
	return nil
}

func SetFirewall(ProductName, fullPath string) {
}
