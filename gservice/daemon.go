package gservice

import (
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/go-glog/glog"
)

type Daemon struct {
	conf *service.Config
	svr  *service.Service
}

func NewDaemon(shut service.Interface, conf *service.Config) *Daemon {
	this := &Daemon{}
	s, e := service.New(shut, conf)
	if e != nil {
		glog.Error("service new failed ", e)
		return nil
	}
	this.conf = conf
	this.svr = &s
	return this
}

func (d *Daemon) control(cmd string) error {
	e := service.Control(*d.svr, cmd)
	if e != nil {
		fmt.Println(cmd, e)
		return e
	}
	return nil
}

func (d *Daemon) Status() (service.Status, error) {
	return (*d.svr).Status()
}
func (d *Daemon) IsRunning() bool {
	if d.svr == nil {
		return false
	}
	status, err := d.Status()
	if err != nil {
		glog.Println(err)
		return false
	}
	//glog.Println("status", status)
	if status == service.StatusRunning {
		glog.Println(d.conf.Name, "is running")
		return true
	} else if status == service.StatusStopped {
		glog.Println(d.conf.Name, "is stopped")
	} else {
		glog.Println(d.conf.Name, "StatusUnknown", status)
	}
	return false
}

func (d *Daemon) Uninstall() error {
	return d.control(service.ControlAction[4])
}
func (d *Daemon) Install() error {
	return d.control(service.ControlAction[3])
}
func (d *Daemon) Restart() error {
	return d.control(service.ControlAction[2])
}
func (d *Daemon) Stop() error {
	return d.control(service.ControlAction[1])
}
func (d *Daemon) Start() error {
	return d.control(service.ControlAction[0])
}
func (d *Daemon) Run() error {
	return (*d.svr).Run()
}
