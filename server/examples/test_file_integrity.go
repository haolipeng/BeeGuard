//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"

	"github.com/haolipeng/BeeGuard/server/internal/analysis"
	"github.com/haolipeng/BeeGuard/server/internal/config"
	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/log"
)

func main() {
	// 加载配置
	cfg, err := config.Load("./conf/server.yaml")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	// 初始化
	log.Init(&cfg.Log)
	db.Init(&cfg.Database)
	defer db.Close()

	fmt.Printf("使用模型: %s\n", cfg.Analysis.OllamaModel)
	fmt.Println("==========================================")
	fmt.Println("文件完整性监控告警 AI 分析测试")
	fmt.Println("==========================================")

	// 清除缓存
	cacheDir := "/tmp/server/test_file_integrity_cache"
	os.RemoveAll(cacheDir)

	// 创建报告目录
	reportDir := "/tmp/server/analysis_reports"
	os.MkdirAll(reportDir, 0755)

	// 创建分析引擎
	engine := analysis.NewEngine(analysis.EngineConfig{
		OllamaURL:   cfg.Analysis.OllamaURL,
		OllamaModel: cfg.Analysis.OllamaModel,
		CacheDir:    cacheDir,
		ReportDir:   reportDir,
	})

	// 获取所有 file_integrity 告警，按主机分组
	database := db.GetDB()
	hosts, err := getFileIntegrityHosts(database)
	if err != nil {
		fmt.Printf("获取文件完整性告警主机失败: %v\n", err)
		return
	}

	if len(hosts) == 0 {
		fmt.Println("未发现 file_integrity 类型的告警")
		return
	}

	fmt.Printf("发现 %d 个主机有文件完整性告警\n\n", len(hosts))

	// 分析结果收集
	var (
		reports    []*analysis.AnalysisReport
		failCount  int
		skipCount  int
		failReason = make(map[string]int)
	)

	// 逐个主机分析
	for i, host := range hosts {
		fmt.Printf("[%d/%d] 分析主机 %s (file_integrity 告警数: %d)...\n", i+1, len(hosts), host.HostIP, host.Count)

		// 获取该主机的 file_integrity 告警
		alerts, err := fetchFileIntegrityAlerts(database, host.HostIP)
		if err != nil {
			fmt.Printf("  获取告警失败: %v\n", err)
			failCount++
			failReason["获取告警失败: "+err.Error()]++
			continue
		}

		if len(alerts) == 0 {
			fmt.Printf("  无告警\n")
			skipCount++
			continue
		}

		fmt.Printf("  获取到 %d 条告警，正在分析...\n", len(alerts))

		// 调用 AI 分析
		analysisCtx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
		result, err := engine.AnalyzeAlerts(analysisCtx, alerts)
		cancel()

		if err != nil {
			fmt.Printf("  分析失败: %v\n", err)
			failCount++
			failReason[err.Error()]++
			continue
		}

		report := &analysis.AnalysisReport{
			AnalysisType:    "file_integrity",
			ScopeKey:        host.HostIP,
			AlertCount:      len(alerts),
			AlertSnapshot:   alerts,
			RiskLevel:       result.RiskLevel,
			AttackPattern:   result.AttackPattern,
			AttackStage:     result.AttackStage,
			Summary:         result.Summary,
			Recommendations: result.Recommendations,
			IOCIndicators:   result.IOCIndicators,
		}

		reports = append(reports, report)
		fmt.Printf("  风险等级: %s | 变更模式: %s\n", report.RiskLevel, report.AttackPattern)
	}

	// 输出汇总报告
	fmt.Println("\n==========================================")
	fmt.Println("文件完整性分析汇总报告")
	fmt.Println("==========================================")
	fmt.Printf("总主机数: %d\n", len(hosts))
	fmt.Printf("成功分析: %d\n", len(reports))
	fmt.Printf("跳过: %d\n", skipCount)
	fmt.Printf("失败: %d\n", failCount)

	if failCount > 0 {
		fmt.Println("\n失败原因:")
		for reason, count := range failReason {
			fmt.Printf("  - %s: %d次\n", reason, count)
		}
	}

	// 按风险等级统计（降噪效果）
	riskStats := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}

	fmt.Println("\n==========================================")
	fmt.Println("详细分析结果")
	fmt.Println("==========================================")

	for _, report := range reports {
		riskStats[report.RiskLevel]++

		fmt.Printf("\n--- 主机: %s ---\n", report.ScopeKey)
		fmt.Printf("告警数量: %d\n", report.AlertCount)
		fmt.Printf("风险等级: %s\n", report.RiskLevel)
		fmt.Printf("变更模式: %s\n", report.AttackPattern)
		fmt.Printf("攻击阶段: %s\n", report.AttackStage)
		fmt.Printf("分析摘要: %s\n", report.Summary)
		if len(report.Recommendations) > 0 {
			fmt.Println("处置建议:")
			for i, rec := range report.Recommendations {
				fmt.Printf("  %d. %s\n", i+1, rec)
			}
		}
		if report.IOCIndicators != nil && len(report.IOCIndicators) > 0 {
			fmt.Printf("IOC指标: %v\n", report.IOCIndicators)
		}
	}

	// 降噪效果统计
	fmt.Println("\n==========================================")
	fmt.Println("降噪效果统计")
	fmt.Println("==========================================")
	lowCount := riskStats["low"]
	threatCount := riskStats["medium"] + riskStats["high"] + riskStats["critical"]
	totalAnalyzed := len(reports)

	fmt.Printf("Low (误报/正常运维):  %d\n", lowCount)
	fmt.Printf("Medium (需关注):      %d\n", riskStats["medium"])
	fmt.Printf("High (高危):          %d\n", riskStats["high"])
	fmt.Printf("Critical (严重):      %d\n", riskStats["critical"])

	if totalAnalyzed > 0 {
		fmt.Printf("\n降噪率: %.1f%% (%d/%d 被标记为正常运维)\n",
			float64(lowCount)/float64(totalAnalyzed)*100, lowCount, totalAnalyzed)
		fmt.Printf("真实告警检出: %d 条 (medium+)\n", threatCount)
	}

	// 保存汇总报告
	summary := map[string]interface{}{
		"test_type":   "file_integrity",
		"total_hosts": len(hosts),
		"analyzed":    len(reports),
		"skipped":     skipCount,
		"failed":      failCount,
		"risk_stats":  riskStats,
		"noise_reduction": map[string]interface{}{
			"low_count":    lowCount,
			"threat_count": threatCount,
			"total":        totalAnalyzed,
		},
		"model":       cfg.Analysis.OllamaModel,
		"analyzed_at": time.Now().Format(time.RFC3339),
		"reports":     reports,
	}

	summaryData, _ := json.MarshalIndent(summary, "", "  ")
	summaryPath := reportDir + "/summary_file_integrity.json"
	os.WriteFile(summaryPath, summaryData, 0644)
	fmt.Printf("\n汇总报告已保存到: %s\n", summaryPath)
}

type fileIntegrityHostCount struct {
	HostIP string
	Count  int64
}

// getFileIntegrityHosts 获取有 file_integrity 告警的主机列表
func getFileIntegrityHosts(database *gorm.DB) ([]fileIntegrityHostCount, error) {
	var hosts []fileIntegrityHostCount

	query := `
		SELECT host_ip, COUNT(*) as count
		FROM v_alert_unified
		WHERE alert_type = 'file_integrity'
		GROUP BY host_ip
		ORDER BY count DESC
	`

	if err := database.Raw(query).Scan(&hosts).Error; err != nil {
		return nil, err
	}

	return hosts, nil
}

// fetchFileIntegrityAlerts 获取主机的 file_integrity 告警
func fetchFileIntegrityAlerts(database *gorm.DB, hostIP string) ([]analysis.AlertContext, error) {
	var alerts []analysis.AlertContext

	query := `
		SELECT alert_type, id, agent_id, host_id, host_name, host_ip,
		       status, alert_time, created_at, updated_at, details
		FROM v_alert_unified
		WHERE alert_type = 'file_integrity' AND host_ip = ?
		ORDER BY alert_time DESC
		LIMIT 20
	`

	if err := database.Raw(query, hostIP).Scan(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}
