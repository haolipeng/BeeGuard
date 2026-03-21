package engine

// ScanResult ClamAV 引擎扫描结果
type ScanResult struct {
	Infected bool   // 是否检出
	VirusName string // 病毒名称（如 "Linux.Trojan.Mirai"）
}

// Engine 扫描引擎接口
type Engine interface {
	// Init 初始化引擎
	Init() error
	// LoadDB 加载病毒数据库
	LoadDB(dbPath string) error
	// ScanFile 扫描单个文件
	ScanFile(path string) (*ScanResult, error)
	// ReloadDB 重新加载病毒数据库（热更新）
	ReloadDB(dbPath string) error
	// Close 关闭引擎，释放资源
	Close()
}
