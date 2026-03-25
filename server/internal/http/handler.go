package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/grpc/handler"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/proto"

	"github.com/gin-gonic/gin"
)

// TaskRequest 任务下发请求
type TaskRequest struct {
	AgentID    string `json:"agent_id"`    // 目标 Agent ID（必填）
	ObjectName string `json:"object_name"` // 目标对象（插件名或 "cloudsec-agent"）
	DataType   int32  `json:"data_type"`   // 任务类型（如 5050=进程采集）
	Data       string `json:"data"`        // 任务参数（JSON）
	Token      string `json:"token"`       // 任务令牌（可选）
}

// TaskResponse 任务下发响应
type TaskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Count   int    `json:"count,omitempty"` // 广播时返回成功发送的数量
}

// AgentResponse Agent 信息响应
type AgentResponse struct {
	AgentID  string    `json:"agent_id"`
	Hostname string    `json:"hostname"`
	IPv4     []string  `json:"ipv4"`
	Version  string    `json:"version"`
	Product  string    `json:"product"`
	LastSeen time.Time `json:"last_seen"`
}

// AgentsListResponse Agent 列表响应
type AgentsListResponse struct {
	Agents []AgentResponse `json:"agents"`
	Total  int             `json:"total"`
}

// SendTask 下发任务给 Agent
// POST /api/task
func (s *Server) SendTask(c *gin.Context) {
	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 验证必填字段
	if req.AgentID == "" {
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Message: "agent_id is required",
		})
		return
	}
	if req.ObjectName == "" {
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Message: "object_name is required",
		})
		return
	}

	// 验证 data_type 是否有效
	if !IsValidTaskDataType(req.DataType) {
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid data_type: %d", req.DataType),
		})
		return
	}

	// 验证 object_name 与 data_type 的匹配
	if IsAgentDataType(req.DataType) {
		if req.ObjectName != ObjectNameAgent {
			c.JSON(http.StatusBadRequest, TaskResponse{
				Success: false,
				Message: fmt.Sprintf("data_type %d requires object_name to be '%s'", req.DataType, ObjectNameAgent),
			})
			return
		}
	}

	// 构建 Command
	cmd := &proto.Command{
		Task: &proto.Task{
			DataType:   req.DataType,
			ObjectName: req.ObjectName,
			Data:       req.Data,
			Token:      req.Token,
		},
	}

	// 单点发送
	result := s.transferServer.SendCommandWithError(req.AgentID, cmd)
	switch result {
	case handler.SendResultSuccess:
		c.JSON(http.StatusOK, TaskResponse{
			Success: true,
			Message: "Task sent to agent",
		})
	case handler.SendResultAgentNotFound:
		c.JSON(http.StatusNotFound, TaskResponse{
			Success: false,
			Message: "Agent not found",
		})
	case handler.SendResultQueueFull:
		c.JSON(http.StatusServiceUnavailable, TaskResponse{
			Success: false,
			Message: "Command queue full",
		})
	}
}

// ListAgents 列出所有在线 Agent
// GET /api/agents
func (s *Server) ListAgents(c *gin.Context) {
	agents := s.transferServer.GetAgents()

	resp := AgentsListResponse{
		Agents: make([]AgentResponse, 0, len(agents)),
		Total:  len(agents),
	}

	for _, agent := range agents {
		resp.Agents = append(resp.Agents, AgentResponse{
			AgentID:  agent.AgentID,
			Hostname: agent.Hostname,
			IPv4:     agent.IPv4,
			Version:  agent.Version,
			Product:  agent.Product,
			LastSeen: agent.LastSeen,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// GetAgent 获取指定 Agent 信息
// GET /api/agents/:id
func (s *Server) GetAgent(c *gin.Context) {
	agentID := c.Param("id")

	agent, ok := s.transferServer.GetAgent(agentID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Agent not found",
		})
		return
	}

	c.JSON(http.StatusOK, AgentResponse{
		AgentID:  agent.AgentID,
		Hostname: agent.Hostname,
		IPv4:     agent.IPv4,
		Version:  agent.Version,
		Product:  agent.Product,
		LastSeen: agent.LastSeen,
	})
}

// PluginConfigRequest 插件配置请求
type PluginConfigRequest struct {
	AgentID string         `json:"agent_id"` // 目标 Agent ID
	Plugins []PluginConfig `json:"plugins"`  // 插件配置列表
}

// PluginConfig 单个插件配置
type PluginConfig struct {
	Name    string `json:"name"`    // 插件名称
	Version string `json:"version"` // 插件版本
}

// SendPluginConfig 下发插件配置给 Agent
// POST /api/config
func (s *Server) SendPluginConfig(c *gin.Context) {
	var req PluginConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if req.AgentID == "" {
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Message: "agent_id is required",
		})
		return
	}

	// 构建 Config 列表
	configs := make([]*proto.Config, 0, len(req.Plugins))
	for _, p := range req.Plugins {
		configs = append(configs, &proto.Config{
			Name:    p.Name,
			Type:    "binary",
			Version: p.Version,
		})
	}

	// 构建 Command
	cmd := &proto.Command{
		Configs: configs,
	}

	// 发送
	result := s.transferServer.SendCommandWithError(req.AgentID, cmd)
	switch result {
	case handler.SendResultSuccess:
		c.JSON(http.StatusOK, TaskResponse{
			Success: true,
			Message: "Plugin config sent to agent",
		})
	case handler.SendResultAgentNotFound:
		c.JSON(http.StatusNotFound, TaskResponse{
			Success: false,
			Message: "Agent not found",
		})
	case handler.SendResultQueueFull:
		c.JSON(http.StatusServiceUnavailable, TaskResponse{
			Success: false,
			Message: "Command queue full",
		})
	}
}

