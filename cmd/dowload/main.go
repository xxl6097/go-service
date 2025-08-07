package main

import (
	"fmt"
	"github.com/xxl6097/go-service/pkg/utils"
)

func main() {
	urls := []string{"https://ghproxy.1888866.xyz/https://github.com/xxl6097/go-service/releases/download/v0.6.49/aatest_v0.6.49_linux_arm64",
		"https://github.moeyy.xyz/https://github.com/xxl6097/go-service/releases/download/v0.6.49/aatest_v0.6.49_linux_arm64",
		"https://gh-proxy.com/https://github.com/xxl6097/go-service/releases/download/v0.6.49/aatest_v0.6.49_linux_arm64",
		"https://ghfast.top/https://github.com/xxl6097/go-service/releases/download/v0.6.49/aatest_v0.6.49_linux_arm64",
		"https://github.com/xxl6097/go-service/releases/download/v0.6.49/aatest_v0.6.49_linux_arm64",
	}
	ileUri := utils.DownloadFileWithCancelByUrls(urls)
	fmt.Println("===>", ileUri)
	//ctx := context.Background()
	//for _, url := range urls {
	//	dstFilePath, err := utils.DownloadWithCancel(ctx, url)
	//	fmt.Println("===>", dstFilePath, err)
	//}
}
