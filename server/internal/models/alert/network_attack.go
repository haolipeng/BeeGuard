package alert

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// NetworkAttack 网络攻击告警实体
type NetworkAttack struct {
	ID                int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID           string           `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostID            *int64           `json:"host_id,omitempty" gorm:"column:host_id"`
	HostName          string           `json:"host_name" gorm:"column:host_name;not null"`
	HostIP            string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	TargetPort        int32            `json:"target_port" gorm:"column:target_port;not null"`
	AttackerIP        string           `json:"attacker_ip" gorm:"column:attacker_ip;not null;index"`
	AttackerLocation  *string          `json:"attacker_location,omitempty" gorm:"column:attacker_location"`
	AttackerCountry   *string          `json:"attacker_country,omitempty" gorm:"column:attacker_country"`
	VulnerabilityName string           `json:"vulnerability_name" gorm:"column:vulnerability_name;not null"`
	VulnerabilityID   *string          `json:"vulnerability_id,omitempty" gorm:"column:vulnerability_id;index"`
	AttackStatus      string           `json:"attack_status" gorm:"column:attack_status;not null"`
	AttackCount       int32            `json:"attack_count" gorm:"column:attack_count;not null"`
	FirstAttackTime   *common.DateTime `json:"first_attack_time,omitempty" gorm:"column:first_attack_time"`
	LastAttackTime    common.DateTime  `json:"last_attack_time" gorm:"column:last_attack_time;not null"`
	AttackPayload     *string          `json:"attack_payload,omitempty" gorm:"column:attack_payload"`
	Status            int16            `json:"status" gorm:"column:status;not null;default:0"`
	WhitelistHit      bool             `json:"whitelist_hit" gorm:"column:whitelist_hit;default:false"`
	WhitelistRuleID   *int64           `json:"whitelist_rule_id,omitempty" gorm:"column:whitelist_rule_id"`
	CreatedAt         common.DateTime  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         common.DateTime  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 alert_network_attack
func (NetworkAttack) TableName() string {
	return "alert_network_attack"
}

// AttackStatus 攻击状态枚举常量
const (
	AttackStatusDetected  = "detected"  // 已检测到网络攻击
	AttackStatusMitigated = "mitigated" // 已缓解网络攻击
)

// NetworkAttackStatus 状态枚举常量
const (
	NetworkAttackStatusPending   = 0 // 待处理
	NetworkAttackStatusProcessed = 1 // 已处理
	NetworkAttackStatusIgnored   = 2 // 已忽略
)

// CommonVulnerabilities 常见漏洞类型枚举常量
const (
	VulnTypeSQLInjection   = "sql_injection"   // SQL注入
	VulnTypeXSS            = "xss"             // 跨站脚本
	VulnTypeRCE            = "rce"             // 远程代码执行
	VulnTypeBufferOverflow = "buffer_overflow" // 缓冲区溢出
	VulnTypeDoS            = "dos"             // 拒绝服务
	VulnTypeDDoS           = "ddos"            // 分布式拒绝服务
	VulnTypeBruteForce     = "brute_force"     // 暴力破解
)
