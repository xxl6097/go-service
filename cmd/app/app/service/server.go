package service

import (
	"context"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/cmd/app/app/service/middle"
	"github.com/xxl6097/go-service/gservice/ukey"
	"github.com/xxl6097/go-service/pkg"
	"log"
	"net/http"
	"os"
	"time"
)

// 处理 GET 请求
func (t *Service) updateHandler(w http.ResponseWriter, r *http.Request) {
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
func (t *Service) versionHandler(w http.ResponseWriter, r *http.Request) {
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
func (t *Service) testHandler(w http.ResponseWriter, r *http.Request) {
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
func (t *Service) handleGet(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(fmt.Sprintf("%s\ntimestamp: %s", pkg.Version(), t.timestamp)))
}

// 处理 GET 请求
func (t *Service) restartHandler(w http.ResponseWriter, r *http.Request) {
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
func (t *Service) uninstallHandler(w http.ResponseWriter, r *http.Request) {
	// 确保只处理 GET 请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintln(w, pkg.Version())
	t.service.Uninstall()
}

//func Serve(t *Service) {
//	// 注册路由
//	defer glog.Flush()
//	http.HandleFunc("/update", t.updateHandler)
//	http.HandleFunc("/version", t.versionHandler)
//	http.HandleFunc("/test", t.testHandler)
//	http.HandleFunc("/get", t.handleGet)
//	http.HandleFunc("/restart", t.restartHandler)
//	http.HandleFunc("/uninstall", t.uninstallHandler)
//
//	// 启动 HTTP 服务器
//	glog.Println("Starting server at http://localhost:8080")
//	glog.Println("update--->http://localhost:8080/update")
//	glog.Println("version--->http://localhost:8080/version")
//	glog.Println("test--->http://localhost:8080/test")
//	glog.Println("restart--->http://localhost:8080/restart")
//	glog.Println("uninstall--->http://localhost:8080/uninstall")
//	if err := http.ListenAndServe(":8080", nil); err != nil {
//		glog.Error("--->", err)
//	}
//}

func Serve(t *Service) {
	// 注册路由
	defer glog.Flush()
	// 创建默认的ServeMux
	mux := http.NewServeMux()
	// 注册路由处理函数
	mux.HandleFunc("/update", t.updateHandler)
	mux.HandleFunc("/version", t.versionHandler)
	mux.HandleFunc("/test", t.testHandler)
	mux.HandleFunc("/get", t.handleGet)
	mux.HandleFunc("/restart", t.restartHandler)
	mux.HandleFunc("/uninstall", t.uninstallHandler)

	// 启动 HTTP 服务器
	glog.Println("Starting server at http://localhost:8080")
	glog.Println("update--->http://localhost:8080/update")
	glog.Println("version--->http://localhost:8080/version")
	glog.Println("test--->http://localhost:8080/test")
	glog.Println("get--->http://localhost:8080/get")
	glog.Println("restart--->http://localhost:8080/restart")
	glog.Println("uninstall--->http://localhost:8080/uninstall")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// 创建服务器配置
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      middle.LoggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Printf("服务器启动，监听端口 %s\n", port)
	log.Fatal(server.ListenAndServe())
}
