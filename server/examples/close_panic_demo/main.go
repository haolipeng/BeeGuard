package main

import (
	"fmt"
	"sync"
)

type AgentInfo struct {
	AgentID   string
	CommandCh chan string
}

// BuggyServer 演示 RUnlock 与 channel 发送之间的竞态窗口
type BuggyServer struct {
	mu     sync.RWMutex
	agents map[string]*AgentInfo
}

func (s *BuggyServer) unregisterAgent(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.agents, agentID)
	fmt.Printf("[unregister] agent %s 已删除\n", agentID)
}

// sendCommand 存在 bug：RUnlock 后到发送之间有竞态窗口
// step1Done/step2Done 用于控制两个 goroutine 的精确交错顺序
func (s *BuggyServer) sendCommand(agentID, cmd string, step1Done, step2Done chan struct{}) {
	s.mu.RLock()
	agent, ok := s.agents[agentID]
	s.mu.RUnlock() // 锁已释放，但 agent 引用仍被持有

	if !ok {
		fmt.Printf("[sendCommand] agent %s 不存在\n", agentID)
		close(step1Done)
		return
	}

	fmt.Println("[sendCommand] 已获取 agent 引用，锁已释放")
	close(step1Done)

	<-step2Done // 等待 channel 被关闭
	fmt.Println("[sendCommand] 尝试向已关闭的 channel 发送...")

	agent.CommandCh <- cmd // panic: send on closed channel
}

// 演示时序：
//  1. sendCommand: RLock → 取 agent 引用 → RUnlock
//  2. Transfer:    Lock → delete agent → Unlock → close(ch)
//  3. sendCommand: 向已关闭的 ch 发送 → PANIC
func main() {
	fmt.Println("=== 演示 close(commandCh) 导致 panic 的竞态条件 ===")

	server := &BuggyServer{agents: make(map[string]*AgentInfo)}
	agentID := "agent-001"
	commandCh := make(chan string, 10)

	//Agent注册
	server.mu.Lock()
	server.agents[agentID] = &AgentInfo{AgentID: agentID, CommandCh: commandCh}
	server.mu.Unlock()

	go func() {
		for cmd := range commandCh {
			fmt.Printf("[receiver] 收到命令: %s\n", cmd)
		}
	}()

	step1Done := make(chan struct{}) // sendCommand 取到引用后通知
	step2Done := make(chan struct{}) // close(ch) 完成后通知
	var wg sync.WaitGroup

	// goroutine A: HTTP API 发送命令
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("\n*** PANIC: %v ***\n", r)
				fmt.Println("根因: RUnlock 到发送之间的窗口中，channel 被关闭")
			}
		}()
		server.sendCommand(agentID, "scan-task-001", step1Done, step2Done)
	}()

	// goroutine B: Transfer 连接断开清理
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-step1Done // 等 sendCommand 释放锁
		fmt.Println("[Transfer] 开始清理...")
		server.unregisterAgent(agentID)
		close(commandCh)
		fmt.Println("[Transfer] channel 已关闭")
		close(step2Done)
	}()

	wg.Wait()
}
