package utils

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/pkg/utils/util"
	"io"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

func RestartWindowsApplication() {
	if IsWindows() {
		// Windows特有重启逻辑
		exe, _ := os.Executable()
		cmd := exec.Command("cmd", "/C", "timeout 2 && "+exe)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			glog.Error(err)
		}
		os.Exit(0)
	}
}

func BlockingFunction[T any](c context.Context, timeout time.Duration, callback func() T) (T, error) {
	ctx, cancel := context.WithTimeout(c, timeout)
	defer cancel()
	resultChan := make(chan T)
	go func() {
		result := callback()
		resultChan <- result
	}()
	var zero T // 声明 T 的零值
	select {
	case res := <-resultChan:
		return res, nil
	case <-ctx.Done():
		return zero, errors.New("timeout")
	}
}
func DynamicSelect[T any](t []T, fun func(context.Context, int, T) T) T {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan T, len(t)) // 缓冲大小等于协程数量
	var wg sync.WaitGroup
	for i, v := range t {
		wg.Add(1)
		go func(ct context.Context, index int, t T, c chan<- T) {
			defer wg.Done()
			c <- fun(ct, index, t)
		}(ctx, i, v, ch)
	}
	var x T
	for i := 0; i < len(t); i++ {
		_, value, ok := reflect.Select([]reflect.SelectCase{{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}})
		r := value.Interface().(T)
		if ok {
			cancel()
			wg.Wait()
			return r
		}
	}
	cancel()
	return x
}
func IsWindows() bool {
	if strings.Compare(runtime.GOOS, "windows") == 0 {
		return true
	}
	return false
}
func IsOpenWRT() bool {
	_, err := os.Stat("/etc/openwrt_release")
	if err == nil {
		return true
	}
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return false
	}
	src := strings.ToLower(string(data))
	dst := strings.ToLower("OpenWRT")
	return strings.Contains(src, dst)
}

func RunChildProcess(executable string, args ...string) error {
	//args = append([]string{executable}, args...)
	//cmd := exec.Command("sudo", args...)
	var cmd *exec.Cmd
	if !IsWindows() {
		arg := []string{executable}
		arg = append(arg, args...)
		cmd = exec.Command("sudo", arg...)
	} else {
		cmd = exec.Command(executable, args...)
	}
	cmd = exec.Command(executable, args...)
	util.SetPlatformSpecificAttrs(cmd)
	glog.Printf("运行子进程 %s %v\n", executable, args)
	return cmd.Start()
}

func RunCmdBySelf(name string, args ...string) error {
	defer glog.Flush()
	//binpath, err := os.Executable()
	//if err != nil {
	//	return err
	//}
	glog.Debug("RunCmdBySelf", name, args)
	err := RunChildProcess(name, args...)
	if err != nil {
		glog.Errorf("RunChildProcess错误: %v\n", err)
		return fmt.Errorf("RunChildProcess错误: %v\n", err)
	}
	glog.Println("子进程启动成功", name)
	return err
}

func RunCmdWithSudo(args ...string) ([]byte, error) {
	glog.Debug("run", args)
	cmd := exec.Command("sudo", args...)
	output, err := cmd.CombinedOutput() // 捕获标准输出和错误
	if err != nil {
		return nil, err
	}
	fmt.Println(string(output)) // 输出：hello world
	return output, err

	//cmd := exec.Command("sudo", args...)
	//cmd.Stdin = os.Stdin
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	//// 执行命令
	//return cmd.Run()
}

func RunCmd(name string, args ...string) string {
	output, err := Cmd(name, args...)
	if err != nil {
		return err.Error()
	}
	return string(output)
}

func Cmd(name string, args ...string) ([]byte, error) {
	glog.Debug("run", name, args)
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput() // 捕获标准输出和错误
	if err != nil {
		return nil, err
	}
	return output, err
}

func RunWithSudo() error {
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

// PrintByteArrayAsConstant 把字节数组以常量字节数组的形式打印出来
func PrintByteArrayAsConstant(bytes []byte) string {
	sb := strings.Builder{}
	sb.WriteString("[]byte{")
	for i, b := range bytes {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("0x%02X", b))
	}
	sb.WriteString("}")
	return sb.String()
}

func ExitAnyKey() {
	fmt.Print("按回车键退出程序...")
	endKey := make([]byte, 1)
	_, _ = os.Stdin.Read(endKey) // 等待用户输入任意内容后按回车
	os.Exit(0)
}

func ExitCountDown(count int) {
	for i := count; i >= 0; i-- {
		fmt.Printf("\r%d秒后退出程序..", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Println()
	os.Exit(0)
}

// GzipCompress 压缩字节数组
func GzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	// 设置压缩级别（BestSpeed, BestCompression, DefaultCompression）
	gz, e := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if e != nil {
		return nil, e
	}
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil { // 必须关闭以写入所有数据
		return nil, err
	}
	return buf.Bytes(), nil
}

// GzipDecompress 解压
func GzipDecompress(compressed []byte) ([]byte, error) {
	r, e := gzip.NewReader(bytes.NewReader(compressed))
	if e != nil {
		return nil, e
	}
	defer r.Close()
	return io.ReadAll(r)
}
