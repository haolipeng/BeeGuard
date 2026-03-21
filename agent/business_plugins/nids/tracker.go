package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nids/log"
)

const (
	cleanupInterval = 5 * time.Minute // 清理扫描间隔
	entryTTL        = 1 * time.Hour   // 条目过期时间
	maxEntries      = 10000           // map 最大条目数
)

// AttackState 攻击状态
type AttackState struct {
	Count         int64
	FirstSeenTime time.Time
	LastSeenTime  time.Time
}

// AttackTracker 攻击状态追踪器（源IP+sid 聚合）
type AttackTracker struct {
	mu     sync.RWMutex
	states map[string]*AttackState // key: "srcIP:sid"
	logger *log.Logger
}

// NewAttackTracker 创建攻击追踪器
func NewAttackTracker(logger *log.Logger) *AttackTracker {
	return &AttackTracker{
		states: make(map[string]*AttackState),
		logger: logger,
	}
}

// StartCleanup 启动后台清理 goroutine，定期删除过期条目
func (t *AttackTracker) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				t.cleanup()
			}
		}
	}()
}

// cleanup 删除 LastSeenTime 超过 entryTTL 的条目
func (t *AttackTracker) cleanup() {
	now := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()

	expired := 0
	for key, state := range t.states {
		if now.Sub(state.LastSeenTime) > entryTTL {
			delete(t.states, key)
			expired++
		}
	}
	if expired > 0 {
		t.logger.Info("AttackTracker cleanup completed", "expired", expired, "remaining", len(t.states))
	}
}

// RecordAttack 记录一次攻击，返回更新后的状态
func (t *AttackTracker) RecordAttack(srcIP string, sid int) *AttackState {
	key := fmt.Sprintf("%s:%d", srcIP, sid)
	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	state, exists := t.states[key]
	if !exists {
		if len(t.states) >= maxEntries {
			t.logger.Warn("AttackTracker capacity limit reached, dropping new entry", "key", key, "maxEntries", maxEntries)
			return &AttackState{}
		}
		state = &AttackState{
			Count:         1,
			FirstSeenTime: now,
			LastSeenTime:  now,
		}
		t.states[key] = state
	} else {
		state.Count++
		state.LastSeenTime = now
	}

	// 返回副本
	return &AttackState{
		Count:         state.Count,
		FirstSeenTime: state.FirstSeenTime,
		LastSeenTime:  state.LastSeenTime,
	}
}
