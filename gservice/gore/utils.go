package gore

import (
	"bufio"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

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
	fmt.Println("文件名:", fileName) // 输出: document.pdf
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

func DownLoad1(url string, args ...string) (string, error) {
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
			dstFile = filepath.Join(os.TempDir(), dstName)
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

func Download(url, dstFile string) error {
	// 要下载的文件的 URL
	// 发送 HTTP GET 请求
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 获取文件的总大小
	totalSize := resp.ContentLength
	if totalSize == -1 {
		fmt.Println("无法获取文件大小，可能不支持 Content-Length 头信息。")
		return fmt.Errorf("无法获取文件大小，可能不支持 Content-Length 头信息。")
	}
	sizeA := float64(resp.ContentLength) / 1024 / 1024
	fmt.Printf("文件大小:%.2fM\n", sizeA)
	// 创建一个本地文件用于保存下载的内容
	file, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建进度写入器实例
	pw := &ProgressWriter{TotalSize: totalSize, Progress: -1}

	// 将响应体的数据复制到本地文件，并通过 ProgressWriter 跟踪进度
	_, err = io.Copy(io.MultiWriter(file, pw), resp.Body)
	if err != nil {
		return fmt.Errorf("下载出错: %v", err)
	}

	fmt.Println("下载完成")
	return nil
}

func Download1(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned HTTP status %v", resp.StatusCode)
	}
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
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
