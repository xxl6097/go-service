package gservice

import (
	"errors"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/go-glog/glog"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Installer struct {
	daemon   *Daemon
	iservice IService
	binDir   string
	binName  string
	binPath  string
}

func NewInstaller(iservice IService, installPath string) *Installer {
	this := Installer{
		binDir:   installPath,
		iservice: iservice,
	}
	conf := iservice.Config()
	//可执行文件名称是取的配置文件配置的名称
	this.binName = conf.Name
	if IsWindows() {
		this.binName = conf.Name + ".exe"
	}

	this.binPath = filepath.Join(this.binDir, this.binName)
	conf.Executable = this.binPath
	_args := make([]string, 0)
	_args = append(_args, "-d")
	if conf.Arguments != nil && len(conf.Arguments) > 0 {
		for i := 0; i < len(conf.Arguments); i++ {
			if conf.Arguments[i] == "-d" {
				panic("不允许有-d参数")
			}
		}
		_args = append(_args, conf.Arguments...)
	}

	conf.Arguments = _args
	this.daemon = NewDaemon(iservice, conf)
	if this.daemon == nil {
		glog.Error("daemon is nil")
		return nil
	}
	return &this
}

func (this *Installer) IsInstalled() bool {
	status, err2 := this.daemon.Status()
	if err2 != nil {
		if status == service.StatusUnknown {
			return false
		}
		if _, err := os.Stat(this.binPath); os.IsNotExist(err) {
			return false
		}
		return true
	}
	return false
}

func (this *Installer) Install() error {
	defer glog.Flush()
	_, err := this.daemon.Status()
	if err == nil {
		glog.Print("检测到", this.binName, "程序已经安装，需要卸载吗?(y/n):")
		var yes string
		fmt.Scanln(&yes)
		if strings.Compare("y", yes) == 0 || strings.Compare("yes", yes) == 0 {
			this.Uninstall()
		} else {
			glog.Debug("结束安装.")
			os.Exit(0)
			return err
		}
	}

	SetFirewall(this.binName, this.binPath)
	SetRLimit()
	if _, err := os.Stat(this.binDir); !os.IsNotExist(err) {
		err5 := os.RemoveAll(this.binDir)
		if err5 != nil {
			glog.Error("删除失败", this.binDir)
		}
	}

	err = os.MkdirAll(this.binDir, 0775)
	if err != nil {
		glog.Printf("MkdirAll %s error:%s", this.binDir, err)
		return err
	}
	var args []string
	if this.iservice != nil {
		args = this.iservice.OnInstall(this.binDir)
	}
	//glog.Println("安装路径：", this.binDir)
	err = os.Chdir(this.binDir)
	if err != nil {
		glog.Println("cd error:", err)
		return err
	}

	//这个地方是取的当前运行的执行文件
	currentBinPath, err := os.Executable()
	if err != nil {
		glog.Fatal("os.Executable() error", err)
		return err
	}
	//glog.Println("可执行程序位置：", binPath)
	src, err := os.Open(currentBinPath) // can not use args[0], on Windows call openp2p is ok(=openp2p.exe)
	if err != nil {
		glog.Printf("os.OpenFile %s error:%s", os.Args[0], err)
		return err
	}
	//将本程序复制到目标为止，目标文件名称为配置文件的名称
	dst, err := os.OpenFile(this.binPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0775)
	if err != nil {
		glog.Printf("os.OpenFile %s error:%s", this.binPath, err)
		return err
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		glog.Printf("文件拷贝失败，错误信息：%s", err)
		return err
	}
	src.Close()
	dst.Close()
	// install system service
	//glog.Println("程序位置:", this.binPath)
	err = this.daemon.Install(args) //.Control("install", this.binPath, []string{"-d"})
	if err == nil {
		glog.Println("服务安装成功!")
	} else {
		glog.Println("服务安装失败，错误信息:", err)
	}
	time.Sleep(time.Second * 2)
	err = this.daemon.Start() //Control("start", this.binPath, []string{"-d"})
	if err != nil {
		glog.Println("服务启动失败，错误信息:", err)
	} else {
		glog.Println("服务启动成功！")
	}
	return err
}

func (this *Installer) Uninstall() error {
	defer glog.Flush()
	if this.daemon.IsRunning() {
		err := this.daemon.Stop() //.Control("stop", "", nil)
		if err != nil {           // service maybe not install
			glog.Println("卸载失败，错误信息：", this.binName, err)
			return err
		}
	}
	_, err := this.daemon.Status()
	if err != nil {
		glog.Println(this.binName, "程序未安装", err)
		return err
	}
	err = this.daemon.Uninstall() //Control("uninstall", "", nil)
	if err != nil {
		glog.Println("服务卸载失败，错误信息：", err)
	} else {
		glog.Println("服务成功卸载！")
	}
	//os.Remove(this.binPath + "0")
	//os.Remove(this.binPath)
	// 尝试删除自身
	glog.Println("尝试删除自身:", this.binDir)
	if err := os.RemoveAll(this.binDir); err != nil {
		fmt.Printf("Error removing executable: %v\n", err)
		time.Sleep(time.Second * 3)
		os.Exit(1)
	} else {
		glog.Println("尝试删除成功")
	}

	return err
}

