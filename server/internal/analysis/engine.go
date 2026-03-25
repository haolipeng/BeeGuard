package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gorm.io/gorm"

	analysismodel "github.com/haolipeng/BeeGuard/server/internal/models/analysis"
	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/log"
)

// Engine 分析引擎
type Engine struct {
	db         *gorm.DB
	cache      *DiskCache
	ollama     *OllamaClient
	reportDir  string
	mu         sync.Mutex
	running    bool
	cancel     context.CancelFunc
}

// EngineConfig 引擎配置
type EngineConfig struct {
	OllamaURL   string        // Ollama服务地址
	OllamaModel string        // 模型名称
	CacheDir    string        // 缓存目录
	CacheTTL    time.Duration // 缓存有效期
	ReportDir   string        // 报告存储目录
}

// NewEngine 创建分析引擎
func NewEngine(cfg EngineConfig) *Engine {
	if cfg.CacheDir == "" {
		cfg.CacheDir = "/tmp/server/analysis_cache"
	}
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 24 * time.Hour
	}
	if cfg.ReportDir == "" {
		cfg.ReportDir = "/tmp/server/analysis_reports"
	}

	// 确保报告目录存在
	os.MkdirAll(cfg.ReportDir, 0755)

	return &Engine{
		db:        db.GetDB(),
		cache:     NewDiskCache(cfg.CacheDir, cfg.CacheTTL),
		ollama:    NewOllamaClient(OllamaConfig{BaseURL: cfg.OllamaURL, Model: cfg.OllamaModel}),
		reportDir: cfg.ReportDir,
	}
}

// AnalyzeByHost 按主机维度分析
func (e *Engine) AnalyzeByHost(ctx context.Context, hostIP string) (*AnalysisReport, error) {
	// 1. 从视图获取该主机的告警
	alerts, err := e.fetchAlertsByHost(ctx, hostIP)
	if err != nil {
		return nil, fmt.Errorf("获取告警失败: %w", err)
	}

	if len(alerts) == 0 {
		return nil, nil
	}

	// 2. 过滤已分析的告警
	alerts = e.cache.FilterAnalyzed(alerts)
	if len(alerts) == 0 {
		log.Debugf("[Engine] 主机 %s 的告警已全部分析过", hostIP)
		return nil, nil
	}

	log.Infof("[Engine] 开始分析主机 %s, 告警数: %d", hostIP, len(alerts))

	// 3. 调用Ollama分析
	result, err := e.ollama.Analyze(ctx, alerts)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 4. 标记为已分析
	e.cache.MarkBatch(alerts)

	// 5. 生成报告
	report := &AnalysisReport{
		AnalysisType:    "host",
		ScopeKey:        hostIP,
		AlertCount:      len(alerts),
		AlertSnapshot:   alerts,
		RiskLevel:       result.RiskLevel,
		AttackPattern:   result.AttackPattern,
		AttackStage:     result.AttackStage,
		Summary:         result.Summary,
		Recommendations: result.Recommendations,
		IOCIndicators:   result.IOCIndicators,
	}

	// 6. 保存所有分析报告（用于测试和分析）
	if err := e.saveReport(report); err != nil {
		log.Warnf("[Engine] 保存报告失败: %v", err)
	}

	log.Infof("[Engine] 分析完成, 主机: %s, 风险等级: %s", hostIP, result.RiskLevel)
	return report, nil
}

