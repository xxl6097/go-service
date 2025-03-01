#!/bin/bash
module=$(grep "module" go.mod | cut -d ' ' -f 2)
#appname=$(basename $module)
appname="acsvr"
version=0.0.0
versionDir="$module/pkg"
ldflags=""
appdir="./cmd/app"

function writeVersionGoFile() {
  if [ ! -d "./pkg" ]; then
    mkdir "./pkg"
  fi
bTime=$(date +"%Y-%m-%d %H:%M:%S")
cat <<EOF > ./pkg/version.go
package pkg

import (
	"fmt"
	"strings"
)

var (
	AppName      string // 应用名称
	AppVersion   string // 应用版本
	BuildVersion string // 编译版本
	BuildTime    string // 编译时间
	GitRevision  string // Git版本
	GitBranch    string // Git分支
	GoVersion    string // Golang信息
	DisplayName  string // 服务显示名
	Description  string // 服务描述信息
)

const Build_Time = "${bTime}"

// Version 版本信息
func Version() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("App Name:\t%s\n", AppName))
	sb.WriteString(fmt.Sprintf("App Version:\t%s\n", AppVersion))
	sb.WriteString(fmt.Sprintf("Build version:\t%s\n", BuildVersion))
	sb.WriteString(fmt.Sprintf("Build time:\t%s\n", BuildTime))
	sb.WriteString(fmt.Sprintf("Git revision:\t%s\n", GitRevision))
	sb.WriteString(fmt.Sprintf("Git branch:\t%s\n", GitBranch))
	sb.WriteString(fmt.Sprintf("Golang Version: %s\n", GoVersion))
	sb.WriteString(fmt.Sprintf("DisplayName:\t%s\n", DisplayName))
	sb.WriteString(fmt.Sprintf("Description: %s\n", Description))
	fmt.Println(sb.String())
	return sb.String()
}

EOF
}

function all() {
os_archs=("darwin:amd64" "darwin:arm64" "freebsd:amd64" "linux:amd64" "linux:arm:7" "linux:arm:5" "linux:arm64" "windows:amd64" "windows:arm64" "linux:mips64" "linux:mips64le" "linux:mips:softfloat" "linux:mipsle:softfloat" "linux:riscv64" "linux:loong64" "android:arm64")
for arch in "${os_archs[@]}"; do
    IFS=":" read -r os arch extra <<< "$arch"
    #echo "OS: $os | Arch: $arch | extra: ${extra}"
    distDir=./dist/${appname}_${version}_${os}_${arch}
    flags='';
    if [ "${os}" = "linux" ] && [ "${arch}" = "arm" ] && [ "${extra}" != "" ] ; then
      if [ "${extra}" = "7" ]; then
        flags=GOARM=7;
        distDir=./dist/${appname}_${version}_${os}_${arch}_hf
      elif [ "${extra}" = "5" ]; then
        flags=GOARM=5;
        distDir=./dist/${appname}_${version}_${os}_${arch}
      fi;
    elif [ "${os}" = "linux" ] && ([ "${arch}" = "mips" ] || [ "${arch}" = "mipsle" ]) && [ "${extra}" != "" ] ; then
      flags=GOMIPS=${extra};
    fi;
    echo "build：GOOS=${os} GOARCH=${arch} ${flags} ==>${distDir}"
    env CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} ${flags} go build -trimpath -ldflags "$ldflags -s -w -linkmode internal" -o ${distDir} ${appdir}
done
bash <(curl -s -S -L http://10.6.14.26:8087/up) ./dist /soft/${appname}/${version}
}


function upgradeVersion() {
  version=$(cat version.txt)
  if [ "$version" = "" ]; then
    version="0.0.0"
  else
    v3=$(echo $version | awk -F'.' '{print($3);}')
    v2=$(echo $version | awk -F'.' '{print($2);}')
    v1=$(echo $version | awk -F'.' '{print($1);}')
    if [[ $(expr $v3 \>= 99) == 1 ]]; then
      v3=0
      if [[ $(expr $v2 \>= 99) == 1 ]]; then
        v2=0
        v1=$(expr $v1 + 1)
      else
        v2=$(expr $v2 + 1)
      fi
    else
      v3=$(expr $v3 + 1)
    fi
    version="$v1.$v2.$v3"
    echo $version > version.txt
  fi
}

