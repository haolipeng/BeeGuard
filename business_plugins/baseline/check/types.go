package check

// RetBaselineInfo 基线检查结果
type RetBaselineInfo struct {
	BaselineId      int            `json:"baseline_id" bson:"baseline_id"`
	BaselineVersion string         `json:"baseline_version" bson:"baseline_version"`
	Status          string         `json:"status" bson:"status"`
	Msg             string         `json:"msg" bson:"msg"`
	CheckList       []RetCheckInfo `json:"check_list" bson:"check_list"`
}

// RetCheckInfo 单个检查项结果
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

// 检查结果状态码
const (
	SuccessCode = 0 // 通过
	FailCode    = 1 // 失败
	ErrorCode   = 2 // 错误
)

// 基线状态
const (
	BaselineStatusError   = "error"
	BaselineStatusSuccess = "success"
)
