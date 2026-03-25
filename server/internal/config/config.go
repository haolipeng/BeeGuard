package config

import (
	"fmt"
	//"log"
	"os"

	"gopkg.in/yaml.v3"
)

// PluginItem 插件配置项
type PluginItem struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// TaskItem 自动下发任务配置项
type TaskItem struct {
	ObjectName string `yaml:"object_name"` // 目标插件名
	DataType   int32  `yaml:"data_type"`   // 任务类型
	Data       string `yaml:"data"`        // 任务参数 JSON（可选）
}

// Config 服务器配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
	GeoIP    GeoIPConfig    `yaml:"geoip"`
	Vuln     VulnConfig     `yaml:"vuln"`
	Analysis AnalysisConfig `yaml:"analysis"`
	Install  InstallConfig  `yaml:"install"`
	Plugins  []PluginItem   `yaml:"plugins"`
	Tasks    []TaskItem     `yaml:"tasks"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	Charset      string `yaml:"charset"`        // 字符集
	ParseTime    bool   `yaml:"parse_time"`     // 是否解析时间
	Loc          string `yaml:"loc"`            // 时区
	PoolSize     int    `yaml:"pool_size"`      // 连接池大小
	GormLogLevel string `yaml:"gorm_log_level"` // GORM 日志级别: silent, error, warn, info (默认: error)
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`   // 允许的来源
	AllowedMethods   []string `yaml:"allowed_methods"`   // 允许的HTTP方法
	AllowedHeaders   []string `yaml:"allowed_headers"`   // 允许的请求头
	AllowCredentials bool     `yaml:"allow_credentials"` // 是否允许携带凭证
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret      string `yaml:"secret"`       // JWT 签名密钥
	ExpireHours int    `yaml:"expire_hours"` // Token 过期时间（小时）
}

// ServerConfig 服务器相关配置
type ServerConfig struct {
	Port           int        `yaml:"port"`              // gRPC 端口
	HttpPort       int        `yaml:"http_port"`         // HTTP API 端口
	MaxRecvMsgSize int        `yaml:"max_recv_msg_size"` // 单位: MB
	MaxSendMsgSize int        `yaml:"max_send_msg_size"` // 单位: MB
	CORS           CORSConfig `yaml:"cors"`              // CORS配置
	JWT            JWTConfig  `yaml:"jwt"`               // JWT配置
}

// LogConfig 日志相关配置
type LogConfig struct {
	Level      string `yaml:"level"`       // debug, info, warn, error
	Dir        string `yaml:"dir"`         // 日志目录，为空则不写文件
	Filename   string `yaml:"filename"`    // 日志文件名，默认 server.log
	MaxSize    int    `yaml:"max_size"`    // 单文件最大 MB，默认 100
	MaxAge     int    `yaml:"max_age"`     // 保留天数，默认 30
	MaxBackups int    `yaml:"max_backups"` // 保留旧文件数，默认 0 不限
}

// GeoIPConfig GeoIP 地理位置查询配置
type GeoIPConfig struct {
	Enabled      bool   `yaml:"enabled"`        // 是否启用
	DBPath       string `yaml:"db_path"`        // 数据库文件路径
	CacheTTL     int    `yaml:"cache_ttl"`      // 缓存过期时间（秒）
	MaxCacheSize int    `yaml:"max_cache_size"` // 最大缓存条目数
}

// VulnConfig 漏洞扫描配置
type VulnConfig struct {
	Enabled        bool   `yaml:"enabled"`         // 是否启用漏洞扫描
	DBDir          string `yaml:"db_dir"`          // 漏洞数据库存储目录
	DBRepository   string `yaml:"db_repository"`   // OCI 仓库地址
	UpdateInterval int    `yaml:"update_interval"` // 漏洞库更新间隔（小时）
	ScanCron       string `yaml:"scan_cron"`       // 定时扫描 cron 表达式
}

// AnalysisConfig AI分析配置
type AnalysisConfig struct {
	Enabled         bool   `yaml:"enabled"`           // 是否启用AI分析
	OllamaURL       string `yaml:"ollama_url"`        // Ollama服务地址
	OllamaModel     string `yaml:"ollama_model"`      // 模型名称
	CacheDir        string `yaml:"cache_dir"`         // 缓存目录
	ReportDir       string `yaml:"report_dir"`        // 报告存储目录
	ScheduleMinutes int    `yaml:"schedule_minutes"`  // 调度间隔（分钟）
}

// InstallConfig Agent 一键安装配置
type InstallConfig struct {
	Enabled    bool   `yaml:"enabled"`     // 是否启用一键安装功能
	PackageDir string `yaml:"package_dir"` // 安装包存放目录
	ServerAddr string `yaml:"server_addr"` // Agent 连接的 gRPC 地址 (ip:port)
}

// 全局配置变量
var AppConfig *Config

