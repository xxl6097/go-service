package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xxl6097/glog/pkg/z"
	"github.com/xxl6097/glog/pkg/zutil"
	"go.uber.org/zap"
)

func DownloadFileWithCancelByUrls(urls []string) string {
	newUrl := DynamicSelect[string](urls, func(ctx context.Context, i int, s string) string {
		var dst string
		select {
		default:
			//tid := GetGoroutineID()
			dstFilePath, err := DownloadWithCancel(ctx, s)
			FileSize(dstFilePath)
			if err == nil {
				z.L().Debug("下载成功", zap.String("本地路径", dstFilePath), zap.String("url", s))
				return dstFilePath
			} else if errors.Is(err, context.Canceled) {
				//fmt.Println("2通道 ", i, err.Error())
				return dst
			} else {
				var netErr net.Error
				if errors.As(err, &netErr) {
					z.L().Error("超时错误", zap.String("本地路径", dstFilePath), zap.Error(netErr))
					time.Sleep(time.Hour)
				}
				<-ctx.Done()
			}
		}
		return dst
	})
	return newUrl
}

func DownloadWithCancel(ctx context.Context, url string, args ...string) (string, error) {
	// 创建可取消的 HTTP 请求
	//glog.Debug("开始下载", url)
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

	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	defer resp.Body.Close()
	var dstFile string
	if args != nil && len(args) > 0 {
		dstFile = args[0]
	}
	tempFolder := GetID() //fmt.Sprintf("%d", time.Now().UnixNano())
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
			dstFile = filepath.Join(zutil.AppHome("temp", "upgrade"), tempFolder, dstName)
		}
	} else {
		dir, f := filepath.Split(dstFile)
		dstFile = filepath.Join(dir, tempFolder, f)
	}
	dir, _ := filepath.Split(dstFile)
	goroutineId := GetGoroutineID()
	_ = ResetDirector(dir)
	// 创建目标文件
	//fmt.Println("os.Create", dstFile)
	outFile, err := os.Create(dstFile)
	if err != nil {
		_ = DeleteAllDirector(dir)
		return "", err
	}
	defer outFile.Close()
	totalSize := resp.ContentLength
	// 分块读取并写入文件
	buf := make([]byte, 4096) // 4KB 缓冲区
	var preProgress float64 = -3.1
	var written int64 // 累计已写入字节数，不依赖文件句柄（避免对已关闭句柄取 Stat 崩溃）
	for {
		select {
		case <-ctx.Done(): // 检查取消信号
			//fmt.Println("下载已取消:", url)
			_ = outFile.Close()
			_ = DeleteAllDirector(dir)
			return "", ctx.Err()
		default:
			n, err1 := resp.Body.Read(buf)
			if err1 != nil && err1 != io.EOF {
				_ = outFile.Close()
				_ = DeleteAllDirector(dir)
				return "", err1
			}
			if n == 0 {
				_ = outFile.Close()
				// 校验下载完整性：仅当服务端给出了有效 Content-Length 时才比对
				// （差量 .patch 等场景 ContentLength 可能为 0/-1，此时跳过校验）。
				// 不一致多为网络截断或磁盘/内存（OpenWRT 的 tmpfs）写满导致的截断文件，
				// 直接返回截断二进制会在后续试跑时崩溃（SIGSEGV）。
				if totalSize > 0 && written != totalSize {
					_ = DeleteAllDirector(dir)
					return "", fmt.Errorf("下载文件不完整：期望 %d 字节，实际 %d 字节（可能磁盘/内存空间不足或网络中断）", totalSize, written)
				}
				z.Println("文件下载完成：", dstFile)
				return dstFile, nil // 正常完成
			}

			w, e := outFile.Write(buf[:n])
			if e != nil {
				_ = outFile.Close()
				_ = DeleteAllDirector(dir)
				return "", fmt.Errorf("写入升级文件失败（可能空间不足）: %w", e)
			}
			written += int64(w)
			if totalSize > 0 {
				progress := float64(written) / float64(totalSize) * 100
				if progress-preProgress > 3 {
					fmt.Printf("[%d]总大小: %.2fMB 已下载: %.2fMB 进度: %.2f%%\n", goroutineId, float64(totalSize)/1e6, float64(written)/1e6, progress)
					preProgress = progress
				}
			}
		}
	}
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

// GetGoroutineID 用于获取当前协程的ID
func GetGoroutineID() uint64 {
	var buf [64]byte
	// 调用runtime.Stack获取当前协程的栈信息
	n := runtime.Stack(buf[:], false)
	// 解析栈信息以提取协程ID
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	var id uint64
	fmt.Sscanf(idField, "%d", &id)
	return id
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