// AnalyzeBySourceIP 按攻击源IP分析
func (e *Engine) AnalyzeBySourceIP(ctx context.Context, sourceIP string) (*AnalysisReport, error) {
	alerts, err := e.fetchAlertsBySourceIP(ctx, sourceIP)
	if err != nil {
		return nil, fmt.Errorf("获取告警失败: %w", err)
	}

	if len(alerts) == 0 {
		return nil, nil
	}

	alerts = e.cache.FilterAnalyzed(alerts)
	if len(alerts) == 0 {
		return nil, nil
	}

	log.Infof("[Engine] 开始分析攻击源 %s, 告警数: %d", sourceIP, len(alerts))

	result, err := e.ollama.Analyze(ctx, alerts)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	e.cache.MarkBatch(alerts)

	report := &AnalysisReport{
		AnalysisType:    "source_ip",
		ScopeKey:        sourceIP,
		AlertCount:      len(alerts),
		AlertSnapshot:   alerts,
		RiskLevel:       result.RiskLevel,
		AttackPattern:   result.AttackPattern,
		AttackStage:     result.AttackStage,
		Summary:         result.Summary,
		Recommendations: result.Recommendations,
		IOCIndicators:   result.IOCIndicators,
	}

	if result.RiskLevel == "medium" || result.RiskLevel == "high" || result.RiskLevel == "critical" {
		e.saveReport(report)
	}

	return report, nil
}

// AnalyzeCriticalAlert 分析单条高危告警
func (e *Engine) AnalyzeCriticalAlert(ctx context.Context, alertType string, alertID int64) (*AnalysisReport, error) {
	if e.cache.IsAnalyzed(alertType, alertID) {
		return nil, nil
	}

	alerts, err := e.fetchAlertByID(ctx, alertType, alertID)
	if err != nil || len(alerts) == 0 {
		return nil, err
	}

	log.Infof("[Engine] 分析高危告警: %s:%d", alertType, alertID)

	result, err := e.ollama.Analyze(ctx, alerts)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	e.cache.MarkAnalyzed(alertType, alertID)

	report := &AnalysisReport{
		AnalysisType:    "single",
		ScopeKey:        fmt.Sprintf("%s:%d", alertType, alertID),
		AlertCount:      1,
		AlertSnapshot:   alerts,
		RiskLevel:       result.RiskLevel,
		AttackPattern:   result.AttackPattern,
		AttackStage:     result.AttackStage,
		Summary:         result.Summary,
		Recommendations: result.Recommendations,
		IOCIndicators:   result.IOCIndicators,
	}

	if result.RiskLevel == "medium" || result.RiskLevel == "high" || result.RiskLevel == "critical" {
		e.saveReport(report)
	}

	return report, nil
}

// ScanAndAnalyze 扫描并分析（定时任务调用）
func (e *Engine) ScanAndAnalyze(ctx context.Context) error {
	log.Info("[Engine] 开始扫描待分析告警...")

	// 1. 扫描告警数量达到阈值的主机
	hosts, err := e.scanHostsWithAlerts(ctx)
	if err != nil {
		log.Warnf("[Engine] 扫描主机失败: %v", err)
	} else {
		for _, host := range hosts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				e.AnalyzeByHost(ctx, host)
			}
		}
	}

	// 2. 扫描高危告警
	criticalAlerts, err := e.scanCriticalAlerts(ctx)
	if err != nil {
		log.Warnf("[Engine] 扫描高危告警失败: %v", err)
	} else {
		for _, a := range criticalAlerts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				e.AnalyzeCriticalAlert(ctx, a.AlertType, a.ID)
			}
		}
	}

	log.Info("[Engine] 扫描分析完成")
	return nil
}

// fetchAlertsByHost 获取主机的告警
func (e *Engine) fetchAlertsByHost(ctx context.Context, hostIP string) ([]AlertContext, error) {
	var alerts []AlertContext

	// 使用 CURRENT_TIMESTAMP AT TIME ZONE 'UTC' 确保使用 UTC 时间进行比较
	query := `
		SELECT alert_type, id, agent_id, host_id, host_name, host_ip,
		       status, alert_time, created_at, updated_at, details
		FROM v_alert_unified
		WHERE host_ip = ?
		  AND alert_time >= (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') - INTERVAL '2 hours'
		ORDER BY alert_time DESC
		LIMIT 20
	`

	log.Infof("[Engine] fetchAlertsByHost: 查询主机 %s 的告警", hostIP)

	if err := e.db.WithContext(ctx).Raw(query, hostIP).Scan(&alerts).Error; err != nil {
		return nil, err
	}
	log.Infof("[Engine] fetchAlertsByHost: 查询到 %d 条告警", len(alerts))

	return alerts, nil
}

