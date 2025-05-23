package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
)

func runWithSudo() error {
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

func main() {
	u, err := user.Current()
	if err != nil {
		fmt.Println("获取用户信息失败:", err)
		return
	}
	fmt.Println("当前系统用户名:", u.Username)
	// ...
}
