package svv

import (
	"github.com/kardianos/service"
	"github.com/xxl6097/go-service/pkg/gs/igs"
)

type Service struct {
	igs.BaseService
}

func (s Service) OnConfig() *service.Config {
	//TODO implement me
	panic("implement me")
}

func (s Service) OnVersion() string {
	//TODO implement me
	panic("implement me")
}

func (s Service) OnRun(service igs.Service) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) GetAny(s2 string) ([]byte, []string) {
	//TODO implement me
	panic("implement me")
}