// fetchAlertsBySourceIP 获取攻击源的告警
func (e *Engine) fetchAlertsBySourceIP(ctx context.Context, sourceIP string) ([]AlertContext, error) {
	var alerts []AlertContext

	query := `
		SELECT alert_type, id, agent_id, host_id, host_name, host_ip,
		       status, alert_time, created_at, updated_at, details
		FROM v_alert_unified
		WHERE (details->>'source_ip' = ? OR details->>'attacker_ip' = ?)
		  AND alert_time >= (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') - INTERVAL '1 hour'
		ORDER BY alert_time DESC
		LIMIT 20
	`

	if err := e.db.WithContext(ctx).Raw(query, sourceIP, sourceIP).Scan(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}

// fetchAlertByID 获取单条告警
func (e *Engine) fetchAlertByID(ctx context.Context, alertType string, alertID int64) ([]AlertContext, error) {
	var alerts []AlertContext

	query := `
		SELECT alert_type, id, agent_id, host_id, host_name, host_ip,
		       status, alert_time, created_at, updated_at, details
		FROM v_alert_unified
		WHERE alert_type = ? AND id = ?
	`

	if err := e.db.WithContext(ctx).Raw(query, alertType, alertID).Scan(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}

// scanHostsWithAlerts 扫描告警数量达到阈值的主机
func (e *Engine) scanHostsWithAlerts(ctx context.Context) ([]string, error) {
	var hosts []string

	query := `
		SELECT host_ip
		FROM v_alert_unified
		WHERE alert_time >= (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') - INTERVAL '2 hours'
		GROUP BY host_ip
		HAVING COUNT(*) >= 2
	`

	if err := e.db.WithContext(ctx).Raw(query).Scan(&hosts).Error; err != nil {
		return nil, err
	}

	return hosts, nil
}

// scanCriticalAlerts 扫描高危告警
func (e *Engine) scanCriticalAlerts(ctx context.Context) ([]AlertContext, error) {
	var alerts []AlertContext

	query := `
		SELECT alert_type, id, agent_id, host_id, host_name, host_ip,
		       status, alert_time, created_at, updated_at, details
		FROM v_alert_unified
		WHERE alert_type IN ('reverse_shell', 'privilege_escalation', 'malware_scan', 'file_integrity')
		  AND alert_time >= (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') - INTERVAL '2 hours'
		LIMIT 10
	`

	if err := e.db.WithContext(ctx).Raw(query).Scan(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}

// saveReport 保存报告到磁盘和数据库
func (e *Engine) saveReport(report *AnalysisReport) error {
	// 1. 保存到磁盘
	filename := fmt.Sprintf("%s_%s_%d.json",
		report.AnalysisType,
		report.ScopeKey,
		time.Now().Unix(),
	)
	// 替换文件名中的特殊字符
	filename = sanitizeFilename(filename)

	filepath := filepath.Join(e.reportDir, filename)

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		log.Warnf("[Engine] 保存报告到磁盘失败: %v", err)
	} else {
		log.Infof("[Engine] 报告已保存到磁盘: %s", filepath)
	}

	// 2. 保存到数据库
	dbReport := &analysismodel.Report{
		AnalysisType:    report.AnalysisType,
		ScopeKey:        report.ScopeKey,
		AlertCount:      report.AlertCount,
		AlertSnapshot:   analysismodel.ToJSONB(report.AlertSnapshot),
		RiskLevel:       report.RiskLevel,
		AttackPattern:   report.AttackPattern,
		AttackStage:     report.AttackStage,
		Summary:         report.Summary,
		Recommendations: analysismodel.ToJSONB(report.Recommendations),
		IOCIndicators:   analysismodel.ToJSONB(report.IOCIndicators),
	}

	if err := e.db.Create(dbReport).Error; err != nil {
		log.Warnf("[Engine] 保存报告到数据库失败: %v", err)
	} else {
		log.Infof("[Engine] 报告已保存到数据库, ID: %d", dbReport.ID)
	}

	return nil
}

// sanitizeFilename 清理文件名
func sanitizeFilename(name string) string {
	result := make([]byte, 0, len(name))
	for _, c := range []byte(name) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '_' || c == '-' || c == '.' || c == ':' {
			result = append(result, c)
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}

// GetReports 获取报告列表
func (e *Engine) GetReports() ([]string, error) {
	files, err := os.ReadDir(e.reportDir)
	if err != nil {
		return nil, err
	}

	var reports []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
			reports = append(reports, f.Name())
		}
	}

	return reports, nil
}

// LoadReport 加载报告
func (e *Engine) LoadReport(filename string) (*AnalysisReport, error) {
	filepath := filepath.Join(e.reportDir, filename)
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var report AnalysisReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}

	return &report, nil
}

