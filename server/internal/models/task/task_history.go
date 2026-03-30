package task

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// TaskHistory Agent 任务历史记录
type TaskHistory struct {
	ID            int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID        string          `json:"task_id" gorm:"column:task_id;type:varchar(64);not null;uniqueIndex"`
	AgentID       string          `json:"agent_id" gorm:"column:agent_id;type:varchar(64);not null;index:idx_task_history_agent_id"`
	HostName      string          `json:"host_name,omitempty" gorm:"column:host_name;type:varchar(128)"`
	HostIP        string          `json:"host_ip,omitempty" gorm:"column:host_ip;type:varchar(256)"`
	TaskType      int32           `json:"task_type" gorm:"column:task_type;not null;index:idx_task_history_task_type"`
	TaskName      string          `json:"task_name" gorm:"column:task_name;type:varchar(128);not null"`
	Parameters    JSONMap         `json:"parameters,omitempty" gorm:"column:parameters;type:jsonb"`
	Status        int16           `json:"status" gorm:"column:status;not null;default:0;index:idx_task_history_status"`
	ResultMessage string          `json:"result_message,omitempty" gorm:"column:result_message;type:text"`
	CreatedAt     common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime;index:idx_task_history_created"`
	UpdatedAt     common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (TaskHistory) TableName() string {
	return "agent_task_history"
}

// JSONMap JSONB 字段类型
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONMap: not []byte")
	}
	return json.Unmarshal(bytes, j)
}

// TaskStatus 任务状态枚举
const (
	TaskStatusSent       = 0 // 已下发
	TaskStatusRunning    = 1 // 执行中
	TaskStatusSuccess    = 2 // 成功
	TaskStatusFailed     = 3 // 失败
	TaskStatusTimeout    = 4 // 超时
)

// TaskTypeInfo 任务类型信息
type TaskTypeInfo struct {
	TaskType   int32  `json:"task_type"`
	Name       string `json:"name"`
	PluginName string `json:"plugin_name"`
}

// SupportedTaskTypes 支持的任务类型列表
var SupportedTaskTypes = []TaskTypeInfo{
	{TaskType: 6050, Name: "快速扫描", PluginName: "scanner"},
	{TaskType: 6053, Name: "全盘扫描", PluginName: "scanner"},
	{TaskType: 6057, Name: "指定路径扫描", PluginName: "scanner"},
	{TaskType: 6010, Name: "检测器配置更新", PluginName: "detector"},
	{TaskType: 8000, Name: "基线检查", PluginName: "baseline"},
	{TaskType: 5050, Name: "端口采集", PluginName: "collector"},
	{TaskType: 5051, Name: "进程采集", PluginName: "collector"},
	{TaskType: 5052, Name: "账户采集", PluginName: "collector"},
	{TaskType: 5053, Name: "软件采集", PluginName: "collector"},
	{TaskType: 5054, Name: "服务采集", PluginName: "collector"},
	{TaskType: 5055, Name: "容器采集", PluginName: "collector"},
	{TaskType: 5056, Name: "镜像采集", PluginName: "collector"},
	{TaskType: 5057, Name: "数据库采集", PluginName: "collector"},
	{TaskType: 5058, Name: "Web服务采集", PluginName: "collector"},
	{TaskType: 5059, Name: "内核模块采集", PluginName: "collector"},
	{TaskType: 5060, Name: "环境变量采集", PluginName: "collector"},
	{TaskType: 5061, Name: "镜像包采集", PluginName: "collector"},
	{TaskType: 5062, Name: "网络连接采集", PluginName: "collector"},
	{TaskType: 1060, Name: "Agent关闭", PluginName: ""},
	{TaskType: 1061, Name: "Agent卸载", PluginName: ""},
}

// GetTaskTypeName 根据 task_type 获取任务名称
func GetTaskTypeName(taskType int32) string {
	for _, t := range SupportedTaskTypes {
		if t.TaskType == taskType {
			return t.Name
		}
	}
	return "未知任务"
}

// GetPluginName 根据 task_type 获取目标插件名
func GetPluginName(taskType int32) string {
	for _, t := range SupportedTaskTypes {
		if t.TaskType == taskType {
			return t.PluginName
		}
	}
	return ""
}

// IsValidTaskType 检查任务类型是否有效
func IsValidTaskType(taskType int32) bool {
	for _, t := range SupportedTaskTypes {
		if t.TaskType == taskType {
			return true
		}
	}
	return false
}
