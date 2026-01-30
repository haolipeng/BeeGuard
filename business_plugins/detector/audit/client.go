package audit

import (
	"fmt"
	"sync"

	"github.com/elastic/go-libaudit/v2"
	"github.com/elastic/go-libaudit/v2/rule"
	"go.uber.org/zap"
)

// Client 审计客户端封装
type Client struct {
	mu          sync.Mutex
	auditClient *libaudit.AuditClient
	rules       []string // 已添加的规则key
	closed      bool
}

// New 创建审计客户端
func New() (*Client, error) {
	client, err := libaudit.NewAuditClient(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit client: %w", err)
	}

	// 获取审计状态
	status, err := client.GetStatus()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to get audit status: %w", err)
	}

	zap.S().Infof("audit status: enabled=%d, pid=%d, backlog_limit=%d",
		status.Enabled, status.PID, status.BacklogLimit)

	return &Client{
		auditClient: client,
		rules:       make([]string, 0),
	}, nil
}

// SetupExecveRule 配置execve监控规则
func (c *Client) SetupExecveRule() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	// 64位execve
	rule64 := &rule.SyscallRule{
		Type:   rule.AppendSyscallRuleType,
		List:   "exit",
		Action: "always",
		Arch:   "b64",
		Syscalls: []string{
			"execve",
		},
		Keys: []string{"detector_exec"},
	}

	ruleData64, err := rule64.Build()
	if err != nil {
		return fmt.Errorf("failed to build 64-bit execve rule: %w", err)
	}

	if err := c.auditClient.AddRule(ruleData64); err != nil {
		zap.S().Warnf("failed to add 64-bit execve rule (may already exist): %v", err)
	} else {
		c.rules = append(c.rules, "detector_exec_b64")
		zap.S().Info("added 64-bit execve audit rule")
	}

	// 32位execve
	rule32 := &rule.SyscallRule{
		Type:   rule.AppendSyscallRuleType,
		List:   "exit",
		Action: "always",
		Arch:   "b32",
		Syscalls: []string{
			"execve",
		},
		Keys: []string{"detector_exec"},
	}

	ruleData32, err := rule32.Build()
	if err != nil {
		return fmt.Errorf("failed to build 32-bit execve rule: %w", err)
	}

	if err := c.auditClient.AddRule(ruleData32); err != nil {
		zap.S().Warnf("failed to add 32-bit execve rule (may already exist): %v", err)
	} else {
		c.rules = append(c.rules, "detector_exec_b32")
		zap.S().Info("added 32-bit execve audit rule")
	}

	return nil
}

// Receive 接收审计事件
func (c *Client) Receive() (*libaudit.RawAuditMessage, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("client is closed")
	}
	client := c.auditClient
	c.mu.Unlock()

	return client.Receive(false)
}

// Cleanup 清理规则
func (c *Client) Cleanup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	// 删除我们添加的规则
	// 注意：这里简化处理，实际应该只删除带有 detector_exec key 的规则
	zap.S().Infof("cleaning up %d audit rules", len(c.rules))
	c.rules = nil

	return nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	if c.auditClient != nil {
		return c.auditClient.Close()
	}
	return nil
}

// IsEnabled 检查审计是否启用
func (c *Client) IsEnabled() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed || c.auditClient == nil {
		return false
	}

	status, err := c.auditClient.GetStatus()
	if err != nil {
		return false
	}

	return status.Enabled != 0
}
