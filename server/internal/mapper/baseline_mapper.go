package mapper

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/baseline"
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// retBaselineInfo agent端上报的基线检查结果（与agent端RetBaselineInfo对齐）
type retBaselineInfo struct {
	BaselineId      string         `json:"baseline_id"`
	BaselineVersion string         `json:"baseline_version"`
	Status          string         `json:"status"`
	Msg             string         `json:"msg"`
	CheckList       []retCheckInfo `json:"check_list"`
}

// retCheckInfo agent端上报的单个检查项结果（与agent端RetCheckInfo对齐）
type retCheckInfo struct {
	CheckId       int    `json:"check_id"`
	Security      string `json:"security"`
	Type          string `json:"type"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Solution      string `json:"solution"`
	TypeCn        string `json:"type_cn"`
	TitleCn       string `json:"title_cn"`
	DescriptionCn string `json:"description_cn"`
	SolutionCn    string `json:"solution_cn"`
	Result        int    `json:"result"`
	Msg           string `json:"msg"`
}

// MapBaselineResult 将agent上报的DataType 8000数据映射为model
// fields["data"] 包含JSON序列化的 RetBaselineInfo
func MapBaselineResult(fields map[string]string, ctx *AgentContext) (*baseline.CheckResult, []*baseline.BaselineCheckDetail) {
	data := fields["data"]
	if data == "" {
		log.Warnf("[BaselineMapper] 基线结果data字段为空")
		return nil, nil
	}

	var info retBaselineInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		log.Errorf("[BaselineMapper] 解析基线结果JSON失败: %v", err)
		return nil, nil
	}

	now := time.Now()
	nowDT := common.DateTime{Time: now}

	// 统计通过/未通过/异常数量
	passed := 0
	failed := 0
	errors := 0
	for _, item := range info.CheckList {
		switch item.Result {
		case 1: // SuccessCode
			passed++
		case 2: // FailCode
			failed++
		default:
			errors++
		}
	}

	// 解析 fields 中的 baseline_id, template_name 和 template_id
	baselineID := fields["baseline_id"]
	templateName := fields["template_name"]
	var templateID int64
	if tidStr := fields["template_id"]; tidStr != "" {
		if tid, err := strconv.ParseInt(tidStr, 10, 64); err == nil {
			templateID = tid
		}
	}

	result := &baseline.CheckResult{
		BaselineID:  baselineID,
		AgentID:     ctx.AgentID,
		TemplateID:  templateID,
		HostIP:      firstIP(ctx.HostIP),
		HostName:    ctx.HostName,
		TotalItems:  len(info.CheckList),
		PassedItems: passed,
		FailedItems: failed,
		ErrorItems:  errors,
		CheckTime:   nowDT,
	}

	// 构建明细列表
	details := make([]*baseline.BaselineCheckDetail, 0, len(info.CheckList))
	for _, item := range info.CheckList {
		// result映射: 1->status=1(通过), 2->status=0(未通过), 其他->status=2(异常)
		var status int16
		switch item.Result {
		case 1:
			status = 1 // 通过
		case 2:
			status = 0 // 未通过
		default:
			status = 2 // 异常
		}

		actualValue := item.Msg
		errorMessage := item.Msg
		hostIP := firstIP(ctx.HostIP)
		hostName := ctx.HostName
		itemName := item.TitleCn
		riskLevel := item.Security

		detail := &baseline.BaselineCheckDetail{
			ItemID:       int64(item.CheckId),
			BaselineID:   &baselineID,
			AgentID:      ctx.AgentID,
			ItemName:     &itemName,
			HostIP:       &hostIP,
			HostName:     &hostName,
			TemplateName: &templateName,
			TemplateID:   int32(templateID),
			RiskLevel:    &riskLevel,
			Status:       status,
			ActualValue:  &actualValue,
			ErrorMessage: &errorMessage,
			CheckTime:    nowDT,
		}
		details = append(details, detail)
	}

	return result, details
}