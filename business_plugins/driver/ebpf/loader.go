// SPDX-License-Identifier: GPL-2.0
package ebpf

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -Wall -Werror" -target bpfel -type execve_event bpf ./bpf/hids.bpf.c -- -I./bpf

import (
	"errors"
	"fmt"
	"os"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"
)

// Loader eBPF程序加载器
type Loader struct {
	objs       *bpfObjects  // eBPF程序和Maps句柄
	links      []link.Link  // Hook点链接（用于detach）
	perfReader *perf.Reader // Perf buffer读取器
}

// NewLoader 创建并加载eBPF程序
func NewLoader() (*Loader, error) {
	// 1. 权限检查：eBPF需要root权限
	if os.Geteuid() != 0 {
		return nil, errors.New("eBPF requires root privileges")
	}

	// 2. 移除memlock限制（防止ENOMEM错误）
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("failed to remove memlock: %w", err)
	}

	// 3. 加载eBPF对象（自动CO-RE重定位）
	objs := &bpfObjects{}
	if err := loadBpfObjects(objs, nil); err != nil {
		return nil, fmt.Errorf("failed to load eBPF objects: %w", err)
	}

	l := &Loader{objs: objs}

	// 4. 附加raw_tracepoint到sched_process_exec
	lnk, err := link.AttachRawTracepoint(link.RawTracepointOptions{
		Name:    "sched_process_exec",
		Program: objs.TpProcExec,
	})
	if err != nil {
		objs.Close()
		return nil, fmt.Errorf("failed to attach raw_tracepoint: %w", err)
	}
	l.links = append(l.links, lnk)

	// 5. 创建perf reader（8页/CPU = 32KB）
	l.perfReader, err = perf.NewReader(objs.Events, 8*4096)
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
