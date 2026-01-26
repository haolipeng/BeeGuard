// Package main 演示 Linux fanotify 文件监控功能
// fanotify 是 Linux 内核提供的文件系统事件通知 API
// 需要 root 权限或 CAP_SYS_ADMIN 能力才能运行
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"unsafe"
)

// fanotify 相关常量
const (
	// fanotify_init flags
	FAN_CLOEXEC         = 0x00000001
	FAN_NONBLOCK        = 0x00000002
	FAN_CLASS_NOTIF     = 0x00000000 // 通知类（默认）
	FAN_CLASS_CONTENT   = 0x00000004 // 内容类（可访问文件内容）
	FAN_CLASS_PRE_CONTENT = 0x00000008 // 预内容类

	FAN_UNLIMITED_QUEUE = 0x00000010
	FAN_UNLIMITED_MARKS = 0x00000020
	FAN_REPORT_TID      = 0x00000100 // 报告线程ID
	FAN_REPORT_FID      = 0x00000200 // 报告文件ID
	FAN_REPORT_DIR_FID  = 0x00000400 // 报告目录ID
	FAN_REPORT_NAME     = 0x00000800 // 报告文件名

	// fanotify_mark flags
	FAN_MARK_ADD          = 0x00000001
	FAN_MARK_REMOVE       = 0x00000002
	FAN_MARK_DONT_FOLLOW  = 0x00000004
	FAN_MARK_ONLYDIR      = 0x00000008
	FAN_MARK_IGNORED_MASK = 0x00000020
	FAN_MARK_MOUNT        = 0x00000010 // 监控整个挂载点
	FAN_MARK_FILESYSTEM   = 0x00000100 // 监控整个文件系统

	// 事件类型
	FAN_ACCESS        = 0x00000001 // 文件被访问
	FAN_MODIFY        = 0x00000002 // 文件被修改
	FAN_ATTRIB        = 0x00000004 // 元数据变更
	FAN_CLOSE_WRITE   = 0x00000008 // 可写文件关闭
	FAN_CLOSE_NOWRITE = 0x00000010 // 只读文件关闭
	FAN_OPEN          = 0x00000020 // 文件被打开
	FAN_MOVED_FROM    = 0x00000040 // 文件被移出
	FAN_MOVED_TO      = 0x00000080 // 文件被移入
	FAN_CREATE        = 0x00000100 // 文件被创建
	FAN_DELETE        = 0x00000200 // 文件被删除
	FAN_DELETE_SELF   = 0x00000400 // 被监控对象删除
	FAN_MOVE_SELF     = 0x00000800 // 被监控对象移动
	FAN_OPEN_EXEC     = 0x00001000 // 文件被执行打开

	FAN_OPEN_PERM     = 0x00010000 // 打开权限事件
	FAN_ACCESS_PERM   = 0x00020000 // 访问权限事件
	FAN_OPEN_EXEC_PERM = 0x00040000 // 执行权限事件

	FAN_EVENT_ON_CHILD = 0x08000000 // 监控子目录事件
	FAN_ONDIR          = 0x40000000 // 目录事件

	// 组合事件
	FAN_CLOSE = FAN_CLOSE_WRITE | FAN_CLOSE_NOWRITE
	FAN_MOVE  = FAN_MOVED_FROM | FAN_MOVED_TO

	// 权限响应
	FAN_ALLOW = 0x01
	FAN_DENY  = 0x02

	// 事件信息类型
	FAN_EVENT_INFO_TYPE_FID      = 1
	FAN_EVENT_INFO_TYPE_DFID     = 2
	FAN_EVENT_INFO_TYPE_DFID_NAME = 3
)

// FanotifyEventMetadata fanotify 事件元数据结构
type FanotifyEventMetadata struct {
	EventLen    uint32 // 事件长度
	Version     uint8  // 版本号
	Reserved    uint8
	MetadataLen uint16 // 元数据长度
	Mask        uint64 // 事件掩码
	Fd          int32  // 文件描述符
	Pid         int32  // 进程ID
}

// FanotifyResponse 权限事件响应结构
type FanotifyResponse struct {
	Fd       int32
	Response uint32
}

// FanotifyMonitor fanotify 监控器
type FanotifyMonitor struct {
	fd          int
	watchPaths  []string
	stopChan    chan struct{}
	permissions bool // 是否处理权限事件
}

// NewFanotifyMonitor 创建新的 fanotify 监控器
func NewFanotifyMonitor(permissions bool) (*FanotifyMonitor, error) {
	var flags uint
	if permissions {
		// 权限类需要 FAN_CLASS_CONTENT 或 FAN_CLASS_PRE_CONTENT
		flags = FAN_CLOEXEC | FAN_CLASS_CONTENT | FAN_UNLIMITED_QUEUE | FAN_UNLIMITED_MARKS
	} else {
		flags = FAN_CLOEXEC | FAN_CLASS_NOTIF | FAN_UNLIMITED_QUEUE | FAN_UNLIMITED_MARKS
	}

	fd, err := fanotifyInit(flags, syscall.O_RDONLY|syscall.O_LARGEFILE)
	if err != nil {
		return nil, fmt.Errorf("fanotify_init 失败: %v (需要 root 权限)", err)
	}

	return &FanotifyMonitor{
		fd:          fd,
		watchPaths:  make([]string, 0),
		stopChan:    make(chan struct{}),
		permissions: permissions,
	}, nil
}

