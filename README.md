go get github.com/kardianos/service
go get -u github.com/xxl6097/glog@v0.1.30
go get -u github.com/xxl6097/go-service@v0.4.13
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-linkmode internal" -o AAServiceApp.exe main.go

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o AAATest1.exe main.go

## 测试截图

go get github.com/kbinani/screenshot

go get -u github.com/inconshreveable/go-update



## goversioninfo

```

go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest

go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
```
