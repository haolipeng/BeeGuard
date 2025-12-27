package check

import (
	"encoding/json"
	"fmt"
)

// Analysis 分析基线检查任务
// 接收任务数据（string 或 int），返回检查结果
func Analysis(data interface{}) (retBaselineInfo RetBaselineInfo, err error) {
	var taskData TaskData

	// 解析任务数据
	switch v := data.(type) {
	case int:
		// 如果直接是 int，作为 baseline_id
		taskData.BaselineId = v
		taskData.CheckIdList = []int{} // 空列表表示检查所有项
	case string:
		// 如果是 string，尝试解析 JSON
		err = json.Unmarshal([]byte(v), &taskData)
		if err != nil {
			retBaselineInfo.Status = BaselineStatusError
			retBaselineInfo.Msg = fmt.Sprintf("parse task data error: %v", err)
			return retBaselineInfo, err
		}
	default:
		retBaselineInfo.Status = BaselineStatusError
		retBaselineInfo.Msg = fmt.Sprintf("unsupported data type: %T", data)
		return retBaselineInfo, fmt.Errorf("unsupported data type: %T", data)
	}

	// 设置基线信息
	retBaselineInfo.BaselineId = taskData.BaselineId
	retBaselineInfo.BaselineVersion = "1.0.0"

	// 模拟检查逻辑 - 生成一些示例检查结果
	if len(taskData.CheckIdList) == 0 {
		// 如果没有指定检查项，生成默认的检查项
		taskData.CheckIdList = []int{1001, 1002, 1003}
	}

	// 为每个检查项生成结果
	for _, checkId := range taskData.CheckIdList {
		retCheckInfo := RetCheckInfo{
			CheckId:       checkId,
			Security:      "medium",
			Type:          "configuration",
			Title:         fmt.Sprintf("Check Item %d", checkId),
			Description:   fmt.Sprintf("Description for check item %d", checkId),
			Solution:      fmt.Sprintf("Solution for check item %d", checkId),
			TypeCn:        "配置检查",
			TitleCn:       fmt.Sprintf("检查项 %d", checkId),
			DescriptionCn: fmt.Sprintf("检查项 %d 的描述", checkId),
			SolutionCn:    fmt.Sprintf("检查项 %d 的解决方案", checkId),
			Result:        SuccessCode, // 模拟通过
			Msg:           "",
		}

		// 模拟一些检查项失败
		if checkId%3 == 0 {
			retCheckInfo.Result = FailCode
			retCheckInfo.Msg = "Check failed: configuration not compliant"
		}

		retBaselineInfo.CheckList = append(retBaselineInfo.CheckList, retCheckInfo)
	}

	// 设置整体状态
	retBaselineInfo.Status = BaselineStatusSuccess
	retBaselineInfo.Msg = "Analysis completed successfully"

	return retBaselineInfo, nil
}