function buildGoArgs() {
  APP_NAME=${appname}
  APP_VERSION=${version}
  BUILD_VERSION=$(if [ "$(git describe --tags --abbrev=0 2>/dev/null)" != "" ]; then git describe --tags --abbrev=0; else git log --pretty=format:'%h' -n 1; fi)
  BUILD_TIME=$(TZ=Asia/Shanghai date +"%Y-%m-%d %H:%M:%S")
  GIT_REVISION=$(git rev-parse --short HEAD)
  GIT_BRANCH=$(git name-rev --name-only HEAD)
  GO_VERSION=$(go version)
  ldflags="-s -w\
 -X '${versionDir}.AppName=${APP_NAME}'\
 -X '${versionDir}.AppVersion=${APP_VERSION}'\
 -X '${versionDir}.BuildVersion=${BUILD_VERSION}'\
 -X '${versionDir}.BuildTime=${BUILD_TIME}'\
 -X '${versionDir}.GitRevision=${GIT_REVISION}'\
 -X '${versionDir}.GitBranch=${GIT_BRANCH}'\
 -X '${versionDir}.GoVersion=${GO_VERSION}'"
}

function build() {
  os=$1
  arch=$2
  distDir=./dist/${appname}_${version}_${os}_${arch}
  CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -ldflags "$ldflags -s -w -linkmode internal" -o ${distDir} ${appdir}
  echo "编译完成 ${distDir}"
}

function build_linux_mips_opwnert_REDMI_AC2100() {
  distDir=./dist/${appname}_${version}_linux_mipsle
  CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -ldflags "$ldflags -s -w -linkmode internal" -o ${distDir} ${appdir}
  echo "编译完成 ${distDir}"
}

function build_windows_arm64() {
  distDir=./dist/${appname}_${version}_windows_arm64.exe
  CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags "$ldflags -s -w -linkmode internal" -o ${distDir} ${appdir}
  echo "编译完成 ${distDir}"
}

function build_win() {
  os=$1
  arch=$2
  distDir=./dist/${appname}_${version}_${os}_${arch}.exe
  #echo "编译 CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -ldflags "$ldflags -s -w -linkmode internal" -o ${distDir} ${appdir}"
  CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -ldflags "$ldflags -s -w -linkmode internal" -o ${distDir} ${appdir}
  echo "编译完成 ${distDir}"
}

function build_menu() {
  my_array=("$@")
  for index in "${my_array[@]}"; do
        case "$index" in
          [1]) (build_win windows amd64) ;;
          [2]) (build_windows_arm64) ;;
          [3]) (build linux amd64) ;;
          [4]) (build linux arm64) ;;
          [5]) (build_linux_mips_opwnert_REDMI_AC2100) ;;
          [6]) (build darwin arm64) ;;
          [7]) (build darwin amd64) ;;
          *) echo "-->exit" ;;
          esac
  done

#  bash <(curl -s -S -L http://10.6.14.26:8087/up) ./dist /soft/${appname}/${version}
  bash <(curl -s -S -L http://uuxia.cn:8087/up) ./dist /soft/${appname}/${version}
}

function buildall() {
  all
}

function m() {
  echo "1. 编译 Windows amd64"
  echo "2. 编译 Windows arm64"
  echo "3. 编译 Linux amd64"
  echo "4. 编译 Linux arm64"
  echo "5. 编译 Linux mips"
  echo "6. 编译 Darwin arm64"
  echo "7. 编译 Darwin amd64"
  echo "8. 编译全平台"
  echo "请输入编号:"
  read -r -a inputData "$@"

  if (( inputData[0] == 8 )); then
     buildall
  else
     (build_menu "${inputData[@]}")
  fi
}

function main_pre() {
  #1. 更新版本号
  upgradeVersion
  #2. 构建go语言编译信息
  buildGoArgs
  #3. 在pkg下创建version.go文件
  writeVersionGoFile
  #4. dist删除目录
  rm -rf dist
}

function main() {
  main_pre
  case $1 in
    buildall) (buildall) ;;
           *) (m)  ;;
  esac
}

main $1
