package main

import (
	"os"
	"path/filepath"
)

// 默认配置文件路径
const (
	defaultConfigPath                          = "config/dangerous_commands.yaml"
	defaultTrustedConfigPath                   = "config/privilege_escalation_whitelist.yaml"
	defaultMaliciousRequestConfigPath          = "config/malicious_request_rules.yaml"
	defaultSensitiveFileConfigPath             = "config/sensitive_file_rules.yaml"
	defaultFileMonitorWhitelistPath            = "config/file_monitor_whitelist.yaml"
	defaultContainerDangerousCommandConfigPath = "config/container_dangerous_commands.yaml"
	defaultContainerSensitiveFileConfigPath    = "config/container_sensitive_file_rules.yaml"
)

func getConfigPath() string {
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, defaultConfigPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}
	return defaultConfigPath
}

func getTrustedConfigPath() string {
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, defaultTrustedConfigPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}
	return defaultTrustedConfigPath
}

func getMaliciousRequestConfigPath() string {
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, defaultMaliciousRequestConfigPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}
	return defaultMaliciousRequestConfigPath
}

func getSensitiveFileConfigPath() string {
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, defaultSensitiveFileConfigPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}
	return defaultSensitiveFileConfigPath
}

func getFileMonitorWhitelistPath() string {
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, defaultFileMonitorWhitelistPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}
	return defaultFileMonitorWhitelistPath
}

func getContainerDangerousCommandConfigPath() string {
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, defaultContainerDangerousCommandConfigPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}
	return defaultContainerDangerousCommandConfigPath
}

func getContainerSensitiveFileConfigPath() string {
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, defaultContainerSensitiveFileConfigPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}
	return defaultContainerSensitiveFileConfigPath
}
