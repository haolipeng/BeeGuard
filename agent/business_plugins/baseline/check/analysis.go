package check

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type RetBaselineInfo struct {
	BaselineId      string         `json:"baseline_id" bson:"baseline_id"`
	BaselineVersion string         `json:"baseline_version" bson:"baseline_version"`
	TemplateName    string         `json:"template_name" bson:"template_name"`
	TemplateId      int            `json:"template_id" bson:"template_id"`
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

type TaskData struct {
	BaselineId   string        `json:"baseline_id"`
	CheckIdList  []int         `json:"check_id_list"`
	BaselineInfo *BaselineInfo `json:"baseline_info,omitempty"` // 服务端下发的完整规则
}

var (
	BaselineStatusError   = "error"
	BaselineStatusSuccess = "success"
)

// AnalysisBaseline start baseline task
func AnalysisBaseline(taskData TaskData) (retBaselineInfo RetBaselineInfo, err error) {

	// analysis params
	checkIdList := taskData.CheckIdList
	retBaselineInfo.BaselineId = taskData.BaselineId

	if taskData.BaselineInfo == nil {
		err = fmt.Errorf("baseline_info is required")
		return retBaselineInfo, err
	}
	baselineInfo := *taskData.BaselineInfo

	retBaselineInfo.BaselineVersion = baselineInfo.BaselineVersion
	retBaselineInfo.TemplateName = baselineInfo.TemplateName
	retBaselineInfo.TemplateId = baselineInfo.TemplateId

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

// Analysis parse task data and run baseline check
func Analysis(data string) (retBaselineInfo RetBaselineInfo, err error) {
	var taskData TaskData
	err = json.Unmarshal([]byte(data), &taskData)
	if err != nil {
		retBaselineInfo.Status = BaselineStatusError
		retBaselineInfo.Msg = err.Error()
		return retBaselineInfo, err
	}

	// start analysis
	retBaselineInfo, err = AnalysisBaseline(taskData)
	if err != nil {
		retBaselineInfo.Status = BaselineStatusError
		retBaselineInfo.Msg = err.Error()
	} else {
		retBaselineInfo.Status = BaselineStatusSuccess
	}
	return retBaselineInfo, err
}
