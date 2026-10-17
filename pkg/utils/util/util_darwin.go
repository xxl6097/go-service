package util

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// DefaultInstallPath macOS 下安装到当前用户的 ~/Library/Application Support/<MarketName>。
// 系统级的 /usr/local 属主是 root:wheel 0755，非 root 进程既建不了目录也写不进去，
// 安装会卡在 EnsureDir 的 permission denied。
var DefaultInstallPath = defaultInstallPath()

// defaultInstallPath 取不到 $HOME 时退回系统路径。
func defaultInstallPath() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, "Library", "Application Support", MarketName)
	}
	return "/usr/local/" + MarketName
}

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
		// ⚠️ 这里只能设 Setsid，**不能再设 Setpgid**。
		//
		// Go 在 darwin 上按「先 setsid、后 setpgid」的顺序设置（见
		// runtime 的 syscall/exec_bsd.go）。setsid 成功后子进程已经是新会话
		// 的首进程（session leader），而 POSIX 规定会话首进程不能再改进程组
		// —— 随后那句 setpgid(0,0) 会失败并返回 EPERM，exec 直接起不来，
		// 报错就是："fork/exec <path>: operation not permitted"。
		//
		// 这个错最坑的地方在于它跟被执行的程序完全无关：升级其实已经成功
		// （二进制已替换），只是随后的「重启子进程」失败，用户看到的是
		// 「升级失败」。单设 Setsid 已能脱离控制终端，进程组随新会话自然
		// 成立，不需要再 setpgid（Linux 侧同样只设了 Setsid）。
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true, // 创建新会话，脱离终端
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
