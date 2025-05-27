package gs

import (
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/internal"
	"github.com/xxl6097/go-service/pkg/gs/igs"
	"github.com/xxl6097/go-service/pkg/utils/util"
	"os"
	"path/filepath"
)

func init() {
	glog.Register(util.MarketName)
}

func Run(srv igs.IService) error {
	if len(os.Args) > 1 {
		glog.LogDefaultLogSetting(fmt.Sprintf("%s.log", os.Args[1]))
	} else {
		binPath := os.Args[0]
		binName := filepath.Base(binPath)
		glog.LogDefaultLogSetting(fmt.Sprintf("%s-app.log", binName))
	}

	if srv == nil {
		return fmt.Errorf("请继承igs.IService接口")
	}
	return internal.NewCore(srv).Run()
}
