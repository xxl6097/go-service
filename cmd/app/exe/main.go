package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/h2non/filetype"
	"github.com/xxl6097/go-service/cmd/app/exe/arch"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func IsWindowsExecutable(path string) bool {
	ext := filepath.Ext(path)
	switch strings.ToLower(ext) {
	case ".exe", ".bat", ".cmd", ".ps1":
		return true
	default:
		return false // 参考网页[5]
	}
}
func IsUnixExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	// 检查权限位是否包含任意执行权限（--x--x--x）
	return info.Mode()&0111 != 0 // 参考网页[2]
}
func CheckFileArchitecture(path string) (string, error) {
	file, _ := os.Open(path)
	defer file.Close()

	header := make([]byte, 4)
	if _, err := file.Read(header); err != nil {
		return "", err
	}

	//fmt.Println(path, string(header))
	// ELF 文件（Linux/Unix）
	if bytes.HasPrefix(header, []byte{0x7F, 'E', 'L', 'F'}) {
		switch header[3] { // 架构标识位
		case 0x01:
			return "x86", nil
		case 0x02:
			return "x86_64", nil
		case 0x28:
			return "ARM64", nil
		}
	}

	// PE 文件（Windows）
	if bytes.HasPrefix(header, []byte{'M', 'Z'}) {
		file.Seek(0x3C, 0) // 定位到PE头偏移量
		peOffset := make([]byte, 4)
		file.Read(peOffset)
		file.Seek(int64(binary.LittleEndian.Uint32(peOffset)), 0)
		peSig := make([]byte, 4)
		file.Read(peSig)
		if bytes.Equal(peSig, []byte{'P', 'E', 0, 0}) {
			return "x86_64", nil // 简化处理，实际需解析更详细信息
		}
	}
	return "", fmt.Errorf("未知文件类型")
}
func CurrentArch() string {
	return runtime.GOARCH // 返回如 "amd64", "arm64" 等
}

// 示例：ELF 文件架构解析
func parseELFArch1(header []byte) string {
	if len(header) < 5 {
		return "unknown"
	}
	switch header[4] {
	case 0x01:
		return "x86"
	case 0x02:
		return "x86_64"
	case 0xB7:
		return "ARM64"
	default:
		return "unknown"
	}
}

// DetectFileOS 返回文件适用的操作系统类型（如 "linux", "windows", "darwin"）
func DetectFileOS(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 读取前 4 字节判断魔数
	header := make([]byte, 4)
	if _, err := file.Read(header); err != nil {
		return "", err
	}

	//fileArch := parseELFArch(header) // 需补充完整文件头读取逻辑
	//fmt.Println(fileArch, path)
	switch {
	case bytes.HasPrefix(header, []byte{0x7F, 'E', 'L', 'F'}):
		return "linux", nil
	case bytes.HasPrefix(header, []byte{'M', 'Z'}):
		// 验证 PE 头
		file.Seek(0x3C, 0) // 定位 PE 头偏移量
		peOffsetBytes := make([]byte, 4)
		if _, err := file.Read(peOffsetBytes); err != nil {
			return "", err
		}
		peOffset := binary.LittleEndian.Uint32(peOffsetBytes)
		file.Seek(int64(peOffset), 0)
		peHeader := make([]byte, 4)
		if _, err := file.Read(peHeader); err != nil {
			return "", err
		}
		if bytes.Equal(peHeader, []byte{'P', 'E', 0, 0}) {
			return "windows", nil
		}
	case binary.LittleEndian.Uint32(header) == 0xFEEDFACE ||
		binary.LittleEndian.Uint32(header) == 0xFEEDFACF:
		return "darwin", nil
	}

	// 若未匹配魔数，尝试通过扩展名判断脚本类文件
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".sh", ".bash":
		return "linux", nil
	case ".bat", ".cmd":
		return "windows", nil
	}

	return "unknown", nil
}

