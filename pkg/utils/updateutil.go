package utils

import (
	"bufio"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-update"
	"os"
	"runtime"
	"strings"
)

func IsMatch(binpath string) error {
	if IsMissMatchOsApp(binpath) {
		return nil
	}
	return fmt.Errorf("安装文件与当前系统不匹配: %s 当前系统：%s/%s", binpath, runtime.GOOS, runtime.GOARCH)
}
func PerformUpdate(newFilePath, targetPath string, patcher bool) error {
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
		//Middler:    IsMatch,
	}
	if patcher {
		opts.Patcher = update.NewBSDiffPatcher()
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

func IsMissMatchOsApp(binPath string) bool {
	if !FileExists(binPath) {
		glog.Error("文件不存在")
		return false
	}
	err := os.Chmod(binPath, 0755)
	if err != nil {
		glog.Error("赋予权限错误", err)
		return false
	}
	o, e := Cmd(binPath, "-v")
	if e != nil {
		glog.Error("cmd运行错误", e)
		return false
	}
	glog.Debug("运行结果", o)
	return true
}

func ExtractCodeBlocks(markdown string) []string {
	var codeBlocks []string
	inCodeBlock := false
	var currentCodeBlock strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(markdown))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				codeBlocks = append(codeBlocks, currentCodeBlock.String())
				currentCodeBlock.Reset()
			}
			inCodeBlock = !inCodeBlock
		} else if inCodeBlock {
			currentCodeBlock.WriteString(line)
			currentCodeBlock.WriteRune('\n')
		}
	}

	return codeBlocks
}
