package agent

import (
	"encoding/json"
	"sync"
	"testing"
)

// TestStateType_String 测试状态类型的字符串表示
func TestStateType_String(t *testing.T) {
	if StateTypeRunning.String() != "running" {
		t.Errorf("Expected 'running', got '%s'", StateTypeRunning.String())
	}
	if StateTypeAbnormal.String() != "abnormal" {
		t.Errorf("Expected 'abnormal', got '%s'", StateTypeAbnormal.String())
	}
}

// TestSetRunning 测试设置运行状态
func TestSetRunning(t *testing.T) {
	SetRunning()
	state, errJSON := State()
	if state != "running" {
		t.Errorf("Expected state 'running', got '%s'", state)
	}
	var errs []string
	if err := json.Unmarshal([]byte(errJSON), &errs); err != nil {
		t.Fatalf("Failed to unmarshal error JSON: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("Expected empty error list, got %v", errs)
	}
}

// TestSetAbnormal 测试设置异常状态
func TestSetAbnormal(t *testing.T) {
	SetAbnormal("test error 1")
	state, errJSON := State()
	if state != "abnormal" {
		t.Errorf("Expected state 'abnormal', got '%s'", state)
	}
	var errs []string
	if err := json.Unmarshal([]byte(errJSON), &errs); err != nil {
		t.Fatalf("Failed to unmarshal error JSON: %v", err)
	}
	if len(errs) != 1 || errs[0] != "test error 1" {
		t.Errorf("Expected ['test error 1'], got %v", errs)
	}

	// 添加多个错误
	SetAbnormal("test error 2")
	state, errJSON = State()
	if state != "abnormal" {
		t.Errorf("Expected state 'abnormal', got '%s'", state)
	}
	if err := json.Unmarshal([]byte(errJSON), &errs); err != nil {
		t.Fatalf("Failed to unmarshal error JSON: %v", err)
	}
	if len(errs) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errs))
	}
	if errs[0] != "test error 1" || errs[1] != "test error 2" {
		t.Errorf("Expected ['test error 1', 'test error 2'], got %v", errs)
	}

	// 设置为运行状态后，错误列表应该清空
	SetRunning()
	state, errJSON = State()
	if state != "running" {
		t.Errorf("Expected state 'running', got '%s'", state)
	}
	if err := json.Unmarshal([]byte(errJSON), &errs); err != nil {
		t.Fatalf("Failed to unmarshal error JSON: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("Expected empty error list after SetRunning, got %v", errs)
	}
}

// TestState_Concurrent 测试并发状态操作的线程安全性
func TestState_Concurrent(t *testing.T) {
	// 重置为运行状态
	SetRunning()

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发设置状态
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				SetRunning()
			} else {
				SetAbnormal("concurrent error")
			}
		}(i)
	}

	wg.Wait()

	// 验证最终状态（应该是 abnormal，因为最后可能被设置为 abnormal）
	state, errJSON := State()
	t.Logf("Final state: %s, errors: %s", state, errJSON)

	// 验证状态是有效的
	if state != "running" && state != "abnormal" {
		t.Errorf("Invalid state: %s", state)
	}

	// 验证错误 JSON 格式有效
	var errs []string
	if err := json.Unmarshal([]byte(errJSON), &errs); err != nil {
		t.Errorf("Invalid error JSON format: %v", err)
	}
}

// TestState_ConcurrentRead 测试并发读取状态的线程安全性
func TestState_ConcurrentRead(t *testing.T) {
	SetRunning()

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发读取状态
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			state, errJSON := State()
			if state != "running" && state != "abnormal" {
				t.Errorf("Invalid state: %s", state)
			}
			var errs []string
			if err := json.Unmarshal([]byte(errJSON), &errs); err != nil {
				t.Errorf("Invalid error JSON format: %v", err)
			}
		}()
	}

	wg.Wait()
}

// TestState_ConcurrentReadWrite 测试并发读写状态的线程安全性
func TestState_ConcurrentReadWrite(t *testing.T) {
	SetRunning()

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// 并发写入
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				SetRunning()
			} else {
				SetAbnormal("error")
			}
		}(i)
	}

	// 并发读取
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			state, errJSON := State()
			if state != "running" && state != "abnormal" {
				t.Errorf("Invalid state: %s", state)
			}
			var errs []string
			if err := json.Unmarshal([]byte(errJSON), &errs); err != nil {
				t.Errorf("Invalid error JSON format: %v", err)
			}
		}()
	}

	wg.Wait()

	// 验证最终状态有效
	state, errJSON := State()
	if state != "running" && state != "abnormal" {
		t.Errorf("Invalid final state: %s", state)
	}
	t.Logf("Final state after concurrent operations: %s, errors: %s", state, errJSON)
}
