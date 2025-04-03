package utils

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

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

// DivideAndCeil 函数用于进行除法并向上取整
func DivideAndCeil(a, b int) int {
	// 将整数转换为 float64 类型进行除法运算
	result := float64(a) / float64(b)
	// 使用 math.Ceil 函数进行向上取整
	result = math.Ceil(result)
	// 将结果转换回整数类型
	return int(result)
}

func Divide(a, b int) int {
	return DivideAndCeil(a, b) * b
}

// ProgressWriter 自定义进度写入器结构体
type ProgressWriter struct {
	TotalSize int64
	Written   int64
	Progress  float64
	Title     string
}

// Write 实现 io.Writer 接口的 Write 方法
func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Written += int64(n)
	// 计算下载进度百分比
	progress := float64(pw.Written) / float64(pw.TotalSize) * 100
	// 使用 \r 覆盖当前行，实现进度动态更新
	if progress >= pw.Progress {
		glog.Printf("%s %.2f%%\n", pw.Title, progress)
		pw.Progress = progress
		pw.Progress += 5
	}
	return n, nil
}

func GetFileNameFromUrl(rawURL string) string {
	parsedURL, _ := url.Parse(rawURL)
	// 提取路径部分并获取文件名
	fileName := path.Base(parsedURL.Path)
	//fmt.Println("文件名:", fileName) // 输出: document.pdf
	return fileName
}

func GetFilenameFromHeader(header http.Header) string {
	contentDisposition := header.Get("Content-Disposition")
	parts := strings.Split(contentDisposition, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "filename=") {
			fileName := strings.TrimPrefix(part, "filename=")
			fileName = strings.Trim(fileName, `"`) // 去除双引号
			return fileName
		}
	}
	return ""
}

func DownLoadBAK(url string, args ...string) (string, error) {
	// 要下载的文件的 URL
	// 发送 HTTP GET 请求
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var dstFile string
	if args != nil && len(args) > 0 {
		dstFile = args[0]
	}
	if dstFile == "" {
		dstName := GetFileNameFromUrl(url)
		if dstName == "" {
			dstName = GetFilenameFromHeader(resp.Header)
		}
		if dstName == "" {
			fileName := time.Now().Unix()
			dstName = fmt.Sprintf("%d", fileName)
		}
		if dstName != "" {
			dstFile = filepath.Join(GetUpgradeDir(), dstName)
		}
	}

	// 获取文件的总大小
	totalSize := resp.ContentLength
	if totalSize == -1 {
		fmt.Println("无法获取文件大小，可能不支持 Content-Length 头信息。")
		return "", fmt.Errorf("无法获取文件大小，可能不支持 Content-Length 头信息。")
	}
	sizeA := float64(resp.ContentLength) / 1024 / 1024
	fmt.Printf("文件大小:%.2fM\n", sizeA)
	// 创建一个本地文件用于保存下载的内容
	file, err := os.Create(dstFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 创建进度写入器实例
	pw := &ProgressWriter{TotalSize: totalSize, Progress: -1, Title: "文件下载："}
	// 将响应体的数据复制到本地文件，并通过 ProgressWriter 跟踪进度
	_, err = io.Copy(io.MultiWriter(file, pw), resp.Body)
	if err != nil {
		return "", fmt.Errorf("下载出错: %v", err)
	}

	fmt.Println("下载完成")
	return dstFile, nil
}

// IsURL 判断给定的字符串是否是一个有效的URL
func IsURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil {
		return false
	}

	return u.Scheme == "http" || u.Scheme == "https"
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
	return strings.Contains(string(data), "OpenWRT")
}

// FileExists 用于判断文件是否存在
func FileExists(filePath string) bool {
	// 调用 os.Stat 函数获取文件信息
	_, err := os.Stat(filePath)
	// 判断是否为文件不存在的错误
	if os.IsNotExist(err) {
		return false
	}
	// 若有其他错误或无错误，认为文件存在
	return true
}

func Copy(srcFile, dstFile string) error {
	src, err := os.Open(srcFile) // can not use args[0], on Windows call openp2p is ok(=openp2p.exe)
	if err != nil {
		fmt.Printf("打开源文件失败：%v\n", err)
		return err
	}
	var fileSize int64
	var fileName string
	finfo, err := src.Stat()
	if err == nil {
		fileSize = finfo.Size()
		fileName = finfo.Name()
	}
	defer src.Close()
	//将本程序复制到目标为止，目标文件名称为配置文件的名称
	dst, err := os.OpenFile(dstFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0775)
	if err != nil {
		fmt.Printf("创建目标文件失败：%v\n", err)
		return err
	}
	defer dst.Close()
	sizeB := float64(fileSize) / 1024 / 1024
	glog.Printf("正在拷贝%s[大小：%.2fMB]到%s\n", fileName, sizeB, dstFile)
	_, err = io.Copy(dst, src)
	if err != nil {
		fmt.Printf("拷贝文件失败：%v\n", err)
		return err
	}
	return nil
}

