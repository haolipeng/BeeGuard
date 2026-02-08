package trusted

// TrustedConfig YAML 配置结构
type TrustedConfig struct {
	Version            string   `yaml:"version"`
	Description        string   `yaml:"description"`
	TrustedExecutables []string `yaml:"trusted_executables"`
	Enabled            bool     `yaml:"enabled"`
	LogFilteredEvents  bool     `yaml:"log_filtered_events"`
}

// ExeItem BPF map 值类型 (必须与 C 结构体完全匹配)
type ExeItem struct {
	Len  int32      // 字符串长度 (不包含 \0)
	Sid  uint32     // 预留字段,设为 0
	Hash uint64     // Murmur OAAT64 哈希值
	Name [2048]byte // 可执行文件路径
}
