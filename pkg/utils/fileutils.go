package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/utils"
	"os"
)

func CheckFileOrDownload(ctx context.Context, fileUrlOrLocal string) (string, error) {
	defer glog.Flush()
	if utils.IsURL(fileUrlOrLocal) {
		filePath, err := DownloadWithCancel(ctx, fileUrlOrLocal)
		if err != nil {
			glog.Error("下载失败", fileUrlOrLocal, err)
			return "", err
		}
		glog.Debug("下载成功", filePath)
		return filePath, nil
	} else if utils.FileExists(fileUrlOrLocal) {
		glog.Debug("检测为本地文件", fileUrlOrLocal)
		return fileUrlOrLocal, nil
	} else {
		glog.Error("无法识别的文件", fileUrlOrLocal)
		return "", errors.New("无法识别的文件" + fileUrlOrLocal)
	}
}

func CheckDirector(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		// 存在，删除
		return nil
	}
	return os.MkdirAll(path, 0755)
}

func ResetDirector(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		// 存在，删除
		err = os.RemoveAll(path)
		if err != nil {
			return err
		}
		return os.MkdirAll(path, 0755)
	}
	return os.MkdirAll(path, 0755)
}

func DeleteAllDirector(filePath string) error {
	defer glog.Flush()
	err := os.RemoveAll(filePath)
	if err != nil {
		msg := fmt.Errorf("删除失败[%v]: %s,%v\n", os.Getpid(), filePath, err)
		glog.Error(msg)
		return msg
	}
	glog.Infof("删除成功[%v]: %s\n", os.Getpid(), filePath)
	return err
}