func tips(title string) {
	str := strings.ReplaceAll(title, "请输入", "")
	str = strings.ReplaceAll(str, "please input", "")
	str = strings.ReplaceAll(str, "：", "")
	str = strings.ReplaceAll(str, ":", "")
	str = fmt.Sprintf("【%s】不允许输入空", str)
	fmt.Println(str)
}
func InputStringEmpty1(title string) string {
	reader := bufio.NewReader(os.Stdin)
	//glog.Print(title)
	fmt.Print(title)
	input, err := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if err != nil {
		return InputString(title)
	}
	//return strings.TrimSpace(input)
	return input
}
func InputStringEmpty(title, defaultString string) string {
	reader := bufio.NewReader(os.Stdin)
	//glog.Print(title)
	fmt.Print(title)
	input, err := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if err != nil {
		return InputString(title)
	}
	if input == "" {
		return defaultString
	}
	//return strings.TrimSpace(input)
	return input
}

func InputString(title string) string {
	reader := bufio.NewReader(os.Stdin)
	//glog.Print(title)
	fmt.Print(title)
	input, err := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if err != nil {
		return InputString(title)
	}
	//return strings.TrimSpace(input)
	if len(input) == 0 {
		tips(title)
		return InputString(title)
	}
	return input
}
func InputInt(title string) int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(title)
	input, err := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if err != nil {
		return InputInt(title)
	}
	if len(input) == 0 {
		tips(title)
		return InputInt(title)
	}
	num, err := strconv.Atoi(input)
	if err != nil {
		return InputInt(title)
	}
	return num
}

func GetInt() int {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if err != nil {
		return GetInt()
	}
	if len(input) == 0 {
		fmt.Println("不允许输入空")
		return GetInt()
	}
	num, err := strconv.Atoi(input)
	if err != nil {
		return GetInt()
	}
	return num
}

func Exit() {
	for i := 5; i >= 0; i-- {
		fmt.Printf("\r%d秒后退出程序..", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("\n")
	os.Exit(0)
}

func Delete(filePath string, args ...string) error {
	var title string
	if args != nil && len(args) > 0 {
		title = args[0]
	}
	if err := os.Remove(filePath); err != nil {
		msg := fmt.Errorf("%s 文件删除失败: %s,%v\n", title, filePath, err)
		glog.Error(msg)
		return msg
	}
	glog.Infof("%s 文件删除成功: %s\n", title, filePath)
	return nil
}

func DeleteAll(filePath string, args ...string) error {
	var title string
	if args != nil && len(args) > 0 {
		title = args[0]
	}
	if err := os.RemoveAll(filePath); err != nil {
		msg := fmt.Errorf("%s 删除失败: %s,%v\n", title, filePath, err)
		glog.Error(msg)
		return msg
	}
	glog.Infof("%s 删除成功: %s\n", title, filePath)
	return nil
}

func GenerateBin(scrFilePath, dstFilePath string, oldBytes, newBytes []byte) error {
	// 打开原文件
	srcFile, err := os.Open(scrFilePath)
	if err != nil {
		return fmt.Errorf("无法打开文件: %v[%s]", err, scrFilePath)
	}
	defer srcFile.Close()

	var srcFileSize int64
	if stat, err := srcFile.Stat(); err == nil {
		srcFileSize = stat.Size()
		sizeB := float64(stat.Size()) / 1024 / 1024
		glog.Printf("%s[大小：%.2fMB]%s\n", stat.Name(), sizeB, dstFilePath)
	}

	tmpFile, err := os.Create(dstFilePath)
	if err != nil {
		return fmt.Errorf("无法创建临时文件: %v[%s]", err, dstFilePath)
	}
	defer tmpFile.Close()

	reader := bufio.NewReader(srcFile)
	prevBuffer := make([]byte, 0)
	isReplace := false
	var indexSize int64
	newFileSize := int64(0)
	tempProgress := -1
	for {
		thisBuffer := make([]byte, Divide(len(oldBytes), 1024))
		n, err2 := reader.Read(thisBuffer)
		if err2 != nil && err2 != io.EOF {
			return fmt.Errorf("读取文件时出错: %v[%s]", err2, scrFilePath)
		}
		indexSize += int64(n)
		thisBuffer = thisBuffer[:n]
		tempBuffer := append(prevBuffer, thisBuffer...)
		index := bytes.Index(tempBuffer, oldBytes)
		if index > -1 {
			glog.Printf("找到位置[%d]了，签名...\n", index)
			isReplace = true
			tempBuffer = bytes.Replace(tempBuffer, oldBytes, newBytes, -1)
		}
		// 写入前一次的
		writeSize, err1 := tmpFile.Write(tempBuffer[:len(prevBuffer)])
		if err1 != nil {
			return fmt.Errorf("1写入临时文件时出错: %v[%s]", err1, dstFilePath)
		}

		newFileSize += int64(writeSize)
		progress := int(float64(indexSize) / float64(srcFileSize) * 100)
		if progress >= tempProgress {
			glog.Printf("程序签名:%v%s\n", progress, "%")
			tempProgress = progress
			tempProgress += 5
		}

		//前一次的+本次的转给 prev
		prevBuffer = tempBuffer[len(prevBuffer):]
		//if err != nil {
		//	break
		//}
		if n == 0 || err2 != nil {
			break // 文件读取完毕
		}
	}
	if len(prevBuffer) > 0 {
		writeSize, err1 := tmpFile.Write(prevBuffer)
		if err1 != nil {
			return fmt.Errorf("2写入临时文件时出错: %v[%s]", err1, dstFilePath)
		}
		newFileSize += int64(writeSize)
		prevBuffer = nil
	}
	glog.Printf("原始文件大小：%d  %s\n", indexSize, scrFilePath)
	glog.Printf("目标文件大小：%d  %s\n", indexSize, dstFilePath)
	// 给文件赋予执行权限（0755）
	errMsg := os.Chmod(dstFilePath, 0755)
	if errMsg != nil {
		return fmt.Errorf("赋予文件执行权限时出错: %v\n", errMsg)
	}
	if !isReplace {
		glog.Printf("oldBytes[%d]--->%v\n", len(oldBytes), oldBytes)
		glog.Printf("newBytes[%d]--->%v\n", len(newBytes), newBytes)
		return errors.New("位置没找到，数据未替换😭")
	}
	err1 := srcFile.Close()
	if err1 != nil {
		glog.Error("srcFile.Close", err1)
	}
	err1 = tmpFile.Close()
	if err1 != nil {
		glog.Error("tmpFile.Close", err1)
	}

	return nil
}

func RestartForWindows() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "restart")
	// 设置进程属性，创建新会话
	if !IsWindows() {
	}
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("Error starting update process: %v\n", err)
	}
	return nil
}

