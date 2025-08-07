go get github.com/kardianos/service
go get -u github.com/xxl6097/glog@v0.1.50
go get -u github.com/xxl6097/go-service@v0.4.13
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest

### 差分包计算
go get -u github.com/kr/binarydist
go install  github.com/kr/binarydist@latest

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-linkmode internal" -o AAServiceApp.exe main.go

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o AAATest1.exe main.go

## 测试截图

go get github.com/kbinani/screenshot

go get -u github.com/inconshreveable/go-update


go get -u github.com/xxl6097/go-update@v0.0.6



## goversioninfo

```

go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest

go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
```


# 删除从最后一个 "-" 开始的所有字符
filename="log4j-core-2.7.jar"
base_pattern="${filename%%-[0-9]*}.jar"  # 结果：log4j-core.jar

# 匹配所有符合基础模式的文件
find ./lib -name "$base_pattern"


### 差分包
go install github.com/xxl6097/go-service/cmd/differ@latest