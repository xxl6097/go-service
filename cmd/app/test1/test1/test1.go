package test1

import (
	"context"
	"errors"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/gservice/gore"
	"github.com/xxl6097/go-service/gservice/ukey"
	"github.com/xxl6097/go-service/pkg"
	"log"
	"net/http"
	"time"
)

type Test1 struct {
	service gore.IGService
}

func (t Test1) GetAny() any {
	return "这是一段测试数据..."
}

func (t Test1) OnInit() *service.Config {
	return &service.Config{
		Name:        pkg.AppName,
		DisplayName: fmt.Sprintf("A AAATest1 Service %s", pkg.AppVersion),
		Description: "A Golang AAATest1 Service..",
	}
}

func (t Test1) OnVersion() string {
	pkg.Version()
	fmt.Println(ukey.GetBuffer())
	return pkg.AppVersion
}

func (t Test1) OnRun(service gore.IGService) error {
	t.service = service
	glog.SetLogFile("./logs", "app.log")
	go Serve(t)
	for {
		glog.Println("run", time.Now().Format("2006-01-02 15:04:05"))
		time.Sleep(time.Second * 10)
	}
}

// 处理 GET 请求
func (t Test1) updateHandler(w http.ResponseWriter, r *http.Request) {
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
	err := t.service.Upgrade(r.Context(), binurl)
	glog.Println("update", err)
	fmt.Fprintln(w, pkg.Version())
}

// 处理 GET 请求
func (t Test1) versionHandler(w http.ResponseWriter, r *http.Request) {
	// 确保只处理 GET 请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pkg.Version()
	fmt.Println(ukey.GetBuffer())
	fmt.Fprintln(w, pkg.Version())
}

// 处理 GET 请求
func (t Test1) testHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	select {
	case <-time.After(10 * time.Second):
		fmt.Println("Operation completed")
		w.Write([]byte("Operation completed"))
	case <-ctx.Done():
		// 客户端断开或超时
		if ctx.Err() == context.Canceled {
			fmt.Println("Client disconnected")
		}
	}
}

// 处理 GET 请求
func (t Test1) restartHandler(w http.ResponseWriter, r *http.Request) {
	// 确保只处理 GET 请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintln(w, pkg.Version())
	//t.service.Restart()
	t.service.RunCmd("restart")
}

// 处理 GET 请求
func (t Test1) uninstallHandler(w http.ResponseWriter, r *http.Request) {
	// 确保只处理 GET 请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintln(w, pkg.Version())
	t.service.Uninstall()
}

func blockingOperation() string {
	count := 0
	for {
		count++
		time.Sleep(time.Second)
		fmt.Println("hello world..")
		if count > 2 {
			return "err"
		}
	}
	return "ok"
}

func Serve(t Test1) {
	// 注册路由
	http.HandleFunc("/update", t.updateHandler)
	http.HandleFunc("/version", t.versionHandler)
	http.HandleFunc("/test", t.testHandler)
	http.HandleFunc("/restart", t.restartHandler)
	http.HandleFunc("/uninstall", t.uninstallHandler)

	// 启动 HTTP 服务器
	fmt.Println("Starting server at :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func BlockingFunction[T any](c context.Context, timeout time.Duration, callback func() T) (T, error) {
	ctx, cancel := context.WithTimeout(c, timeout)
	defer cancel()
	resultChan := make(chan T)
	go func() {
		result := callback()
		resultChan <- result
	}()
	var zero T // 声明 T 的零值
	select {
	case res := <-resultChan:
		return res, nil
	case <-ctx.Done():
		return zero, errors.New("timeout")
	}
}

func blockingHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resultChan := make(chan string)
	go func() {
		result := blockingOperation()
		resultChan <- result
	}()

	select {
	case res := <-resultChan:
		w.Write([]byte(res))
	case <-ctx.Done():
		http.Error(w, "处理超时", http.StatusGatewayTimeout)
	}
}

// 处理 GET 请求
func testHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	select {
	case <-time.After(10 * time.Second):
		fmt.Println("Operation completed")
		w.Write([]byte("Operation completed"))
	case <-ctx.Done():
		// 客户端断开或超时
		//if ctx.Err() == context.Canceled {
		//}
		fmt.Println("Client disconnected", ctx.Err())
	}
}
func ServeTesting() {
	// 注册路由
	http.HandleFunc("/blocking", blockingHandler)
	http.HandleFunc("/test", testHandler)

	// 启动 HTTP 服务器
	fmt.Println("Starting server at :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
