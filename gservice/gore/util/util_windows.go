package util

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"os/exec"
	"strings"
	"syscall"
)

const (
	DefaultInstallPath = "C:\\Program Files"
	//	defaultBinName     = "AAServiceApp.exe"
)

func SetPlatformSpecificAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP, // 新建进程组
		HideWindow:    true,                             // 隐藏窗口（后台运行）
	}
}

func getOsName() (osName string) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return
	}
	defer k.Close()
	pn, _, err := k.GetStringValue("ProductName")
	if err == nil {
		osName = pn
	}
	return
}

func SetRLimit() error {
	return nil
}

func SetFirewall(ProductName, fullPath string) {
	//fullPath, err := filepath.Abs(os.Args[0])
	//if err != nil {
	//	fmt.Println("add firewall error:", err)
	//	return
	//}
	isXP := false
	osName := getOsName()
	if strings.Contains(osName, "XP") || strings.Contains(osName, "2003") {
		isXP = true
	}
	if isXP {
		exec.Command("cmd.exe", `/c`, fmt.Sprintf(`netsh firewall del allowedprogram "%s"`, fullPath)).Run()
		exec.Command("cmd.exe", `/c`, fmt.Sprintf(`netsh firewall add allowedprogram "%s" "%s" ENABLE`, ProductName, fullPath)).Run()
	} else { // win7 or later
		exec.Command("cmd.exe", `/c`, fmt.Sprintf(`netsh advfirewall firewall del rule name="%s"`, ProductName)).Run()
		exec.Command("cmd.exe", `/c`, fmt.Sprintf(`netsh advfirewall firewall add rule name="%s" dir=in action=allow program="%s" enable=yes`, ProductName, fullPath)).Run()
	}
}
