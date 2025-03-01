package test

import (
	"fmt"
	"github.com/inconshreveable/go-update"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore"
	"github.com/xxl6097/go-service/pkg"
	"log"
	"net/http"
	"time"
)

type Test struct {
	ins gore.Install
}

func (t Test) OnUpgrade(s string, s2 string) (bool, []string) {
	//TODO implement me
	return false, nil
}

func (t Test) OnVersion() string {
	pkg.Version()
	return pkg.AppVersion
}

func (t Test) OnConfig() *service.Config {
	return &service.Config{
		Name:        pkg.AppName,
		DisplayName: fmt.Sprintf("A AAATest1 Service %s", pkg.AppVersion),
		Description: "A Golang AAATest1 Service..",
	}
}

func (t Test) OnInstall(s string) (bool, []string) {
	return true, []string{}
}

func (t Test) OnRun(i gore.Install) error {
	glog.Println("--->OnRun", i)
	t.ins = i
	glog.SetLogFile("./logs", "app.log")
	go Serve(t)
	for {
		glog.Println("run", time.Now().Format("2006-01-02 15:04:05"))
		time.Sleep(time.Second * 10)
	}
}

func (t Test) doUpdate(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = update.Apply(resp.Body, update.Options{})
	if err != nil {
		// error handlingg
		glog.Error(err)
	}

	return err
}

// 处理 GET 请求
func (t Test) updateHandler(w http.ResponseWriter, r *http.Request) {
	// 确保只处理 GET 请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取查询参数
	queryParams := r.URL.Query()

	glog.Println("update", r.URL.Path)
	// 获取单个参数值
	binurl := queryParams.Get("binurl")
	// 返回响应
	response := fmt.Sprintf("Hello, %s", binurl)
	glog.Println("update", response)
	err := t.ins.Upgrade(binurl)
	glog.Println("update", err)
	fmt.Fprintln(w, pkg.Version())
}

// 处理 GET 请求
func (t Test) versionHandler(w http.ResponseWriter, r *http.Request) {
	// 确保只处理 GET 请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pkg.Version()
	fmt.Fprintln(w, pkg.Version())
}

// 处理 GET 请求
func (t Test) restartHandler(w http.ResponseWriter, r *http.Request) {
	// 确保只处理 GET 请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintln(w, pkg.Version())
	t.ins.Restart()
}

// 处理 GET 请求
func (t Test) uninstallHandler(w http.ResponseWriter, r *http.Request) {
	// 确保只处理 GET 请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintln(w, pkg.Version())
	t.ins.Uninstall()
}

func Serve(t Test) {
	// 注册路由
	http.HandleFunc("/update", t.updateHandler)
	http.HandleFunc("/version", t.versionHandler)
	http.HandleFunc("/restart", t.restartHandler)
	http.HandleFunc("/uninstall", t.uninstallHandler)

	// 启动 HTTP 服务器
	fmt.Println("Starting server at :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