func IsCompatibleWithSystem(path string) bool {
	// 步骤1：基础执行权限检查
	if runtime.GOOS == "windows" && !IsWindowsExecutable(path) {
		return false
	} else if runtime.GOOS != "windows" && !IsUnixExecutable(path) {
		return false // 参考网页[2][5]
	}

	// 步骤2：文件架构与系统匹配
	fileArch, err := CheckFileArchitecture(path)
	fmt.Println(path, fileArch, err)
	if err != nil || fileArch != CurrentArch() {
		return false // 参考网页[7]
	}

	// 步骤3：操作系统类型匹配（如Linux文件不能在Windows运行）
	//fileOS := DetectFileOS(path) // 需实现文件OS检测（如ELF为Linux，PE为Windows）
	fileOS, _ := DetectFileOS(path)
	fmt.Println(path, fileOS)
	return fileOS == runtime.GOOS
}
func parseMachOArch(header []byte) string {
	if len(header) < 8 {
		return "unknown"
	}
	cpuType := binary.LittleEndian.Uint32(header[4:8])
	switch cpuType {
	case 0x01000007:
		return "x86_64"
	case 0x0100000C:
		return "ARM64"
	}
	return "unknown"
}
func parsePEArch(file *os.File) string {
	file.Seek(0x3C, 0)
	peOffsetBytes := make([]byte, 4)
	file.Read(peOffsetBytes)
	peOffset := binary.LittleEndian.Uint32(peOffsetBytes)
	file.Seek(int64(peOffset)+4, 0) // 定位到 Machine 字段
	machine := make([]byte, 2)
	file.Read(machine)
	switch binary.LittleEndian.Uint16(machine) {
	case 0x8664:
		return "x86_64"
	case 0x014C:
		return "x86"
	}
	return "unknown"
}
func parseELFArch(header []byte) string {
	if len(header) < 18 {
		return "unknown"
	}
	fmt.Println("header:", string(header[:4]), hex.EncodeToString(header[:5]), hex.EncodeToString(header[16:18]))
	switch header[4] {
	case 0x01:
		return "x86"
	case 0x02:
		return "x86_64"
	}
	switch binary.LittleEndian.Uint16(header[16:18]) {
	case 0x3E:
		return "x86_64"
	case 0xB7:
		return "ARM64"
	}
	return "unknown"
}
func DetectFileArch(path string) string {
	buf, _ := os.ReadFile(path)
	kind, _ := filetype.Match(buf)
	var arch string
	switch kind.Extension {
	case "elf":
		arch = parseELFArch(buf)
		break
	case "exe", "dll":
		file, _ := os.Open(path)
		defer file.Close()
		arch = parsePEArch(file)
		break
	case "macho":
		arch = parseMachOArch(buf)
		break
	}
	fmt.Println("--->", kind.Extension, arch, path)
	return arch
}
func IsArchCompatible(filePath string) bool {
	fileArch := DetectFileArch(filePath)
	currentArch := runtime.GOARCH // 如 "amd64", "arm64"
	return fileArch == currentArch
}
func main() {
	//b := IsExecutableFile(os.Args[1])
	//fmt.Println(b, os.Args[1])

	// 解析符号链接真实路径
	//realPath, _ := filepath.EvalSymlinks(os.Args[1])
	//IsCompatibleWithSystem(realPath) // 参考网页[3]

	//if IsCompatibleWithSystem(os.Args[1]) {
	//	fmt.Println("文件可在当前系统运行", os.Args[1])
	//} else {
	//	fmt.Println("不兼容：架构或系统类型不匹配", os.Args[1])
	//}
	dir := "./release"
	files, err := os.ReadDir(dir)
	if err == nil {
		for _, file := range files {
			binpath := filepath.Join(dir, file.Name())
			//fileOS, _ := DetectFileOS(binpath)
			//fmt.Println(fileOS, binpath)
			info, e := arch.CheckExecutableArchitecture(binpath)
			if e != nil {
				fmt.Println("错误", binpath, e, info)
			} else {
				fmt.Printf("%v %v %s\n", info.Compatible, info, binpath)
			}

			//fmt.Printf("文件格式: %s 目标架构: %s 架构兼容: %v %s\n", info.Format, info.TargetArch, info.Compatible, binpath)

		}
	}
}
