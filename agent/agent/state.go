package agent

import (
	"encoding/json"
	"sync"
)

// 状态类型
type StateType int32

const (
	StateTypeRunning StateType = iota
	StateTypeAbnormal
)

var stateTypeMap = map[StateType]string{
	StateTypeRunning:  "running",
	StateTypeAbnormal: "abnormal",
}

var (
	mu           = &sync.Mutex{}
	currentState = StateTypeRunning
	abnormalErrs = []string{}
)

func (x StateType) String() string {
	return stateTypeMap[x]
}

// SetRunning 设置 Agent 状态为运行中
func SetRunning() {
	mu.Lock()
	defer mu.Unlock()
	currentState = StateTypeRunning
	abnormalErrs = []string{}
}

// SetAbnormal 设置 Agent 状态为异常，并记录错误信息
func SetAbnormal(err string) {
	mu.Lock()
	defer mu.Unlock()
	currentState = StateTypeAbnormal
	abnormalErrs = append(abnormalErrs, err)
}

// State 返回当前状态和错误信息（JSON 格式）
func State() (string, string) {
	mu.Lock()
	defer mu.Unlock()
	err, _ := json.Marshal(abnormalErrs)
	return currentState.String(), string(err)
}
