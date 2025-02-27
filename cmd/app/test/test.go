package test

import (
	"fmt"
	"github.com/inconshreveable/go-update"
	"github.com/kardianos/service"
	"github.com/kbinani/screenshot"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore"
	"github.com/xxl6097/go-service/pkg"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Test struct {
	ins gore.Install
}

func (t Test) OnVersion() string {
	pkg.Version()
	return pkg.AppVersion
}

func (t Test) OnConfig() *service.Config {
	return &service.Config{
		Name:        "aatest",
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
		Screenshot()
		// 替换为你的.exe文件的路径
		time.Sleep(time.Second * 5)
	}
}

func Screenshot() {
	n := screenshot.NumActiveDisplays()
	bindir, _ := os.Getwd()
	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			glog.Error(bounds, err)
		} else {
			fileName := fmt.Sprintf("%d_%dx%d.png", i, bounds.Dx(), bounds.Dy())
			fileName += filepath.Join(bindir, fileName)
			file, _ := os.Create(fileName)
			defer file.Close()
			png.Encode(file, img)

			glog.Errorf("#%d : %v \"%s\"\n", i, bounds, fileName)
		}
		glog.Flush()
	}
}

func doUpdate(url string) error {
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
	err := doUpdate(binurl)
	glog.Println("update", response, err)
	fmt.Fprintln(w, response)
	if err != nil {
		fmt.Fprintln(w, fmt.Sprintf("Error:%v", err))
		return
	} else {
		go func() {
			time.Sleep(time.Millisecond * 1000)
			glog.Println("准备重启")
			err = t.ins.Restart()
			glog.Println("准备成功")
		}()
	}
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

func Serve(t Test) {
	// 注册路由
	http.HandleFunc("/update", t.updateHandler)
	http.HandleFunc("/version", t.versionHandler)
	http.HandleFunc("/restart", t.restartHandler)

	// 启动 HTTP 服务器
	fmt.Println("Starting server at :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
