package utils

import (
	"bufio"
	"fmt"
	"github.com/inconshreveable/go-update"
	"os"
)

func PerformUpdate(newFilePath, targetPath string) error {
	file, err := os.Open(newFilePath)
	if err != nil {
		return fmt.Errorf("升级文件打开失败【%s】: %v", newFilePath, err)
	}
	defer func() {
		_ = file.Close()
	}()
	// Windows需要管理员权限
	opts := update.Options{
		TargetPath: targetPath, // 当前可执行文件路径
		Patcher:    update.NewBSDiffPatcher(),
	}

	//opts.CheckPermissions()
	//opts := update.Options{
	//	TargetPath: os.Args[0], // 当前可执行文件路径
	//}
	// 使用 bufio.NewReader 创建带缓冲的读取器
	if err = update.Apply(bufio.NewReader(file), opts); err != nil {
		if e := update.RollbackError(err); e != nil {
			return fmt.Errorf("更新失败且无法回滚: %w", e)
		}
		return fmt.Errorf("更新失败: %w", err)
	}
	return nil
}