// DetectorConfigRequest 检测器配置下发请求
type DetectorConfigRequest struct {
	AgentID string                 `json:"agent_id" binding:"required"`
	Service string                 `json:"service" binding:"required"` // ssh/ftp
	Config  map[string]interface{} `json:"config" binding:"required"`
}

// SendDetectorConfig 下发检测器配置
// POST /api/detector/config
func (s *Server) SendDetectorConfig(c *gin.Context) {
	var req DetectorConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 验证 service
	if req.Service != "ssh" && req.Service != "ftp" {
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Message: "Invalid service, must be 'ssh' or 'ftp'",
		})
		return
	}

	// 序列化配置为JSON
	configData, err := json.Marshal(req.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Message: "Failed to serialize config: " + err.Error(),
		})
		return
	}

	// 构建Command
	cmd := &proto.Command{
		Task: &proto.Task{
			DataType:   DataTypeDetectorConfigUpdate, // 6010
			ObjectName: req.Service,                  // "ssh" 或 "ftp"
			Data:       string(configData),
			Token:      fmt.Sprintf("detector-config-%d", time.Now().UnixNano()),
		},
	}

	// 发送
	result := s.transferServer.SendCommandWithError(req.AgentID, cmd)
	switch result {
	case handler.SendResultSuccess:
		c.JSON(http.StatusOK, TaskResponse{
			Success: true,
			Message: "Detector config sent to agent",
		})
	case handler.SendResultAgentNotFound:
		c.JSON(http.StatusNotFound, TaskResponse{
			Success: false,
			Message: "Agent not found",
		})
	case handler.SendResultQueueFull:
		c.JSON(http.StatusServiceUnavailable, TaskResponse{
			Success: false,
			Message: "Command queue full",
		})
	}
}

// UninstallAgentRequest 卸载 Agent 请求
type UninstallAgentRequest struct {
	AgentID string `json:"agent_id" binding:"required"`
}

// UninstallAgent 远程卸载 Agent
// POST /api/agent/uninstall
func (s *Server) UninstallAgent(c *gin.Context) {
	var req UninstallAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warnf("[API] 卸载Agent请求参数无效: %v", err)
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	log.Infof("[API] 收到卸载Agent请求: agent_id=%s", req.AgentID)

	// 构建卸载命令
	cmd := &proto.Command{
		Task: &proto.Task{
			DataType:   DataTypeAgentUninstall, // 1061
			ObjectName: ObjectNameAgent,        // "cloudsec-agent"
		},
	}

	// 发送卸载命令
	result := s.transferServer.SendCommandWithError(req.AgentID, cmd)
	switch result {
	case handler.SendResultSuccess:
		log.Infof("[API] 卸载命令��发成功: agent_id=%s", req.AgentID)
		c.JSON(http.StatusOK, TaskResponse{
			Success: true,
			Message: "Uninstall command sent to agent",
		})
	case handler.SendResultAgentNotFound:
		log.Warnf("[API] 卸载命令下发失败，Agent未找到: agent_id=%s", req.AgentID)
		c.JSON(http.StatusNotFound, TaskResponse{
			Success: false,
			Message: "Agent not found",
		})
	case handler.SendResultQueueFull:
		log.Warnf("[API] 卸载命令下发失败，命令队列已满: agent_id=%s", req.AgentID)
		c.JSON(http.StatusServiceUnavailable, TaskResponse{
			Success: false,
			Message: "Command queue full",
		})
	}
}
