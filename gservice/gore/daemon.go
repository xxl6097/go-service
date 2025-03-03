package gore

import (
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
)

type Daemon struct {
	conf *service.Config
	svr  service.Service
	shut service.Interface
}

func NewDaemon(shut service.Shutdowner, conf *service.Config) (*Daemon, error) {
	this := &Daemon{
		shut: shut,
	}
	s, e := service.New(shut, conf)
	if e != nil {
		return nil, fmt.Errorf("service new err", e)
	}
	if s == nil {
		return nil, fmt.Errorf("create service is nil")
	}
	this.conf = conf
	this.svr = s
	service.Interactive()
	return this, nil
}

func (d *Daemon) control(cmd string) error {
	e := service.Control(d.svr, cmd)
	if e != nil {
		fmt.Println(cmd, e)
		return e
	}
	return nil
}

func (d *Daemon) Status() (service.Status, error) {
	return d.svr.Status()
}
func (d *Daemon) IsRunning() bool {
	if d.svr == nil {
		return false
	}
	status, err := d.Status()
	if err != nil {
		//glog.Println(err)
		return false
	}
	if status == service.StatusRunning {
		//glog.Println(d.conf.Name, "is running")
		return true
	} else if status == service.StatusStopped {
		//glog.Println(d.conf.Name, "is stopped")
	} else {
		//glog.Println(d.conf.Name, "StatusUnknown", status)
	}
	return false
}

func (d *Daemon) Uninstall() error {
	return d.control(service.ControlAction[4])
}

//	func (d *Daemon) Install(args []string) error {
//		if d.conf.Option == nil {
//			d.conf.Option = make(map[string]interface{})
//		}
//		d.conf.Option["Interactive"] = true
//		if args != nil && len(args) > 0 {
//			d.conf.Arguments = append(d.conf.Arguments, args...)
//			glog.Flush()
//			s, e := service.New(d.shut, d.conf)
//			if e != nil {
//				glog.Error("service new failed ", e)
//				return nil
//			}
//			d.svr = s
//		}
//		return d.control(service.ControlAction[3])
//	}
func (d *Daemon) Install() error {
	defer glog.Flush()
	if d.conf.Option == nil {
		d.conf.Option = make(map[string]interface{})
	}
	d.conf.Option["Interactive"] = true
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
	return d.svr.Run()
}

//func (d *Daemon) GetService() service.Service {
//	return d.svr
//}
