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

	// LogDirectory 日志目录
	LogDirectory string `yaml:"log_directory"`

	// RetryMaxCount 最大重试次数
	RetryMaxCount int `yaml:"retry_max_count"`

	// RetryInterval 重试间隔（秒）
	RetryInterval int `yaml:"retry_interval"`

	// Standalone standalone 模式配置
	Standalone *StandaloneConfig `yaml:"standalone,omitempty"`

	// Log 日志配置
	Log *LogConfig `yaml:"log,omitempty"`
}

// StandaloneConfig standalone 模式配置
type StandaloneConfig struct {
	// Enabled 是否启用 standalone 模式
	Enabled bool `yaml:"enabled"`

	// Output 输出方式: "stderr" (控制台输出) 或文件路径（如 "/tmp/results.json"）
	Output string `yaml:"output"`

	// Plugins 指定加载的插件列表，为空则加载全部
	Plugins []string `yaml:"plugins,omitempty"`

	// FlushInterval 刷新间隔（秒）
	FlushInterval int `yaml:"flush_interval"`
}

// LogConfig 日志配置
type LogConfig struct {
	// Level 日志级别: debug/info/warn/error，默认 "info"
	Level string `yaml:"level"`

	// File 日志文件路径，空或 "stderr" 则输出到 stderr
	File string `yaml:"file"`

	// MaxSize 单文件最大 MB，默认 10
	MaxSize int `yaml:"max_size"`

	// MaxBackups 保留旧文件数，默认 5
	MaxBackups int `yaml:"max_backups"`

	// Compress 是否压缩旧文件，默认 false
	Compress bool `yaml:"compress"`
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
	// 0. 环境变量覆盖 server 地址（由 systemd EnvironmentFile 注入）
	if serverEnv := os.Getenv("SPECIFIED_SERVER"); serverEnv != "" {
		cfg.Server = serverEnv
	}

	// 1. Standalone 模式下 server 不是必须的
	isStandalone := cfg.Standalone != nil && cfg.Standalone.Enabled
	if !isStandalone && cfg.Server == "" {
		return fmt.Errorf("server address is required (or enable standalone mode)")
	}

	// 2. 设置 standalone 模式默认值
	if isStandalone {
		if cfg.Standalone.Output == "" {
			cfg.Standalone.Output = "stderr"
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
		cfg.PluginsDirectory = "/opt/cloudsec/agent/plugins"
	}

	if cfg.LogDirectory == "" {
		cfg.LogDirectory = "/opt/cloudsec/agent/logs"
	}

	if cfg.RetryMaxCount <= 0 {
		cfg.RetryMaxCount = 10 // 默认最大重试 10 次
	}

	if cfg.RetryInterval <= 0 {
		cfg.RetryInterval = 5 // 默认重试间隔 5 秒
	}

	// 4. 设置日志默认值
	if cfg.Log == nil {
		cfg.Log = &LogConfig{}
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.MaxSize <= 0 {
		cfg.Log.MaxSize = 10
	}
	if cfg.Log.MaxBackups <= 0 {
		cfg.Log.MaxBackups = 5
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
func SetStandalone(enabled bool, output string, plugins []string) error {
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
	if len(plugins) > 0 {
		globalConfig.Standalone.Plugins = plugins
	}

	// 设置默认值
	if globalConfig.Standalone.Output == "" {
		globalConfig.Standalone.Output = "stderr"
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
