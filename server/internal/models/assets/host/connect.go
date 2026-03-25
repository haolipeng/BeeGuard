package host

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Connect 出站连接事件记录表
type Connect struct {
	ID       int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID  string `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostName string `json:"host_name" gorm:"column:host_name"`
	HostIP   string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`

	// 进程信息
	PID  int `json:"pid" gorm:"column:pid;not null"` // 进程ID（线程ID）
	TGID int `json:"tgid" gorm:"column:tgid"`        // 线程组ID（进程ID）
	PPID int `json:"ppid" gorm:"column:ppid"`        // 父进程ID
	UID  int `json:"uid" gorm:"column:uid"`          // 用户ID

	Comm    string `json:"comm" gorm:"column:comm"`         // 进程名（最多16字节）
	ExePath string `json:"exe_path" gorm:"column:exe_path"` // 可执行文件完整路径

	// 连接信息
	Protocol   string `json:"protocol" gorm:"column:protocol"`       // 协议类型（tcp/udp）
	RemoteIP   string `json:"remote_ip" gorm:"column:remote_ip"`     // 远端IP地址
	RemotePort int    `json:"remote_port" gorm:"column:remote_port"` // 远端端口

	PidTree string `json:"pid_tree" gorm:"column:pid_tree"` // 进程树（预留）

	EventTime common.DateTime `json:"event_time" gorm:"column:event_time;not null;index"` // 事件发生时间
	CreatedAt common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (Connect) TableName() string {
	return "event_connect"
}
