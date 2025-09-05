package srv

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/assets"
	"github.com/xxl6097/go-service/assets/we"
	"github.com/xxl6097/go-service/cmd/app/app/wx"
	"github.com/xxl6097/go-service/pkg"
	"github.com/xxl6097/go-service/pkg/github"
	"github.com/xxl6097/go-service/pkg/github/model"
	"github.com/xxl6097/go-service/pkg/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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
	Msg  string `json:"msg,omitempty"`
	Data any    `json:"data,omitempty"`
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
	glog.Println("patchUpdate", response)
	if t.gs == nil {
		return []byte(response), fmt.Errorf("gs is nil")
	}
	err := t.gs.Upgrade(ctx, binurl)
	glog.Println("patchUpdate err", err)
	return []byte(pkg.AppVersion), err
}

func GetVersion() map[string]interface{} {
	hostName, _ := os.Hostname()
	return map[string]interface{}{
		"hostName":    hostName,
		"appName":     pkg.AppName,
		"appVersion":  pkg.AppVersion,
		"buildTime":   pkg.BuildTime,
		"gitRevision": pkg.GitRevision,
		"gitBranch":   pkg.GitBranch,
		"goVersion":   pkg.GoVersion,
		"displayName": pkg.DisplayName,
		"description": pkg.Description,
		"osType":      pkg.OsType,
		"arch":        pkg.Arch,
	}
}

// 处理 GET 请求
func (t *Service) versionHandler() ([]byte, error) {
	return []byte(fmt.Sprintf("\r\n%s", pkg.Version())), nil
}

// 处理 GET 请求
func (t *Service) apiVersion(w http.ResponseWriter, r *http.Request) {
	version := GetVersion()
	bb, err := json.Marshal(version)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bb)
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
func (t *Service) handleGet() (any, error) {
	return []byte(fmt.Sprintf("%s\ntimestamp: %s", pkg.Version(), t.timestamp)), nil
}

// 处理 GET 请求
func (t *Service) handleDelete() (any, error) {
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
func (t *Service) restartHandler() (any, error) {
	//err := t.gs.RunCmd("restart")
	err := t.gs.Restart()
	return nil, err
}

// 处理 GET 请求
func (t *Service) handlePanic() (any, error) {
	panic(fmt.Sprintf("handlePanic %v", time.Now().Format(time.DateTime)))
	os.Exit(-1)
	return nil, nil
}

// 处理 GET 请求
func (t *Service) handleNull() (any, error) {
	var testPoint *model.Node
	glog.Println(testPoint.FilePath)
	return nil, nil
}

// 处理 GET 请求
func (t *Service) uninstallHandler() (any, error) {
	err := t.gs.UnInstall()
	return nil, err
}

// 处理 GET 请求
func (t *Service) checkVersionHandler() (any, error) {
	//
	github.Api().SetName("xxl6097", "go-service")
	data, err := github.Api().CheckUpgrade(pkg.BinName)
	return data, err
}

// 处理 GET 请求
func (t *Service) confirmUpgrade(r *http.Request, data any) (any, error) {
	if data == nil {
		return nil, fmt.Errorf("msg.Data is nil")
	}
	switch v := data.(type) {
	case map[string]interface{}:
		if v["data"] == nil {
			return nil, fmt.Errorf("msg.Data[\"data\"] is nil")
		}
		if value, ok := v["data"].(string); ok {
			fmt.Println(value)
			urls := github.Api().GetProxyUrls(value)
			fileUri := utils.DownloadFileWithCancelByUrls(urls)
			if fileUri != "" {
				glog.Debug("升级文件", fileUri)
				err := t.gs.Upgrade(r.Context(), fileUri)
				return nil, err
			} else {
				glog.Debug("文件地址空fileUri", fileUri)
				return nil, fmt.Errorf("文件地址空fileUri=%s", fileUri)
			}
		} else {
			return nil, fmt.Errorf("msg.Data[\"data\"] is nil")
		}
	}

	return data, nil
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
	var data any
	defer func() {
		res := Result{
			Code: 0,
		}
		if data != nil {
			switch v := data.(type) {
			case []byte:
				res.Data = string(v)
			case string:
				res.Data = v
			case error:
				res.Data = v.Error()
			case map[string]interface{}:
				res.Data = v
			default:
				res.Data = fmt.Sprintf("%v", v)
			}
		}
		if errMsg != nil {
			res.Code = -1
			res.Msg = fmt.Sprintf("%s", errMsg)
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
	subRouter.Handle("/favicon.ico", http.FileServer(we.FileSystem)).Methods("GET")
	subRouter.PathPrefix("/static/").Handler(
		assets.MakeHTTPGzipHandler(http.StripPrefix("/static/", http.FileServer(we.FileSystem))),
	).Methods("GET")
	subRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusMovedPermanently)
	})
}

// 与你后台配置一致的 Token
const yourToken = "het002402"

