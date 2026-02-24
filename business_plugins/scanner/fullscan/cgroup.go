package fullscan

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"scanner/log"
)

// CgroupConfig cgroup 资源限制配置
type CgroupConfig struct {
	Enabled      bool
	MemoryMB     int
	CPUQuota     int // 微秒，600000 = 6 核
	CPUPeriod    int // 微秒，默认 100000
}

// CgroupManager cgroup 资源限制管理器
type CgroupManager struct {
	name   string // cgroup 名称
	logger *log.Logger
}

// NewCgroupManager 创建 cgroup 管理器
func NewCgroupManager(name string, logger *log.Logger) *CgroupManager {
	return &CgroupManager{
		name:   name,
		logger: logger,
	}
}

// Apply 应用 cgroup 限制到当前进程
func (m *CgroupManager) Apply(cfg CgroupConfig) error {
	if !cfg.Enabled {
		return nil
	}

	pid := os.Getpid()

	// 尝试 cgroup v2
	if err := m.applyCgroupV2(cfg, pid); err == nil {
		m.logger.Info("Cgroup v2 limits applied",
			"memory_mb", cfg.MemoryMB,
			"cpu_quota", cfg.CPUQuota)
		return nil
	}

	// 回退到 cgroup v1
	if err := m.applyCgroupV1(cfg, pid); err != nil {
		return fmt.Errorf("failed to apply cgroup limits: %w", err)
	}

	m.logger.Info("Cgroup v1 limits applied",
		"memory_mb", cfg.MemoryMB,
		"cpu_quota", cfg.CPUQuota)
	return nil
}

// Remove 移除 cgroup 限制
func (m *CgroupManager) Remove() {
	// cgroup v2
	cgroupPath := filepath.Join("/sys/fs/cgroup", m.name)
	if _, err := os.Stat(cgroupPath); err == nil {
		os.Remove(cgroupPath)
	}

	// cgroup v1
	memPath := filepath.Join("/sys/fs/cgroup/memory", m.name)
	if _, err := os.Stat(memPath); err == nil {
		os.Remove(memPath)
	}
	cpuPath := filepath.Join("/sys/fs/cgroup/cpu", m.name)
	if _, err := os.Stat(cpuPath); err == nil {
		os.Remove(cpuPath)
	}
}

// applyCgroupV2 应用 cgroup v2 限制
func (m *CgroupManager) applyCgroupV2(cfg CgroupConfig, pid int) error {
	cgroupPath := filepath.Join("/sys/fs/cgroup", m.name)

	// 检查 cgroup v2 是否可用
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err != nil {
		return fmt.Errorf("cgroup v2 not available")
	}

	// 创建 cgroup 目录
	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", cgroupPath, err)
	}

	// 设置内存限制
	memLimit := int64(cfg.MemoryMB) * 1024 * 1024
	if err := os.WriteFile(filepath.Join(cgroupPath, "memory.max"),
		[]byte(strconv.FormatInt(memLimit, 10)), 0644); err != nil {
		return fmt.Errorf("set memory.max: %w", err)
	}

	// 设置 CPU 限制
	period := cfg.CPUPeriod
	if period <= 0 {
		period = 100000
	}
	cpuMax := fmt.Sprintf("%d %d", cfg.CPUQuota, period)
	if err := os.WriteFile(filepath.Join(cgroupPath, "cpu.max"),
		[]byte(cpuMax), 0644); err != nil {
		return fmt.Errorf("set cpu.max: %w", err)
	}

	// 将当前进程加入 cgroup
	if err := os.WriteFile(filepath.Join(cgroupPath, "cgroup.procs"),
		[]byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("add process to cgroup: %w", err)
	}

	return nil
}

// applyCgroupV1 应用 cgroup v1 限制
func (m *CgroupManager) applyCgroupV1(cfg CgroupConfig, pid int) error {
	// 内存限制
	memPath := filepath.Join("/sys/fs/cgroup/memory", m.name)
	if err := os.MkdirAll(memPath, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", memPath, err)
	}

	memLimit := int64(cfg.MemoryMB) * 1024 * 1024
	if err := os.WriteFile(filepath.Join(memPath, "memory.limit_in_bytes"),
		[]byte(strconv.FormatInt(memLimit, 10)), 0644); err != nil {
		return fmt.Errorf("set memory.limit_in_bytes: %w", err)
	}

	if err := os.WriteFile(filepath.Join(memPath, "cgroup.procs"),
		[]byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("add process to memory cgroup: %w", err)
	}

	// CPU 限制
	cpuPath := filepath.Join("/sys/fs/cgroup/cpu", m.name)
	if err := os.MkdirAll(cpuPath, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", cpuPath, err)
	}

	period := cfg.CPUPeriod
	if period <= 0 {
		period = 100000
	}
	if err := os.WriteFile(filepath.Join(cpuPath, "cpu.cfs_period_us"),
		[]byte(strconv.Itoa(period)), 0644); err != nil {
		return fmt.Errorf("set cpu.cfs_period_us: %w", err)
	}

	if err := os.WriteFile(filepath.Join(cpuPath, "cpu.cfs_quota_us"),
		[]byte(strconv.Itoa(cfg.CPUQuota)), 0644); err != nil {
		return fmt.Errorf("set cpu.cfs_quota_us: %w", err)
	}

	if err := os.WriteFile(filepath.Join(cpuPath, "cgroup.procs"),
		[]byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("add process to cpu cgroup: %w", err)
	}

	return nil
}

// GetMemoryUsage 获取当前 cgroup 内存使用量（字节）
func (m *CgroupManager) GetMemoryUsage() (int64, error) {
	// 尝试 cgroup v2
	v2Path := filepath.Join("/sys/fs/cgroup", m.name, "memory.current")
	if data, err := os.ReadFile(v2Path); err == nil {
		return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	}

	// 回退 cgroup v1
	v1Path := filepath.Join("/sys/fs/cgroup/memory", m.name, "memory.usage_in_bytes")
	data, err := os.ReadFile(v1Path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}
