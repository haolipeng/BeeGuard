package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/db/repository"
	"github.com/haolipeng/BeeGuard/server/internal/grpc/handler"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/proto"

	"github.com/gin-gonic/gin"
)

// BaselineCheckRequest 基线检测任务请求
type BaselineCheckRequest struct {
	AgentIDs    []string `json:"agent_ids" binding:"required"` // 目标 agent 列表
	BaselineID  string   `json:"baseline_id"`                  // 检测批次ID（前端task_id）
	CheckIDList []int    `json:"check_id_list"`                // 可选，检查项过滤
	TemplateID  int64    `json:"template_id"`                  // 服务端模板 ID（必填）
}

// BaselineCheckResponse 基线检测任务下发响应
type BaselineCheckResponse struct {
	Success int               `json:"success"` // 成功数
	Failed  int               `json:"failed"`  // 失败数
	Results []AgentSendResult `json:"results"` // 每个 agent 的结果
}

// AgentSendResult 单个 agent 的发送结果
type AgentSendResult struct {
	AgentID string `json:"agent_id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// baselineTaskData 下发给 agent 的任务数据（与 agent 端 TaskData 对齐）
type baselineTaskData struct {
	BaselineId   string                `json:"baseline_id"`
	CheckIdList  []int                 `json:"check_id_list,omitempty"`
	BaselineInfo *baselineInfoForAgent `json:"baseline_info,omitempty"`
}

// baselineInfoForAgent 服务端构建的完整基线规则（与 agent 端 BaselineInfo 对齐）
type baselineInfoForAgent struct {
	BaselineVersion string              `json:"baseline_version"`
	TemplateName    string              `json:"template_name"`
	TemplateId      int                 `json:"template_id"`
	CheckList       []checkInfoForAgent `json:"check_list"`
}

// checkInfoForAgent 单个检查项（与 agent 端 CheckInfo 对齐）
type checkInfoForAgent struct {
	CheckId       int             `json:"check_id"`
	TitleCn       string          `json:"title_cn"`
	Security      string          `json:"security"`
	TypeCn        string          `json:"type_cn"`
	DescriptionCn string          `json:"description_cn"`
	SolutionCn    string          `json:"solution_cn"`
	Check         json.RawMessage `json:"check"` // BaselineCheck JSON
}

// SendBaselineCheck 下发基线检测任务
// POST /api/baseline/check
func (s *Server) SendBaselineCheck(c *gin.Context) {
	var req BaselineCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// 参数校验: template_id 必填
	if req.TemplateID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "template_id is required",
		})
		return
	}

	// 构建 Task.Data
	var taskData baselineTaskData

	// 从数据库加载自定义模板和检查项
	baselineRepo := repository.NewBaselineRepository()
	ctx := context.Background()

	template, err := baselineRepo.GetTemplate(ctx, req.TemplateID)
	if err != nil || template == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("Template not found: %d", req.TemplateID),
		})
		return
	}

	items, err := baselineRepo.ListCheckItemsByTemplateID(ctx, req.TemplateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to load check items: " + err.Error(),
		})
		return
	}

	// 构建完整的 BaselineInfo
	checkList := make([]checkInfoForAgent, 0, len(items))
	for _, item := range items {
		ci := checkInfoForAgent{
			CheckId:       int(item.ID),
			TitleCn:       item.ItemName,
			Security:      item.RiskLevel,
			TypeCn:        item.Category,
			DescriptionCn: item.FixSuggestion,
			SolutionCn:    item.FixSuggestion,
			Check:         json.RawMessage(item.CheckRules),
		}
		checkList = append(checkList, ci)
	}

	version := ""
	if template.Version != nil {
		version = *template.Version
	}
	templateName := ""
	if template.TemplateName != "" {
		templateName = template.TemplateName
	}
	taskData = baselineTaskData{
		BaselineId: req.BaselineID,
		BaselineInfo: &baselineInfoForAgent{
			BaselineVersion: version,
			TemplateName:    templateName,
			TemplateId:      int(req.TemplateID),
			CheckList:       checkList,
		},
	}

	// 序列化任务数据
	taskDataJSON, err := json.Marshal(taskData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to serialize task data: " + err.Error(),
		})
		return
	}

	// 构建 Command
	cmd := &proto.Command{
		Task: &proto.Task{
			DataType:   DataTypeBaselineCheck,
			ObjectName: ObjectNameBaseline,
			Data:       string(taskDataJSON),
			Token:      fmt.Sprintf("baseline-%d", time.Now().UnixNano()),
		},
	}

	// 遍历 agent_ids 逐个发送
	resp := BaselineCheckResponse{
		Results: make([]AgentSendResult, 0, len(req.AgentIDs)),
	}

	for _, agentID := range req.AgentIDs {
		result := s.transferServer.SendCommandWithError(agentID, cmd)
		var agentResult AgentSendResult
		agentResult.AgentID = agentID

		switch result {
		case handler.SendResultSuccess:
			agentResult.Success = true
			agentResult.Message = "Task sent"
			resp.Success++
		case handler.SendResultAgentNotFound:
			agentResult.Success = false
			agentResult.Message = "Agent not found"
			resp.Failed++
		case handler.SendResultQueueFull:
			agentResult.Success = false
			agentResult.Message = "Command queue full"
			resp.Failed++
		}

		resp.Results = append(resp.Results, agentResult)
	}

	log.Infof("[API] 基线检测任务下发完成: success=%d failed=%d agents=%v",
		resp.Success, resp.Failed, req.AgentIDs)

	c.JSON(http.StatusOK, resp)
}
