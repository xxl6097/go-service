# 升级后服务无法自动重启 — 修复说明

## 背景

`gfs` 服务通过 `POST /fs/upgrade` 触发自升级。升级流程末尾会「拉起一个独立的 restart 子进程去执行 `systemctl restart`」，旧进程本应随后优雅退出。

但线上日志显示：旧进程成功停止（`gfs.service: Succeeded`），新进程却**从未被拉起**（无 `Starting/Started gfs.service`），服务停在停止态。

### 根因

`internal/helper.go` 的 `update()` 在拉起 restart 子进程后，**紧接着同步执行了 `systemctl stop` 停掉自己**。而 kardianos 生成的 systemd unit 未显式设置 `KillMode`，默认是 `KillMode=control-group`——`systemctl stop gfs.service` 会向本服务 cgroup 内**所有**进程发 SIGTERM，其中包括：

1. 刚启动 1ms 的 `systemd-run` 辅助进程（还没来得及把 restart 子进程放进独立 scope），被一起杀掉 → restart 子进程根本没被拉起；
2. `systemctl stop` 客户端进程自身 → 所以日志里出现 `stop失败 … signal: terminated`。

结果就是「旧服务停了、`systemctl restart` 从未被发出」。

---

## 改动一：`internal/helper.go` — `update()`

去掉拉起重启子进程后那步自杀式 `stopService()`。restart 子进程在独立 scope 里执行 `systemctl restart`（stop+start），当前进程由那次 SIGTERM 触发优雅退出。

### 原来

```go
	err = this.RunCMD("restart")
	if !utils.IsWindows() {
		_ = this.stopService()
	}
	return err
```

### 现在

```go
	// 只拉起重启子进程，不要在这里 stopService()：当前服务自行 systemctl stop
	// （KillMode=control-group）会把同一 cgroup 里的 systemd-run 辅助进程、以及
	// systemctl 客户端自身一并 SIGTERM，restart 子进程还没被拉起服务就停在停止态了。
	// restart 子进程在独立 scope 里执行 systemctl restart（stop+start），
	// 当前进程会由那次 SIGTERM 触发优雅退出。
	err = this.RunCMD("restart")
	return err
```

---

## 改动二：`pkg/utils/sysutil.go` — `RunChildProcess()`

加固回退路径：`systemd-run` 不可用或启动失败时，不再静默回退到普通 fork（普通 fork 逃不出 service cgroup，restart 触发 stop 时同样会连坐），而是直接报错，让调用方知道重启没能被安全拉起。

同时把 `useSystemdRun()` 拆成 `isSystemd()`（是否 systemd 环境）+ `useSystemdRun()`（是否有可用的 `systemd-run`），用于区分「systemd 但缺 systemd-run」和「非 systemd」。

### 原来

```go
	if useSystemdRun() {
		runArgs := []string{
			"--scope",
			"--collect", // 运行结束后自动清理该 transient unit
			fmt.Sprintf("--unit=%s-restart-%d", filepath.Base(executable), time.Now().UnixNano()),
			executable,
		}
		runArgs = append(runArgs, args...)
		cmd := exec.Command("systemd-run", runArgs...)
		util.SetPlatformSpecificAttrs(cmd)
		fmt.Printf("运行子进程(systemd-run) %s %v\n", executable, args)
		if err := cmd.Start(); err == nil {
			return nil
		} else {
			// systemd-run 启动失败则回退到普通方式
			z.Errorf("systemd-run 启动失败，回退普通子进程: %v", err)
		}
	}
	cmd := exec.Command(executable, args...)
	util.SetPlatformSpecificAttrs(cmd)
	fmt.Printf("运行子进程 %s %v\n", executable, args)
	return cmd.Start()
}

// useSystemdRun 判断当前是否为可用 systemd-run 的 Linux systemd 环境。
func useSystemdRun() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	// systemd 运行时目录存在即视为 systemd 系统
	if _, err := os.Stat("/run/systemd/system"); err != nil {
		return false
	}
	if _, err := exec.LookPath("systemd-run"); err != nil {
		return false
	}
	return true
}
```

### 现在

```go
	if useSystemdRun() {
		runArgs := []string{
			"--scope",
			"--collect", // 运行结束后自动清理该 transient unit
			fmt.Sprintf("--unit=%s-restart-%d", filepath.Base(executable), time.Now().UnixNano()),
			executable,
		}
		runArgs = append(runArgs, args...)
		cmd := exec.Command("systemd-run", runArgs...)
		util.SetPlatformSpecificAttrs(cmd)
		fmt.Printf("运行子进程(systemd-run) %s %v\n", executable, args)
		if err := cmd.Start(); err == nil {
			return nil
		} else {
			// systemd-run 启动失败不能回退到普通 fork：普通 fork 的子进程仍在这个
			// service cgroup 里，restart 触发 systemctl stop 时会被连带 SIGTERM，
			// restart 照样不生效。直接报错，让调用方知道重启没能被安全拉起。
			return fmt.Errorf("systemd-run 启动失败: %w", err)
		}
	}
	// systemd 环境但缺少 systemd-run：普通 fork 逃不出 service cgroup，同上会连坐。
	if isSystemd() {
		return errors.New("systemd 环境缺少 systemd-run，无法安全拉起重启子进程")
	}
	cmd := exec.Command(executable, args...)
	util.SetPlatformSpecificAttrs(cmd)
	fmt.Printf("运行子进程 %s %v\n", executable, args)
	return cmd.Start()
}

// isSystemd 判断是否运行在 systemd 管理的 Linux 系统上。
func isSystemd() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	// systemd 运行时目录存在即视为 systemd 系统
	if _, err := os.Stat("/run/systemd/system"); err != nil {
		return false
	}
	return true
}

// useSystemdRun 判断当前是否为可用 systemd-run 的 Linux systemd 环境。
func useSystemdRun() bool {
	if !isSystemd() {
		return false
	}
	if _, err := exec.LookPath("systemd-run"); err != nil {
		return false
	}
	return true
}
```

---

## 影响范围

| 平台 | 行为 |
| --- | --- |
| Linux + systemd（含 systemd-run） | 修复：restart 子进程正常在独立 scope 里执行 `systemctl restart`，旧进程由 SIGTERM 优雅退出 |
| Linux + systemd（缺 systemd-run） | 由「静默回退→重启不生效」改为「明确报错」，避免停在停止态 |
| Linux（OpenWRT / SysV 等非 systemd） | 不变：普通 fork + `Setsid`，无 cgroup 连坐问题 |
| macOS（launchd） | 不变：普通 fork + `Setsid` + `launchctl restart` |
| Windows | 不变：走 `RestartWindowsApplication`，原 `stopService()` 本就不执行 |

## 验证

- `go build ./internal/... ./pkg/utils/` 通过
- `go vet ./internal/ ./pkg/utils/` 无新增问题（`helper.go:98`、`netutil` 为历史遗留，与本次改动无关）
