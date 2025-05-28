package github

import (
	"encoding/json"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/pkg/github/model"
	"github.com/xxl6097/go-service/pkg/utils"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

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

func (this *githubApi) Upgrade(fullName string, fn func(string, string, string)) *githubApi {
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
			this.data = map[string]interface{}{
				"fullUrl":      fullUrl,
				"patchUrl":     patchUrl,
				"releaseNotes": releaseNote,
			}
		}
	}

	return this
}

func (this *githubApi) GetProxyUrls(fileUrl string) []string {
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
