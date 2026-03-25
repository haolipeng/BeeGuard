// SPDX-License-Identifier: GPL-2.0
package ebpf

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -Wall -Werror -D__TARGET_ARCH_x86" -target amd64 -type execve_event -type commit_creds_event -type connect_event -type dns_event -type stdio_path_buf -type file_event -type mount_event bpf ./bpf/hids.bpf.c -- -I./bpf

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/btf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"
)

// Logger 日志接口，供 BTF 探测使用
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
}

// Loader eBPF程序加载器
type Loader struct {
	objs       *bpfObjects  // eBPF程序和Maps句柄
	links      []link.Link  // Hook点链接（用于detach）
	perfReader *perf.Reader // Perf buffer读取器
}

// getKernelRelease 获取当前内核版本（uname -r）
func getKernelRelease() (string, error) {
	var utsname syscall.Utsname
	if err := syscall.Uname(&utsname); err != nil {
		return "", fmt.Errorf("uname failed: %w", err)
	}
	// Convert [65]int8 to string, trim null bytes
	release := make([]byte, 0, len(utsname.Release))
	for _, b := range utsname.Release {
		if b == 0 {
			break
		}
		release = append(release, byte(b))
	}
	return string(bytes.TrimSpace(release)), nil
}

// probeBTF 探测 BTF 可用性：先尝试原生内核 BTF，再回退到打包的 BTF 文件
// 返回 nil 表示使用原生 BTF（让 cilium/ebpf 自动处理）；
// 返回非 nil *btf.Spec 表示使用打包的外部 BTF。
func probeBTF(btfDir string, logger Logger) (*btf.Spec, error) {
	// 尝试加载原生 BTF
	_, err := btf.LoadKernelSpec()
	if err == nil {
		logger.Info("Using native kernel BTF from /sys/kernel/btf/vmlinux")
		return nil, nil
	}

	logger.Warn("Native kernel BTF not available, trying bundled BTF...", "error", err)

	// 获取内核版本
	release, err := getKernelRelease()
	if err != nil {
		return nil, fmt.Errorf("failed to get kernel release: %w", err)
	}

	// 在 btfDir 中查找匹配的 BTF 文件
	btfPath := filepath.Join(btfDir, release+".btf")
	if _, err := os.Stat(btfPath); err != nil {
		return nil, fmt.Errorf("no BTF available for kernel %s (looked in %s)", release, btfDir)
	}

	spec, err := btf.LoadSpec(btfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load BTF from %s: %w", btfPath, err)
	}

	logger.Info("Using bundled BTF", "path", btfPath, "kernel", release)
	return spec, nil
}

