package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// 判断是否为常见可执行扩展名
func IsWindowsExecutable(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".exe" || ext == ".bat" || ext == ".cmd" || ext == ".ps1"
}

// 补充文件头签名验证（以PE文件为例）
func IsPEExecutable(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 2)
	if _, err := file.Read(header); err != nil {
		return false
	}
	return string(header) == "MZ" // DOS头部魔数[4](@ref)
}

// 判断文件是否对任意用户可执行
func IsUnixExecutable(mode os.FileMode) bool {
	return mode&0111 != 0 // 0111 表示 owner/group/other 任一有执行权限[1,2](@ref)
}

// 使用示例
func CheckExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() { // 排除目录
		return false
	}
	return IsUnixExecutable(info.Mode())
}
func IsExecutableFile(path string) bool {
	if runtime.GOOS == "windows" {
		return IsWindowsExecutable(path) && IsPEExecutable(path)
	} else {
		return CheckExecutable(path)
	}
}
