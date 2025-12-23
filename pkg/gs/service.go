package gs

import (
	"fmt"

	"github.com/xxl6097/go-service/internal"
	"github.com/xxl6097/go-service/pkg/gs/igs"
)

func InitLog(everyType int) {
	internal.InitLog(everyType)
}

func Run(srv igs.IService) error {
	if srv == nil {
		return fmt.Errorf("请继承igs.IService接口")
	}
	return internal.NewCore(srv).Run()
}
