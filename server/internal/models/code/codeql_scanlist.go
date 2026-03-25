package code

import (
	"time"

	"gorm.io/gorm"
)

// RepoScanList 仓库扫描列表模型
type RepoScanList struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	RepoID      int64          `json:"repo_id" gorm:"not null"`                                                       // 仓库ID
	RepoName    string         `json:"repo_name" gorm:"not null;size:255"`                                            // 仓库名称
	RuleID      *int64         `json:"rule_id,omitempty" gorm:"index"`                                                // 规则ID
	RuleName    *string        `json:"rule_name,omitempty" gorm:"type:text"`                                          // 规则名称
	Severity    string         `json:"severity" gorm:"size:255"`                                                      // 漏洞严重级别
	Confidence  string         `json:"confidence" gorm:"not null;type:enum('HIGH','MEDIUM','LOW')"`                   // 漏洞置信度
	FilePath    *string        `json:"file_path,omitempty" gorm:"size:500"`                                           // 文件路径
	StartLine   *int64         `json:"start_line,omitempty"`                                                          // 开始行号
	EndLine     *int64         `json:"end_line,omitempty"`                                                            // 结束行号
	StartColumn *int64         `json:"start_column,omitempty" gorm:"size:255"`                                        // 开始列
	EndColumn   *int64         `json:"end_column,omitempty" gorm:"size:255"`                                          // 结束列
	CodeSnippet *string        `json:"code_snippet,omitempty" gorm:"type:text"`                                       // 问题代码片段
	Message     *string        `json:"message,omitempty" gorm:"type:text"`                                            // 消息描述
	Remediation *string        `json:"remediation,omitempty" gorm:"type:text"`                                        // 修复建议
	Language    *string        `json:"language,omitempty" gorm:"size:100"`                                            // 编程语言
	CodeFlows   *string        `json:"code_flows,omitempty" gorm:"type:text"`                                         // 代码流
	Related     *string        `json:"related,omitempty" gorm:"type:text"`                                            // 相关信息
	ScanTime    *time.Time     `json:"scan_time,omitempty" gorm:"type:datetime(3)"`                                   // 扫描时间
	Status      string         `json:"status" gorm:"type:enum('UNFIXED','FIXED','FALSE_POSITIVE');default:'UNFIXED'"` // 漏洞状态
	FixedTime   *time.Time     `json:"fixed_time,omitempty"`                                                          // 修复时间
	ProjectName *string        `json:"project_name,omitempty" gorm:"size:255"`                                        // 项目名称
	Branch      *string        `json:"branch,omitempty" gorm:"size:64"`                                               // 代码分支
	CommitID    *string        `json:"commit_id,omitempty" gorm:"size:64"`                                            // 提交ID
	ScanType    string         `json:"scan_type" gorm:"not null;size:50"`                                             // 扫描类型
	StartedAt   *time.Time     `json:"started_at,omitempty"`                                                          // 开始时间
	FinishedAt  *time.Time     `json:"finished_at,omitempty"`                                                         // 完成时间
	Result      *string        `json:"result,omitempty" gorm:"type:text"`                                             // 扫描结果
	ErrorMsg    *string        `json:"error_msg,omitempty" gorm:"type:text"`                                          // 错误消息
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"` // 软删除时间
}

// TableName 指定表名
func (RepoScanList) TableName() string {
	return "codeql_scan_results"
}
