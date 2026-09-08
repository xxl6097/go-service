#!/bin/bash
# 使用 garble 符号混淆交叉编译两个程序（macOS / Windows / Linux，共 6 个平台）
#   考勤统计 cmd/attendance  →  dist/workhours*
#   用户书签 cmd/userbook    →  dist/userbook-*
# 依赖：garble（go install mvdan.cc/garble@latest）

set -e
cd "$(dirname "$0")"

# garble 必须在 PATH 中
if ! command -v garble >/dev/null 2>&1; then
  echo "未找到 garble，正在安装..."
  go install mvdan.cc/garble@latest
  export PATH="$PATH:$(go env GOPATH)/bin"
fi

mkdir -p dist-garble

# garble 交叉编译：需要把 -ldflags 放在 build 之后
# garble 参数：-literals 混淆字符串字面量，-seed=random 每次随机种子
# 首次编译会混淆 Go 标准库与依赖，耗时较长（后续有缓存会快很多）。
# 用法：build <GOOS> <GOARCH> <包路径> <输出文件>
build() {
  local goos=$1 goarch=$2 pkg=$3 out=$4
  echo "=== ${pkg} ${goos}/${goarch} → ${out} ==="
  CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch \
    garble -literals -tiny -seed=random \
      build -trimpath -ldflags "-s -w -buildid=" -buildvcs=false \
      -o "$out" "$pkg"
}

# 平台列表：3 系统 × 2 架构
platforms=(
  "darwin  arm64"
#  "darwin  amd64"
#  "windows amd64"
#  "linux   amd64"
#  "linux   arm64"
)

for plat in "${platforms[@]}"; do
  read -r goos goarch <<< "$plat"
  ext=""
  [ "$goos" = "windows" ] && ext=".exe"
  build "$goos" "$goarch" ./cmd/app/app "dist-garble/app-${goos}-${goarch}${ext}"
  #build "$goos" "$goarch" ./cmd/userbook  "dist-garble/userbook-${goos}-${goarch}${ext}"
done

echo ""
echo "=== 产物 ==="
ls -lh dist-garble/
