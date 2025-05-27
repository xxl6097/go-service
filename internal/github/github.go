package github

import (
	"encoding/json"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/internal/github/model"
	"github.com/xxl6097/go-service/pkg/utils"
	"io"
	"net/http"
	"os"
	"strings"
)

var GithuApiHost = "https://api.github.com/repos/xxl6097/go-frp-panel/releases/latest"

type githubapi struct {
	result  *model.GitHubModel
	proxies []string
}

func request() ([]byte, error) {
	glog.Debug("request", GithuApiHost)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", GithuApiHost, nil)
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
func Request() *githubapi {
	defer func() {
		if err := recover(); err != nil {
			glog.Debug(err)
		}
	}()
	this := &githubapi{}
	body, err := request()
	var result model.GitHubModel
	err = json.Unmarshal(body, &result)
	if err != nil {
		panic(fmt.Errorf("github请求失败 %v", err))
	}
	if this.result == nil {
		panic("github请求结果空～")
	}
	this.proxies = utils.ParseMarkdownCodeToStringArray(result.Body)
	return this
}

func (this *githubapi) Upgrade(fullName string, fn func(string, string, string)) *githubapi {
	oldVersion := utils.GetVersionByFileName(fullName)
	hasNewVersion := utils.CompareVersions(this.result.TagName, oldVersion)
	glog.Debug("最新版本:", this.result.TagName)
	glog.Debug("本地版本:", oldVersion)
	if hasNewVersion > 0 {
		newVersionAppName := utils.ReplaceNewVersionBinName(fullName, this.result.TagName)
		var fullUpUrl, patchUpUrl string
		patchName := fmt.Sprintf("%s.patch", newVersionAppName)
		for _, asset := range this.result.Assets {
			if strings.Compare(strings.ToLower(asset.Name), strings.ToLower(newVersionAppName)) == 0 {
				fullUpUrl = asset.BrowserDownloadUrl
			} else if strings.Compare(strings.ToLower(asset.Name), strings.ToLower(patchName)) == 0 {
				patchUpUrl = asset.BrowserDownloadUrl
			}
		}

		if hasNewVersion != 1 {
			//版本之间只有相差一个版本号才能差量升级
			patchUpUrl = ""
		}
		index := strings.Index(this.result.Body, "---")
		releaseNote := this.result.Body
		if index > 0 {
			releaseNote = releaseNote[:index]
		}
		if fn != nil {
			fn(patchUpUrl, fullUpUrl, releaseNote)
		}
	}

	return this
}
