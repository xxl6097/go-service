package main

import (
	"fmt"
	"path/filepath"

	"github.com/xxl6097/glog/pkg/zutil"
	"github.com/xxl6097/go-service/pkg/utils"
)

func main() {
	inFilePath := "/root/files/aatest_v0.6.43_linux_arm64"

	dstFile := filepath.Join(zutil.TempDir(), filepath.Base(inFilePath))
	fmt.Println(dstFile)
	outFilePath := filepath.Join(zutil.AppHome("temp", "sign", utils.GetID()), filepath.Base(inFilePath))
	fmt.Println("--", outFilePath)
}
