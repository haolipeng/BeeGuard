package analysis

import "encoding/json"

// AlertContext 告警上下文（从视图读取）
type AlertContext struct {
	AlertType string          `json:"alert_type"`
	ID        int64           `json:"id"`
	AgentID   string          `json:"agent_id"`
	HostID    *int64          `json:"host_id,omitempty"`
	HostName  string          `json:"host_name"`
	HostIP    string          `json:"host_ip"`
	Status    int16           `json:"status"`
	AlertTime string          `json:"alert_time"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	Details   json.RawMessage `json:"details"`
}

// AnalysisResult AI分析结果
type AnalysisResult struct {
	RiskLevel       string              `json:"risk_level"`
	AttackPattern   string              `json:"attack_pattern"`
	AttackStage     string              `json:"attack_stage"`
	Summary         string              `json:"summary"`
	Recommendations []string            `json:"recommendations"`
	IOCIndicators   map[string][]string `json:"ioc_indicators"`
}

// AnalysisReport 分析报告（有意义时存储）
type AnalysisReport struct {
	AnalysisType    string              `json:"analysis_type"`    // host / source_ip / time_range
	ScopeKey        string              `json:"scope_key"`        // 主机IP / 攻击源IP / 时间范围
	AlertCount      int                 `json:"alert_count"`
	AlertSnapshot   []AlertContext      `json:"alert_snapshot"`
	RiskLevel       string              `json:"risk_level"`
	AttackPattern   string              `json:"attack_pattern"`
	AttackStage     string              `json:"attack_stage"`
	Summary         string              `json:"summary"`
	Recommendations []string            `json:"recommendations"`
	IOCIndicators   map[string][]string `json:"ioc_indicators"`
}

// HostAlertGroup 按主机聚合的告警组
type HostAlertGroup struct {
	HostIP         string          `json:"host_ip"`
	HostName       string          `json:"host_name"`
	AlertCount     int             `json:"alert_count"`
	Alerts         []AlertContext  `json:"alerts"`
	FirstAlertTime string          `json:"first_alert_time"`
	LastAlertTime  string          `json:"last_alert_time"`
}

// OllamaRequest Ollama API请求
type OllamaRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	Stream   bool   `json:"stream"`
	Options  struct {
		Temperature float64 `json:"temperature,omitempty"`
		NumPredict  int     `json:"num_predict,omitempty"`
	} `json:"options,omitempty"`
}

// OllamaResponse Ollama API响应
type OllamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	// 最终响应时的统计信息
	TotalDuration      int64 `json:"total_duration,omitempty"`
	LoadDuration       int64 `json:"load_duration,omitempty"`
	PromptEvalCount    int   `json:"prompt_eval_count,omitempty"`
	EvalCount          int   `json:"eval_count,omitempty"`
}
