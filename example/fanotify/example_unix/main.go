// Package main 演示使用 golang.org/x/sys/unix 包进行 fanotify 文件监控
// 安装: go get golang.org/x/sys/unix
// 运行: sudo go run main.go
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// 核心文件监控配置
type WatchConfig struct {
	Path      string
	Mask      uint64
	MountWide bool // 是否监控整个挂载点
}

// FanotifyMonitor 基于 unix 包的 fanotify 监控器
type FanotifyMonitor struct {
	fd       int
	stopChan chan struct{}
}

// NewFanotifyMonitor 创建监控器
func NewFanotifyMonitor() (*FanotifyMonitor, error) {
	// 初始化 fanotify
	// FAN_CLASS_NOTIF: 通知类
	// FAN_CLOEXEC: exec 时自动关闭
	// FAN_NONBLOCK: 非阻塞模式
	fd, err := unix.FanotifyInit(
		unix.FAN_CLASS_NOTIF|unix.FAN_CLOEXEC,
		unix.O_RDONLY,
	)
	if err != nil {
		return nil, fmt.Errorf("fanotify_init 失败: %v (需要 root 权限)", err)
	}

	return &FanotifyMonitor{
		fd:       fd,
		stopChan: make(chan struct{}),
	}, nil
}

// AddWatch 添加文件/目录监控
func (m *FanotifyMonitor) AddWatch(config WatchConfig) error {
	// 检查路径是否存在
	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		return fmt.Errorf("路径不存在: %s", config.Path)
	}

	var flags uint
	if config.MountWide {
		flags = unix.FAN_MARK_ADD | unix.FAN_MARK_MOUNT
	} else {
		flags = unix.FAN_MARK_ADD
	}

	err := unix.FanotifyMark(
		m.fd,
		flags,
		config.Mask,
		unix.AT_FDCWD,
		config.Path,
	)
	if err != nil {
		return fmt.Errorf("fanotify_mark 失败: %v", err)
	}

	mountStr := ""
	if config.MountWide {
		mountStr = " (挂载点)"
	}
	fmt.Printf("已添加监控: %s%s\n", config.Path, mountStr)
	return nil
}

// Start 开始监控事件
func (m *FanotifyMonitor) Start() error {
	buf := make([]byte, 4096)

	fmt.Println("\n开始监控文件事件 (按 Ctrl+C 退出)...\n")

	for {
		select {
		case <-m.stopChan:
			return nil
		default:
		}

		// 读取事件
		n, err := unix.Read(m.fd, buf)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			if err == syscall.EAGAIN {
				continue
			}
			return fmt.Errorf("读取事件失败: %v", err)
		}

		if n > 0 {
			m.processEvents(buf[:n])
		}
	}
}

// processEvents 处理事件数据
func (m *FanotifyMonitor) processEvents(data []byte) {
	// FanotifyEventMetadata 结构体大小
	metadataSize := int(unsafe.Sizeof(unix.FanotifyEventMetadata{}))

	offset := 0
	for offset+metadataSize <= len(data) {
		// 解析事件元数据
		event := (*unix.FanotifyEventMetadata)(unsafe.Pointer(&data[offset]))

		// 验证事件长度
		if event.Event_len < uint32(metadataSize) {
			break
		}

		// 处理事件
		m.handleEvent(event)

		// 关闭文件描述符（非常重要）
		if event.Fd >= 0 {
			unix.Close(int(event.Fd))
		}

		offset += int(event.Event_len)
	}
}

// handleEvent 处理单个事件
func (m *FanotifyMonitor) handleEvent(event *unix.FanotifyEventMetadata) {
	// 获取文件路径
	filePath := m.getFilePath(int(event.Fd))

	// 获取进程信息
	procName := m.getProcessName(int(event.Pid))

	// 解析事件类型
	eventDesc := m.getEventDescription(event.Mask)

	fmt.Printf("[事件] %s\n", eventDesc)
	fmt.Printf("  文件: %s\n", filePath)
	fmt.Printf("  进程: %s (PID: %d)\n", procName, event.Pid)
	fmt.Println()
}

// getFilePath 通过 fd 获取文件路径
func (m *FanotifyMonitor) getFilePath(fd int) string {
	if fd < 0 {
		return "<无文件描述符>"
	}

	linkPath := fmt.Sprintf("/proc/self/fd/%d", fd)
	path, err := os.Readlink(linkPath)
	if err != nil {
		return "<无法获取路径>"
	}
	return path
}

