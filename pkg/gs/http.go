package gs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/pkg/github"
	"github.com/xxl6097/go-service/pkg/gs/igs"
	"github.com/xxl6097/go-service/pkg/utils"
	"github.com/xxl6097/go-service/pkg/utils/util"
)

var pool = &sync.Pool{
	New: func() interface{} { return make([]byte, 32*1024) },
}

func update(srv igs.Service, w http.ResponseWriter, r *http.Request) {
	res, f := Response(r)
	defer f(w)
	if srv == nil {
		res.Error("srv is nil")
		return
	}
	ctx := r.Context()
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	updir := glog.AppHome()
	_, _, free, _ := util.GetDiskUsage(updir)
	if free < utils.GetSelfSize()*2 {
		if err := utils.ClearTemp(); err != nil {
			glog.Println("/tmp清空失败:", err)
		} else {
			glog.Println("/tmp清空完成")
		}
	}

	var newFilePath string
	switch r.Method {
	case "PUT", "put":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			res.Response(400, fmt.Sprintf("read request body error: %v", err))
			glog.Warnf("%s", res.Msg)
			return
		}
		if len(body) == 0 {
			res.Response(400, "升级URL空的哦～")
			glog.Warnf("%s", res.Msg)
			return
		}
		binUrl := string(body)
		glog.Debugf("升级URL地址: %s", binUrl)
		newUrl := utils.DownloadFileWithCancelByUrls(github.Api().GetProxyUrls(binUrl))
		newFilePath = newUrl
		break
	case "POST", "post":
		// 获取上传的文件
		file, handler, err := r.FormFile("file")
		if err != nil {
			res.Error("body no file")
			return
		}
		defer file.Close()
		dstFilePath := filepath.Join(glog.AppHome("temp", "upgrade"), handler.Filename)
		//dstFilePath 名称为上传文件的原始名称
		dst, err := os.Create(dstFilePath)
		if err != nil {
			res.Error(fmt.Sprintf("create file %s error: %v", handler.Filename, err))
			return
		}
		buf := pool.Get().([]byte)
		defer pool.Put(buf)
		_, err = io.CopyBuffer(dst, file, buf)
		_ = dst.Close()
		if err != nil {
			res.Error(err.Error())
			return
		}
		newFilePath = dstFilePath
		break
	default:
		res.Error("位置请求方法")
	}
	if newFilePath != "" {
		glog.Debugf("开始升级 %s", newFilePath)
		err := srv.Upgrade(ctx, newFilePath)
		glog.Debug("---->升级", err)
		if err == nil {
			res.Ok("升级成功～")
		} else {
			res.Error(fmt.Sprintf("更新失败～%v", err))
		}

	}
}

func checkVersion(name string, w http.ResponseWriter, r *http.Request) {
	res, f := Response(r)
	defer f(w)
	if name == "" {
		name = glog.AppName()
	}
	data, err := github.Api().CheckUpgrade(name)
	if err != nil {
		res.Err(err)
	} else {
		glog.Debug("version:", data)
		res.Any(data)
	}
}

func ApiCheckVersion(binName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		checkVersion(binName, w, r)
	}
}

func ApiUpdate(srv igs.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		update(srv, w, r)
	}
}
