package utils

import (
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore/util"
	"os/exec"
)

func RunChildProcess(executable string, args ...string) error {
	//args = append([]string{executable}, args...)
	//cmd := exec.Command("sudo", args...)
	var cmd *exec.Cmd
	if !IsWindows() {
		arg := []string{executable}
		arg = append(arg, args...)
		cmd = exec.Command("sudo", arg...)
	} else {
		cmd = exec.Command(executable, args...)
	}
	cmd = exec.Command(executable, args...)
	util.SetPlatformSpecificAttrs(cmd)
	glog.Printf("运行子进程 %s %v\n", executable, args)
	return cmd.Start()
}
