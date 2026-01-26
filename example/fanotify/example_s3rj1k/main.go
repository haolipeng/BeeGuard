// Package main 演示使用 s3rj1k/go-fanotify 库进行文件监控
// 安装: go get github.com/s3rj1k/go-fanotify/fanotify
// 运行: sudo go run main.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/s3rj1k/go-fanotify/fanotify"
)

func main() {
	// 检查 root 权限
	if os.Geteuid() != 0 {
		log.Fatal("此程序需要 root 权限运行，请使用 sudo")
	}

	fmt.Println("=== 使用 s3rj1k/go-fanotify 库的文件监控示例 ===")
	fmt.Println()

	// 初始化 fanotify
	// FAN_CLASS_NOTIF: 通知模式
	// FAN_CLOEXEC: exec 时自动关闭
	notify, err := fanotify.Initialize(
		fanotify.FAN_CLASS_NOTIF|fanotify.FAN_CLOEXEC,
		os.O_RDONLY,
	)
	if err != nil {
		log.Fatalf("初始化 fanotify 失败: %v", err)
	}
	defer notify.File.Close()

	// 要监控的核心文件和目录
	watchPaths := []struct {
		path  string
		mount bool // 是否监控整个挂载点
	}{
		{"/etc/passwd", false},
		{"/etc/shadow", false},
		{"/etc/sudoers", false},
		{"/etc/ssh", false},
		{"/tmp", true}, // 监控 /tmp 挂载点
	}

	// 定义监控的事件类型
	eventMask := uint64(
		fanotify.FAN_OPEN |
			fanotify.FAN_ACCESS |
			fanotify.FAN_MODIFY |
			fanotify.FAN_CLOSE_WRITE |
			fanotify.FAN_CLOSE_NOWRITE,
	)

	// 添加监控标记
	for _, wp := range watchPaths {
		// 检查路径是否存在
		if _, err := os.Stat(wp.path); os.IsNotExist(err) {
			fmt.Printf("跳过不存在的路径: %s\n", wp.path)
			continue
		}

		var markFlags uint
		if wp.mount {
			markFlags = fanotify.FAN_MARK_ADD | fanotify.FAN_MARK_MOUNT
		} else {
			markFlags = fanotify.FAN_MARK_ADD
		}

		err = notify.Mark(markFlags, eventMask, -1, wp.path)
		if err != nil {
			log.Printf("添加监控失败 %s: %v", wp.path, err)
			continue
		}
		fmt.Printf("已添加监控: %s (挂载点: %v)\n", wp.path, wp.mount)
	}

	fmt.Println()
	fmt.Println("开始监控文件事件 (按 Ctrl+C 退出)...")
	fmt.Println()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 创建停止通道
	stopChan := make(chan struct{})

	go func() {
		<-sigChan
		fmt.Println("\n收到退出信号，正在停止监控...")
		close(stopChan)
	}()

	// 事件处理循环
	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
			}

			// 获取事件
			event, err := notify.GetEvent()
			if err != nil {
				log.Printf("获取事件失败: %v", err)
				continue
			}

			if event == nil {
				continue
			}

			// 获取文件路径
			path, err := event.GetPath()
			if err != nil {
				path = "<无法获取路径>"
			}

			// 获取进程信息
			procName := getProcessName(event.Pid)

			// 解析事件类型
			eventDesc := parseEventMask(event.Mask)

			fmt.Printf("[事件] %s\n", eventDesc)
			fmt.Printf("  文件: %s\n", path)
			fmt.Printf("  进程: %s (PID: %d)\n", procName, event.Pid)
			fmt.Println()

			// 关闭事件（释放文件描述符）
			event.Close()
		}
	}()

	<-stopChan
	fmt.Println("监控已停止")
}

// getProcessName 获取进程名称
func getProcessName(pid int32) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return "<未知进程>"
	}
	// 移除换行符
	name := string(data)
	if len(name) > 0 && name[len(name)-1] == '\n' {
		name = name[:len(name)-1]
	}
	return name
}

// parseEventMask 解析事件掩码
func parseEventMask(mask uint64) string {
	var events []string

	if mask&fanotify.FAN_ACCESS != 0 {
		events = append(events, "访问")
	}
	if mask&fanotify.FAN_MODIFY != 0 {
		events = append(events, "修改")
	}
	if mask&fanotify.FAN_OPEN != 0 {
		events = append(events, "打开")
	}
	if mask&fanotify.FAN_CLOSE_WRITE != 0 {
		events = append(events, "关闭(写)")
	}
	if mask&fanotify.FAN_CLOSE_NOWRITE != 0 {
		events = append(events, "关闭(读)")
	}

	if len(events) == 0 {
		return fmt.Sprintf("未知事件(0x%x)", mask)
	}

	result := ""
	for i, e := range events {
		if i > 0 {
			result += ", "
		}
		result += e
	}
	return result
}