// LoadConfig 加载配置到全局变量
func LoadConfig() {
	cfg, err := Load("conf/server.yaml")
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}
	AppConfig = cfg
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:           50051,
			HttpPort:       8080,
			MaxRecvMsgSize: 16,
			MaxSendMsgSize: 16,
			CORS: CORSConfig{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
				AllowCredentials: true,
			},
			JWT: JWTConfig{
				Secret:      "server-default-jwt-secret",
				ExpireHours: 24,
			},
		},
		Database: DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "postgres",
			Password:     "",
			Database:     "server",
			PoolSize:     10,
			GormLogLevel: "error", // 默认只记录错误
		},
		Log: LogConfig{
			Level:    "info",
			Dir:      "/opt/cloudsec/server/logs",
			Filename: "server.log",
			MaxSize:  100,
			MaxAge:   30,
		},
		Vuln: VulnConfig{
			Enabled:        false,
			DBDir:          "/opt/cloudsec/server/data/trivy-db", //漏洞数据库文件默认路径
			DBRepository:   "ghcr.io/aquasecurity/trivy-db:2", //db默认仓库名称
			UpdateInterval: 24,
			ScanCron:       "0 2 * * *",
		},
		Analysis: AnalysisConfig{
			Enabled:         false,
			OllamaURL:       "http://localhost:11434",
			OllamaModel:     "qwen3.5:0.8b",
			CacheDir:        "/tmp/server/analysis_cache",
			ReportDir:       "/tmp/server/analysis_reports",
			ScheduleMinutes: 30,
		},
		Install: InstallConfig{
			Enabled:    false,
			PackageDir: "/opt/cloudsec/server/packages",
		},
	}
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("无效的 gRPC 端口号: %d", c.Server.Port)
	}
	if c.Server.HttpPort <= 0 || c.Server.HttpPort > 65535 {
		return fmt.Errorf("无效的 HTTP 端口号: %d", c.Server.HttpPort)
	}
	if c.Server.Port == c.Server.HttpPort {
		return fmt.Errorf("gRPC 端口和 HTTP 端口不能相同")
	}
	if c.Server.MaxRecvMsgSize <= 0 {
		return fmt.Errorf("max_recv_msg_size 必须大于 0")
	}
	if c.Server.MaxSendMsgSize <= 0 {
		return fmt.Errorf("max_send_msg_size 必须大于 0")
	}

	// 验证 JWT 配置
	if c.Server.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret 不能为空")
	}
	if c.Server.JWT.ExpireHours <= 0 {
		c.Server.JWT.ExpireHours = 24
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Log.Level] {
		return fmt.Errorf("无效的日志级别: %s", c.Log.Level)
	}

	// 验证数据库配置
	if c.Database.Host == "" {
		return fmt.Errorf("数据库 host 不能为空")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("无效的数据库端口号: %d", c.Database.Port)
	}
	if c.Database.User == "" {
		return fmt.Errorf("数据库 user 不能为空")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("数据库名称不能为空")
	}
	if c.Database.Charset == "" {
		c.Database.Charset = "utf8mb4"
	}
	if c.Database.Loc == "" {
		c.Database.Loc = "Local"
	}
	if c.Database.PoolSize <= 0 {
		c.Database.PoolSize = 10
	}
	// 验证 GORM 日志级别（如果配置文件中没有该字段或为空，默认为 error）
	validGormLevels := map[string]bool{"silent": true, "error": true, "warn": true, "info": true}
	if c.Database.GormLogLevel == "" {
		c.Database.GormLogLevel = "error" // 默认值：当配置文件中没有 gorm_log_level 字段时，默认为 error
	} else if !validGormLevels[c.Database.GormLogLevel] {
		return fmt.Errorf("无效的 GORM 日志级别: %s (有效值: silent, error, warn, info)", c.Database.GormLogLevel)
	}

	// 验证 GeoIP 配置
	if c.GeoIP.Enabled {
		if c.GeoIP.DBPath == "" {
			return fmt.Errorf("geoip.db_path 不能为空")
		}
		// 检查文件是否存在（非严格要求，允许运行时下载）
		if _, err := os.Stat(c.GeoIP.DBPath); err != nil {
			// 注意：这里使用 fmt.Printf 因为 log 包可能未初始化
			fmt.Printf("警告: GeoIP 数据库文件不存在: %s (将在运行时尝试使用)\n", c.GeoIP.DBPath)
		}
	}

	// 验证漏洞扫描配置
	if c.Vuln.Enabled {
		if c.Vuln.DBDir == "" {
			return fmt.Errorf("vuln.db_dir 不能为空")
		}
		if c.Vuln.DBRepository == "" {
			c.Vuln.DBRepository = "ghcr.io/aquasecurity/trivy-db:2"
		}
		if c.Vuln.UpdateInterval <= 0 {
			c.Vuln.UpdateInterval = 24
		}
		if c.Vuln.ScanCron == "" {
			c.Vuln.ScanCron = "0 2 * * *"
		}
	}

	// 验证一键安装配置
	if c.Install.Enabled {
		if c.Install.PackageDir == "" {
			return fmt.Errorf("install.package_dir 不能为空")
		}
	}

	return nil
}
