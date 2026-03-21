package engine

import (
	"sync"
	"time"
)

// IPEvents 单个IP的事件记录
type IPEvents struct {
	Events    []time.Time // 事件时间戳列表
	LastAlert time.Time   // 上次告警时间(用于去重)
}

// SlidingWindow 滑动时间窗口
type SlidingWindow struct {
	mu         sync.RWMutex
	events     map[string]*IPEvents // IP -> 事件列表
	timeframe  time.Duration        // 时间窗口
	threshold  int                  // 触发阈值
	ignoreTime time.Duration        // 告警抑制时间
}

// NewSlidingWindow 创建新的滑动窗口
func NewSlidingWindow(timeframe, ignoreTime time.Duration, threshold int) *SlidingWindow {
	return &SlidingWindow{
		events:     make(map[string]*IPEvents),
		timeframe:  timeframe,
		threshold:  threshold,
		ignoreTime: ignoreTime,
	}
}

// WindowResult 窗口检查结果
type WindowResult struct {
	Triggered bool      // 是否触发
	Count     int       // 事件数量
	FirstSeen time.Time // 首次事件时间
	LastSeen  time.Time // 最后事件时间
}

// Check 检查事件是否触发告警
func (w *SlidingWindow) Check(ip string, eventTime time.Time) *WindowResult {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 获取或创建IP事件记录
	ipEvents, exists := w.events[ip]
	if !exists {
		ipEvents = &IPEvents{Events: []time.Time{}}
		w.events[ip] = ipEvents
	}

	// 添加当前事件
	ipEvents.Events = append(ipEvents.Events, eventTime)

	// 清理过期事件(超出时间窗口)
	cutoff := eventTime.Add(-w.timeframe)
	validEvents := make([]time.Time, 0, len(ipEvents.Events))
	for _, t := range ipEvents.Events {
		if t.After(cutoff) || t.Equal(cutoff) {
			validEvents = append(validEvents, t)
		}
	}
	ipEvents.Events = validEvents

	// 检查是否达到阈值
	if len(validEvents) >= w.threshold {
		// 检查告警抑制
		if time.Since(ipEvents.LastAlert) < w.ignoreTime {
			return &WindowResult{
				Triggered: false,
				Count:     len(validEvents),
				FirstSeen: validEvents[0],
				LastSeen:  validEvents[len(validEvents)-1],
			}
		}

		// 更新最后告警时间
		ipEvents.LastAlert = time.Now()

		return &WindowResult{
			Triggered: true,
			Count:     len(validEvents),
			FirstSeen: validEvents[0],
			LastSeen:  validEvents[len(validEvents)-1],
		}
	}

	return &WindowResult{
		Triggered: false,
		Count:     len(validEvents),
	}
}

// Cleanup 清理过期的IP记录(减少内存占用)
func (w *SlidingWindow) Cleanup() {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-w.timeframe * 2) // 保留2倍时间窗口内的记录

	for ip, ipEvents := range w.events {
		// 如果所有事件都过期了，删除该IP记录
		if len(ipEvents.Events) == 0 {
			delete(w.events, ip)
			continue
		}

		// 检查最后一个事件是否过期
		lastEvent := ipEvents.Events[len(ipEvents.Events)-1]
		if lastEvent.Before(cutoff) {
			delete(w.events, ip)
		}
	}
}

// Stats 返回窗口统计信息
func (w *SlidingWindow) Stats() (ipCount int, totalEvents int) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	ipCount = len(w.events)
	for _, ipEvents := range w.events {
		totalEvents += len(ipEvents.Events)
	}
	return
}
