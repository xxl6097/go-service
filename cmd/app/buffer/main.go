package main

import (
	"bytes"
	"os"
)

func main() {
	cfgBuffer := bytes.Repeat([]byte{byte(0x2B)}, 1024)
	_ = os.WriteFile("./assets/buffer/buffer", cfgBuffer, 0644)
	//fmt.Println("Buffer is ", string(buffer.BufferData))
}
