package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ScannerConfig 主配置结构
type ScannerConfig struct {
	Scanner ScannerSection `yaml:"scanner"`
}

// ScannerSection scanner 配置段
type ScannerSection struct {
	Engine   EngineConfig   `yaml:"engine"`
	Cronjob  CronjobConfig  `yaml:"cronjob"`
	ScanDirs []ScanDir      `yaml:"scan_dirs"`
	Filter   FilterConfig   `yaml:"filter"`
	Fullscan FullscanConfig `yaml:"fullscan"`
	Cgroup   CgroupConfig   `yaml:"cgroup"`
}

// EngineConfig 扫描引擎配置
type EngineConfig struct {
	DBPath      string `yaml:"db_path"`
	MaxFileSize int64  `yaml:"max_file_size"`
	MaxScanTime int    `yaml:"max_scan_time"`
}

// CronjobConfig 定时扫描配置
type CronjobConfig struct {
	DirScanInterval  string `yaml:"dir_scan_interval"`
	ProcScanInterval string `yaml:"proc_scan_interval"`
	Throttle         string `yaml:"throttle"`
}

// ParseDirScanInterval 解析目录扫描间隔
func (c *CronjobConfig) ParseDirScanInterval() time.Duration {
	d, err := time.ParseDuration(c.DirScanInterval)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

// ParseProcScanInterval 解析进程扫描间隔
func (c *CronjobConfig) ParseProcScanInterval() time.Duration {
	d, err := time.ParseDuration(c.ProcScanInterval)
	if err != nil {
		return 1 * time.Hour
	}
	return d
}

// ParseThrottle 解析节流间隔
func (c *CronjobConfig) ParseThrottle() time.Duration {
	d, err := time.ParseDuration(c.Throttle)
	if err != nil {
		return 1 * time.Second
	}
	return d
}

// ScanDir 扫描目录配置
type ScanDir struct {
	Path     string `yaml:"path"`
	MaxDepth int    `yaml:"max_depth"`
}

// FilterConfig 文件过滤配置
type FilterConfig struct {
	PathWhitelist []string `yaml:"path_whitelist"`
	SkipFileTypes []string `yaml:"skip_file_types"`
	MinFileSize   int64    `yaml:"min_file_size"`
	MaxFileSize   int64    `yaml:"max_file_size"`
}

// FullscanConfig 全盘扫描配置
type FullscanConfig struct {
	MaxWorkers    int    `yaml:"max_workers"`
	MaxMemoryMB   int    `yaml:"max_memory_mb"`
	MaxCPUPercent int    `yaml:"max_cpu_percent"`
	QuickTimeout  string `yaml:"quick_timeout"`
	FullTimeout   string `yaml:"full_timeout"`
}

// ParseQuickTimeout 解析快速扫描超时
func (c *FullscanConfig) ParseQuickTimeout() time.Duration {
	d, err := time.ParseDuration(c.QuickTimeout)
	if err != nil {
		return 1 * time.Hour
	}
	return d
}

// ParseFullTimeout 解析全盘扫描超时
func (c *FullscanConfig) ParseFullTimeout() time.Duration {
	d, err := time.ParseDuration(c.FullTimeout)
	if err != nil {
		return 48 * time.Hour
	}
	return d
}

// CgroupConfig Cgroup 资源限制配置
type CgroupConfig struct {
	Enabled           bool `yaml:"enabled"`
	NormalMemoryMB    int  `yaml:"normal_memory_mb"`
	NormalCPUQuota    int  `yaml:"normal_cpu_quota"`
	FullscanMemoryMB  int  `yaml:"fullscan_memory_mb"`
	FullscanCPUQuota  int  `yaml:"fullscan_cpu_quota"`
}

// LoadConfig 加载 YAML 配置文件
func LoadConfig(path string) (*ScannerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	cfg := defaultConfig()

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return cfg, nil
}

// defaultConfig 返回默认配置
func defaultConfig() *ScannerConfig {
	return &ScannerConfig{
		Scanner: ScannerSection{
			Engine: EngineConfig{
				DBPath:      "/var/lib/clamav",
				MaxFileSize: 18874368,
				MaxScanTime: 5,
			},
			Cronjob: CronjobConfig{
				DirScanInterval:  "24h",
				ProcScanInterval: "1h",
				Throttle:         "1s",
			},
			ScanDirs: []ScanDir{
				{Path: "/root", MaxDepth: 3},
				{Path: "/bin", MaxDepth: 2},
				{Path: "/sbin", MaxDepth: 2},
				{Path: "/usr/bin", MaxDepth: 2},
				{Path: "/usr/sbin", MaxDepth: 2},
				{Path: "/usr/local", MaxDepth: 3},
				{Path: "/etc", MaxDepth: 2},
				{Path: "/var/www", MaxDepth: 20},
			},
			Filter: FilterConfig{
				PathWhitelist: []string{"/dev", "/proc", "/sys", "/boot", "/opt/cloudsec/agent"},
				SkipFileTypes: []string{"video", "audio", "image"},
				MinFileSize:   4,
				MaxFileSize:   18874368,
			},
			Fullscan: FullscanConfig{
				MaxWorkers:    6,
				MaxMemoryMB:   512,
				MaxCPUPercent: 600,
				QuickTimeout:  "1h",
				FullTimeout:   "48h",
			},
			Cgroup: CgroupConfig{
				Enabled:          true,
				NormalMemoryMB:   180,
				NormalCPUQuota:   10000,
				FullscanMemoryMB: 512,
				FullscanCPUQuota: 600000,
			},
		},
	}
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	if path := os.Getenv("SCANNER_CONFIG_PATH"); path != "" {
		return path
	}

	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, "config/scanner.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return "config/scanner.yaml"
}

// resolveDBPath 解析数据库路径（相对路径基于可执行文件目录）
func resolveDBPath(dbPath string) string {
	if filepath.IsAbs(dbPath) {
		return dbPath
	}

	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		return filepath.Join(dir, dbPath)
	}

	return dbPath
}
