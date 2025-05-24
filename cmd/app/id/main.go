package main

import (
	"fmt"
	"sync"
	"time"
)

// IDGenerator 结构体用于生成唯一ID
type IDGenerator struct {
	mu        sync.Mutex
	lastTime  int64
	sequence  int32
	machineID int32
}

// NewIDGenerator 创建一个新的ID生成器实例
func NewIDGenerator(machineID int32) *IDGenerator {
	return &IDGenerator{
		lastTime:  time.Now().UnixNano() / 1e6, // 转换为毫秒
		sequence:  0,
		machineID: machineID,
	}
}

// NextID 生成下一个唯一ID
// ID结构: 41位时间戳 + 10位机器ID + 12位序列号
func (g *IDGenerator) NextID() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixNano() / 1e6 // 毫秒时间戳

	if now == g.lastTime {
		g.sequence = (g.sequence + 1) & 0xFFF // 序列号范围0-4095
		if g.sequence == 0 {
			// 序列号溢出，等待下一毫秒
			for now <= g.lastTime {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTime = now

	// 组装ID: 时间戳部分 | 机器ID部分 | 序列号部分
	id := ((now - 1609459200000) << 22) | // 减去一个固定时间戳，扩展可用年限
		(int64(g.machineID) << 12) |
		int64(g.sequence)

	return id
}

func main() {
	// 初始化ID生成器，传入机器ID(0-1023)
	generator := NewIDGenerator(1000)

	go func() {

		// 生成并打印10个唯一ID
		for i := 0; i < 5; i++ {
			id := generator.NextID()
			fmt.Printf("2生成的唯一ID: %d\n", id)
		}
	}()

	go func() {
		// 生成并打印10个唯一ID
		for i := 0; i < 5; i++ {
			id := generator.NextID()
			fmt.Printf("3生成的唯一ID: %d\n", id)
		}
	}()

	go func() {
		// 生成并打印10个唯一ID
		for i := 0; i < 5; i++ {
			id := generator.NextID()
			fmt.Printf("3生成的唯一ID: %d\n", id)
		}
	}()

	go func() {
		// 生成并打印10个唯一ID
		for i := 0; i < 5; i++ {
			id := generator.NextID()
			fmt.Printf("3生成的唯一ID: %d\n", id)
		}
	}()

	for {

	}
}
