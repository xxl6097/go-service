package netutil

import (
	"fmt"
	"net"
	"time"
)

// TcpPing 检测指定IP:port的TCP连通性
// ip: 目标IP，例如 "127.0.0.1"
// port: 端口，例如 8080
// timeout: 超时时间，建议 1~3秒
// return: isOpen 是否通, err 错误信息
func TcpPing(ip string, port int, timeout time.Duration) (bool, error) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		// 连接失败：端口不通/主机不可达/超时
		return false, fmt.Errorf("connect %s failed: %w", addr, err)
	}
	defer conn.Close()
	return true, nil
}

// BatchTcpPing 批量检测多个ip端口，同步串行
func BatchTcpPing(list []struct {
	IP   string
	Port int
}, timeout time.Duration) map[string]bool {
	result := make(map[string]bool)
	for _, item := range list {
		ok, _ := TcpPing(item.IP, item.Port, timeout)
		key := fmt.Sprintf("%s:%d", item.IP, item.Port)
		result[key] = ok
	}
	return result
}
