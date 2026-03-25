package analysis

import (
	"context"
	"sync"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/log"
)

// Scheduler 分析调度器
type Scheduler struct {
	engine       *Engine
	interval     time.Duration
	mu           sync.Mutex
	running      bool
	cancel       context.CancelFunc
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	Interval time.Duration // 扫描间隔，默认 30 分钟
}

// NewScheduler 创建调度器
func NewScheduler(engine *Engine, cfg SchedulerConfig) *Scheduler {
	if cfg.Interval == 0 {
		cfg.Interval = 30 * time.Minute
	}

	return &Scheduler{
		engine:   engine,
		interval: cfg.Interval,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.running = true

	go s.run(ctx)

	log.Infof("[Scheduler] 已启动, 扫描间隔: %v", s.interval)
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	if s.cancel != nil {
		s.cancel()
	}
	s.running = false

	log.Info("[Scheduler] 已停止")
}

// run 运行循环
func (s *Scheduler) run(ctx context.Context) {
	// 启动后立即执行一次
	s.tick(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

// tick 执行一次扫描
func (s *Scheduler) tick(ctx context.Context) {
	start := time.Now()

	if err := s.engine.ScanAndAnalyze(ctx); err != nil {
		log.Warnf("[Scheduler] 扫描失败: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed > time.Minute {
		log.Infof("[Scheduler] 本次扫描耗时: %v", elapsed)
	}
}

// Trigger 手动触发一次分析
func (s *Scheduler) Trigger(ctx context.Context) error {
	return s.engine.ScanAndAnalyze(ctx)
}

// IsRunning 检查是否正在运行
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
