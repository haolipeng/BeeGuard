package alert

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// BruteForce 暴力破解告警实体
type BruteForce struct {
	ID              int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID         string           `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostID          *int64           `json:"host_id,omitempty" gorm:"column:host_id"`
	HostName        string           `json:"host_name" gorm:"column:host_name;not null"`
	HostIP          string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	SourceIP        string           `json:"source_ip" gorm:"column:source_ip;not null;index"`
	SourceLocation  *string          `json:"source_location,omitempty" gorm:"column:source_location"`
	AttackType      string           `json:"attack_type" gorm:"column:attack_type;not null"`
	TargetIP        string           `json:"target_ip" gorm:"column:target_ip;not null"`
	TargetPort      *int32           `json:"target_port,omitempty" gorm:"column:target_port"`
	Username        string           `json:"username" gorm:"column:username;not null"`
	AttemptCount    int32            `json:"attempt_count" gorm:"column:attempt_count;not null"`
	Result          string           `json:"result" gorm:"column:result;not null"`
	AttackTime      common.DateTime  `json:"attack_time" gorm:"column:attack_time;not null"`
	FirstAttackTime *common.DateTime `json:"first_attack_time,omitempty" gorm:"column:first_attack_time"`
	Status          int16            `json:"status" gorm:"column:status;not null;default:0"`
	WhitelistHit    bool             `json:"whitelist_hit" gorm:"column:whitelist_hit;default:false"`
	WhitelistRuleID *int64           `json:"whitelist_rule_id,omitempty" gorm:"column:whitelist_rule_id"`
	IsBlocked       *int16           `json:"is_blocked,omitempty" gorm:"column:is_blocked;default:0"`
	ProcessTime     *common.DateTime `json:"process_time,omitempty" gorm:"column:process_time"`
	Processor       *string          `json:"processor,omitempty" gorm:"column:processor"`
	Remark          *string          `json:"remark,omitempty" gorm:"column:remark"`
	CreatedAt       common.DateTime  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       common.DateTime  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 alert_brute_force
func (BruteForce) TableName() string {
	return "alert_brute_force"
}

// AttackType 攻击类型枚举常量
const (
	AttackTypeSSH    = "ssh"    // SSH暴力破解
	AttackTypeFTP    = "ftp"    // FTP暴力破解
	AttackTypeRDP    = "rdp"    // RDP暴力破解
	AttackTypeMySQL  = "mysql"  // MySQL暴力破解
	AttackTypeRedis  = "redis"  // Redis暴力破解
	AttackTypeWeb    = "web"    // Web登录暴力破解
	AttackTypeSMB    = "smb"    // SMB暴力破解
	AttackTypeTelnet = "telnet" // Telnet暴力破解
)

// BruteForceStatus 状态枚举常量
const (
	BruteForceStatusPending   = 0 // 待处理
	BruteForceStatusProcessed = 1 // 已处理
	BruteForceStatusIgnored   = 2 // 已忽略
)

// BruteForceResult 攻击结果枚举常量
const (
	BruteForceResultFailed  = "failed"  // 暴力破解未成功
	BruteForceResultSuccess = "success" // 暴力破解后成功登录
)

// IsBlocked 是否封禁枚举常量
const (
	NotBlocked = 0 // 未封禁
	Blocked    = 1 // 已封禁
)
