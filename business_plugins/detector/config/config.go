package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 全局配置
type Config struct {
	SSH        SSHConfig        `yaml:"ssh" json:"ssh"`
	FTP        FTPConfig        `yaml:"ftp" json:"ftp"`
	SSHAnomaly SSHAnomalyConfig `yaml:"ssh_anomaly" json:"ssh_anomaly"`
}

// SSHConfig SSH检测配置
type SSHConfig struct {
	Enabled   bool     `yaml:"enabled" json:"enabled"`
	LogPaths  []string `yaml:"log_paths" json:"log_paths"`
	Rules     []Rule   `yaml:"rules" json:"rules"`
	Whitelist []string `yaml:"whitelist" json:"whitelist"`
}

// FTPConfig FTP检测配置
type FTPConfig struct {
	Enabled   bool     `yaml:"enabled" json:"enabled"`
	LogPaths  []string `yaml:"log_paths" json:"log_paths"`
	Rules     []Rule   `yaml:"rules" json:"rules"`
	Whitelist []string `yaml:"whitelist" json:"whitelist"`
}

// SSHAnomalyConfig SSH异常登录检测配置
type SSHAnomalyConfig struct {
	Enabled      bool          `yaml:"enabled" json:"enabled"`
	LogPaths     []string      `yaml:"log_paths" json:"log_paths"`
	AnomalyRules []AnomalyRule `yaml:"anomaly_rules" json:"anomaly_rules"`
	AlertLevel   int           `yaml:"alert_level" json:"alert_level"`
	IgnoreTime   int           `yaml:"ignore_time" json:"ignore_time"` // 告警抑制时间(秒)
}

// TimeRange 时间段定义
type TimeRange struct {
	Start string `yaml:"start" json:"start"` // 开始时间，格式 "HH:MM"，如 "09:00"
	End   string `yaml:"end" json:"end"`     // 结束时间，格式 "HH:MM"，如 "18:00"
}

// AnomalyRule 异常登录白名单规则
type AnomalyRule struct {
	Name        string      `yaml:"name" json:"name"`
	Description string      `yaml:"description" json:"description"`
	Enabled     bool        `yaml:"enabled" json:"enabled"`
	IPs         []string    `yaml:"ips" json:"ips"`                   // IP白名单（单IP列表）
	TimeRanges  []TimeRange `yaml:"time_ranges" json:"time_ranges"`   // 允许的时间段列表，为空表示全天允许
	Users       []string    `yaml:"users" json:"users"`               // 允许的用户列表，为空表示不限制用户
}

// Rule 检测规则
type Rule struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Pattern     string `yaml:"pattern" json:"pattern"`
	Action      string `yaml:"action" json:"action"`
	Frequency   int    `yaml:"frequency" json:"frequency"`
	Timeframe   int    `yaml:"timeframe" json:"timeframe"`
	Level       int    `yaml:"level" json:"level"`
	Ignore      int    `yaml:"ignore" json:"ignore"`
	GroupBy     string `yaml:"group_by" json:"group_by"`
}

// Load 从指定目录加载配置
func Load(configDir string) (*Config, error) {
	cfg := &Config{}

	// 加载SSH暴力破解检测配置
	sshConfigPath := filepath.Join(configDir, "ssh_brute_force.yaml")
	if _, err := os.Stat(sshConfigPath); err == nil {
		sshCfg, err := loadSSHConfig(sshConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load ssh config: %w", err)
		}
		cfg.SSH = *sshCfg
	} else {
		// 使用默认配置
		cfg.SSH = defaultSSHConfig()
	}

	// 加载FTP暴力破解检测配置
	ftpConfigPath := filepath.Join(configDir, "ftp_brute_force.yaml")
	if _, err := os.Stat(ftpConfigPath); err == nil {
		ftpCfg, err := loadFTPConfig(ftpConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load ftp config: %w", err)
		}
		cfg.FTP = *ftpCfg
	} else {
		// 使用默认配置
		cfg.FTP = defaultFTPConfig()
	}

	// 加载SSH异常登录检测配置
	sshAnomalyConfigPath := filepath.Join(configDir, "ssh_anomaly_login.yaml")
	if _, err := os.Stat(sshAnomalyConfigPath); err == nil {
		sshAnomalyCfg, err := loadSSHAnomalyConfig(sshAnomalyConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load ssh_anomaly_login config: %w", err)
		}
		cfg.SSHAnomaly = *sshAnomalyCfg
	} else {
		// 使用默认配置（功能关闭）
		cfg.SSHAnomaly = defaultSSHAnomalyConfig()
	}

	return cfg, nil
}

