package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	dstPath := "/User/uuxia/acfrp_arm64.exe"
	fileName := filepath.Base(dstPath)
	fmt.Println(fileName)
}
