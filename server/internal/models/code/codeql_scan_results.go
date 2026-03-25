package code

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// CodeqlScanResults 代码审计仓库结果列表实体
type CodeqlScanResults struct {
	ID          int64            `json:"id" gorm:"primaryKey;not null;autoIncrement"` // 主键ID
	RepoID      int64            `json:"repo_id" gorm:"not null"`
	RepoName    string           `json:"repo_name" gorm:"not null"` // 关联仓库名称
	RuleID      *int64           `json:"rule_id"`                  // 规则索引
	RuleName    *string          `json:"rule_name"`                  // 规则名称
	Severity    string           `json:"severity"`                   // 严重性
	Confidence  string           `json:"confidence"`                 // 置信度
	FilePath    *string          `json:"file_path"`                  // 文件地址
	StartLine   *int64           `json:"start_line"`                 // 开始行
	EndLine     *int64           `json:"end_line"`                   // 结束行
	CodeSnippet *string          `json:"code_snippet"`               // 代码片段
	Message     *string          `json:"message"`                    // 消息
	CodeFlows   *string          `json:"code_flows"`                 // 代码流
	Related     *string          `json:"related"`                    // 相关信息
	Remediation *string          `json:"remediation"`                // 修复建议
	Language    *string          `json:"language"`                   // 编程语言
	ScanTime    *common.DateTime `json:"scan_time,omitempty"`        // 扫描时间
	Status      string           `json:"status" gorm:"not null"`     // 状态
	FixedTime   *common.DateTime `json:"fixed_time,omitempty"`       // 修复时间
	ProjectName *string          `json:"project_name"`               // 项目名称
	Branch      *string          `json:"branch"`                     // 分支
	CommitID    *string          `json:"commit_id"`                  // 提交ID
	CreatedAt   common.DateTime  `json:"created_at"`                 // 创建时间
	UpdatedAt   common.DateTime  `json:"updated_at"`                 // 更新时间
	DeletedAt   *common.DateTime `json:"deleted_at,omitempty"`       // 删除时间
	ScanType    string           `json:"scan_type" gorm:"not null"`  // 扫描类型
	StartedAt   *common.DateTime `json:"started_at,omitempty"`       // 开始时间
	FinishedAt  *common.DateTime `json:"finished_at,omitempty"`      // 完成时间
	Result      *string          `json:"result"`                     // 结果
	ErrorMsg    *string          `json:"error_msg"`                  // 错误信息
}

// TableName 指定表名为 repos_scan_result
//
//	func (ReposScanResult) TableName() string {
//		return "repos_scan_result"
//	}
func (CodeqlScanResults) TableName() string {
	return "codeql_scan_results"
}
