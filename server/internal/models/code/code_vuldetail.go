package code

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// CodeVulDetail 代码漏洞详情实体
type CodeVulDetail struct {
	ID            int32            `json:"id" gorm:"primaryKey;not null;autoIncrement"` // 主键ID
	ScanResultsID *int32           `json:"scan_results_id"`                             // codeql_scan_results关联id
	Path          *string          `json:"path"`                                        // 文件路径
	Code          string           `json:"code" gorm:"not null"`                        // 有漏洞的代码文件
	CreatedAt     common.DateTime  `json:"created_at"`                                  // 创建时间
	UpdatedAt     common.DateTime  `json:"updated_at"`                                  // 更新时间
	DeletedAt     *common.DateTime `json:"deleted_at,omitempty"`                        // 删除时间
}

// TableName 指定表名为 code_vuldetail
func (CodeVulDetail) TableName() string {
	return "code_vuldetail"
}
