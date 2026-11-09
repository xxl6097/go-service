package main

import (
	"bytes"
	"os"
)

func main() {
	// 占位区长度：必须 >= rawKey(39) + 4 + 签名数据长度，且留足余量。
	// 历史值 1024 太小，配置稍大就会被"静默撑大"，破坏可执行文件结构。
	cfgBuffer := bytes.Repeat([]byte{byte(0x2B)}, 8192)
	_ = os.WriteFile("./assets/buffer/buffer", cfgBuffer, 0644)
	//fmt.Println("Buffer is ", string(buffer.BufferData))
}
