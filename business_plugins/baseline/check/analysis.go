package check

import (
	"baseline/infra"
	"encoding/json"
	"fmt"
	"strconv"
)

type RetBaselineInfo struct {
	BaselineId      int            `json:"baseline_id" bson:"baseline_id"`
	BaselineVersion string         `json:"baseline_version" bson:"baseline_version"`
	Status          string         `json:"status" bson:"status"`
	Msg             string         `json:"msg" bson:"msg"`
	CheckList       []RetCheckInfo `json:"check_list" bson:"check_list"`
}

type RetCheckInfo struct {
	CheckId       int    `json:"check_id" bson:"check_id"`
	Security      string `json:"security" bson:"security"`
	Type          string `json:"type" bson:"type"`
	Title         string `json:"title" bson:"title"`
	Description   string `json:"description" bson:"description"`
	Solution      string `json:"solution" bson:"solution"`
	TypeCn        string `json:"type_cn" bson:"type_cn"`
	TitleCn       string `json:"title_cn" bson:"title_cn"`
	DescriptionCn string `json:"description_cn" bson:"description_cn"`
	SolutionCn    string `json:"solution_cn" bson:"solution_cn"`
	Result        int    `json:"result" bson:"result"`
	Msg           string `json:"msg" bson:"msg"`
}

// TaskData 任务数据结构
type TaskData struct {
	BaselineId  int   `json:"baseline_id"`
	CheckIdList []int `json:"check_id_list"`
}

// 基线状态
const (
	BaselineStatusError   = "error"
	BaselineStatusSuccess = "success"
)

// get baselin config info
func getBaselineConfigData(baselineId int) (baselineInfo BaselineInfo, err error) {

	// bind config file
	var yamlPath string
	if baselineId < 6000 {
		yamlPath = fmt.Sprintf("config/linux/%d.yaml", baselineId)
	} else {
		yamlPath = fmt.Sprintf("config/container/%d.yaml", baselineId)
	}
	err = infra.BindYaml(yamlPath, &baselineInfo)
	if err != nil {
		return
	}
	return
}

func AnalysisBaseline(taskData TaskData) (retBaselineInfo RetBaselineInfo, err error) {
	// analysis params
	baselineId := taskData.BaselineId
	checkIdList := taskData.CheckIdList
	retBaselineInfo.BaselineId = baselineId

	baselineInfo, err := getBaselineConfigData(baselineId)
	if err != nil {
		infra.Loger.Println("getBaselineConfigData error:", err)
		return retBaselineInfo, err
	}

	retBaselineInfo.BaselineVersion = baselineInfo.BaselineVersion

	// get and analysis check rule
	taskCheckIdMap := make(map[int]int)
	for _, checkId := range checkIdList {
		taskCheckIdMap[checkId] = 0
	}
	for _, checkInfo := range baselineInfo.CheckList {
		if len(checkIdList) != 0 {
			if _, ok := taskCheckIdMap[checkInfo.CheckId]; !ok {
				continue
			}
		}
		var retcheckInfo RetCheckInfo
		retcheckInfo.CheckId = checkInfo.CheckId
		retcheckInfo.Security = checkInfo.Security
		retcheckInfo.TypeCn = checkInfo.TypeCn
		retcheckInfo.TitleCn = checkInfo.TitleCn
		retcheckInfo.DescriptionCn = checkInfo.DescriptionCn
		retcheckInfo.SolutionCn = checkInfo.SolutionCn
		retcheckInfo.Type = checkInfo.Type
		retcheckInfo.Title = checkInfo.Title
		retcheckInfo.Description = checkInfo.Description
		retcheckInfo.Solution = checkInfo.Solution
		ifPass, err := AnalysisRule(checkInfo.Check)
		if err != nil {
			retcheckInfo.Result = ErrorCode
			errCode, _ := strconv.Atoi(err.Error()[:2])
			switch errCode {
			case ErrorFile:
				retcheckInfo.Result = ErrorFile
				retcheckInfo.Msg = err.Error()[3:]
			case ErrorConfigWrite:
				retcheckInfo.Result = ErrorConfigWrite
				retcheckInfo.Msg = err.Error()[3:]
			default:
				retcheckInfo.Result = ErrorCode
				retcheckInfo.Msg = err.Error()
			}
		} else {
			if ifPass {
				retcheckInfo.Result = SuccessCode
			} else {
				retcheckInfo.Result = FailCode
			}
		}
		retBaselineInfo.CheckList = append(retBaselineInfo.CheckList, retcheckInfo)
	}
	return retBaselineInfo, err
}

// Analysis 分析基线检查任务
// 接收任务数据（string 或 int），返回检查结果
func Analysis(data interface{}) (retBaselineInfo RetBaselineInfo, err error) {
	var taskData TaskData

	// 解析任务数据
	switch v := data.(type) {
	case int:
		// 如果直接是 int，作为 baseline_id
		taskData.BaselineId = v
	case string:
		// 如果是 string，尝试解析参数
		err = json.Unmarshal([]byte(v), &taskData)
		if err != nil {
			retBaselineInfo.Status = BaselineStatusError
			retBaselineInfo.Msg = fmt.Sprintf("parse task data error: %v", err)
			return retBaselineInfo, err
		}
	default:
		// 不支持其他数据类型
		retBaselineInfo.Status = BaselineStatusError
		retBaselineInfo.Msg = fmt.Sprintf("unsupported data type: %T", data)
		return retBaselineInfo, fmt.Errorf("unsupported data type: %T", data)
	}

	// 调用基线分析的函数
	retBaselineInfo, err = AnalysisBaseline(taskData)
	if err != nil {
		retBaselineInfo.Status = BaselineStatusError
		retBaselineInfo.Msg = err.Error()
	} else {
		retBaselineInfo.Status = BaselineStatusSuccess
	}

	return retBaselineInfo, nil
}