// loadSSHConfig 加载SSH配置文件
func loadSSHConfig(path string) (*SSHConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		SSH SSHConfig `yaml:"ssh"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	// 设置默认值
	setSSHDefaults(&wrapper.SSH)

	return &wrapper.SSH, nil
}

// defaultSSHConfig 返回默认SSH配置
func defaultSSHConfig() SSHConfig {
	cfg := SSHConfig{
		Enabled: true,
		LogPaths: []string{
			"/var/log/auth.log",
			"/var/log/secure",
		},
		Rules: []Rule{
			{
				Name:        "auth_failure_brute_force",
				Description: "SSH认证失败暴力破解检测",
				Pattern:     `Failed (password|publickey) for .* from (\S+)`,
				Action:      "failed",
				Frequency:   6,
				Timeframe:   120,
				Level:       10,
				Ignore:      60,
				GroupBy:     "source_ip",
			},
			{
				Name:        "invalid_user_brute_force",
				Description: "SSH非法用户暴力破解检测",
				Pattern:     `(Invalid|Illegal) user .* from (\S+)`,
				Action:      "invalid_user",
				Frequency:   6,
				Timeframe:   120,
				Level:       10,
				Ignore:      60,
				GroupBy:     "source_ip",
			},
		},
		Whitelist: []string{
			"127.0.0.1",
			"::1",
		},
	}
	return cfg
}

// setSSHDefaults 设置SSH配置默认值
func setSSHDefaults(cfg *SSHConfig) {
	if len(cfg.LogPaths) == 0 {
		cfg.LogPaths = []string{
			"/var/log/auth.log",
			"/var/log/secure",
		}
	}

	for i := range cfg.Rules {
		if cfg.Rules[i].Frequency == 0 {
			cfg.Rules[i].Frequency = 6
		}
		if cfg.Rules[i].Timeframe == 0 {
			cfg.Rules[i].Timeframe = 120
		}
		if cfg.Rules[i].Level == 0 {
			cfg.Rules[i].Level = 10
		}
		if cfg.Rules[i].Ignore == 0 {
			cfg.Rules[i].Ignore = 60
		}
		if cfg.Rules[i].GroupBy == "" {
			cfg.Rules[i].GroupBy = "source_ip"
		}
	}
}

// IsWhitelisted 检查IP是否在白名单中
func (c *SSHConfig) IsWhitelisted(ip string) bool {
	for _, w := range c.Whitelist {
		if w == ip {
			return true
		}
	}
	return false
}

// ParseSSHConfigFromJSON 从 JSON 字符串解析 SSH 配置
func ParseSSHConfigFromJSON(data string) (*SSHConfig, error) {
	var wrapper struct {
		SSH SSHConfig `json:"ssh"`
	}
	if err := json.Unmarshal([]byte(data), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}

	// 设置默认值
	setSSHDefaults(&wrapper.SSH)

	return &wrapper.SSH, nil
}

// loadFTPConfig 加载FTP配置文件
func loadFTPConfig(path string) (*FTPConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		FTP FTPConfig `yaml:"ftp"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	// 设置默认值
	setFTPDefaults(&wrapper.FTP)

	return &wrapper.FTP, nil
}

