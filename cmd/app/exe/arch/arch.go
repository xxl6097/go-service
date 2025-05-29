package arch

import (
	"bytes"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Architecture 表示系统架构类型
type Architecture string

// Architecture 表示系统架构类型

const (
	ArchAMD64    Architecture = "amd64"
	Arch386      Architecture = "386"
	ArchARM      Architecture = "arm"
	ArchARM64    Architecture = "arm64"
	ArchPPC64    Architecture = "ppc64"
	ArchPPC64LE  Architecture = "ppc64le"
	ArchS390X    Architecture = "s390x"
	ArchMIPS     Architecture = "mips"
	ArchMIPSLE   Architecture = "mipsle"
	ArchMIPS64   Architecture = "mips64"
	ArchMIPS64LE Architecture = "mips64le"
	ArchLoong64  Architecture = "loong64"
	ArchRiscv64  Architecture = "riscv64"
	ArchUnknown  Architecture = "unknown"
)

// ExecutableFormat 表示可执行文件格式
type ExecutableFormat string

const (
	FormatELF     ExecutableFormat = "ELF"
	FormatPE      ExecutableFormat = "PE"
	FormatMachO   ExecutableFormat = "Mach-O"
	FormatScript  ExecutableFormat = "Script"
	FormatUnknown ExecutableFormat = "Unknown"
)

// FileInfo 包含可执行文件的信息
type FileInfo struct {
	Format       ExecutableFormat
	TargetArch   Architecture
	OsType       string
	Compatible   bool
	ErrorMessage string
}

// CheckExecutableArchitecture 检查可执行文件的架构兼容性
func CheckExecutableArchitecture(filePath string) (FileInfo, error) {
	var info FileInfo
	currentArch := Architecture(runtime.GOARCH)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return info, fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close()

	// 读取文件头部以确定格式
	magic := make([]byte, 8)
	if _, err := file.Read(magic); err != nil {
		return info, fmt.Errorf("无法读取文件头部: %v", err)
	}
	osType, osErr := detectFileOS(filePath, file, magic)
	// 重置文件指针
	if _, err := file.Seek(0, 0); err != nil {
		return info, fmt.Errorf("无法重置文件指针: %v", err)
	}

	info.OsType = osType

	// 确定文件格式并检查架构
	switch {
	case bytes.HasPrefix(magic, []byte{0x7F, 'E', 'L', 'F'}):
		info.Format = FormatELF
		arch, err := checkELFArchitecture(file)
		if err != nil {
			return info, fmt.Errorf("检查ELF架构失败: %v", err)
		}
		info.TargetArch = arch
		info.Compatible = arch == currentArch
		if osErr != nil {
			info.Compatible = false
		} else {
			if strings.Compare(strings.ToLower(osType), strings.ToLower(runtime.GOOS)) != 0 {
				info.Compatible = false
				osErr = fmt.Errorf("当前操作系统:%s/%s,目标文件类型:%s/%s", runtime.GOOS, runtime.GOARCH, osType, arch)
			}
		}
		if !info.Compatible {
			if osErr != nil {
				info.ErrorMessage = osErr.Error()
			} else {
				info.ErrorMessage = fmt.Sprintf("ELF目标架构(%s)与当前系统架构(%s)不匹配 %v", arch, currentArch, osErr)
			}
		}

	case bytes.HasPrefix(magic, []byte{'M', 'Z'}):
		info.Format = FormatPE
		arch, err := checkPEArchitecture(file)
		if err != nil {
			return info, fmt.Errorf("检查PE架构失败: %v", err)
		}
		info.TargetArch = arch
		info.Compatible = arch == currentArch
		if osErr != nil {
			info.Compatible = false
		} else {
			if strings.Compare(strings.ToLower(osType), strings.ToLower(runtime.GOOS)) != 0 {
				info.Compatible = false
				osErr = fmt.Errorf("os type %s is not supported", osType)
			}
		}
		if !info.Compatible {
			if osErr != nil {
				info.ErrorMessage = osErr.Error()
			} else {
				info.ErrorMessage = fmt.Sprintf("PE目标架构(%s)与当前系统架构(%s)不匹配 %v", arch, currentArch, osErr)
			}
		}

	case bytes.HasPrefix(magic, []byte{'\xFE', '\xED', '\xFA', '\xCE'}) ||
		bytes.HasPrefix(magic, []byte{'\xCE', '\xFA', '\xED', '\xFE'}) ||
		bytes.HasPrefix(magic, []byte{'\xFE', '\xED', '\xFA', '\xCF'}) ||
		bytes.HasPrefix(magic, []byte{'\xCF', '\xFA', '\xED', '\xFE'}):
		info.Format = FormatMachO
		arch, err := checkMachOArchitecture(file)
		if err != nil {
			return info, fmt.Errorf("检查Mach-O架构失败: %v", err)
		}
		info.TargetArch = arch
		info.Compatible = arch == currentArch
		if osErr != nil {
			info.Compatible = false
		} else {
			if strings.Compare(strings.ToLower(osType), strings.ToLower(runtime.GOOS)) != 0 {
				info.Compatible = false
				osErr = fmt.Errorf("os type %s is not supported", osType)
			}
		}
		if !info.Compatible {
			if osErr != nil {
				info.ErrorMessage = osErr.Error()
			} else {
				info.ErrorMessage = fmt.Sprintf("Mach-O目标架构(%s)与当前系统架构(%s)不匹配 %v", arch, currentArch, osErr)
			}
		}

	case bytes.HasPrefix(magic, []byte{'#', '!'}):
		info.Format = FormatScript
		arch, err := checkMachOArchitecture(file)
		if err != nil {
			return info, fmt.Errorf("检查Mach-O架构失败: %v", err)
		}
		// 脚本架构取决于解释器，通常是与系统架构一致
		info.TargetArch = currentArch
		info.Compatible = arch == currentArch
		if osErr != nil {
			info.Compatible = false
		} else {
			if strings.Compare(strings.ToLower(osType), strings.ToLower(runtime.GOOS)) != 0 {
				info.Compatible = false
				osErr = fmt.Errorf("os type %s is not supported", osType)
			}
		}
		if !info.Compatible {
			if osErr != nil {
				info.ErrorMessage = osErr.Error()
			} else {
				info.ErrorMessage = fmt.Sprintf("Mach-O目标架构(%s)与当前系统架构(%s)不匹配 %v", arch, currentArch, osErr)
			}
		}

	default:
		info.Format = FormatUnknown
		info.TargetArch = ArchUnknown
		info.Compatible = false
		info.ErrorMessage = "无法识别的文件格式"
	}

	return info, nil
}

// checkELFArchitecture 检查ELF文件的架构
func checkELFArchitecture(file *os.File) (Architecture, error) {
	elfFile, err := elf.Open(file.Name())
	if err != nil {
		return ArchUnknown, err
	}
	defer elfFile.Close()

	switch elfFile.Machine {
	case elf.EM_X86_64:
		return ArchAMD64, nil
	case elf.EM_386:
		return Arch386, nil
	case elf.EM_ARM:
		return ArchARM, nil
	case elf.EM_AARCH64:
		return ArchARM64, nil
	case elf.EM_PPC64:
		if elfFile.Class == elf.ELFCLASS64 && elfFile.Data == elf.ELFDATA2LSB {
			return ArchPPC64LE, nil
		}
		return ArchPPC64, nil
	case elf.EM_S390:
		return ArchS390X, nil
	case elf.EM_LOONGARCH:
		return ArchLoong64, nil
	case elf.EM_MIPS:
		if elfFile.Class == elf.ELFCLASS64 {
			if elfFile.Data == elf.ELFDATA2LSB {
				return ArchMIPS64LE, nil // MIPS64 小端序
			}
			return ArchMIPS64, nil // MIPS64 大端序
		}
		// MIPS32 架构
		if elfFile.Data == elf.ELFDATA2LSB {
			return ArchMIPSLE, nil // MIPS32 小端序
		}
		return ArchMIPS, nil // MIPS32 大端序
	case elf.EM_RISCV:
		return ArchRiscv64, nil
	default:
		return ArchUnknown, fmt.Errorf("不支持的ELF架构: %v", elfFile)
	}
}

// checkPEArchitecture 检查PE文件的架构
func checkPEArchitecture(file *os.File) (Architecture, error) {
	peFile, err := pe.Open(file.Name())
	if err != nil {
		return ArchUnknown, err
	}
	defer peFile.Close()

	switch peFile.Machine {
	case pe.IMAGE_FILE_MACHINE_AMD64:
		return ArchAMD64, nil
	case pe.IMAGE_FILE_MACHINE_I386:
		return Arch386, nil
	case pe.IMAGE_FILE_MACHINE_ARM:
		return ArchARM, nil
	case pe.IMAGE_FILE_MACHINE_ARM64:
		return ArchARM64, nil
	default:
		return ArchUnknown, fmt.Errorf("不支持的PE架构: %v", peFile.Machine)
	}
}

// checkMachOArchitecture 检查Mach-O文件的架构
func checkMachOArchitecture(file *os.File) (Architecture, error) {
	machoFile, err := macho.Open(file.Name())
	if err != nil {
		return ArchUnknown, err
	}
	defer machoFile.Close()

	switch machoFile.Cpu {
	case macho.Cpu386:
		return Arch386, nil
	case macho.CpuAmd64:
		return ArchAMD64, nil
	case macho.CpuArm:
		return ArchARM, nil
	case macho.CpuArm64:
		return ArchARM64, nil
	default:
		return ArchUnknown, fmt.Errorf("不支持的Mach-O架构: %v", machoFile.Cpu)
	}
}

// 判断ELF文件的目标操作系统
func DetectELFOS(filePath string) (string, error) {
	elfFile, err := elf.Open(filePath)
	if err != nil {
		fmt.Println("3----->", err)
		return "", err
	}
	defer elfFile.Close()

	switch elfFile.OSABI {
	case elf.ELFOSABI_NONE, elf.ELFOSABI_LINUX:
		return "Linux", nil
	case elf.ELFOSABI_FREEBSD:
		return "FreeBSD", nil
	case elf.ELFOSABI_SOLARIS:
		return "Solaris", nil
	case elf.ELFOSABI_HPUX:
		return "HP-UX", nil
	default:
		return fmt.Sprintf("Unknown ABI: %d", elfFile.OSABI), nil
	}
}

func detectLinuxOS(filePath string) (string, error) {
	elfFile, err := elf.Open(filePath)
	if err != nil {
		return "", nil
	}
	defer elfFile.Close()
	switch elfFile.OSABI {
	case elf.ELFOSABI_NONE, elf.ELFOSABI_LINUX:
		return "linux", nil
	case elf.ELFOSABI_FREEBSD:
		return "freebsd", nil
	case elf.ELFOSABI_MODESTO:
		return "modesto", nil
	case elf.ELFOSABI_OPENBSD:
		return "openbsd", nil
	case elf.ELFOSABI_OPENVMS:
		return "openvms", nil
	case elf.ELFOSABI_FENIXOS:
		return "fenixos", nil
	case elf.ELFOSABI_CLOUDABI:
		return "cloudabi", nil
	case elf.ELFOSABI_NETBSD:
		return "netbsd", nil
	case elf.ELFOSABI_STANDALONE:
		return "standalone", nil
	case elf.ELFOSABI_SOLARIS:
		return "solaris", nil
	case elf.ELFOSABI_HPUX:
		return "hpux", nil
	default:
		return "unknown", fmt.Errorf("Unknown ABI: %d", elfFile.OSABI)
	}
}

// detectFileOS 返回文件适用的操作系统类型（如 "linux", "windows", "darwin"）
func detectFileOS(path string, file *os.File, header []byte) (string, error) {
	if path == "" {
		return "", errors.New("path is nil")
	}
	if header == nil {
		return "", errors.New("header is nil")
	}
	if file == nil {
		return "", errors.New("file is nil")
	}
	if len(header) < 5 {
		return "", errors.New("header is too short")
	}
	switch {
	case bytes.HasPrefix(header, []byte{0x7F, 'E', 'L', 'F'}):
		//return "linux", nil
		ostype, e := detectLinuxOS(path)
		if e != nil {
			return ostype, e
		}
		return ostype, nil
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

	return "unknown", errors.New("unknown")
}
