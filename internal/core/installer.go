package core

import (
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/pkg/gs/igs"
	"github.com/xxl6097/go-service/pkg/ukey"
	"github.com/xxl6097/go-service/pkg/utils"
	"path/filepath"
	"strings"
)

// Install 原始的安装文件未被删除
func Install(g igs.IService, binPath, installBinPath string) error {
	if gs, ok := g.(igs.Installer); ok {
		return gs.OnInstall(binPath, installBinPath)
	} else {
		cfg := getAny(filepath.Dir(installBinPath), g)
		if cfg != nil {
			newFilePath, e := ukey.SignFileBySelfKey(cfg, binPath)
			if e != nil {
				return e
			}
			return manualInstall(newFilePath, installBinPath)
		}
	}
	return manualInstall(binPath, installBinPath)
}

func getAny(binDir string, g igs.IService) []byte {
	if g == nil {
		glog.Error("igs.IService is nil")
		return nil
	}
	buffer := g.GetAny(binDir)
	if buffer == nil {
		cfg := map[string]any{"path": binDir}
		bb, err := ukey.StructToGob(cfg)
		if err != nil {
			return []byte(err.Error())
		}
		buffer = bb
	}
	return buffer
}

//func getAny(binDir string, g igs.IService) any {
//	if g == nil {
//		glog.Error("igs.IService is nil")
//		return nil
//	}
//	cfg := g.GetAny(filepath.Dir(binDir))
//	if cfg == nil {
//		cfg = ukey.KeyBuffer{}
//	}
//	switch v := cfg.(type) {
//	case ukey.KeyBuffer:
//		v.MenuDisable = true
//		return v
//	case *ukey.KeyBuffer:
//		v.MenuDisable = true
//		return v
//	}
//	return cfg
//}

func manualInstall(binPath, installBinPath string) error {
	if binPath == "" || installBinPath == "" {
		return fmt.Errorf("安装文件空 binPath:%v installBinPath:%v", binPath, installBinPath)
	}
	if strings.Compare(strings.ToLower(binPath), strings.ToLower(installBinPath)) == 0 {
		return fmt.Errorf("当前文件与安装文件路径一致，不允许安装 binPath:%v installBinPath:%v", binPath, installBinPath)
	} else {
		defer func() {
			_ = utils.DeleteAllDirector(filepath.Dir(filepath.Dir(binPath)))
		}()
	}
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
