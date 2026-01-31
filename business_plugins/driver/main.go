// SPDX-License-Identifier: GPL-2.0
package main

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	businessplugins "business_plugins/lib"
	"driver/ebpf"
	"driver/events"
	"driver/log"
)

func main() {
	// 1. 初始化客户端（FD 3/4通信）
	client := businessplugins.New()
	defer client.Close()

	// 2. 初始化日志组件
	logger := log.New()
	logger.Info("Starting eBPF driver plugin...")

	// 3. 加载eBPF程序
	loader, err := ebpf.NewLoader()
	if err != nil {
		logger.Fatal("Failed to load eBPF program", "error", err)
		os.Exit(1)
	}
	defer loader.Close()

	logger.Info("eBPF program loaded successfully")

	// 4. 启动事件读取循环
	go func() {
		for {
			// 4.1 从perf buffer读取事件（阻塞）
			rec, err := loader.Read()
			if err != nil {
				if errors.Is(err, syscall.EINTR) {
					// 被信号中断，继续
					continue
				}
				logger.Error("Failed to read from perf buffer", "error", err)
				continue
			}

			// 4.2 检查丢失事件
			if rec.LostSamples > 0 {
				logger.Warn("Lost samples", "count", rec.LostSamples, "cpu", rec.CPU)
			}

			// 4.3 反序列化事件
			var evt events.ExecveEvent
			if err := evt.UnmarshalBinary(rec.RawSample); err != nil {
				logger.Error("Failed to unmarshal event", "error", err)
				continue
			}

			// 4.4 转换为protobuf格式并发送到Agent
			record := evt.ToRecord()
			if err := client.SendRecord(record); err != nil {
				logger.Error("Failed to send record to agent", "error", err)
			}
		}
	}()

	// 5. 等待退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("Received termination signal, shutting down...")
}
