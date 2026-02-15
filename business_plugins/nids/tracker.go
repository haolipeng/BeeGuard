package main

import (
	"fmt"
	"sync"
	"time"
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
}

// NewAttackTracker 创建攻击追踪器
func NewAttackTracker() *AttackTracker {
	return &AttackTracker{
		states: make(map[string]*AttackState),
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
