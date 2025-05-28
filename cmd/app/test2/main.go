package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/xxl6097/go-service/pkg/utils"
)

func main() {
	newProxy := []string{"http://uuxia.cn:8087/soft/windows/geek.exe",
		"http://uuxia.cn:8087/soft/windows/Git-2.43.0-64-bit.exe",
		"http://uuxia.cn:8087/soft/windows/DG5611580_x64.zip",
		"http://uuxia.cn:8087/soft/mm-wiki-v0.2.1-windows-amd64.tar.gz",
	}
	newUrl := utils.DynamicSelect[string](newProxy, func(ctx context.Context, i int, s string) string {
		var dst string
		select {
		default:
			tid := utils.GetGoroutineID()
			fmt.Println("1通道 ", i, s, tid)
			dstFilePath, err := utils.DownloadWithCancel(ctx, s)
			if err == nil {
				return dstFilePath
			} else if errors.Is(err, context.Canceled) {
				//fmt.Println("2通道 ", i, err.Error())
				return dst
			} else {
				fmt.Println("3通道 ", i, err.Error())
			}
		}
		return dst
	})
	fmt.Println("下载完成", newUrl)
	select {}
}
