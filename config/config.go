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

	// PluginsDirectory 插件目录
	PluginsDirectory string `yaml:"plugins_directory"`

	// RetryMaxCount 最大重试次数
	RetryMaxCount int `yaml:"retry_max_count"`

	// RetryInterval 重试间隔（秒）
	RetryInterval int `yaml:"retry_interval"`

	// Standalone standalone 模式配置
	Standalone *StandaloneConfig `yaml:"standalone,omitempty"`
}

// StandaloneConfig standalone 模式配置
type StandaloneConfig struct {
	// Enabled 是否启用 standalone 模式
	Enabled bool `yaml:"enabled"`

	// Output 输出方式: "log" (zap日志) 或 "file" (JSON文件)
	Output string `yaml:"output"`

	// OutputPath JSON文件输出路径（当 Output 为 "file" 时生效）
	OutputPath string `yaml:"output_path"`

	// Plugins 指定加载的插件列表，为空则加载全部
	Plugins []string `yaml:"plugins,omitempty"`

	// FlushInterval 刷新间隔（秒）
	FlushInterval int `yaml:"flush_interval"`
}

var (
	// globalConfig 全局配置实例
	globalConfig *Config
	// initOnce 确保配置只初始化一次
	initOnce sync.Once
	// initErr 初始化错误
	initErr error
	// customConfigPath 命令行指定的配置文件路径
	customConfigPath string
)

const (
	// DefaultConfigFile 默认配置文件名称
	DefaultConfigFile = "agent.yaml"
)

// SetConfigPath 设置配置文件路径（供命令行参数使用）
func SetConfigPath(path string) {
	customConfigPath = path
}

// GetConfigPath 获取配置文件路径
// 优先级：命令行参数 > 默认路径 > 当前目录
func GetConfigPath() string {
	// 1. 命令行指定的路径优先级最高
	if customConfigPath != "" {
		return customConfigPath
	}

	// 2. 尝试默认路径（/etc/cloudsec-agent/agent.yaml）
	defaultPath := filepath.Join("/etc", "cloudsec-agent", DefaultConfigFile)
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}

	// 3. 回退到当前目录
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
	// 1. Standalone 模式下 server 不是必须的
	isStandalone := cfg.Standalone != nil && cfg.Standalone.Enabled
	if !isStandalone && cfg.Server == "" {
		return fmt.Errorf("server address is required (or enable standalone mode)")
	}

	// 2. 设置 standalone 模式默认值
	if isStandalone {
		if cfg.Standalone.Output == "" {
			cfg.Standalone.Output = "log"
		}
		if cfg.Standalone.OutputPath == "" {
			cfg.Standalone.OutputPath = "/tmp/cloudsec-detection-results.json"
		}
		if cfg.Standalone.FlushInterval <= 0 {
			cfg.Standalone.FlushInterval = 1 // 默认 1 秒
		}
	}

	// 3. 设置默认值
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 30 // 默认 30 秒
	}

	if cfg.WorkingDirectory == "" {
		cfg.WorkingDirectory = "/var/run/cloudsec-agent"
	}

	if cfg.PluginsDirectory == "" {
		cfg.PluginsDirectory = "/opt/cloudsec/plugins"
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

// SetStandalone 设置 standalone 模式配置（供命令行参数使用）
// 在 Init() 之后调用
func SetStandalone(enabled bool, output, outputPath string, plugins []string) error {
	if globalConfig == nil {
		return errors.New("config not initialized, call Init() first")
	}

	if !enabled {
		return nil
	}

	if globalConfig.Standalone == nil {
		globalConfig.Standalone = &StandaloneConfig{}
	}

	globalConfig.Standalone.Enabled = true

	if output != "" {
		globalConfig.Standalone.Output = output
	}
	if outputPath != "" {
		globalConfig.Standalone.OutputPath = outputPath
	}
	if len(plugins) > 0 {
		globalConfig.Standalone.Plugins = plugins
	}

	// 设置默认值
	if globalConfig.Standalone.Output == "" {
		globalConfig.Standalone.Output = "log"
	}
	if globalConfig.Standalone.FlushInterval <= 0 {
		globalConfig.Standalone.FlushInterval = 1
	}

	return nil
}

// IsStandalone 检查是否为 standalone 模式
func IsStandalone() bool {
	if globalConfig == nil {
		return false
	}
	return globalConfig.Standalone != nil && globalConfig.Standalone.Enabled
}
