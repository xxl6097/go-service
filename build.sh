#!/bin/bash
module=$(grep "module" go.mod | cut -d ' ' -f 2)
options=("windows:amd64" "windows:arm64" "linux:amd64" "linux:arm64" "linux:arm:7" "linux:arm:5" "linux:mips64" "linux:mips64le" "linux:mips:softfloat" "linux:mipsle:softfloat" "linux:riscv64" "linux:loong64" "darwin:amd64" "darwin:arm64" "freebsd:amd64" "android:arm64")
#options=("linux:amd64" "windows:amd64")
version=$(git tag -l "v[0-99]*.[0-99]*.[0-99]*" --sort=-creatordate | head -n 1)
versionDir="$module/pkg"

function writeVersionGoFile() {
  if [ ! -d "./pkg" ]; then
    mkdir "./pkg"
  fi
cat <<EOF > ./pkg/version.go
package pkg
import (
	"fmt"
	"strings"
	"runtime"
)
func init() {
	OsType = runtime.GOOS
	Arch = runtime.GOARCH
}
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
	OsType       string // 操作系统
	Arch         string // cpu类型
	BinName      string // 运行文件名称，包含平台架构
)
// Version 版本信息
func Version() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%-15s: %-5s\n", "App Name", AppName))
	sb.WriteString(fmt.Sprintf("%-15s: %-5s\n", "App Version", AppVersion))
	sb.WriteString(fmt.Sprintf("%-15s: %-5s\n", "Build version", BuildVersion))
	sb.WriteString(fmt.Sprintf("%-15s: %-5s\n", "Build time", BuildTime))
	sb.WriteString(fmt.Sprintf("%-15s: %-5s\n", "Git revision", GitRevision))
	sb.WriteString(fmt.Sprintf("%-15s: %-5s\n", "Git branch", GitBranch))
	sb.WriteString(fmt.Sprintf("%-15s: %-5s\n", "Golang Version", GoVersion))
	sb.WriteString(fmt.Sprintf("%-15s: %-5s\n", "DisplayName", DisplayName))
	sb.WriteString(fmt.Sprintf("%-15s: %-5s\n", "Description", Description))
	sb.WriteString(fmt.Sprintf("%-15s: %-5s\n", "OsType", OsType))
	sb.WriteString(fmt.Sprintf("%-15s: %-5s\n", "Arch", Arch))
	sb.WriteString(fmt.Sprintf("%-15s: %-5s\n", "BinName", BinName))
	fmt.Println(sb.String())
	return sb.String()
}
EOF
}

# shellcheck disable=SC2120
function buildgo() {
  builddir=$1
  appname=$2
  version=$3
  appdir=$4
  os=$5
  arch=$6
  extra=$7
  dstFilePath=${builddir}/${appname}_${version}_${os}_${arch}
  flags='';
  if [ "${os}" = "linux" ] && [ "${arch}" = "arm" ] && [ "${extra}" != "" ] ; then
    if [ "${extra}" = "7" ]; then
      flags=GOARM=7;
      dstFilePath=${builddir}/${appname}_${version}_${os}_${arch}hf
    elif [ "${extra}" = "5" ]; then
      flags=GOARM=5;
      dstFilePath=${builddir}/${appname}_${version}_${os}_${arch}
    fi;
  elif [ "${os}" = "windows" ] ; then
    dstFilePath=${builddir}/${appname}_${version}_${os}_${arch}.exe
    if [ "${arch}" = "amd64" ]; then
        go generate ${appdir}
    fi
  elif [ "${os}" = "linux" ] && ([ "${arch}" = "mips" ] || [ "${arch}" = "mipsle" ]) && [ "${extra}" != "" ] ; then
    flags=GOMIPS=${extra};
  fi;
  #echo "build：GOOS=${os} GOARCH=${arch} ${flags} ==> ${dstFilePath}"
  printf "build：GOOS=%-7s GOARCH=%-8s ==> %s\n" ${os} ${arch} ${dstFilePath}

  filename=$(basename "$dstFilePath")
  binName="-X '${versionDir}.BinName=${filename}'"
  #echo "--->env CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} ${flags} go build -trimpath -ldflags "$ldflags $binName -linkmode internal" -o ${dstFilePath} ${appdir}"
  env CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} ${flags} go build -trimpath -ldflags "$ldflags $binName -linkmode internal" -o ${dstFilePath} ${appdir}
  if [ "${os}" = "windows" ] ; then
    if [ "${arch}" = "amd64" ]; then
        rm -rf ${appdir}/resource.syso
    fi
  fi;
}

