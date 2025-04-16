package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 生成 HMAC-SHA1 签名
func generateHMAC(key, message string) string {
	// 将密钥和消息转换为字节切片
	keyBytes := []byte(key)
	messageBytes := []byte(message)
	// 1. 创建 HMAC 实例，使用 SHA1 算法
	mac := hmac.New(sha1.New, keyBytes)
	// 2. 写入消息内容
	mac.Write(messageBytes)
	// 3. 计算签名摘要
	signatureBytes := mac.Sum(nil)
	// 4. 将二进制签名转换为 Base64 字符串
	return base64.StdEncoding.EncodeToString(signatureBytes)
}

func generateParams(params map[string]string) string {
	var sb strings.Builder
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for index, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(fmt.Sprintf("%v", params[k]))
		if index != len(keys)-1 {
			sb.WriteString("&")
		}
	}
	fmt.Println(sb.String())
	return sb.String()
}
func generateMethod(params map[string]string, method, host string) string {
	var sb strings.Builder
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sb.WriteString(method)
	sb.WriteString(host)
	sb.WriteString("?")
	sb.WriteString(generateParams(params))
	fmt.Println(sb.String())
	return sb.String()
}

func main() {
	baseUrl := "https://iot-cloudapi.clife.cn/v5x/open/api/device/list"
	secretKey := "81120c674bf0451b829833d04e342520"
	secretId := "70e8765e8a78475bb0ff0489e47da1ff"
	productId := "13571"
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)

	params := map[string]string{}
	params["timestamp"] = timestamp
	params["secretId"] = secretId
	params["productId"] = productId
	data := generateMethod(params, "POST", baseUrl)
	fmt.Println("加密数据", data)
	fmt.Println("加密密钥", secretKey)
	sign := generateHMAC(secretKey, data)
	fmt.Println("密钥", sign)
	params["sign"] = sign
	baseUrl += fmt.Sprintf("?%s", generateParams(params))
	client := &http.Client{}
	req, _ := http.NewRequest("POST", baseUrl, nil)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bb, _ := io.ReadAll(resp.Body)
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response:", string(bb))

}
