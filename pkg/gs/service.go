package gs

import (
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/internal"
	"github.com/xxl6097/go-service/pkg/gs/igs"
	"github.com/xxl6097/go-service/pkg/utils/util"
)

func init() {
	glog.Register(util.MarketName)
}

func Run(srv igs.IService) error {
	if srv == nil {
		return fmt.Errorf("请继承igs.IService接口")
	}
	return internal.NewCore(srv).Run()
}