// NewLoader 创建并加载eBPF程序
func NewLoader(btfDir string, logger Logger) (*Loader, error) {
	// 1. 权限检查：eBPF需要root权限
	if os.Geteuid() != 0 {
		return nil, errors.New("eBPF requires root privileges")
	}

	// 2. 移除memlock限制（防止ENOMEM错误）
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("failed to remove memlock: %w", err)
	}

	// 3. 探测 BTF 可用性
	btfSpec, err := probeBTF(btfDir, logger)
	if err != nil {
		return nil, fmt.Errorf("BTF probe failed: %w", err)
	}

	// 4. 加载eBPF对象（自动CO-RE重定位）
	objs := &bpfObjects{}
	var opts *ebpf.CollectionOptions
	if btfSpec != nil {
		opts = &ebpf.CollectionOptions{
			Programs: ebpf.ProgramOptions{
				KernelTypes: btfSpec,
			},
		}
	}
	if err := loadBpfObjects(objs, opts); err != nil {
		return nil, fmt.Errorf("failed to load eBPF objects: %w", err)
	}

	l := &Loader{objs: objs}

	// 5. 附加raw_tracepoint到sched_process_exec
	lnk, err := link.AttachRawTracepoint(link.RawTracepointOptions{
		Name:    "sched_process_exec",
		Program: objs.TpProcExec,
	})
	if err != nil {
		objs.Close()
		return nil, fmt.Errorf("failed to attach raw_tracepoint: %w", err)
	}
	l.links = append(l.links, lnk)

	// 6. 附加kprobe到commit_creds（提权检测）
	kpLink, err := link.Kprobe("commit_creds", objs.KpCommitCreds, nil)
	if err != nil {
		l.Close()
		return nil, fmt.Errorf("failed to attach kprobe to commit_creds: %w", err)
	}
	l.links = append(l.links, kpLink)

	// 7. 附加raw_tracepoint到sys_exit（网络系统调用返回处理）
	sysExitLink, err := link.AttachRawTracepoint(link.RawTracepointOptions{
		Name:    "sys_exit",
		Program: objs.TpSysExit,
	})
	if err != nil {
		l.Close()
		return nil, fmt.Errorf("failed to attach raw_tracepoint/sys_exit: %w", err)
	}
	l.links = append(l.links, sysExitLink)

	// 9. 附加kprobe到security_inode_create（文件创建监控）
	fileCreateLink, err := link.Kprobe("security_inode_create", objs.KpInodeCreate, nil)
	if err != nil {
		l.Close()
		return nil, fmt.Errorf("failed to attach kprobe to security_inode_create: %w", err)
	}
	l.links = append(l.links, fileCreateLink)

	// 10. 附加kprobe到security_inode_rename（文件重命名监控）
	fileRenameLink, err := link.Kprobe("security_inode_rename", objs.KpInodeRename, nil)
	if err != nil {
		l.Close()
		return nil, fmt.Errorf("failed to attach kprobe to security_inode_rename: %w", err)
	}
	l.links = append(l.links, fileRenameLink)

	// 11. 附加kprobe到security_inode_unlink（文件删除监控）
	fileUnlinkLink, err := link.Kprobe("security_inode_unlink", objs.KpInodeUnlink, nil)
	if err != nil {
		l.Close()
		return nil, fmt.Errorf("failed to attach kprobe to security_inode_unlink: %w", err)
	}
	l.links = append(l.links, fileUnlinkLink)

	// 12. 创建perf reader（32页/CPU = 128KB，扩展后的 execve_event ~1.4KB）
	l.perfReader, err = perf.NewReader(objs.Events, 32*4096)
	if err != nil {
		l.Close()
		return nil, fmt.Errorf("failed to create perf reader: %w", err)
	}

	return l, nil
}

// Read 从perf buffer读取一个事件（阻塞）
func (l *Loader) Read() (perf.Record, error) {
	return l.perfReader.Read()
}

// Close 清理资源：detach hook点、关闭perf reader、卸载eBPF程序
func (l *Loader) Close() error {
	var errs []error

	// 1. Detach所有hook点
	for _, lnk := range l.links {
		if err := lnk.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close link: %w", err))
		}
	}

	// 2. 关闭perf reader
	if l.perfReader != nil {
		if err := l.perfReader.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close perf reader: %w", err))
		}
	}

	// 3. 卸载eBPF程序和Maps
	if l.objs != nil {
		if err := l.objs.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close eBPF objects: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// GetTrustedExesMap 返回 trusted_exes BPF map 句柄
// 供用户态程序填充可信任可执行文件列表
func (l *Loader) GetTrustedExesMap() *ebpf.Map {
	return l.objs.TrustedExes
}

// GetFileTrustedExesMap 返回 file_trusted_exes BPF map 句柄
// 供用户态程序填充文件监控白名单
func (l *Loader) GetFileTrustedExesMap() *ebpf.Map {
	return l.objs.FileTrustedExes
}

// GetRootMntnsMap 返回 root_mntns BPF map 句柄
// 供用户态程序写入宿主机的 mount 命名空间 ID
func (l *Loader) GetRootMntnsMap() *ebpf.Map {
	return l.objs.RootMntns
}