func (this *Installer) Upgrade() error {
	if this.daemon.IsRunning() {
		err := this.daemon.Stop() //.Control("stop", "", nil)
		if err != nil {           // service maybe not install
			glog.Println("服务停止失败，错误信息：", err)
			return err
		}
	}
	_, err := this.daemon.Status()
	if err == nil {
		err := this.daemon.Uninstall()
		if err != nil {
			glog.Println("服务卸载失败，错误信息：", err)
			return err
		} else {
			glog.Println("服务成功卸载！")
		}
	}

	if len(os.Args) <= 1 {
		glog.Error("参数错误，请重新配置参数")
		return errors.New("参数错误，请重新配置参数")
	}
	if strings.Compare(os.Args[1], "upgrade") != 0 {
		glog.Error("参数错误，请重新配置参数")
		return errors.New("参数错误，请重新配置参数")
	}

	var binUrl string
	if this.iservice != nil {
		binUrl = this.iservice.OnUpgrade()
	}

	if binUrl == "" {
		if len(os.Args) <= 2 {
			glog.Error("参数错误，请重新配置参数")
			return errors.New("参数错误，请重新配置参数")
		}
		binUrl = os.Args[2]
	}

	if !IsURL(binUrl) {
		glog.Error("参数错误，请输入正确的URL", binUrl)
		return errors.New("参数错误，请输入正确的URL")
	}
	//删除可执行文件
	if _, err := os.Stat(this.binPath); !os.IsNotExist(err) {
		err := os.Remove(this.binPath)
		if err != nil {
			glog.Error("删除失败L", this.binPath)
			return err
		}
	}

	glog.Debug("下载文件", binUrl)
	err = download(binUrl, this.binPath)
	if err != nil {
		glog.Error("下载失败", err)
		return err
	}
	glog.Debug("下载成功.", this.binPath)
	err = os.Chmod(this.binPath, 0755)
	if err == nil {
		glog.Debug(this.binPath, "赋予0755权限成功")
	} else {
		glog.Error(this.binPath, "赋予0755权限失败", err)
	}

	var args []string
	if this.iservice != nil {
		args = this.iservice.OnInstall(this.binDir)
	}

	err = this.daemon.Install(args) //.Control("install", this.binPath, []string{"-d"})
	if err == nil {
		glog.Println("服务升级成功!")
	} else {
		glog.Println("服务升级失败，错误信息:", err)
	}
	time.Sleep(time.Second * 2)
	err = this.daemon.Start()
	if err != nil {
		glog.Println("服务启动失败，错误信息:", err)
	} else {
		glog.Println("服务启动成功！")
	}
	return err
}

func (this *Installer) Upgradebak() {
	if this.daemon.IsRunning() {
		err := this.daemon.Stop() //.Control("stop", "", nil)
		if err != nil {           // service maybe not install
			glog.Println("服务停止失败，错误信息：", err)
			return
		}
	}
	if len(os.Args) <= 2 {
		glog.Error("参数错误，请重新配置参数")
		return
	}
	if strings.Compare(os.Args[1], "upgrade") != 0 {
		glog.Error("参数错误，请重新配置参数")
		return
	}
	binUrl := os.Args[2]
	if !IsURL(binUrl) {
		glog.Error("参数错误，请输入正确的URL", binUrl)
		return
	}
	if _, err := os.Stat(this.binPath); !os.IsNotExist(err) {
		errs := os.Remove(this.binPath)
		if errs != nil {
			glog.Error("删除失败L", this.binPath)
			return
		}
	}

	err1 := download(binUrl, this.binPath)
	if err1 != nil {
		glog.Error("下载失败", err1)
		return
	}
	glog.Error(this.binPath, "下载成功.")
	err := this.daemon.Start()
	if err != nil {
		glog.Println("服务启动失败，错误信息:", err)
	} else {
		glog.Println("服务启动成功！")
	}
}

func (this *Installer) InstallByFilename() {
	defer glog.Flush()
	//glog.Println("installByFilename", os.Args[0])
	targetPath := os.Args[0]
	args := []string{"install"}
	env := os.Environ()
	cmd := exec.Command(targetPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = env
	err := cmd.Run()
	if err != nil {
		glog.Error(targetPath, args)
		glog.Error("install by filename, start process error:", err)
		//return
	}
	for i := 10; i > 0; i-- {
		fmt.Printf("\r%d秒后退出程序..", i)
		time.Sleep(1 * time.Second)
	}
	os.Exit(0)
}

func (this *Installer) Restart() error {
	defer glog.Flush()
	defer glog.Println("restart end")
	glog.Println("重启...")
	err := this.daemon.Restart() //Control("restart", "", nil)
	if err != nil {
		glog.Println("服务重启失败，错误信息：", err)
	} else {
		glog.Println("服务重启成功!")
	}

	return err
}

func (this *Installer) StartService() error {
	defer glog.Flush()
	glog.Println("start")
	defer glog.Println("start end")
	err := this.daemon.Start() //Control("start", "", nil)
	if err != nil {
		glog.Println("start system service error:", err)
	} else {
		glog.Println("start system service ok.")
	}
	return err
}
func (this *Installer) StopService() error {
	defer glog.Flush()
	glog.Println("stop")
	defer glog.Println("stop end")
	err := this.daemon.Stop() //.Control("stop", "", nil)
	if err != nil {
		glog.Println("stop system service error:", err)
	} else {
		glog.Println("stop system service ok.")
	}
	return err
}
func (this *Installer) Run() error {
	return this.daemon.Run()
}

func download(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned HTTP status %v", resp.StatusCode)
	}
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// IsURL 判断给定的字符串是否是一个有效的URL
func IsURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil {
		return false
	}

	return u.Scheme == "http" || u.Scheme == "https"
}
func IsWindows() bool {
	if strings.Compare(runtime.GOOS, "windows") == 0 {
		return true
	}
	return false
}