// getProcessName 获取进程名称
func (m *FanotifyMonitor) getProcessName(pid int) string {
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	data, err := os.ReadFile(commPath)
	if err != nil {
		return "<未知进程>"
	}
	return string(bytes.TrimSpace(data))
}

// getEventDescription 获取事件描述
func (m *FanotifyMonitor) getEventDescription(mask uint64) string {
	var events []string

	if mask&unix.FAN_ACCESS != 0 {
		events = append(events, "访问")
	}
	if mask&unix.FAN_MODIFY != 0 {
		events = append(events, "修改")
	}
	if mask&unix.FAN_CLOSE_WRITE != 0 {
		events = append(events, "关闭(写)")
	}
	if mask&unix.FAN_CLOSE_NOWRITE != 0 {
		events = append(events, "关闭(读)")
	}
	if mask&unix.FAN_OPEN != 0 {
		events = append(events, "打开")
	}
	if mask&unix.FAN_OPEN_EXEC != 0 {
		events = append(events, "执行")
	}
	if mask&unix.FAN_ATTRIB != 0 {
		events = append(events, "属性变更")
	}
	if mask&unix.FAN_CREATE != 0 {
		events = append(events, "创建")
	}
	if mask&unix.FAN_DELETE != 0 {
		events = append(events, "删除")
	}
	if mask&unix.FAN_MOVED_FROM != 0 {
		events = append(events, "移出")
	}
	if mask&unix.FAN_MOVED_TO != 0 {
		events = append(events, "移入")
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

// Stop 停止监控
func (m *FanotifyMonitor) Stop() {
	close(m.stopChan)
}

// Close 关闭监控器
func (m *FanotifyMonitor) Close() error {
	return unix.Close(m.fd)
}

func main() {
	// 检查 root 权限
	if os.Geteuid() != 0 {
		fmt.Println("错误: 此程序需要 root 权限运行")
		fmt.Println("请使用: sudo go run main.go")
		os.Exit(1)
	}

	fmt.Println("=== 使用 golang.org/x/sys/unix 的 Fanotify 监控示例 ===")
	fmt.Println()

	// 创建监控器
	monitor, err := NewFanotifyMonitor()
	if err != nil {
		log.Fatalf("创建监控器失败: %v", err)
	}
	defer monitor.Close()

	// 定义要监控的核心系统文件
	watchConfigs := []WatchConfig{
		// 核心配置文件
		{
			Path:      "/etc/passwd",
			Mask:      unix.FAN_OPEN | unix.FAN_MODIFY | unix.FAN_CLOSE_WRITE,
			MountWide: false,
		},
		{
			Path:      "/etc/shadow",
			Mask:      unix.FAN_OPEN | unix.FAN_MODIFY | unix.FAN_CLOSE_WRITE,
			MountWide: false,
		},
		{
			Path:      "/etc/sudoers",
			Mask:      unix.FAN_OPEN | unix.FAN_MODIFY | unix.FAN_CLOSE_WRITE,
			MountWide: false,
		},
		// SSH 配置目录
		{
			Path:      "/etc/ssh",
			Mask:      unix.FAN_OPEN | unix.FAN_MODIFY | unix.FAN_CLOSE_WRITE,
			MountWide: false,
		},
		// 定时任务
		{
			Path:      "/etc/crontab",
			Mask:      unix.FAN_OPEN | unix.FAN_MODIFY | unix.FAN_CLOSE_WRITE,
			MountWide: false,
		},
		// /tmp 目录（监控整个挂载点）
		{
			Path:      "/tmp",
			Mask:      unix.FAN_OPEN | unix.FAN_MODIFY | unix.FAN_CREATE | unix.FAN_DELETE,
			MountWide: true,
		},
	}

	// 添加监控
	for _, config := range watchConfigs {
		if err := monitor.AddWatch(config); err != nil {
			log.Printf("警告: %v", err)
		}
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n收到退出信号，正在停止监控...")
		monitor.Stop()
	}()

	// 开始监控
	if err := monitor.Start(); err != nil {
		log.Printf("监控错误: %v", err)
	}

	fmt.Println("监控已停止")
}
