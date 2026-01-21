package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// Server 服务器地址（格式：host:port）
	Server string `yaml:"server"`

	// ConnectTimeout 连接超时时间（秒）
	ConnectTimeout int `yaml:"connect_timeout"`

	// WorkingDirectory Agent 工作目录
	WorkingDirectory string `yaml:"working_directory"`

	// RetryMaxCount 最大重试次数
	RetryMaxCount int `yaml:"retry_max_count"`

	// RetryInterval 重试间隔（秒）
	RetryInterval int `yaml:"retry_interval"`
}

var (
	// globalConfig 全局配置实例
	globalConfig *Config
	// initOnce 确保配置只初始化一次
	initOnce sync.Once
	// initErr 初始化错误
	initErr error
)

const (
	// DefaultConfigFile 默认配置文件名称
	DefaultConfigFile = "config.yaml"
)

// GetConfigPath 获取配置文件路径
// 优先级：默认路径 > 当前目录
func GetConfigPath() string {
	// 1. 尝试默认路径（/etc/cloudsec-agent/config.yaml）
	defaultPath := filepath.Join("/etc", "cloudsec-agent", DefaultConfigFile)
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}

	// 2. 回退到当前目录
	return DefaultConfigFile
}

// LoadFromFile 从文件加载配置
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return &cfg, nil
}

// Load 加载配置（自动查找配置文件路径）
func Load() (*Config, error) {
	path := GetConfigPath()
	return LoadFromFile(path)
}

// ValidateAndSetDefaults 验证配置并设置默认值
func ValidateAndSetDefaults(cfg *Config) error {
	// 1. 验证必填配置项
	if cfg.Server == "" {
		return fmt.Errorf("server address is required")
	}

	// 2. 设置默认值
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 30 // 默认 30 秒
	}

	if cfg.WorkingDirectory == "" {
		cfg.WorkingDirectory = "/var/run/cloudsec-agent"
	}

	if cfg.RetryMaxCount <= 0 {
		cfg.RetryMaxCount = 10 // 默认最大重试 10 次
	}

	if cfg.RetryInterval <= 0 {
		cfg.RetryInterval = 5 // 默认重试间隔 5 秒
	}

	return nil
}

// LoadAndValidate 加载并验证配置（包含默认值设置）
func LoadAndValidate() (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	if err := ValidateAndSetDefaults(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Init 初始化全局配置（在程序启动时调用）
func Init() error {
	initOnce.Do(func() {
		cfg, err := LoadAndValidate()
		if err != nil {
			initErr = err
			return
		}
		globalConfig = cfg
	})
	return initErr
}

// Get 获取全局配置，其他模块调用
func Get() (*Config, error) {
	if globalConfig == nil {
		return nil, errors.New("config not initialized, call Init() first")
	}
	return globalConfig, nil
}
