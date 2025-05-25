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

func ResetDirector(path string) error {
	// 检查目录是否存在
	if _, err := os.Stat(path); err == nil {
		// 存在，删除
		err = os.RemoveAll(path)
		if err != nil {
			return err
		}
		return os.MkdirAll(path, 0755)
	} else if !os.IsNotExist(err) {
		// 其他错误
		return err
	}
	// 不存在，创建
	return os.MkdirAll(path, 0755)
}

func DeleteAllDirector(filePath string) error {
	defer glog.Flush()
	err := os.RemoveAll(filePath)
	if err != nil {
		msg := fmt.Errorf("删除失败: %s,%v\n", filePath, err)
		glog.Error(msg)
		return msg
	}
	glog.Infof("删除成功: %s\n", filePath)
	return err
}