# builddir：输出目录
# appname：应用名称
# version：应用版本
# appdir：main.go目录
# disname：显示名
# describe：描述
function buildMenu() {
  builddir=$1
  appname=$2
  version=$3
  appdir=$4
  disname=$5
  describe=$6
  ldflags=$(buildLdflags $appname $disname $describe)
  PS3="请选择需要编译的平台："
  select arch in "${options[@]}"; do
      if [[ -n "$arch" ]]; then
        IFS=":" read -r os arch extra <<< "$arch"
        buildgo $builddir $appname $version $appdir $os $arch $extra
        return $?
      else
        echo "输入无效，请重新选择。"
      fi
  done
}

# builddir：输出目录
# appname：应用名称
# version：应用版本
# appdir：main.go目录
# disname：显示名
# describe：描述
function buildAll() {
  builddir=$1
  appname=$2
  version=$3
  appdir=$4
  disname=$5
  describe=$6
  ldflags=$(buildLdflags $appname $disname $describe)
  for arch in "${options[@]}"; do
      IFS=":" read -r os arch extra <<< "$arch"
      buildgo $builddir $appname $version $appdir $os $arch $extra
  done
  #wait
}

function build() {
  #echo "---->$1 $2 $3 $4 $5 $6 $7"
  if [ $7 -eq 1 ]; then
    buildMenu $1 $2 $3 $4 $5 $6
  else
    buildAll $1 $2 $3 $4 $5 $6
  fi
}

function buildLdflags() {
  #os_name=$(uname -s)
  #echo "os type $os_name"
  appname=$1
  DisplayName=$2
  Description=$3
  APP_NAME=${appname}
  #BUILD_VERSION=$(if [ "$(git describe --tags --abbrev=0 2>/dev/null)" != "" ]; then git describe --tags --abbrev=0; else git log --pretty=format:'%h' -n 1; fi)
  BUILD_TIME=$(TZ=Asia/Shanghai date "+%Y-%m-%d %H:%M:%S")
  GIT_REVISION=$(git rev-parse --short HEAD)
  #GIT_BRANCH=$(git name-rev --name-only HEAD)
  #GIT_BRANCH=$(git tag -l "v[0-99]*.[0-99]*.[0-99]*" --sort=-creatordate | head -n 1)
  GO_VERSION=$(go version)
  # shellcheck disable=SC2089
  local ldflags="-s -w\
 -X '${versionDir}.DisplayName=${DisplayName}_${version}'\
 -X '${versionDir}.Description=${Description}'\
 -X '${versionDir}.AppName=${APP_NAME}'\
 -X '${versionDir}.AppVersion=${version}'\
 -X '${versionDir}.BuildVersion=${version}'\
 -X '${versionDir}.BuildTime=${BUILD_TIME}'\
 -X '${versionDir}.GitRevision=${GIT_REVISION}'\
 -X '${versionDir}.GitBranch=${version}'\
 -X '${versionDir}.GoVersion=${GO_VERSION}'"
  echo "$ldflags"
}

function buildFrps() {
    appname="acfrps"
    appdir="./cmd/frps"
    DisplayName="AcFrps网络代理程序"
    Description="一款基于GO语言的网络代理服务程序"
    builddir="./release/frps"
    rm -rf ${builddir}
    build $builddir $appname "$version" $appdir $DisplayName $Description "$1"
}

function showBuildDir() {
  # 检查是否输入路径参数
  if [ -z "$1" ]; then
      echo "用法: $0 <路径>"
      exit 1
  fi

  # 验证路径是否存在且为目录
  if [ ! -d "$1" ]; then
      echo "错误: 路径 '$1' 不存在或不是目录！"
      exit 1
  fi

  # 获取指定路径下的所有直接子目录（非递归）
  dirs=()
  while IFS= read -r dir; do
      dirs+=("$dir")
  done < <(find "$1" -maxdepth 1 -type d ! -path "$1" | sort)

  # 检查是否有子目录
  if [ ${#dirs[@]} -eq 0 ]; then
      echo "路径 '$1' 下没有子目录！"
      exit 0
  fi

  # 生成交互式菜单
  echo "请选择要操作的目录："
  PS3="输入序号 (1-${#dirs[@]}): "
  select dir in "${dirs[@]}"; do
      if [[ -n "$dir" ]] && [[ $REPLY -ge 1 && $REPLY -le ${#dirs[@]} ]]; then
          echo "您选择的目录是: $dir"
          break
#          return $dir
      else
          echo "无效输入！请输入有效序号。"
      fi
  done
}
# shellcheck disable=SC2120
function buildDir() {
  showBuildDir ./cmd/app
  builddir="./release/${dir}"
  appname=$(basename "$dir")
  appdir=${dir}
  disname="${dir}应用程序"
  describe="一款基于GO语言的${dir}程序"
  rm -rf ${builddir}
  buildMenu $builddir $appname "$version" $appdir $disname $describe
}

function main() {
  buildDir
}

main