package gore

import "github.com/kardianos/service"

type IGService interface {
	Restart() error
	Upgrade(string) error
	Uninstall() error
}

type GService interface {
	OnVersion() string
	OnConfig() *service.Config
	OnRun(IGService) error
}