// AddWatch 添加监控路径
func (m *FanotifyMonitor) AddWatch(path string, mask uint64) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// 检查路径是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("路径不存在: %s", absPath)
	}

	flags := FAN_MARK_ADD

	err = fanotifyMark(m.fd, uint(flags), mask, -1, absPath)
	if err != nil {
		return fmt.Errorf("fanotify_mark 失败: %v", err)
	}

	m.watchPaths = append(m.watchPaths, absPath)
	fmt.Printf("已添加监控: %s\n", absPath)
	return nil
}

// AddMountWatch 添加挂载点监控
func (m *FanotifyMonitor) AddMountWatch(path string, mask uint64) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	flags := FAN_MARK_ADD | FAN_MARK_MOUNT

	err = fanotifyMark(m.fd, uint(flags), mask, -1, absPath)
	if err != nil {
		return fmt.Errorf("fanotify_mark 挂载点失败: %v", err)
	}

	m.watchPaths = append(m.watchPaths, absPath)
	fmt.Printf("已添加挂载点监控: %s\n", absPath)
	return nil
}

// Start 开始监控
func (m *FanotifyMonitor) Start() error {
	buf := make([]byte, 4096)

	fmt.Println("\n开始监控文件事件 (按 Ctrl+C 退出)...\n")

	for {
		select {
		case <-m.stopChan:
			return nil
		default:
		}

		n, err := syscall.Read(m.fd, buf)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return fmt.Errorf("读取事件失败: %v", err)
		}

		if n > 0 {
			m.processEvents(buf[:n])
		}
	}
}

// processEvents 处理事件
func (m *FanotifyMonitor) processEvents(buf []byte) {
	reader := bytes.NewReader(buf)

	for reader.Len() >= int(unsafe.Sizeof(FanotifyEventMetadata{})) {
		var event FanotifyEventMetadata
		if err := binary.Read(reader, binary.LittleEndian, &event); err != nil {
			break
		}

		// 跳过额外的事件数据
		extraLen := int(event.EventLen) - int(unsafe.Sizeof(FanotifyEventMetadata{}))
		if extraLen > 0 {
			reader.Seek(int64(extraLen), 1)
		}

		// 处理事件
		m.handleEvent(&event)

		// 关闭文件描述符（非常重要！）
		if event.Fd >= 0 {
			syscall.Close(int(event.Fd))
		}
	}
}

// handleEvent 处理单个事件
func (m *FanotifyMonitor) handleEvent(event *FanotifyEventMetadata) {
	// 获取文件路径
	filePath := m.getFilePath(event.Fd)

	// 获取进程信息
	procName := m.getProcessName(event.Pid)

	// 构建事件描述
	eventDesc := m.getEventDescription(event.Mask)

	// 处理权限事件
	if m.permissions && (event.Mask&(FAN_OPEN_PERM|FAN_ACCESS_PERM|FAN_OPEN_EXEC_PERM) != 0) {
		// 默认允许访问，实际应用中可以添加策略判断
		allowed := m.checkPermission(filePath, event.Pid)
		m.respondPermission(event.Fd, allowed)
		if allowed {
			eventDesc += " [已允许]"
		} else {
			eventDesc += " [已拒绝]"
		}
	}

	fmt.Printf("[事件] %s\n", eventDesc)
	fmt.Printf("  文件: %s\n", filePath)
	fmt.Printf("  进程: %s (PID: %d)\n", procName, event.Pid)
	fmt.Println()
}

