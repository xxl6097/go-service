package utils

//import (
//	"bytes"
//	"debug/elf"
//	"debug/macho"
//	"debug/pe"
//	"fmt"
//	"io"
//	"os"
//	"os/exec"
//	"os/user"
//	"path/filepath"
//	"runtime"
//	"strings"
//	"syscall"
//)
//
//// IsRunnableOnSystem 判断文件是否适合当前系统运行
//func IsRunnableOnSystem(filePath string) (bool, string, error) {
//	// 检查文件是否存在
//	if _, err := os.Stat(filePath); os.IsNotExist(err) {
//		return false, "文件不存在", err
//	}
//
//	// 打开文件进行检查
//	file, err := os.Open(filePath)
//	if err != nil {
//		return false, "无法打开文件", err
//	}
//	defer file.Close()
//
//	// 读取文件前32字节用于魔数检查
//	magic := make([]byte, 32)
//	n, err := file.Read(magic)
//	if err != nil && err != io.EOF {
//		return false, "读取文件失败", err
//	}
//	magic = magic[:n]
//
//	// 获取当前系统信息
//	currentOS := runtime.GOOS
//	currentArch := runtime.GOARCH
//
//	// 根据魔数判断文件类型并检查兼容性
//	switch {
//	case bytes.HasPrefix(magic, []byte{0x7F, 'E', 'L', 'F'}): // ELF文件
//		return checkELF(filePath, currentOS, currentArch)
//	case bytes.HasPrefix(magic, []byte{'M', 'Z'}): // PE/COFF文件 (Windows)
//		return checkPE(filePath, currentOS)
//	case bytes.HasPrefix(magic, []byte{'\xFE', '\xED', '\xFA', '\xCE'}) ||
//		bytes.HasPrefix(magic, []byte{'\xCE', '\xFA', '\xED', '\xFE'}) ||
//		bytes.HasPrefix(magic, []byte{'\xFE', '\xED', '\xFA', '\xCF'}) ||
//		bytes.HasPrefix(magic, []byte{'\xCF', '\xFA', '\xED', '\xFE'}): // Mach-O文件
//		return checkMachO(filePath, currentOS, currentArch)
//	case isShellScript(magic): // 脚本文件
//		return checkScript(filePath, currentOS)
//	default:
//		// 尝试使用系统命令检查
//		return checkWithSystemCommand(filePath, currentOS)
//	}
//}
//
//// 检查ELF文件兼容性
//func checkELF(filePath, currentOS, currentArch string) (bool, string, error) {
//	// 非Linux系统不支持ELF
//	if currentOS != "linux" && currentOS != "freebsd" && currentOS != "openbsd" && currentOS != "netbsd" {
//		return false, "当前系统不支持ELF可执行文件", nil
//	}
//
//	f, err := elf.Open(filePath)
//	if err != nil {
//		return false, "无法解析ELF文件", err
//	}
//	defer f.Close()
//
//	// 检查架构兼容性
//	archMatch := false
//	switch f.Machine {
//	case elf.EM_X86_64:
//		archMatch = currentArch == "amd64"
//	case elf.EM_386:
//		archMatch = currentArch == "386"
//	case elf.EM_ARM:
//		archMatch = currentArch == "arm"
//	case elf.EM_AARCH64:
//		archMatch = currentArch == "arm64"
//	}
//
//	if !archMatch {
//		return false, fmt.Sprintf("ELF架构(%v)与当前系统架构(%s)不匹配", f.Machine, currentArch), nil
//	}
//
//	// 检查操作系统ABI
//	if f.OSABI != elf.ELFOSABI_NONE && f.OSABI != elf.ELFOSABI_LINUX {
//		return false, fmt.Sprintf("ELF ABI(%v)与当前系统不兼容", f.OSABI), nil
//	}
//
//	// 检查文件权限
//	return hasExecutePermission(filePath)
//}
//
//// 检查PE文件兼容性
//func checkPE(filePath, currentOS string) (bool, string, error) {
//	// 非Windows系统不支持PE
//	if currentOS != "windows" {
//		return false, "当前系统不支持Windows可执行文件", nil
//	}
//
//	f, err := pe.Open(filePath)
//	if err != nil {
//		return false, "无法解析PE文件", err
//	}
//	defer f.Close()
//
//	// 检查子系统类型
//	if f.OptionalHeader != nil {
//		winHeader, ok := f.OptionalHeader.(*pe.OptionalHeader64)
//		if !ok {
//			winHeader32, ok32 := f.OptionalHeader.(*pe.OptionalHeader32)
//			if !ok32 {
//				return false, "无法确定PE文件子系统类型", nil
//			}
//			if winHeader32.Subsystem != pe.SubsystemWindowsCUI && winHeader32.Subsystem != pe.SubsystemWindowsGUI {
//				return false, "不支持的PE子系统类型", nil
//			}
//		} else {
//			if winHeader.Subsystem != pe.SubsystemWindowsCUI && winHeader.Subsystem != pe.SubsystemWindowsGUI {
//				return false, "不支持的PE子系统类型", nil
//			}
//		}
//	}
//
//	// 检查架构兼容性
//	archMatch := false
//	switch f.Machine {
//	case pe.IMAGE_FILE_MACHINE_AMD64:
//		archMatch = runtime.GOARCH == "amd64"
//	case pe.IMAGE_FILE_MACHINE_I386:
//		archMatch = runtime.GOARCH == "386"
//	case pe.IMAGE_FILE_MACHINE_ARM64:
//		archMatch = runtime.GOARCH == "arm64"
//	}
//
//	if !archMatch {
//		return false, fmt.Sprintf("PE架构(%v)与当前系统架构(%s)不匹配", f.Machine, runtime.GOARCH), nil
//	}
//
//	return true, "", nil
//}
//
//// 检查Mach-O文件兼容性
//func checkMachO(filePath, currentOS, currentArch string) (bool, string, error) {
//	// 非macOS系统不支持Mach-O
//	if currentOS != "darwin" {
//		return false, "当前系统不支持macOS可执行文件", nil
//	}
//
//	f, err := macho.Open(filePath)
//	if err != nil {
//		return false, "无法解析Mach-O文件", err
//	}
//	defer f.Close()
//
//	// 检查架构兼容性
//	archMatch := false
//	switch f.Cpu {
//	case macho.Cpu386:
//		archMatch = currentArch == "386"
//	case macho.CpuAmd64:
//		archMatch = currentArch == "amd64"
//	case macho.CpuArm:
//		archMatch = currentArch == "arm"
//	case macho.CpuArm64:
//		archMatch = currentArch == "arm64"
//	}
//
//	if !archMatch {
//		return false, fmt.Sprintf("Mach-O架构(%v)与当前系统架构(%s)不匹配", f.Cpu, currentArch), nil
//	}
//
//	return true, "", nil
//}
//
//// 检查脚本文件
//func checkScript(filePath, currentOS string) (bool, string, error) {
//	// 读取文件前128字节检查shebang
//	file, err := os.Open(filePath)
//	if err != nil {
//		return false, "无法打开文件", err
//	}
//	defer file.Close()
//
//	shebang := make([]byte, 128)
//	n, err := file.Read(shebang)
//	if err != nil && err != io.EOF {
//		return false, "读取文件失败", err
//	}
//	shebang = shebang[:n]
//
//	// 检查shebang行
//	if bytes.HasPrefix(shebang, []byte{'#'}) {
//		// 提取解释器路径
//		shebangLine := string(shebang[:bytes.IndexByte(shebang, '\n')])
//		if !strings.HasPrefix(shebangLine, "#!") {
//			return false, "无效的shebang行", nil
//		}
//
//		interpreter := strings.TrimSpace(shebangLine[2:])
//		if interpreter == "" {
//			return false, "缺少解释器路径", nil
//		}
//
//		// 检查解释器是否存在
//		if currentOS != "windows" {
//			// 在Unix-like系统上检查解释器是否存在于PATH中
//			_, err := exec.LookPath(filepath.Base(interpreter))
//			if err != nil {
//				return false, fmt.Sprintf("解释器'%s'不存在或不可执行", interpreter), err
//			}
//		} else {
//			// 在Windows上，检查是否有兼容的shell
//			shells := []string{"cmd.exe", "powershell.exe", "bash.exe"}
//			found := false
//			for _, shell := range shells {
//				_, err := exec.LookPath(shell)
//				if err == nil {
//					found = true
//					break
//				}
//			}
//			if !found {
//				return false, "未找到兼容的shell解释器", nil
//			}
//		}
//
//		// 检查文件权限
//		return hasExecutePermission(filePath)
//	}
//
//	return false, "不是有效的脚本文件", nil
//}
//
//// 检查文件是否有可执行权限
//func hasExecutePermission(filePath string) (bool, string, error) {
//	info, err := os.Stat(filePath)
//	if err != nil {
//		return false, "无法获取文件信息", err
//	}
//
//	// Windows系统不依赖文件权限位
//	if runtime.GOOS == "windows" {
//		return true, "", nil
//	}
//
//	// 检查文件权限位
//	mode := info.Mode()
//	if mode&0111 != 0 {
//		return true, "", nil
//	}
//
//	// 检查当前用户是否为文件所有者并拥有执行权限
//	currentUser, err := user.Current()
//	if err != nil {
//		return false, "无法获取当前用户信息", err
//	}
//
//	fileOwner, err := user.LookupId(fmt.Sprintf("%d", info.Sys().(*syscall.Stat_t).Uid))
//	if err != nil {
//		return false, "无法获取文件所有者信息", err
//	}
//
//	if currentUser.Uid == fileOwner.Uid && (mode&0100 != 0) {
//		return true, "", nil
//	}
//
//	return false, "文件没有可执行权限", nil
//}
//
//// 判断是否为shell脚本
//func isShellScript(magic []byte) bool {
//	return bytes.HasPrefix(magic, []byte("#!"))
//}
//
//// 使用系统命令检查
//func checkWithSystemCommand(filePath, currentOS string) (bool, string, error) {
//	if currentOS == "windows" {
//		// Windows上检查文件扩展名
//		ext := strings.ToLower(filepath.Ext(filePath))
//		validExts := map[string]bool{
//			".exe":  true,
//			".com":  true,
//			".bat":  true,
//			".cmd":  true,
//			".ps1":  true,
//			".vbs":  true,
//			".js":   true, // 假设存在Node.js环境
//			".py":   true, // 假设存在Python环境
//			".rb":   true, // 假设存在Ruby环境
//			".sh":   true, // 假设存在Bash环境
//			".bat":  true,
//			".cmd":  true,
//			".vbs":  true,
//			".wsf":  true,
//			".msc":  true,
//			".cpl":  true,
//			".scr":  true,
//			".msp":  true,
//			".appx": true,
//		}
//
//		if validExts[ext] {
//			return true, "", nil
//		}
//
//		return false, "未知的Windows可执行文件类型", nil
//	}
//
//	// Unix-like系统使用file命令检查MIME类型
//	cmd := exec.Command("file", "--brief", "--mime-type", filePath)
//	output, err := cmd.Output()
//	if err != nil {
//		return false, "无法确定文件类型", err
//	}
//
//	mimeType := strings.TrimSpace(string(output))
//	if strings.HasPrefix(mimeType, "application/x-executable") ||
//		strings.HasPrefix(mimeType, "application/x-sharedlib") ||
//		strings.HasPrefix(mimeType, "text/x-script.") {
//		return true, "", nil
//	}
//
//	return false, fmt.Sprintf("不支持的文件类型: %s", mimeType), nil
//}
