package main

import (
	"fmt"
	"os"
	"os/exec"
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
	if err := runWithSudo(); err != nil {
		fmt.Printf("获取管理员权限失败: %v\n", err)
		os.Exit(1)
	}

	// 执行需要管理员权限的操作
	fmt.Println("已获取管理员权限，正在执行敏感操作...")
	// ...
}
