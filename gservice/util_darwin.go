package gservice

import (
	"bytes"
	"os/exec"
	"strings"
	"syscall"
)

const (
	defaultInstallPath = "/usr/local"
	// defaultBinName     = "AAServiceApp"
)

func execOutput(name string, args ...string) string {
	cmdGetOsName := exec.Command(name, args...)
	var cmdOut bytes.Buffer
	cmdGetOsName.Stdout = &cmdOut
	cmdGetOsName.Run()
	return cmdOut.String()
}
func getOsName() (osName string) {
	//fmt.Println(AppConfig.ProductName)
	output := execOutput("sw_vers", "-productVersion")
	osName = "Mac OS X " + strings.TrimSpace(output)
	return
}

func SetRLimit() error {
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}
	limit.Cur = 65536
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}
	return nil
}

func SetFirewall(ProductName, fullPath string) {
}
