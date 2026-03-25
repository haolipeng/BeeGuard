package code

import (
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/models/common"

	"gorm.io/gorm"
)

// Repos 仓库实体
type Repos struct {
	RepoID               int64           `json:"repo_id" gorm:"primaryKey"`
	RepoName             string          `json:"repo_name" gorm:"not null"`       // 仓库名称
	RepoURL              string          `json:"repo_url" gorm:"not null"`        // 仓库地址
	PullMethod           string          `json:"pull_method"`                     // 拉取方式
	IsPrivate            bool            `json:"is_private"`                      // 是否私有
	CodeqlRules          *string         `json:"codeql_rules"`
	Description          string          `json:"description"`           // 描述
	CodeHash             string          `json:"code_hash"`             // 代码哈希
	Owner                string          `json:"owner"`                 // 所有者
	Branch               string          `json:"branch"`                // 分支
	LocalPath            string          `json:"local_path"`            // 本地路径
	Language             string          `json:"language"`              // 编程语言
	ScanFrequency        string          `json:"scan_frequency"`        // 扫描频率
	TotalVulnerabilities int64           `json:"total_vulnerabilities"` // 漏洞总数
	CriticalCount        int64           `json:"critical_count"`        // 严重漏洞数量
	HighCount            int64           `json:"high_count"`            // 高危漏洞数量
	MediumCount          int64           `json:"medium_count"`          // 中危漏洞数量
	LowCount             int64           `json:"low_count"`             // 低危漏洞数量
	ScanStartTime        *time.Time      `json:"scan_start_time"`       // 扫描开始时间
	ScanEndTime          *time.Time      `json:"scan_end_time"`         // 扫描结束时间
	LastScanTime         *time.Time      `json:"last_scan_time"`        // 最后扫描时间
	Status               string          `json:"status"`                // 状态
	Deleted              bool            `json:"deleted"`
	CreatedAt            common.DateTime `json:"created_at"`
	UpdatedAt            common.DateTime `json:"updated_at"`
	DeletedAt            gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}
