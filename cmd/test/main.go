package main

import (
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/pkg/utils"
	"path/filepath"
)

func main() {
	inFilePath := "/root/files/aatest_v0.6.43_linux_arm64"

	dstFile := filepath.Join(glog.TempDir(), filepath.Base(inFilePath))
	fmt.Println(dstFile)
	outFilePath := filepath.Join(glog.AppHome("temp", "sign", utils.GetID()), filepath.Base(inFilePath))
	fmt.Println("--", outFilePath)
}