// getFilePath 通过 fd 获取文件路径
func (m *FanotifyMonitor) getFilePath(fd int32) string {
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
func (m *FanotifyMonitor) getProcessName(pid int32) string {
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

	if mask&FAN_ACCESS != 0 {
		events = append(events, "访问")
	}
	if mask&FAN_MODIFY != 0 {
		events = append(events, "修改")
	}
	if mask&FAN_ATTRIB != 0 {
		events = append(events, "属性变更")
	}
	if mask&FAN_CLOSE_WRITE != 0 {
		events = append(events, "关闭(写)")
	}
	if mask&FAN_CLOSE_NOWRITE != 0 {
		events = append(events, "关闭(读)")
	}
	if mask&FAN_OPEN != 0 {
		events = append(events, "打开")
	}
	if mask&FAN_OPEN_EXEC != 0 {
		events = append(events, "执行")
	}
	if mask&FAN_OPEN_PERM != 0 {
		events = append(events, "打开权限请求")
	}
	if mask&FAN_ACCESS_PERM != 0 {
		events = append(events, "访问权限请求")
	}
	if mask&FAN_OPEN_EXEC_PERM != 0 {
		events = append(events, "执行权限请求")
	}
	if mask&FAN_MOVED_FROM != 0 {
		events = append(events, "移出")
	}
	if mask&FAN_MOVED_TO != 0 {
		events = append(events, "移入")
	}
	if mask&FAN_CREATE != 0 {
		events = append(events, "创建")
	}
	if mask&FAN_DELETE != 0 {
		events = append(events, "删除")
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

// checkPermission 检查权限（示例：简单的白名单机制）
func (m *FanotifyMonitor) checkPermission(filePath string, pid int32) bool {
	// 示例：拒绝对特定敏感文件的访问
	sensitiveFiles := []string{
		"/etc/shadow",
		"/etc/gshadow",
	}

	for _, sf := range sensitiveFiles {
		if filePath == sf {
			procName := m.getProcessName(pid)
			// 只允许特定进程访问
			if procName != "passwd" && procName != "sudo" && procName != "su" {
				fmt.Printf("  ⚠️  拒绝进程 %s 访问敏感文件: %s\n", procName, filePath)
				return false
			}
		}
	}
	return true
}

// respondPermission 响应权限请求
func (m *FanotifyMonitor) respondPermission(fd int32, allow bool) {
	response := FanotifyResponse{
		Fd: fd,
	}
	if allow {
		response.Response = FAN_ALLOW
	} else {
		response.Response = FAN_DENY
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, response)
	syscall.Write(m.fd, buf.Bytes())
}

// Stop 停止监控
func (m *FanotifyMonitor) Stop() {
	close(m.stopChan)
}

// Close 关闭监控器
func (m *FanotifyMonitor) Close() error {
	return syscall.Close(m.fd)
}

// fanotifyInit 系统调用封装
func fanotifyInit(flags uint, eventFlags uint) (int, error) {
	fd, _, errno := syscall.Syscall(
		syscall.SYS_FANOTIFY_INIT,
		uintptr(flags),
		uintptr(eventFlags),
		0,
	)
	if errno != 0 {
		return -1, errno
	}
	return int(fd), nil
}

// fanotifyMark 系统调用封装
func fanotifyMark(fd int, flags uint, mask uint64, dirFd int, path string) error {
	pathBytes, err := syscall.BytePtrFromString(path)
	if err != nil {
		return err
	}

	_, _, errno := syscall.Syscall6(
		syscall.SYS_FANOTIFY_MARK,
		uintptr(fd),
		uintptr(flags),
		uintptr(mask),
		uintptr(dirFd),
		uintptr(unsafe.Pointer(pathBytes)),
		0,
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func main() {
	// 检查是否 root 权限
	if os.Geteuid() != 0 {
		fmt.Println("错误: 此程序需要 root 权限运行")
		fmt.Println("请使用: sudo go run main.go")
		os.Exit(1)
	}

	fmt.Println("=== Linux Fanotify 文件监控示例 ===")
	fmt.Println()

	// 创建监控器（设置 true 启用权限事件处理）
	monitor, err := NewFanotifyMonitor(false)
	if err != nil {
		fmt.Printf("创建监控器失败: %v\n", err)
		os.Exit(1)
	}
	defer monitor.Close()

	// 定义要监控的核心系统文件和目录
	watchTargets := []struct {
		path string
		mask uint64
	}{
		// 核心配置文件
		{"/etc/passwd", FAN_OPEN | FAN_MODIFY | FAN_CLOSE_WRITE},
		{"/etc/shadow", FAN_OPEN | FAN_MODIFY | FAN_CLOSE_WRITE},
		{"/etc/sudoers", FAN_OPEN | FAN_MODIFY | FAN_CLOSE_WRITE},
		{"/etc/ssh", FAN_OPEN | FAN_MODIFY | FAN_CLOSE_WRITE},

		// 系统启动脚本
		{"/etc/crontab", FAN_OPEN | FAN_MODIFY | FAN_CLOSE_WRITE},
		{"/etc/cron.d", FAN_OPEN | FAN_MODIFY | FAN_CLOSE_WRITE},

		// 系统二进制目录（监控执行）
		{"/usr/bin", FAN_OPEN_EXEC | FAN_CLOSE},
		{"/usr/sbin", FAN_OPEN_EXEC | FAN_CLOSE},
	}

	// 添加监控
	for _, target := range watchTargets {
		if _, err := os.Stat(target.path); err == nil {
			if err := monitor.AddWatch(target.path, target.mask); err != nil {
				fmt.Printf("警告: 添加监控失败 %s: %v\n", target.path, err)
			}
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
		fmt.Printf("监控错误: %v\n", err)
	}

	fmt.Println("监控已停止")
}
