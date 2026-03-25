package analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/log"
)

// OllamaClient Ollama API客户端
type OllamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
	timeout    time.Duration
}

// OllamaConfig Ollama配置
type OllamaConfig struct {
	BaseURL string        // Ollama服务地址，默认 http://localhost:11434
	Model   string        // 模型名称，默认 qwen2.5:0.5b
	Timeout time.Duration // 请求超时，默认 60秒
}

// NewOllamaClient 创建Ollama客户端
func NewOllamaClient(cfg OllamaConfig) *OllamaClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:11434"
	}
	if cfg.Model == "" {
		cfg.Model = "qwen2.5:0.5b"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 300 * time.Second
	}

	return &OllamaClient{
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		timeout: cfg.Timeout,
	}
}

// Analyze 分析告警，带重试机制
func (c *OllamaClient) Analyze(ctx context.Context, alerts []AlertContext) (*AnalysisResult, error) {
	var prompt string
	if allFileIntegrity(alerts) {
		prompt = c.buildFileIntegrityPrompt(alerts)
	} else {
		prompt = c.buildPrompt(alerts)
	}
	alertsJSON, _ := json.MarshalIndent(alerts, "", "  ")

	var lastErr error
	var lastResponse string

	// 最多重试5次
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			log.Warnf("[Ollama] 第%d次重试分析...", attempt)
			// 构建重试提示词，告诉AI上次解析失败
			prompt = c.buildRetryPrompt(string(alertsJSON), lastResponse, lastErr)
		}

		response, err := c.generate(ctx, prompt)
		if err != nil {
			lastErr = err
			continue // 请求失败，重试
		}

		result, err := c.parseResult(response)
		if err == nil {
			// 解析成功，返回结果
			if attempt > 0 {
				log.Infof("[Ollama] 第%d次重试后解析成功", attempt)
			}
			return result, nil
		}

		// 解析失败，记录错误和响应，继续重试
		lastErr = err
		lastResponse = response
		log.Warnf("[Ollama] 解析AI响应失败 (尝试 %d/5): %v", attempt+1, err)
	}

	// 5次都失败了，直接存储原始结果
	log.Errorf("[Ollama] 5次尝试后仍无法解析AI响应，将存储原始结果")
	return c.createFallbackResult(lastResponse, lastErr)
}

// buildRetryPrompt 构建重试提示词
func (c *OllamaClient) buildRetryPrompt(alertsJSON string, lastResponse string, parseErr error) string {
	return fmt.Sprintf(`你是一个专业的安全分析专家。请分析以下安全告警数据。

## 告警数据
%s

## 上次的分析结果无法解析
上次返回的内容：
%s

解析错误：%v

## 重新分析要求
请重新分析，并确保返回**有效的JSON格式**，不要包含任何其他内容（如markdown标记、解释文字等）。

## 输出格式（必须是有效的JSON）
{
  "risk_level": "low",
  "attack_pattern": "攻击模式描述",
  "attack_stage": "攻击阶段",
  "summary": "分析摘要",
  "recommendations": ["建议1", "建议2"],
  "ioc_indicators": {
    "ips": ["可疑IP"],
    "domains": ["可疑域名"],
    "files": ["可疑文件路径"]
  }
}

只输出JSON，不要输出其他内容。`, alertsJSON, lastResponse, parseErr)
}

// createFallbackResult 创建降级存储的结果
func (c *OllamaClient) createFallbackResult(rawResponse string, parseErr error) (*AnalysisResult, error) {
	// 尝试从原始响应中提取有用信息
	result := c.extractPartialResult(rawResponse, rawResponse)

	// 如果提取失败，使用默认值
	if result.RiskLevel == "" {
		result.RiskLevel = "low"
	}
	if result.AttackPattern == "" {
		result.AttackPattern = "unknown"
	}
	if result.AttackStage == "" {
		result.AttackStage = "unknown"
	}
	if result.Summary == "" {
		result.Summary = fmt.Sprintf("AI响应解析失败: %v", parseErr)
	}
	if result.Recommendations == nil {
		result.Recommendations = []string{"请人工检查原始AI响应"}
	}

	// 将原始响应存储在IOC中，方便后续查看
	if result.IOCIndicators == nil {
		result.IOCIndicators = make(map[string][]string)
	}
	result.IOCIndicators["_raw_response"] = []string{rawResponse}
	result.IOCIndicators["_parse_error"] = []string{parseErr.Error()}

	return result, nil
}

// allFileIntegrity 判断告警是否全部为 file_integrity 类型
func allFileIntegrity(alerts []AlertContext) bool {
	if len(alerts) == 0 {
		return false
	}
	for _, a := range alerts {
		if a.AlertType != "file_integrity" {
			return false
		}
	}
	return true
}

