package task

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/haolipeng/BeeGuard/server/internal/db/repository"
	"github.com/haolipeng/BeeGuard/server/internal/grpc/handler"
	taskModel "github.com/haolipeng/BeeGuard/server/internal/models/task"
	"github.com/haolipeng/BeeGuard/server/proto"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler Agent 任务管理控制器
type Handler struct {
	DB             *gorm.DB
	Repo           *repository.TaskRepository
	TransferServer *handler.TransferServer
}

// SendTaskRequest 下发任务请求
type SendTaskRequest struct {
	AgentID    string                 `json:"agent_id"`
	AgentIDs   []string               `json:"agent_ids"`
	TaskType   int32                  `json:"task_type" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
}

// SendTask 下发任务到 Agent
func (h *Handler) SendTask(c *gin.Context) {
	var req SendTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !taskModel.IsValidTaskType(req.TaskType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的任务类型"})
		return
	}

	// 获取目标 agent 列表
	agentIDs := req.AgentIDs
	if req.AgentID != "" && len(agentIDs) == 0 {
		agentIDs = []string{req.AgentID}
	}
	if len(agentIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请指定目标Agent"})
		return
	}

	taskName := taskModel.GetTaskTypeName(req.TaskType)
	pluginName := taskModel.GetPluginName(req.TaskType)

	// 序列化参数
	var paramData string
	if req.Parameters != nil {
		paramBytes, _ := json.Marshal(req.Parameters)
		paramData = string(paramBytes)
	}

	var results []gin.H
	for _, agentID := range agentIDs {
		taskID := uuid.New().String()

		// 获取 Agent 信息
		agentInfo, exists := h.TransferServer.GetAgent(agentID)
		var hostName, hostIP string
		if exists {
			hostName = agentInfo.Hostname
			if len(agentInfo.IPv4) > 0 {
				hostIP = agentInfo.IPv4[0]
			}
		}

		// 构造 gRPC Command
		cmd := &proto.Command{
			Task: &proto.Task{
				DataType:   req.TaskType,
				ObjectName: pluginName,
				Token:      taskID,
				Data:       paramData,
			},
		}

		// 发送到 Agent
		success := h.TransferServer.SendCommand(agentID, cmd)

		// 记录任务历史
		task := &taskModel.TaskHistory{
			TaskID:     taskID,
			AgentID:    agentID,
			HostName:   hostName,
			HostIP:     hostIP,
			TaskType:   req.TaskType,
			TaskName:   taskName,
			Parameters: taskModel.JSONMap(req.Parameters),
			Status:     taskModel.TaskStatusSent,
		}
		if !success {
			task.Status = taskModel.TaskStatusFailed
			task.ResultMessage = "Agent不在线或命令队列已满"
		}
		_ = h.Repo.Create(c.Request.Context(), task)

		results = append(results, gin.H{
			"task_id":  taskID,
			"agent_id": agentID,
			"success":  success,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}

// ListHistory 查询任务历史列表
func (h *Handler) ListHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}

	agentID := c.Query("agent_id")
	taskTypeStr := c.Query("task_type")
	statusStr := c.Query("status")

	var taskType int32
	if taskTypeStr != "" {
		v, _ := strconv.Atoi(taskTypeStr)
		taskType = int32(v)
	}
	var status int16 = -1
	if statusStr != "" {
		v, _ := strconv.Atoi(statusStr)
		status = int16(v)
	}

	tasks, total, err := h.Repo.List(c.Request.Context(), page, limit, agentID, taskType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务历史失败"})
		return
	}

	totalPages := int(total / int64(limit))
	if total%int64(limit) > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tasks,
		"pagination": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_count":  total,
			"per_page":     limit,
			"has_next":     page < totalPages,
			"has_prev":     page > 1,
		},
	})
}

// GetHistory 获取任务详情
func (h *Handler) GetHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	task, err := h.Repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": task})
}

// GetTaskTypes 获取支持的任务类型列表
func (h *Handler) GetTaskTypes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": taskModel.SupportedTaskTypes})
}
