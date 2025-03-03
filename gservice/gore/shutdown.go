package gore

import "github.com/kardianos/service"

type shutdown interface {
	GService
	OnStop(service.Service)
	OnShutDown(service.Service)
}

func ShutDown(g GService, svr service.Service) {
	if gs, ok := g.(shutdown); ok {
		gs.OnShutDown(svr)
	}
}
func Stop(g GService, svr service.Service) {
	if gs, ok := g.(shutdown); ok {
		gs.OnStop(svr)
	}
}