// generate 调用Ollama生成
func (c *OllamaClient) generate(ctx context.Context, prompt string) (string, error) {
	req := OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}
	req.Options.Temperature = 0.3
	req.Options.NumPredict = 4096 // 增加输出token限制

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	url := c.baseURL + "/api/generate"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("请求Ollama失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama返回错误: %s, %s", resp.Status, string(respBody))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	log.Infof("[Ollama] 生成完成, 耗时: %v, tokens: %d", time.Since(start), ollamaResp.EvalCount)

	return ollamaResp.Response, nil
}

// buildPrompt 构建分析提示词
func (c *OllamaClient) buildPrompt(alerts []AlertContext) string {
	alertsJSON, _ := json.MarshalIndent(alerts, "", "  ")

	return fmt.Sprintf(`你是一个专业的安全分析专家。请分析以下安全告警数据，判断是否存在真实的安全威胁。

## 告警数据
%s

## 分析要求
1. 判断整体风险等级（low/medium/high/critical）
2. 识别攻击模式（如：暴力破解、Web攻击、横向移动、数据窃取等）
3. 判断攻击阶段（recon/initial_access/execution/persistence/privilege_escalation/lateral_movement/data_exfiltration）
4. 简要分析摘要（100字以内）
5. 给出处置建议

## 输出格式（必须是有效的JSON）
{
  "risk_level": "low",
  "attack_pattern": "攻击模式描述",
  "attack_stage": "攻击阶段",
  "summary": "分析摘要",
  "recommendations": ["建议1", "建议2"],
  "ioc_indicators": {
    "ips": ["可疑IP"],
    "domains": ["可疑域名"],
    "files": ["可疑文件路径"]
  }
}

注意：
- 如果是正常运维操作或误报，risk_level设为low
- 只输出JSON，不要输出其他内容`, alertsJSON)
}

// buildFileIntegrityPrompt 构建文件完整性监控专用分析提示词
func (c *OllamaClient) buildFileIntegrityPrompt(alerts []AlertContext) string {
	alertsJSON, _ := json.MarshalIndent(alerts, "", "  ")

	return fmt.Sprintf(`你是一个专业的安全分析专家，擅长主机文件完整性监控（File Integrity Monitoring）告警分析。请分析以下文件完整性告警数据，判断是否存在真实的安全威胁。

## 告警数据
%s

## 文件完整性分析指导

### 常见误报/正常运维场景（应标记为 low）：
1. **包管理器操作**：apt/yum/dnf/pip/npm 等安装、更新、卸载时产生的文件变更
2. **配置管理工具**：ansible/puppet/chef/saltstack 推送配置导致的文件修改
3. **日志轮转**：logrotate 产生的日志文件变更
4. **系统更新**：系统补丁安装导致的二进制文件和库文件更新
5. **合法管理操作**：root/管理员用户通过 vi/vim/nano/sed 编辑配置文件
6. **证书更新**：SSL/TLS 证书自动续期导致的文件变更
7. **临时文件**：/tmp、/var/tmp 下的文件变更

### 可疑/高危场景（应标记为 medium/high/critical）：
1. **系统二进制篡改**：/usr/bin、/usr/sbin、/bin、/sbin 下的核心命令被替换（非包管理器操作）
2. **SSH配置修改**：/etc/ssh/sshd_config 被修改、authorized_keys 被新增或篡改
3. **Crontab篡改**：/etc/crontab、/var/spool/cron/ 下文件被非管理员修改
4. **Rootkit特征**：系统工具（ls/ps/netstat/ss）被替换、LD_PRELOAD相关文件变更
5. **后门植入**：在 /usr/local/bin、/opt 等目录新增可执行文件
6. **PAM模块篡改**：/etc/pam.d/ 下文件被修改
7. **启动项修改**：systemd unit文件、/etc/init.d/ 脚本被修改或新增
8. **密码/认证文件**：/etc/passwd、/etc/shadow、/etc/sudoers 被修改
9. **内核模块**：/lib/modules/ 下新增或修改 .ko 文件

### 分析要点：
- 结合 details 中的 rule_type（core_file/config_file/system_file）评估文件重要性
- 结合 details 中的 threat_action（add/modify/delete）评估操作类型
- 关注操作进程（process）和用户（user）是否合法
- 多个文件同时变更可能暗示批量篡改
- 关注文件权限变更（特别是 SUID/SGID 位被设置）

## 分析要求
1. 判断整体风险等级（low/medium/high/critical）
2. 识别变更模式（如：正常运维、包管理器更新、可疑文件篡改、后门植入、Rootkit安装等）
3. 判断攻击阶段（若为恶意操作）：persistence/privilege_escalation/execution/defense_evasion
4. 简要分析摘要（100字以内）
5. 给出处置建议

## 输出格式（必须是有效的JSON）
{
  "risk_level": "low",
  "attack_pattern": "变更模式描述",
  "attack_stage": "攻击阶段或normal_operation",
  "summary": "分析摘要",
  "recommendations": ["建议1", "建议2"],
  "ioc_indicators": {
    "ips": [],
    "domains": [],
    "files": ["被修改的可疑文件路径"]
  }
}

注意：
- 如果是包管理器更新、配置管理工具操作、日志轮转等正常运维操作，risk_level设为low
- 如果无法确定是否为恶意操作，risk_level设为medium并建议人工复核
- 只输出JSON，不要输出其他内容`, alertsJSON)
}

