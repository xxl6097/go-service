package gs

import (
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/internal"
	"github.com/xxl6097/go-service/pkg/gs/igs"
	"os"
)

func Run(srv igs.IService) error {
	if len(os.Args) > 1 {
		glog.LogDefaultLogSetting(fmt.Sprintf("%s.log", os.Args[1]))
	} else {
		glog.LogDefaultLogSetting("app.log")
	}
	if srv == nil {
		return fmt.Errorf("请继承igs.IService接口")
	}
	return internal.NewCore(srv).Run()
}