// CheckSignature 验证微信请求签名
// 参数: signature, timestamp, nonce 来自请求URL, token 为你在微信后台配置的令牌
// 返回值: 验证通过返回 true，否则返回 false
func CheckSignature(signature, timestamp, nonce, token string) bool {
	// 1. 将 token, timestamp, nonce 放入切片
	params := []string{token, timestamp, nonce}
	// 2. 按字典序排序
	sort.Strings(params)
	// 3. 拼接成一个字符串
	var combinedStr string
	for _, s := range params {
		combinedStr += s
	}
	// 4. 对拼接后的字符串进行 sha1 加密
	hasher := sha1.New()
	hasher.Write([]byte(combinedStr))
	calculatedSignature := fmt.Sprintf("%x", hasher.Sum(nil)) // %x 表示格式化为小写十六进制
	// 5. 将加密后的字符串与 signature 对比
	return calculatedSignature == signature
}

func wechatMessageHandler(w http.ResponseWriter, userMsg wx.EventMessage) {
	// 1. 构造回复消息
	replyMsg := wx.CreateTextResponse(
		userMsg.FromUserName, // 接收方：发送消息的用户OpenID
		userMsg.ToUserName,   // 发送方：公众号ID
		"您好，这是自动回复：\r\n"+
			"姓名：夏小力\r\n"+
			"性别：男\r\n"+
			"工作：码农", // 回复内容
	)

	// 2. 将结构体序列化为XML字节切片
	xmlData, err := xml.Marshal(replyMsg)
	if err != nil {
		log.Printf("Error marshaling XML response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 3. 设置响应头并返回XML
	w.Header().Set("Content-Type", "application/xml") // 务必设置为 application/xml[1,2](@ref)
	w.Write(xmlData)
	log.Println("Text response sent successfully.")
}

func decodeWxMes(w http.ResponseWriter, body []byte) error {
	var baseMsg struct {
		MsgType string `xml:"MsgType"`
	}
	if err := xml.Unmarshal(body, &baseMsg); err != nil {
		// 处理错误
		return err
	}
	switch baseMsg.MsgType {
	case "text":
		var textMsg wx.TextMessage
		err := xml.Unmarshal(body, &textMsg)
		fmt.Println(textMsg)
		return err
	case "event":
		var eventMsg wx.EventMessage
		err := xml.Unmarshal(body, &eventMsg)
		fmt.Println(eventMsg)
		wechatMessageHandler(w, eventMsg)
		return err
	case "image":
	}

	return nil
}

// AppID:wxbe2c2961b236427f
// AppSecret:667fc391b1ca8f4c58d1b5f224356ad5
// http://v.uuxia.cn/api/wx/push
// het002402
// T23zgdkUCPCxaJmTvsqhYRidouCQXPnBgLq2qTdTtjK
func apiPush(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s %s\n", r.Method, r.URL.String(), r.Proto)
	queryParams := r.URL.Query()
	echostr := queryParams.Get("echostr")
	nonce := queryParams.Get("nonce")
	openid := queryParams.Get("openid")
	signature := queryParams.Get("signature")
	timestamp := queryParams.Get("timestamp")
	fmt.Println(echostr, openid, timestamp, signature)

	switch r.Method {
	case http.MethodPost:
		ok := CheckSignature(signature, timestamp, nonce, yourToken)
		fmt.Println(ok)
		// 验证签名
		if ok {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close() // 确保关闭 Body
			err = decodeWxMes(w, body)
			fmt.Printf("%s %v\n", string(body), err)
			// 若确认此次GET请求来自微信服务器，请原样返回 echostr 参数内容
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(echostr))
			log.Println("微信服务器验证成功")
		} else {
			// 校验失败
			w.WriteHeader(http.StatusForbidden)
			log.Println("微信服务器验证失败: 签名无效")
			return
		}
		break
	case http.MethodGet:
		ok := CheckSignature(signature, timestamp, echostr, yourToken)
		fmt.Println(ok)
		// 验证签名
		if ok {
			// 若确认此次GET请求来自微信服务器，请原样返回 echostr 参数内容
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(echostr))
			log.Println("微信服务器验证成功")
		} else {
			// 校验失败
			w.WriteHeader(http.StatusForbidden)
			log.Println("微信服务器验证失败: 签名无效")
			return
		}
		break
	default:
		break
	}
}

func Server(p int, t *Service) {
	router := mux.NewRouter() // 创建路由器实例[1,5](@ref)

	staticPrefix := "/log/"
	baseDir := glog.AppHome()
	router.PathPrefix(staticPrefix).Handler(http.StripPrefix(staticPrefix, http.FileServer(http.Dir(baseDir))))

	// 注册路由处理函数
	router.HandleFunc("/api/cmd", t.apiCommand)
	router.HandleFunc("/api/version", t.apiVersion)
	router.HandleFunc("/api/wx/push", apiPush)
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

func (this *Service) handleMessage(body []byte, r *http.Request) (any, error) {
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
	case "panic":
		return this.handlePanic()
	case "null":
		return this.handlePanic()
	case "uninstall":
		return this.uninstallHandler()
	case "checkversion":
		return this.checkVersionHandler()
	case "confirm-upgrade":
		return this.confirmUpgrade(r, msg.Data)
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
