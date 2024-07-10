package gservice

import "github.com/kardianos/service"

type IService interface {
	service.Interface
	Config() *service.Config
	Version() string
	OnInstall()
}
