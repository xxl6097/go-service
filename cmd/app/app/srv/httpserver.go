package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/assets"
	_ "github.com/xxl6097/go-service/assets/we"
	"github.com/xxl6097/go-service/internal/github"
	"github.com/xxl6097/go-service/pkg"
	"github.com/xxl6097/go-service/pkg/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var logQueue = NewLogQueue()

func init() {
	glog.Hook(func(bytes []byte) {
		logQueue.AddMessage(string(bytes[2:]))
	})
}

type Message[T any] struct {
	Action string `json:"action,omitempty"`
	Data   T      `json:"data,omitempty"`
}

type Result struct {
	Code int    `json:"code"`
	Data string `json:"data,omitempty"`
}

// 处理 GET 请求
func (t *Service) updateHandler(binurl string, ctx context.Context) ([]byte, error) {
	response := fmt.Sprintf("Hello, %s", binurl)
	glog.Println("update", response)
	if t.gs == nil {
		return []byte(response), fmt.Errorf("gs is nil")
	}
	err := t.gs.Upgrade(ctx, binurl)
	glog.Println("update", err)
	return []byte(pkg.AppVersion), err
}

// 处理 GET 请求
func (t *Service) patchUpdateHandler(binurl string, ctx context.Context) ([]byte, error) {
	response := fmt.Sprintf("Hello, %s", binurl)
	glog.Println("patchUpdateHandler", response)
	if t.gs == nil {
		return []byte(response), fmt.Errorf("gs is nil")
	}
	err := t.gs.Upgrade(ctx, binurl)
	glog.Println("patchUpdateHandler", err)
	return []byte(pkg.AppVersion), err
}

// 处理 GET 请求
func (t *Service) versionHandler() ([]byte, error) {
	return []byte(fmt.Sprintf("\r\n%s", pkg.Version())), nil
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
func (t *Service) handleGet() ([]byte, error) {
	return []byte(fmt.Sprintf("%s\ntimestamp: %s", pkg.Version(), t.timestamp)), nil
}

// 处理 GET 请求
func (t *Service) handleDelete() ([]byte, error) {
	// 获取当前可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	// 确保路径是绝对路径
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return nil, err
	}
	err = os.Remove(exePath)
	if err != nil {
		return []byte("失败"), err
	} else {
		return []byte("成功"), err
	}
}

// 处理 GET 请求
func (t *Service) restartHandler() ([]byte, error) {
	//err := t.gs.RunCmd("restart")
	err := t.gs.Restart()
	return nil, err
}

// 处理 GET 请求
func (t *Service) uninstallHandler() ([]byte, error) {
	err := t.gs.UnInstall()
	return nil, err
}

// 处理 GET 请求
func (t *Service) checkVersionHandler() ([]byte, error) {
	err := github.Request("xxl6097", "go-service").Upgrade(pkg.BinName, func(pathUrl string, fullUrl string, releaseNotes string, api *github.GithubApi) {
		glog.Debug(pathUrl, fullUrl, releaseNotes)
		if pathUrl != "" {
			glog.Debug("差量升级")
			fileLocalUri := utils.DownloadFileWithCancelByUrls(api.GetProxyUrls(pathUrl))
			err := t.gs.Upgrade(context.Background(), fileLocalUri)
			//err := t.gs.Upgrade(context.Background(), pathUrl)
			glog.Debug("差量升级结果", err)
		} else if fullUrl != "" {
			glog.Debug("全量升级")
			//err := t.gs.Upgrade(context.Background(), utils.DownloadFileWithCancelByUrls(api.GetUrl(fullUrl)))
			fileLocalUri := utils.DownloadFileWithCancelByUrls(api.GetProxyUrls(fullUrl))
			err := t.gs.Upgrade(context.Background(), fileLocalUri)
			glog.Debug("全量升级结果", err)
		}

	}).Result()
	return nil, err
}

// 处理 GET 请求
func (t *Service) handleLog() ([]byte, error) {
	return []byte(fmt.Sprintf("\r\n%s", pkg.Version())), nil
}

// /api/shutdown
func (this *Service) ApiClear() ([]byte, error) {
	binPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	binDir := filepath.Dir(binPath)
	clientsDir := filepath.Join(binDir, "clients")
	err = utils.DeleteAllDirector(clientsDir)
	err = utils.DeleteAllDirector(glog.AppHome())
	if err != nil {
		return nil, err
	} else {
		return []byte("删除成功"), nil
	}
}

func (this *Service) ApiCMD(arg string) ([]byte, error) {
	if arg == "" {
		return nil, fmt.Errorf("arg is empty")
	}
	args := strings.Split(arg, " ")
	if args == nil || len(args) == 0 {
		return nil, fmt.Errorf("args is empty")
	}
	glog.Infof("args: %s", args)
	return utils.RunCmdWithSudo(args...)
}

func (this *Service) ApiSelfCMD(arg string) ([]byte, error) {
	if arg == "" {
		return nil, fmt.Errorf("arg is empty")
	}
	args := strings.Split(arg, " ")
	if args == nil || len(args) == 0 {
		return nil, fmt.Errorf("args is empty")
	}
	glog.Infof("args: %s", args)
	err := this.gs.RunCMD(args...)
	return nil, err
}