// parseResult 解析AI响应
func (c *OllamaClient) parseResult(response string) (*AnalysisResult, error) {
	// 尝试提取JSON部分
	jsonStr := c.extractJSON(response)

	var result AnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// 解析失败，尝试修复JSON
		fixedJSON := c.fixJSON(jsonStr)
		if err := json.Unmarshal([]byte(fixedJSON), &result); err != nil {
			// 仍然失败，尝试从部分响应中提取信息
			partialResult := c.extractPartialResult(response, jsonStr)
			if partialResult.RiskLevel == "" {
				return nil, fmt.Errorf("解析AI响应失败: %w", err)
			}
			result = *partialResult
		}
	}

	// 验证并规范化风险等级
	result.RiskLevel = c.normalizeRiskLevel(result.RiskLevel)

	return &result, nil
}

// extractJSON 从响应中提取JSON
func (c *OllamaClient) extractJSON(response string) string {
	response = strings.TrimSpace(response)

	// 处理markdown代码块: ```json ... ``` 或 ``` ... ```
	if strings.HasPrefix(response, "```") {
		// 找到第一个换行后的内容
		if idx := strings.Index(response, "\n"); idx != -1 {
			response = response[idx+1:]
		}
		// 移除结尾的 ```
		if idx := strings.LastIndex(response, "```"); idx != -1 {
			response = response[:idx]
		}
		response = strings.TrimSpace(response)
	}

	// 查找第一个 { 和最后一个 }
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start != -1 && end != -1 && end > start {
		return response[start : end+1]
	}

	return response
}

// fixJSON 尝试修复JSON
func (c *OllamaClient) fixJSON(jsonStr string) string {
	// 尝试补全不完整的JSON
	// 统计未闭合的括号
	braceCount := 0
	bracketCount := 0
	inString := false
	escape := false

	for _, ch := range jsonStr {
		if escape {
			escape = false
			continue
		}
		if ch == '\\' {
			escape = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch ch {
		case '{':
			braceCount++
		case '}':
			braceCount--
		case '[':
			bracketCount++
		case ']':
			bracketCount--
		}
	}

	// 补全缺失的闭合符号
	result := jsonStr
	for bracketCount > 0 {
		result += "]"
		bracketCount--
	}
	for braceCount > 0 {
		result += "}"
		braceCount--
	}

	return result
}

// extractPartialResult 从部分响应中提取结果
func (c *OllamaClient) extractPartialResult(response, jsonStr string) *AnalysisResult {
	// 尝试提取关键字段
	result := &AnalysisResult{
		RiskLevel:     "low",
		AttackPattern: "unknown",
		AttackStage:   "unknown",
		Summary:       "AI响应解析失败",
	}

	// 尝试提取风险等级
	if strings.Contains(response, "critical") || strings.Contains(response, "严重") {
		result.RiskLevel = "critical"
	} else if strings.Contains(response, "high") || strings.Contains(response, "高危") {
		result.RiskLevel = "high"
	} else if strings.Contains(response, "medium") || strings.Contains(response, "中危") {
		result.RiskLevel = "medium"
	}

	// 尝试提取摘要
	if idx := strings.Index(response, `"summary"`); idx != -1 {
		// 找到摘要值
		start := strings.Index(response[idx:], `:"`)
		if start != -1 {
			start = idx + start + 2
			end := strings.Index(response[start:], `"`)
			if end > 0 {
				result.Summary = response[start : start+end]
			}
		}
	}

	// 尝试提取攻击模式
	if idx := strings.Index(response, `"attack_pattern"`); idx != -1 {
		start := strings.Index(response[idx:], `:"`)
		if start != -1 {
			start = idx + start + 2
			end := strings.Index(response[start:], `"`)
			if end > 0 {
				result.AttackPattern = response[start : start+end]
			}
		}
	}

	return result
}

// normalizeRiskLevel 规范化风险等级
func (c *OllamaClient) normalizeRiskLevel(level string) string {
	switch strings.ToLower(level) {
	case "critical", "高", "严重":
		return "critical"
	case "high", "高危":
		return "high"
	case "medium", "中", "中危":
		return "medium"
	case "low", "低", "低危":
		return "low"
	default:
		return "low"
	}
}

// CheckHealth 检查Ollama服务健康状态
func (c *OllamaClient) CheckHealth(ctx context.Context) error {
	url := c.baseURL + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama服务不可达: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama服务异常: %s", resp.Status)
	}

	return nil
}

// ListModels 列出可用模型
func (c *OllamaClient) ListModels(ctx context.Context) ([]string, error) {
	url := c.baseURL + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	names := make([]string, len(result.Models))
	for i, m := range result.Models {
		names[i] = m.Name
	}

	return names, nil
}