// Stats 获取统计信息
func (e *Engine) Stats() map[string]interface{} {
	reports, _ := e.GetReports()
	return map[string]interface{}{
		"cache":  e.cache.Stats(),
		"ollama": e.ollama.baseURL + " (" + e.ollama.model + ")",
		"reports_count": len(reports),
		"report_dir": e.reportDir,
	}
}

// AnalyzeAlerts 直接分析告警（用于测试，不限时间窗口）
func (e *Engine) AnalyzeAlerts(ctx context.Context, alerts []AlertContext) (*AnalysisResult, error) {
	if len(alerts) == 0 {
		return nil, nil
	}

	log.Infof("[Engine] 开始分析 %d 条告警", len(alerts))

	result, err := e.ollama.Analyze(ctx, alerts)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	log.Infof("[Engine] 分析完成, 风险等级: %s", result.RiskLevel)
	return result, nil
}

// GetReportsFromDB 从数据库获取报告列表
func (e *Engine) GetReportsFromDB(ctx context.Context, page, pageSize int, riskLevel, analysisType string) ([]analysismodel.Report, int64, error) {
	var reports []analysismodel.Report
	var total int64

	query := e.db.Model(&analysismodel.Report{})

	if riskLevel != "" {
		query = query.Where("risk_level = ?", riskLevel)
	}
	if analysisType != "" {
		query = query.Where("analysis_type = ?", analysisType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// GetReportFromDB 从数据库获取单个报告
func (e *Engine) GetReportFromDB(ctx context.Context, id int64) (*analysismodel.Report, error) {
	var report analysismodel.Report
	if err := e.db.First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

// DeleteReportFromDB 从数据库删除报告
func (e *Engine) DeleteReportFromDB(ctx context.Context, id int64) error {
	return e.db.Delete(&analysismodel.Report{}, id).Error
}

// GetDBReportStats 获取数据库报告统计
func (e *Engine) GetDBReportStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 总数
	var total int64
	e.db.Model(&analysismodel.Report{}).Count(&total)
	stats["total"] = total

	// 按风险等级统计
	var riskStats []struct {
		RiskLevel string
		Count     int64
	}
	e.db.Model(&analysismodel.Report{}).Select("risk_level, count(*) as count").Group("risk_level").Scan(&riskStats)
	for _, rs := range riskStats {
		stats[rs.RiskLevel] = rs.Count
	}

	// 按分析类型统计
	var typeStats []struct {
		AnalysisType string
		Count        int64
	}
	e.db.Model(&analysismodel.Report{}).Select("analysis_type, count(*) as count").Group("analysis_type").Scan(&typeStats)
	for _, ts := range typeStats {
		stats["type_"+ts.AnalysisType] = ts.Count
	}

	return stats, nil
}
