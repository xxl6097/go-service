#!/usr/bin/env bash
# go-proxy-doctor.sh — 诊断并修复 Go 拉取 GitHub 依赖慢/失败
#
# 用法:
#   ./go-proxy-doctor.sh           # 只探测 + 打印建议，不改任何配置
#   ./go-proxy-doctor.sh --apply   # 备份现有 go env 后应用推荐配置

set -uo pipefail

APPLY=0
[[ "${1:-}" == "--apply" ]] && APPLY=1

green()  { printf "\033[32m%s\033[0m\n" "$1"; }
red()    { printf "\033[31m%s\033[0m\n" "$1"; }
yellow() { printf "\033[33m%s\033[0m\n" "$1"; }
blue()   { printf "\033[34m%s\033[0m\n" "$1"; }
bold()   { printf "\033[1m%s\033[0m\n" "$1"; }

MIRRORS=(
  "https://goproxy.cn"
  "https://mirrors.tencent.com/go"
  "https://mirrors.aliyun.com/goproxy"
  "https://proxy.golang.com.cn"
  "https://goproxy.io"
)
PROBE="/github.com/gin-gonic/gin/@v/list"

bold "=============================================="
bold " 1. 当前 Go 模块配置"
bold "=============================================="
command -v go >/dev/null 2>&1 || { red "未找到 go 命令，请先安装 Go"; exit 1; }
go version
echo
go env GOPROXY GOSUMDB GOPRIVATE GONOPROXY GONOSUMDB GOFLAGS GOTOOLCHAIN

CUR_PROXY="$(go env GOPROXY)"
if [[ "$CUR_PROXY" == *","* && "$CUR_PROXY" != *"|"* ]]; then
  echo
  yellow "⚠  GOPROXY 使用逗号分隔：只有 404/410 才回退，"
  yellow "   遇到超时或 5xx 会直接终止整个下载 —— 这是最常见的失败原因。"
fi
if ! git config --global --get-regexp 'url\.|http\.' >/dev/null 2>&1; then
  echo
  yellow "⚠  未配置任何 git 代理 / URL 重写，direct 兜底等于裸连 github.com。"
fi

echo
bold "=============================================="
bold " 2. 镜像可达性探测"
bold "=============================================="
OK_MIRRORS=()
for m in "${MIRRORS[@]}"; do
  out=$(curl -sS -o /dev/null -w "%{http_code} %{time_total}" --max-time 8 "${m}${PROBE}" 2>/dev/null)
  if [[ -z "$out" ]]; then out="000 0.000"; fi
  code=${out%% *}
  t=${out##* }
  if [[ "$code" == "200" ]]; then
    green "$(printf '%-36s %s %6ss' "$m" "$code" "$t")"
    OK_MIRRORS+=("$m")
  else
    red "$(printf '%-36s %s %6ss' "$m" "$code" "$t")"
  fi
done
[[ ${#OK_MIRRORS[@]} -eq 0 ]] && OK_MIRRORS=("${MIRRORS[@]}")

echo
bold "=============================================="
bold " 3. GitHub 直连探测（direct 兜底路径）"
bold "=============================================="
http_out=$(curl -sS -o /dev/null -w "%{http_code} %{time_total}" --max-time 10 https://github.com/ 2>/dev/null)
if [[ -z "$http_out" ]]; then
  red "HTTPS  github.com            不可达"
else
  printf "HTTPS  %-22s %s %ss\n" "github.com" "${http_out%% *}" "${http_out##* }"
fi

if command -v nc >/dev/null 2>&1; then
  if nc -z -G 5 ssh.github.com 443 >/dev/null 2>&1; then
    green "SSH    ssh.github.com:443   可达"
  else
    red   "SSH    ssh.github.com:443   不可达"
  fi
fi

echo
bold "=============================================="
bold " 4. 推荐配置"
bold "=============================================="
PROXY_LIST=""
for m in "${OK_MIRRORS[@]}"; do PROXY_LIST+="${m}|"; done
PROXY_LIST+="direct"

echo "go env -w GOPROXY=\"${PROXY_LIST}\""
echo "go env -w GOSUMDB=off"
echo "# 若上面第 1 节列出的 shell 脚本里有 export GOPROXY，记得同步改掉那一行"
echo
echo "# git 兜底（二选一）："
echo "# A. 有本地 HTTP 代理时，只对 github 生效，不影响 cnb.cool"
echo "git config --global http.https://github.com/.proxy http://127.0.0.1:7890"
echo "# B. 走 SSH 443，穿透性最好"
echo "git config --global url.\"ssh://git@ssh.github.com:443/\".insteadOf \"https://github.com/\""

echo
if [[ $APPLY -eq 1 ]]; then
  bold "=============================================="
  bold " 5. 应用配置"
  bold "=============================================="
  BACKUP="$HOME/.goenv.backup.$(date +%Y%m%d%H%M%S)"
  go env > "$BACKUP"
  blue ">> 已备份当前 go env 到 $BACKUP"
  go env -w GOPROXY="$PROXY_LIST"
  go env -w GOSUMDB=off
  go env -w GOTOOLCHAIN=local
  green ">> 已应用"
  echo
  echo "GOPROXY = $(go env GOPROXY)"
  echo "GOSUMDB = $(go env GOSUMDB)"
  echo "GOTOOLCHAIN = $(go env GOTOOLCHAIN)"
  echo
  yellow "如需回滚：go env -w GOPROXY=\"\$(grep '^GOPROXY=' $BACKUP | cut -d= -f2-)\""
else
  yellow "（未做任何修改。确认无误后加 --apply 执行）"
fi