func EnsureDir(path string) error {
	// 检查目录是否存在
	if _, err := os.Stat(path); err == nil {
		// 存在，删除
		err = os.RemoveAll(path)
		if err != nil {
			return err
		}
		return os.MkdirAll(path, 0755)
	} else if !os.IsNotExist(err) {
		// 其他错误
		return err
	}
	// 不存在，创建
	return os.MkdirAll(path, 0755)
}

func GetUpgradeDir() string {
	return glog.GetCrossPlatformDataDir("upgrade")
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

func DownloadFileWithCancel(ctx context.Context, url string, args ...string) (string, error) {
	// 创建可取消的 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	// 创建 HTTP 客户端
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	var dstFile string
	if args != nil && len(args) > 0 {
		dstFile = args[0]
	}
	tempFolder := fmt.Sprintf("%d", time.Now().Unix())
	if dstFile == "" {
		dstName := GetFileNameFromUrl(url)
		if dstName == "" {
			dstName = GetFilenameFromHeader(resp.Header)
		}
		if dstName == "" {
			fileName := time.Now().Unix()
			dstName = fmt.Sprintf("%d", fileName)
		}
		if dstName != "" {
			dstFile = filepath.Join(GetUpgradeDir(), tempFolder, dstName)
		}
	} else {
		dir, f := filepath.Split(dstFile)
		dstFile = filepath.Join(dir, tempFolder, f)
	}
	dir, _ := filepath.Split(dstFile)
	EnsureDir(dir)
	// 创建目标文件
	outFile, err := os.Create(dstFile)
	if err != nil {
		return "", err
	}
	defer outFile.Close()
	totalSize := resp.ContentLength
	// 分块读取并写入文件
	buf := make([]byte, 4096) // 4KB 缓冲区
	var preProgress float64 = -3.1
	for {
		select {
		case <-ctx.Done(): // 检查取消信号
			fmt.Println("下载已取消:", url)
			dir, _ := filepath.Split(dstFile)
			DeleteAll(dir, "下载已取消")
			return "", ctx.Err()
		default:
			n, err := resp.Body.Read(buf)
			if err != nil && err != io.EOF {
				return "", err
			}
			if n == 0 {
				fmt.Println("文件路径：", dstFile)
				return dstFile, nil // 正常完成
			}

			if _, err := outFile.Write(buf[:n]); err != nil {
				return "", err
			}
			fileSize := getFileSize(outFile)
			progress := float64(fileSize) / float64(totalSize) * 100
			if progress-preProgress > 3 {
				fmt.Printf("[%s]总大小: %.2fMB 已下载: %.2fMB 进度: %.2f%%\n", tempFolder, float64(totalSize)/1e6, float64(fileSize)/1e6, progress)
				preProgress = progress
			}
		}
	}

}

func getFileSize(f *os.File) int64 {
	info, _ := f.Stat()
	return info.Size()
}
