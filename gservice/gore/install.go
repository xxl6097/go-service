package gore

import (
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/ukey"
	"github.com/xxl6097/go-service/gservice/utils"
	"path/filepath"
)

type Installer interface {
	GService
	//OnInstall arg1: 当前运行bin文件路径，arg2安装的目标bin文件路径
	OnInstall(string, string) error
}

// DefaultInstaller 默认实现，但是需要用户继承一些方法
type DefaultInstaller interface {
	BaseService
}

func Install(g GService, binPath, installBinPath string) error {
	if gs, ok := g.(Installer); ok {
		glog.Println("------Install---1", binPath, installBinPath)
		return gs.OnInstall(binPath, installBinPath)
	} else if gss, okk := g.(DefaultInstaller); okk {
		glog.Println("------Install---2", binPath, installBinPath)
		cfg := gss.GetAny(filepath.Dir(installBinPath))
		if cfg != nil {
			glog.Println("------Install---3", binPath, installBinPath)
			return signInstall(cfg, binPath, installBinPath)
		}
	}
	glog.Println("------Install--4", binPath, installBinPath)
	return manualInstall(binPath, installBinPath)
}

func signInstall(cfg any, binPath, installBinPath string) error {
	//1、获取用户的配置信息；
	//2、加密用户信息，构建二进制常量；
	//3、从原始二进制文件中查询特征码，替换为二进制常量；
	//4、替换后，将二进制文件复制到安装指定的目录
	newBufferBytes, err := ukey.GenConfig(cfg, false)
	if err != nil {
		return fmt.Errorf("构建签名信息错误: %v", err)
	}
	//安装程序，需要对程序进行签名，那么需要传入两个参数：
	//1、最原始的key；
	//2、需写入的data
	buffer := ukey.GetBuffer()
	glog.Printf("buffer大小 %d\n", len(buffer))
	err = utils.GenerateBin(binPath, installBinPath, buffer, newBufferBytes)
	if err != nil {
		return fmt.Errorf("签名错误: %v", err)
	}
	if utils.FileExists(binPath) {
		_ = utils.Delete(binPath, "旧运行文件")
	}
	return nil
}

func manualInstall(binPath, installBinPath string) error {
	err := utils.Copy(binPath, installBinPath)
	if err != nil {
		glog.Printf("文件拷贝失败，错误信息：%s", err)
		return err
	}
	return nil
}

//安装、升级、卸载、开启、关闭、重启分别定义接口并继承BaseService
//1、安装，无实现，执行默认行为1（复制）；有实现，且参数合法，则执行子类行为，否则执行默认行为1
//2、作为go-service，默认行为2需要做签名功能
