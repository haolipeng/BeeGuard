package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// NIDSConfig 主配置结构
type NIDSConfig struct {
	Interface     string              `yaml:"interface"`
	BPFFilter     string              `yaml:"bpf_filter"`
	Snaplen       int32               `yaml:"snaplen"`
	TCPReassembly TCPReassemblyConfig `yaml:"tcp_reassembly"`
	RulesFile     string              `yaml:"rules_file"`
}

// TCPReassemblyConfig TCP 流重组配置
type TCPReassemblyConfig struct {
	MaxBufferSize int           `yaml:"max_buffer_size"`
	MaxStreams     int           `yaml:"max_streams"`
	StreamTimeout time.Duration `yaml:"stream_timeout"`
}

// LoadConfig 加载 YAML 配置文件
func LoadConfig(path string) (*NIDSConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	cfg := &NIDSConfig{
		// 默认值
		Interface: "lo",
		BPFFilter: "tcp port 80 or tcp port 8080",
		Snaplen:   65535,
		TCPReassembly: TCPReassemblyConfig{
			MaxBufferSize: 262144,
			MaxStreams:     10000,
			StreamTimeout: 120 * time.Second,
		},
		RulesFile: "config/nids.rules",
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return cfg, nil
}

// getConfigPath 获取配置文件路径
// 优先使用环境变量 NIDS_CONFIG_PATH，否则查找可执行文件同目录，最后当前目录
func getConfigPath() string {
	if path := os.Getenv("NIDS_CONFIG_PATH"); path != "" {
		return path
	}

	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, "config/nids.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return "config/nids.yaml"
}

// getRulesPath 获取规则文件路径
// 如果配置中指定的路径是相对路径，则相对于可执行文件目录
func getRulesPath(cfg *NIDSConfig) string {
	if filepath.IsAbs(cfg.RulesFile) {
		return cfg.RulesFile
	}

	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		rulesPath := filepath.Join(dir, cfg.RulesFile)
		if _, err := os.Stat(rulesPath); err == nil {
			return rulesPath
		}
	}

	return cfg.RulesFile
}
