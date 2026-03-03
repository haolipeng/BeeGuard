package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// resolveExePath 补全可执行文件的完整路径
// eBPF 在 kprobe 上下文中 dentry 遍历可能失败，仅返回文件名
// 通过 /proc/<pid>/exe readlink 获取完整路径
func resolveExePath(tgid uint32, ebpfPath string) string {
	if len(ebpfPath) > 0 && ebpfPath[0] == '/' {
		return ebpfPath
	}
	link, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", tgid))
	if err == nil {
		return link
	}
	return ebpfPath
}

// resolveParentComm 读取父进程名称
func resolveParentComm(ppid uint32) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", ppid))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// resolveParentUID 读取父进程的 UID（通过 /proc/<ppid>/status 中的 Uid 行）
func resolveParentUID(ppid uint32) string {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", ppid))
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
			break
		}
	}
	return ""
}

// resolveUsername 将 UID 解析为用户名（通过 /etc/passwd）
func resolveUsername(uid uint32) string {
	uidStr := fmt.Sprintf("%d", uid)
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return uidStr
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) >= 3 && parts[2] == uidStr {
			return parts[0]
		}
	}
	return uidStr
}

// buildPidTree 在用户态构建进程链字符串，格式: "PID<comm<PID<comm<..."，最多 8 层
func buildPidTree(tgid uint32, comm string) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("%d<%s", tgid, comm))
	pid := tgid
	for i := 0; i < 7; i++ {
		ppid := readPPid(pid)
		if ppid == 0 || ppid == pid {
			break
		}
		parentComm := resolveParentComm(ppid)
		if parentComm == "" {
			break
		}
		buf.WriteString(fmt.Sprintf("<%d<%s", ppid, parentComm))
		pid = ppid
	}
	return buf.String()
}

// readProcCmdline 从 /proc/<tgid>/cmdline 读取干净的命令行
// /proc/pid/cmdline 中各参数以 NULL 分隔，这里替换为空格
// 进程可能已退出，此时返回空字符串
func readProcCmdline(tgid uint32) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", tgid))
	if err != nil || len(data) == 0 {
		return ""
	}
	// 去掉尾部的 NULL 字节
	for len(data) > 0 && data[len(data)-1] == 0 {
		data = data[:len(data)-1]
	}
	// 将 NULL 分隔符替换为空格
	for i := range data {
		if data[i] == 0 {
			data[i] = ' '
		}
	}
	return string(data)
}

// readPPid 从 /proc/<pid>/status 读取父进程 PID
func readPPid(pid uint32) uint32 {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PPid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var ppid uint32
				if n, _ := fmt.Sscanf(fields[1], "%d", &ppid); n == 1 {
					return ppid
				}
			}
			break
		}
	}
	return 0
}