// defaultFTPConfig 返回默认FTP配置
func defaultFTPConfig() FTPConfig {
	cfg := FTPConfig{
		Enabled: true,
		LogPaths: []string{
			"/var/log/vsftpd.log",
			"/var/log/xferlog",
		},
		Rules: []Rule{
			{
				Name:        "auth_failure_brute_force",
				Description: "FTP认证失败暴力破解检测",
				Action:      "failed",
				Frequency:   6,
				Timeframe:   120,
				Level:       10,
				Ignore:      60,
				GroupBy:     "source_ip",
			},
			{
				Name:        "multiple_connection_attempt",
				Description: "FTP多次连接尝试检测",
				Action:      "connect",
				Frequency:   10,
				Timeframe:   60,
				Level:       10,
				Ignore:      60,
				GroupBy:     "source_ip",
			},
		},
		Whitelist: []string{
			"127.0.0.1",
			"::1",
		},
	}
	return cfg
}

// setFTPDefaults 设置FTP配置默认值
func setFTPDefaults(cfg *FTPConfig) {
	if len(cfg.LogPaths) == 0 {
		cfg.LogPaths = []string{
			"/var/log/vsftpd.log",
			"/var/log/xferlog",
		}
	}

	for i := range cfg.Rules {
		if cfg.Rules[i].Frequency == 0 {
			cfg.Rules[i].Frequency = 6
		}
		if cfg.Rules[i].Timeframe == 0 {
			cfg.Rules[i].Timeframe = 120
		}
		if cfg.Rules[i].Level == 0 {
			cfg.Rules[i].Level = 10
		}
		if cfg.Rules[i].Ignore == 0 {
			cfg.Rules[i].Ignore = 60
		}
		if cfg.Rules[i].GroupBy == "" {
			cfg.Rules[i].GroupBy = "source_ip"
		}
	}
}

// IsWhitelisted 检查IP是否在FTP白名单中
func (c *FTPConfig) IsWhitelisted(ip string) bool {
	for _, w := range c.Whitelist {
		if w == ip {
			return true
		}
	}
	return false
}

// ParseFTPConfigFromJSON 从 JSON 字符串解析 FTP 配置
func ParseFTPConfigFromJSON(data string) (*FTPConfig, error) {
	var wrapper struct {
		FTP FTPConfig `json:"ftp"`
	}
	if err := json.Unmarshal([]byte(data), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}

	// 设置默认值
	setFTPDefaults(&wrapper.FTP)

	return &wrapper.FTP, nil
}

// loadSSHAnomalyConfig 加载SSH异常登录检测配置文件
func loadSSHAnomalyConfig(path string) (*SSHAnomalyConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		SSHAnomaly SSHAnomalyConfig `yaml:"ssh_anomaly_login"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	// 设置默认值
	setSSHAnomalyDefaults(&wrapper.SSHAnomaly)

	return &wrapper.SSHAnomaly, nil
}

// defaultSSHAnomalyConfig 返回默认SSH异常登录配置（功能关闭）
func defaultSSHAnomalyConfig() SSHAnomalyConfig {
	return SSHAnomalyConfig{
		Enabled: false,
		LogPaths: []string{
			"/var/log/auth.log",
			"/var/log/secure",
		},
		AnomalyRules: []AnomalyRule{},
		AlertLevel:   8,
		IgnoreTime:   300,
	}
}

// setSSHAnomalyDefaults 设置SSH异常登录配置默认值
func setSSHAnomalyDefaults(cfg *SSHAnomalyConfig) {
	if len(cfg.LogPaths) == 0 {
		cfg.LogPaths = []string{
			"/var/log/auth.log",
			"/var/log/secure",
		}
	}
	if cfg.AlertLevel == 0 {
		cfg.AlertLevel = 8
	}
	if cfg.IgnoreTime == 0 {
		cfg.IgnoreTime = 300
	}
}

// ParseSSHAnomalyConfigFromJSON 从 JSON 字符串解析 SSH 异常登录配置
func ParseSSHAnomalyConfigFromJSON(data string) (*SSHAnomalyConfig, error) {
	var wrapper struct {
		SSHAnomaly SSHAnomalyConfig `json:"ssh_anomaly_login"`
	}
	if err := json.Unmarshal([]byte(data), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}

	// 设置默认值
	setSSHAnomalyDefaults(&wrapper.SSHAnomaly)

	return &wrapper.SSHAnomaly, nil
}
