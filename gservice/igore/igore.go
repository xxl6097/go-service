package igore

import (
	"github.com/kardianos/service"
	"github.com/xxl6097/go-service/gservice/gore"
)

type GS interface {
	OnVersion() string
	Config() *service.Config
	OnInstall(string) []any
	OnRun(i gore.Install) error
}

type GlobGS interface {
	GS
	Glob() error
}

func Glob(gs GS) error {
	if ggs, ok := gs.(GlobGS); ok {
		return ggs.Glob()
	}
	return manualGlob(gs)
}

func manualGlob(gs GS) error {
	return nil
}

type test struct {
	GlobGS
}

//func (this *test) Version() string {
//	return "test"
//}
//
//func (this *test) Glob() error {
//	return nil
//}

func NewTest() GS {
	return &test{}
}

func Test001() {
	Glob(NewTest())
}
