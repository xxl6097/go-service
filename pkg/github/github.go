package github

import (
	"encoding/json"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/pkg/github/model"
	"github.com/xxl6097/go-service/pkg/utils"
	"golang.org/x/mod/modfile"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func init() {
	LoadGithubKey()
}

func LoadGithubKey() {
	fpath := filepath.Join(glog.AppHome("obj"), "githubKey.dat")
	obj, err := utils.LoadWithGob[model.GithubKey](fpath)
	if err == nil && obj.ClientId != "" && obj.ClientSecret != "" {
		os.Setenv("GITHUB_CLIENT_ID", obj.ClientId)
		os.Setenv("GITHUB_CLIENT_SECRET", obj.ClientSecret)
	}
}

var (
	instance *githubApi
	once     sync.Once
)

type githubApi struct {
	result  *model.GitHubModel
	proxies []string
	data    any
	err     error
}

// Api 返回单例实例
func Api() *githubApi {
	once.Do(func() {
		instance = &githubApi{} // 初始化逻辑
		fmt.Println("github api Singleton instance created")
	})
	return instance
}

func request(githubUser, repoName string) ([]byte, error) {
	baseUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", githubUser, repoName)
	glog.Debug("request", baseUrl)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseUrl, nil)
	clientId := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if clientId != "" || clientSecret != "" {
		req.SetBasicAuth(clientId, clientSecret) // 自动 Base64 编码
	}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("请求失败:%v\n", err)
		return nil, err
	}
	defer resp.Body.Close() // 必须关闭响应体 [1,5,8](@ref)
	if resp.StatusCode != 200 {
		glog.Error(resp.StatusCode, resp.Status)
		return nil, fmt.Errorf("请求失败 %v %v", resp.StatusCode, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		glog.Error("github请求失败", err)
		return nil, err
	}
	return body, nil
}
func (this *githubApi) Request(githubUser, repoName string) *githubApi {
	//this := &GithubApi{}
	body, err := request(githubUser, repoName)
	var result model.GitHubModel
	err = json.Unmarshal(body, &result)
	if err != nil {
		this.err = fmt.Errorf("github请求失败 %v", err)
	}
	this.result = &result
	if this.result == nil {
		this.err = fmt.Errorf("github请求结果空~")
	}
	glog.Debug("TagName", this.result.TagName)
	this.proxies = utils.ParseMarkdownCodeToStringArray(result.Body)
	return this
}

func (this *githubApi) DefaultRequest() *githubApi {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
	}()
	data, err := os.ReadFile("go.mod")
	if err != nil {
		panic(err)
	}

	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		panic(err)
	}
	if !strings.Contains(modFile.Module.Mod.Path, "github.com") {
		panic(fmt.Sprintf("module name not Contains github.com"))
	}
	userName := filepath.Base(filepath.Dir(modFile.Module.Mod.Path))
	repoName := filepath.Base(modFile.Module.Mod.Path)
	return this.Request(userName, repoName)
}

func (this *githubApi) CheckUpgrade(fullName string, fn func(string, string, string)) *githubApi {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
	}()
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
	}()
	if this.result == nil {
		this.DefaultRequest()
	}
	oldVersion := utils.GetVersionByFileName(fullName)
	hasNewVersion := utils.CompareVersions(this.result.TagName, oldVersion)
	glog.Debug("最新版本:", this.result.TagName)
	glog.Debug("本地版本:", oldVersion)
	if hasNewVersion > 0 {
		newVersionAppName := utils.ReplaceNewVersionBinName(fullName, this.result.TagName)
		var fullUrl, patchUrl string
		patchName := fmt.Sprintf("%s.patch", newVersionAppName)
		for _, asset := range this.result.Assets {
			if strings.Compare(strings.ToLower(asset.Name), strings.ToLower(newVersionAppName)) == 0 {
				fullUrl = asset.BrowserDownloadUrl
			} else if strings.Compare(strings.ToLower(asset.Name), strings.ToLower(patchName)) == 0 {
				patchUrl = asset.BrowserDownloadUrl
			}
		}

		if hasNewVersion != 1 {
			//版本之间只有相差一个版本号才能差量升级
			patchUrl = ""
		}
		index := strings.Index(this.result.Body, "---")
		releaseNote := this.result.Body
		if index > 0 {
			releaseNote = releaseNote[:index]
		}
		if fn != nil {
			fn(patchUrl, fullUrl, releaseNote)
		}

		this.data = map[string]interface{}{
			"fullUrl":      fullUrl,
			"patchUrl":     patchUrl,
			"releaseNotes": fmt.Sprintf("### ✅ 新版本\r\n* %s\r\n%s", this.result.TagName, releaseNote),
		}
	} else {
		this.err = fmt.Errorf("【%s】已是最新版本～", oldVersion)
	}

	return this
}

func (this *githubApi) GetProxyUrls(fileUrl string) []string {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
	}()
	if this.result == nil {
		this.DefaultRequest()
	}
	newProxy := make([]string, 0)
	if this.proxies == nil || len(this.proxies) <= 0 {
		newProxy = append(newProxy, fileUrl)
	} else {
		for _, proxy := range this.proxies {
			newUrl := fmt.Sprintf("%s%s", proxy, fileUrl)
			newProxy = append(newProxy, newUrl)
		}
	}
	return newProxy
}

func (this *githubApi) Result() (any, error) {
	return this.data, this.err
}
func (this *githubApi) GetModel() *model.GitHubModel {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
	}()
	if this.result == nil {
		this.DefaultRequest()
	}
	return this.result
}

func (this *githubApi) GetDownloadUrl(fn func(string, *model.Assets) bool) string {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
	}()
	if this.result == nil {
		this.DefaultRequest()
	}
	for _, asset := range this.result.Assets {
		//if strings.Compare(strings.ToLower(asset.Name), strings.ToLower(name)) == 0 {
		//	return asset.BrowserDownloadUrl
		//}
		if fn != nil && fn(this.result.TagName, &asset) {
			return asset.BrowserDownloadUrl
		}
	}
	return ""
}
