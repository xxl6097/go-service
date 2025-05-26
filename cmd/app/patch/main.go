package main

import (
	"bufio"
	"bytes"
	"github.com/kr/binarydist"
	"os"
)

func main() {
	// 生成补丁示例（需配合bsdiff等工具）
	patch := new(bytes.Buffer)
	oldBytes, _ := os.Open("/Users/uuxia/Downloads/a/aatest_v0.5.9_linux_amd64")
	newBytes, _ := os.Open("/Users/uuxia/Desktop/work/code/github/golang/go-service/release/aatest_v0.5.10_linux_amd64")

	err := binarydist.Diff(bufio.NewReader(oldBytes), bufio.NewReader(newBytes), patch)
	if err != nil {
		return
	}
	os.WriteFile("./update.patch", patch.Bytes(), 0644)

}
