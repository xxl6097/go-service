package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println(os.Args)
	fmt.Println(filepath.Dir(os.Args[1]))
	fmt.Println(filepath.Ext(os.Args[1]))
	appName := filepath.Base(os.Args[1])
	if ext := filepath.Ext(appName); ext != "" {
		appName = strings.TrimSuffix(appName, ext)
	}
	if strings.Contains(appName, "_") {
		arr := strings.Split(appName, "_")
		if arr != nil && len(arr) > 0 {
			appName = arr[0]
		}
	}
	if strings.Contains(appName, "-") {
		arr := strings.Split(appName, "-")
		if arr != nil && len(arr) > 0 {
			appName = arr[0]
		}
	}
	if strings.Contains(appName, ".") {
		arr := strings.Split(appName, ".")
		if arr != nil && len(arr) > 0 {
			appName = arr[0]
		}
	}
	fmt.Println(appName)
}
