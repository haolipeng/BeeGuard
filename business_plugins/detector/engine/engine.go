package engine

import (
	"fmt"
	"sync"
	"time"

	businessplugins "business_plugins/lib"

	"gitlab.myinterest.top/security/agent/business_plugins/detector/watcher"
	"go.uber.org/zap"
)

// Engine 检测引擎
type Engine struct {
	client    *businessplugins.Client
	detectors []Detector
	watchers  []*watcher.Watcher
	done      chan struct{}
	wg        sync.WaitGroup
}

// New 创建检测引擎
func New(client *businessplugins.Client) *Engine {
	return &Engine{
		client:    client,
		detectors: []Detector{},
		watchers:  []*watcher.Watcher{},
		done:      make(chan struct{}),
	}
}

// Register 注册检测器
func (e *Engine) Register(d Detector) {
	e.detectors = append(e.detectors, d)
}

// Run 启动检测引擎
func (e *Engine) Run() {
	zap.S().Info("detection engine starting...")

	// 为每个检测器创建日志监控器
	for _, d := range e.detectors {
		detector := d // 闭包捕获
		logPaths := detector.LogPaths()

		if len(logPaths) == 0 {
			zap.S().Warnf("detector %s has no log paths configured", detector.Name())
			continue
		}

		// 创建日志行处理函数
		handler := func(line string) {
			e.processLine(detector, line)
		}

		// 创建并启动监控器
		w := watcher.New(logPaths, handler)
		if err := w.Start(); err != nil {
			zap.S().Errorf("failed to start watcher for %s: %v", detector.Name(), err)
			continue
		}

		e.watchers = append(e.watchers, w)
		zap.S().Infof("detector %s started, watching %d log files", detector.Name(), len(logPaths))
	}

	// 启动定期清理协程
	e.wg.Add(1)
	go e.cleanupLoop()

	zap.S().Info("detection engine started")
}

// processLine 处理单行日志
func (e *Engine) processLine(d Detector, line string) {
	// 解析日志行
	event := d.Parse(line)
	if event == nil {
		return // 不匹配任何规则
	}

	// 检查是否触发告警
	alert := d.Check(event)
	if alert == nil {
		return // 未达到告警阈值
	}

	// 发送告警
	e.sendAlert(d, alert)
}

// sendAlert 发送告警记录
func (e *Engine) sendAlert(d Detector, alert *Alert) {
	zap.S().Warnf("ALERT: %s brute force detected from %s, count=%d, rule=%s",
		alert.Service, alert.SourceIP, alert.Count, alert.RuleName)

	rec := &businessplugins.Record{
		DataType:  int32(d.DataType()),
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"alert_type":  alert.AlertType,
				"service":     alert.Service,
				"rule_name":   alert.RuleName,
				"description": alert.Description,
				"source_ip":   alert.SourceIP,
				"target_user": alert.TargetUser,
				"count":       fmt.Sprintf("%d", alert.Count),
				"timeframe":   fmt.Sprintf("%d", alert.Timeframe),
				"first_seen":  fmt.Sprintf("%d", alert.FirstSeen),
				"last_seen":   fmt.Sprintf("%d", alert.LastSeen),
				"level":       fmt.Sprintf("%d", alert.Level),
			},
		},
	}

	if err := e.client.SendRecord(rec); err != nil {
		zap.S().Errorf("failed to send alert: %v", err)
	}
}

// cleanupLoop 定期清理过期数据
func (e *Engine) cleanupLoop() {
	defer e.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-e.done:
			return
		case <-ticker.C:
			// 这里可以添加清理逻辑，如清理滑动窗口中的过期数据
			zap.S().Debug("cleanup tick")
		}
	}
}

// Stop 停止检测引擎
func (e *Engine) Stop() {
	zap.S().Info("detection engine stopping...")
	close(e.done)

	// 停止所有监控器
	for _, w := range e.watchers {
		w.Stop()
	}

	e.wg.Wait()
	zap.S().Info("detection engine stopped")
}
