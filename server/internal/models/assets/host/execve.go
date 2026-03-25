package host

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Execve execve事件记录表
type Execve struct {
	ID       int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID  string `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostName string `json:"host_name" gorm:"column:host_name"`
	HostIP   string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`

	// execve事件字段
	PID  int `json:"pid" gorm:"column:pid;not null"` // 进程ID（线程ID）
	TGID int `json:"tgid" gorm:"column:tgid"`        // 线程组ID（进程ID）
	PPID int `json:"ppid" gorm:"column:ppid"`         // 父进程ID
	PGID int `json:"pgid" gorm:"column:pgid"`         // 进程组ID
	UID  int `json:"uid" gorm:"column:uid"`           // 用户ID

	Comm    string `json:"comm" gorm:"column:comm"`         // 进程名（最多16字节）
	ExePath string `json:"exe_path" gorm:"column:exe_path"` // 可执行文件完整路径
	Args    string `json:"args" gorm:"column:args"`         // 命令行参数

	EventTime common.DateTime `json:"event_time" gorm:"column:event_time;not null;index"` // 事件发生时间
	CreatedAt common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (Execve) TableName() string {
	return "event_execve"
}
