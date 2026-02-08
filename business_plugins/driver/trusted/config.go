package trusted

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "config/trusted_executables.yaml"

// LoadConfig 加载并验证配置文件
func LoadConfig(path string) (*TrustedConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg TrustedConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// 验证条目必须是绝对路径且长度 < 256（与 eBPF 哈希函数限制一致）
	for _, entry := range cfg.TrustedExecutables {
		if len(entry) == 0 {
			return nil, fmt.Errorf("empty entry in trusted_executables")
		}
		if len(entry) >= 256 {
			return nil, fmt.Errorf("entry too long (max 255): %s", entry)
		}
		if !strings.HasPrefix(entry, "/") {
			return nil, fmt.Errorf("entry must be absolute path: %s", entry)
		}
	}

	return &cfg, nil
}
