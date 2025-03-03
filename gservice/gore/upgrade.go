package gore

import (
	"errors"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/ukey"
	"github.com/xxl6097/go-service/gservice/utils"
	"os"
	"time"
)

type Upgrade interface {
	GService
	OnUpgrade(string, string) error
}

// DefaultUpgrade 默认实现，但是需要用户继承一些方法
type DefaultUpgrade interface {
	BaseService
}

func Update(g GService, installBinPath, fileUrlOrLocalPath string) error {
	if gs, ok := g.(Upgrade); ok {
		return gs.OnUpgrade(installBinPath, fileUrlOrLocalPath)
	} else if gss, okk := g.(DefaultUpgrade); okk {
		cfg := gss.GetAny()
		if cfg != nil {
			return signUpdate(installBinPath, installBinPath)
		}
	}
	return manualUpgrade(installBinPath, fileUrlOrLocalPath)
}

func signUpdate(binPath, newFileUrlOrLocalPath string) error {
	//1、读取老文件特征数据；
	//2、下载新文件
	//3、替换新文件特征数据
	//4、数据写到安装目录地址（oldBinPath）
	cfgBufferBytes := ukey.GetCfgBufferFromFile(binPath)
	if cfgBufferBytes == nil {
		err := fmt.Errorf("读取原文件配置信息失败 %s", binPath)
		glog.Error(err)
		return err
	}
	glog.Debug("获取配置数据成功", len(cfgBufferBytes))

	err := utils.Delete(binPath, "旧运行文件")
	if err != nil {
		return err
	}
	var newFilePath string
	if utils.FileExists(newFileUrlOrLocalPath) {
		newFilePath = newFileUrlOrLocalPath
	} else if utils.IsURL(newFileUrlOrLocalPath) {
		glog.Debug("下载文件", newFileUrlOrLocalPath)
		temp, err := utils.DownLoad(newFileUrlOrLocalPath)
		if err != nil {
			glog.Error("下载失败", err)
			return err
		}
		glog.Debug("下载成功.", temp)
		newFilePath = temp
	}
	if newFilePath != "" {
		oldBuffer := ukey.GetBuffer()
		err := utils.GenerateBin(newFilePath, binPath, oldBuffer, cfgBufferBytes)
		if err != nil {
			glog.Error("签名错误：", err)
			return err
		}
		return nil
	} else {
		return fmt.Errorf("新文件错误～❌")
	}
}

func manualUpgrade(installBinPath string, fileUrlOrLocalPath string) error {
	time.Sleep(100 * time.Millisecond)
	err := utils.Delete(installBinPath, "删除旧版")
	if err != nil {
		glog.Errorf("旧版删除失败 %v\n", err)
		return err
	}
	if utils.FileExists(fileUrlOrLocalPath) {
		newPath := fileUrlOrLocalPath
		glog.Debugf("拷贝新版 %s==>%s\n", newPath, installBinPath)
		err = utils.Copy(newPath, installBinPath)
		if err != nil {
			glog.Error("拷贝失败", err)
			return err
		} else {
			glog.Debugf("新版拷贝成功 %s==>%s\n", newPath, installBinPath)
			err = os.Remove(newPath)
			if err != nil {
				glog.Error("删除安装文件失败", err)
			}
			return nil
		}

	} else if utils.IsURL(fileUrlOrLocalPath) {
		glog.Debug("下载新版本", fileUrlOrLocalPath)
		_, err = utils.DownLoad(fileUrlOrLocalPath, installBinPath)
		if err != nil {
			glog.Error("下载失败", err)
			return err
		}
		glog.Debug("下载成功.", installBinPath)
		return nil
	} else {
		msg := fmt.Sprintf("参数错误，请输入正确的URL %s", fileUrlOrLocalPath)
		glog.Error(msg)
		return errors.New(msg)
	}
}
