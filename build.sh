#!/bin/bash
#修改为自己的应用名称
appname=AAATest
DisplayName=AAATest
Description="基于Go语言的服务程序，可安装和管理第三方应用程序，可运行于Windows、Linux、Macos、Openwrt等各类操作系统。"
version=0.0.0

function tag() {
    version=$(getversion)
    echo "current version:${version}"
    git add .
    git commit -m "release v${version}"
    git tag -a v$version -m "release v${version}"
    git push origin v$version
    echo $version >version.txt
}


function build() {
  rm -rf bin
  os=$1
  arch=$2
  CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -ldflags "$ldflags -s -w -linkmode internal" -o ./bin/${appname}_v${version}_${os}_${arch} ./cmd/app
  bash <(curl -s -S -L http://uuxia.cn:8086/up) ./bin/${appname}_v${version}_${os}_${arch}
}

function build_win() {
  rm -rf bin
  os=$1
  arch=$2
  CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -ldflags "$ldflags -s -w -linkmode internal" -o ./bin/${appname}_v${version}_${os}_${arch}.exe ./cmd/app
  bash <(curl -s -S -L http://uuxia.cn:8086/up) ./bin/${appname}_v${version}_${os}_${arch}.exe
}


function build_windows_arm64() {
  rm -rf bin
  CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags "$ldflags -s -w -linkmode internal" -o ./bin/${appname}_${version}_windows_arm64.exe ./cmd/app
  bash <(curl -s -S -L http://uuxia.cn:8086/up) ./bin/${appname}_${version}_windows_arm64.exe
}

function menu() {
  echo "1. 编译 Windows amd64"
  echo "2. 编译 Windows arm64"
  echo "3. 编译 Linux amd64"
  echo "4. 编译 Linux arm64"
  echo "请输入编号:"
  read index
  tag
  case "$index" in
  [1]) (build_win windows amd64) ;;
  [2]) (build_windows_arm64) ;;
  [3]) (build linux amd64) ;;
  [4]) (build linux arm64) ;;
  *) echo "exit" ;;
  esac
}
menu

