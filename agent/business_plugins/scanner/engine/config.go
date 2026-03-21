package engine

const (
	// DefaultMaxFileSize 默认最大扫描文件大小（18MB）
	DefaultMaxFileSize = 18874368

	// DefaultMaxScanTime 默认单文件扫描超时（秒）
	DefaultMaxScanTime = 5

	// DefaultScanWorkers 默认扫描线程数
	DefaultScanWorkers = 6

	// DefaultMaxMemoryMB 默认最大内存限制（MB）
	DefaultMaxMemoryMB = 512

	// ClamAVScanOptionStandard 标准扫描选项
	ClamAVScanOptionStandard = 0

	// ClamAVScanOptionAllMatch 扫描所有匹配
	ClamAVScanOptionAllMatch = 1
)
