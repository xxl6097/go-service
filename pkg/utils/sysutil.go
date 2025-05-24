package utils

import (
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/utils"
	"os"
	"os/exec"
)

func RestartWindowsApplication() {
	if utils.IsWindows() {
		// Windows特有重启逻辑
		exe, _ := os.Executable()
		cmd := exec.Command("cmd", "/C", "timeout 2 && "+exe)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			glog.Error(err)
		}
		os.Exit(0)
	}
}
