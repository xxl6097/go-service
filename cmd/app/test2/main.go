package main

import (
	"context"
	"fmt"
	"github.com/xxl6097/go-service/gservice/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func DownloadFileWithCancel(ctx context.Context, url string, args ...string) error {
	// 创建可取消的 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	// 创建 HTTP 客户端
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	var dstFile string
	if args != nil && len(args) > 0 {
		dstFile = args[0]
	}
	if dstFile == "" {
		dstName := utils.GetFileNameFromUrl(url)
		if dstName == "" {
			dstName = utils.GetFilenameFromHeader(resp.Header)
		}
		if dstName == "" {
			fileName := time.Now().Unix()
			dstName = fmt.Sprintf("%d", fileName)
		}
		if dstName != "" {
			dstFile = filepath.Join(utils.GetUpgradeDir(), dstName)
		}
	}
	// 创建目标文件
	outFile, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	fmt.Println("文件路径：", dstFile)
	totalSize := resp.ContentLength
	// 分块读取并写入文件
	buf := make([]byte, 4096) // 4KB 缓冲区
	for {
		select {
		case <-ctx.Done(): // 检查取消信号
			fmt.Println("\n下载已取消")
			return ctx.Err()
		default:
			n, err := resp.Body.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 {
				return nil // 正常完成
			}

			if _, err := outFile.Write(buf[:n]); err != nil {
				return err
			}
			fileSize := getFileSize(outFile)
			progress := float64(fileSize) / float64(totalSize) * 100
			fmt.Printf("总大小: %.2fMB 已下载: %.2fMB 进度: %.2f%%\r", float64(totalSize)/1e6, float64(fileSize)/1e6, progress)
		}
	}

}

func getFileSize(f *os.File) int64 {
	info, _ := f.Stat()
	return info.Size()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动下载协程
	go func() {
		err := DownloadFileWithCancel(ctx, "http://uuxia.cn:8087/soft/windows/HotPE_Client_V0.3.240201.7z")
		if err != nil {
			fmt.Println("下载错误:", err)
		}
	}()

	// 模拟用户5秒后取消操作
	time.Sleep(3 * time.Second)
	fmt.Println("取消下载")
	cancel() // 触发取消

	// 等待操作完成
	time.Sleep(1 * time.Second)
}
