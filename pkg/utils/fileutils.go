package utils

import (
	"context"
	"errors"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/utils"
)

func CheckFileOrDownload(ctx context.Context, fileUrlOrLocal string) (string, error) {
	if utils.IsURL(fileUrlOrLocal) {
		filePath, err := utils.DownloadFileWithCancel(ctx, fileUrlOrLocal)
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
