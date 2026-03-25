package analysis

import (
	"context"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/log"
)

var (
	defaultEngine    *Engine
	defaultScheduler *Scheduler
)

// Config 分析模块配置
type Config struct {
	// Ollama配置
	OllamaURL   string
	OllamaModel string // 默认 qwen3.5:0.8b

	// 缓存配置
	CacheDir string // 默认 /tmp/server/analysis_cache
	CacheTTL time.Duration // 默认 24小时

	// 报告配置
	ReportDir string // 默认 /tmp/server/analysis_reports

	// 调度配置
	ScheduleInterval time.Duration // 默认 30分钟
	AutoStart        bool          // 是否自动启动调度器
}

// Init 初始化分析模块
func Init(cfg Config) error {
	// 设置默认值
	if cfg.OllamaModel == "" {
		cfg.OllamaModel = "qwen3.5:0.8b"
	}

	// 创建引擎
	defaultEngine = NewEngine(EngineConfig{
		OllamaURL:   cfg.OllamaURL,
		OllamaModel: cfg.OllamaModel,
		CacheDir:    cfg.CacheDir,
		CacheTTL:    cfg.CacheTTL,
		ReportDir:   cfg.ReportDir,
	})

	// 检查Ollama连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := defaultEngine.ollama.CheckHealth(ctx); err != nil {
		log.Warnf("[Analysis] Ollama服务检查失败: %v, 分析功能可能不可用", err)
	} else {
		log.Infof("[Analysis] Ollama服务正常, 模型: %s", cfg.OllamaModel)
	}

	// 创建调度器
	defaultScheduler = NewScheduler(defaultEngine, SchedulerConfig{
		Interval: cfg.ScheduleInterval,
	})

	// 自动启动
	if cfg.AutoStart {
		if err := defaultScheduler.Start(); err != nil {
			log.Warnf("[Analysis] 启动调度器失败: %v", err)
		}
	}

	log.Info("[Analysis] 模块初始化完成")
	return nil
}

// GetEngine 获取引擎实例
func GetEngine() *Engine {
	return defaultEngine
}

// GetScheduler 获取调度器实例
func GetScheduler() *Scheduler {
	return defaultScheduler
}

// AnalyzeByHost 分析指定主机的告警
func AnalyzeByHost(ctx context.Context, hostIP string) (*AnalysisReport, error) {
	if defaultEngine == nil {
		return nil, nil
	}
	return defaultEngine.AnalyzeByHost(ctx, hostIP)
}

// AnalyzeBySourceIP 分析指定攻击源的告警
func AnalyzeBySourceIP(ctx context.Context, sourceIP string) (*AnalysisReport, error) {
	if defaultEngine == nil {
		return nil, nil
	}
	return defaultEngine.AnalyzeBySourceIP(ctx, sourceIP)
}

// AnalyzeCriticalAlert 分析高危告警
func AnalyzeCriticalAlert(ctx context.Context, alertType string, alertID int64) (*AnalysisReport, error) {
	if defaultEngine == nil {
		return nil, nil
	}
	return defaultEngine.AnalyzeCriticalAlert(ctx, alertType, alertID)
}

// TriggerAnalysis 手动触发分析
func TriggerAnalysis(ctx context.Context) error {
	if defaultScheduler == nil {
		return nil
	}
	return defaultScheduler.Trigger(ctx)
}

// Stats 获取统计信息
func Stats() map[string]interface{} {
	if defaultEngine == nil {
		return nil
	}
	return defaultEngine.Stats()
}

// Stop 停止分析模块
func Stop() {
	if defaultScheduler != nil {
		defaultScheduler.Stop()
	}
	log.Info("[Analysis] 模块已停止")
}