// 处理 GET 请求
func (t *Service) handleSudo() ([]byte, error) {
	if err := utils.RunWithSudo(); err != nil {
		msg := fmt.Sprintf("获取管理员权限失败: %v\n", err)
		glog.Println(msg)
		return []byte(msg), err
	}
	msg := "已获取管理员权限，正在执行敏感操作..."
	glog.Println(msg)
	return []byte(msg), nil
}

func (this *Service) apiCommand(w http.ResponseWriter, r *http.Request) {
	var errMsg error
	var data []byte
	defer func() {
		var msg string
		if data != nil {
			msg = fmt.Sprintf("data: %s", string(data))
		}
		if errMsg != nil {
			msg = fmt.Sprintf("%s err: %s", msg, errMsg)
		}

		res := Result{
			Code: 0,
			Data: msg,
		}
		jsonData, err := json.Marshal(res)
		if err != nil {
			glog.Errorf("json marshal err: %v", err)
			return
		}
		//glog.Debug(string(jsonData))
		glog.Debug("Code", res.Code)
		_, _ = w.Write(jsonData)
	}()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg = fmt.Errorf("body读取失败 %v", err)
		return
	}
	if body == nil {
		errMsg = fmt.Errorf("body is nil")
		return
	}

	data, errMsg = this.handleMessage(body, r)
}

func addStatic(subRouter *mux.Router) {
	subRouter.Handle("/favicon.ico", http.FileServer(assets.FileSystem)).Methods("GET")
	subRouter.PathPrefix("/static/").Handler(
		assets.MakeHTTPGzipHandler(http.StripPrefix("/static/", http.FileServer(assets.FileSystem))),
	).Methods("GET")
	subRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusMovedPermanently)
	})
}

func Server(p int, t *Service) {
	router := mux.NewRouter() // 创建路由器实例[1,5](@ref)
	assets.Load("")

	staticPrefix := "/log/"
	baseDir := glog.AppHome()
	router.PathPrefix(staticPrefix).Handler(http.StripPrefix(staticPrefix, http.FileServer(http.Dir(baseDir))))

	// 注册路由处理函数
	router.HandleFunc("/api/cmd", t.apiCommand)
	router.HandleFunc("/api/sse-stream", SseHandler(logQueue))

	addStatic(router.NewRoute().Subrouter())

	if p <= 0 {
		p = 9090
	}

	port := fmt.Sprintf(":%d", p)
	address := "http://localhost" + port
	// 启动 HTTP 服务器
	glog.Printf("Starting server at %s\n", address)

	fmt.Println(address)
	// 启动服务器
	err := http.ListenAndServe(port, router)
	glog.Fatal(err)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to Home Page"))
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("User API Endpoint"))
}

func (this *Service) handleMessage(body []byte, r *http.Request) ([]byte, error) {
	var msg Message[map[string]interface{}]
	err := json.Unmarshal(body, &msg)
	if err != nil {
		return nil, fmt.Errorf("解析Json对象失败 %v", err)
	}
	glog.Debugf("body:%s", string(body))
	switch msg.Action {
	case "update":
		if msg.Data == nil {
			return nil, fmt.Errorf("msg.Data is nil")
		}
		if msg.Data["data"] == nil {
			return nil, fmt.Errorf("msg.Data[\"data\"] is nil")
		}
		if v, ok := msg.Data["data"].(string); ok {
			return this.updateHandler(v, r.Context())
		} else {
			return nil, fmt.Errorf("msg.Data[\"data\"] is nil")
		}
	case "patch":
		if msg.Data == nil {
			return nil, fmt.Errorf("msg.Data is nil")
		}
		if msg.Data["data"] == nil {
			return nil, fmt.Errorf("msg.Data[\"data\"] is nil")
		}
		if v, ok := msg.Data["data"].(string); ok {
			return this.patchUpdateHandler(v, r.Context())
		} else {
			return nil, fmt.Errorf("msg.Data[\"data\"] is nil")
		}
	case "version":
		return this.versionHandler()
	case "sudo":
		return this.handleSudo()
	case "get":
		return this.handleGet()
	case "delete":
		return this.handleDelete()
	case "restart":
		return this.restartHandler()
	case "uninstall":
		return this.uninstallHandler()
	case "checkversion":
		return this.checkVersionHandler()
	case "log":
		return this.handleLog()
	case "clear":
		return this.ApiClear()
	case "self":
		if msg.Data == nil {
			return nil, fmt.Errorf("msg.Data is nil")
		}
		if msg.Data["data"] == nil {
			return nil, fmt.Errorf("msg.Data[\"data\"] is nil")
		}
		if v, ok := msg.Data["data"].(string); ok {
			return this.ApiSelfCMD(v)
		} else {
			return nil, fmt.Errorf("msg.Data[\"data\"] is nil")
		}
	case "cmd":
		if msg.Data == nil {
			return nil, fmt.Errorf("msg.Data is nil")
		}
		if msg.Data["data"] == nil {
			return nil, fmt.Errorf("msg.Data[\"data\"] is nil")
		}
		if v, ok := msg.Data["data"].(string); ok {
			return this.ApiCMD(v)
		} else {
			return nil, fmt.Errorf("msg.Data[\"data\"] is nil")
		}
	}
	return nil, nil
}
