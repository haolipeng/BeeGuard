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

	// 清除缓存
	cacheDir := "/tmp/server/test_all_cache"
	os.RemoveAll(cacheDir)

	// 创建分析引擎
	engine := analysis.NewEngine(analysis.EngineConfig{
		OllamaURL:   cfg.Analysis.OllamaURL,
		OllamaModel: cfg.Analysis.OllamaModel,
		CacheDir:    cacheDir,
		ReportDir:   "/tmp/server/analysis_reports",
	})

	// 获取所有有告警的主机
	ctx := context.Background()
	hosts, err := getHostsWithAlerts(ctx)
	if err != nil {
		fmt.Printf("获取主机列表失败: %v\n", err)
		return
	}

	fmt.Printf("发现 %d 个主机有告警\n\n", len(hosts))

	database := db.GetDB()

	// 分析结果收集
	var (
		reports    []*analysis.AnalysisReport
		failCount  int
		skipCount  int
		failReason = make(map[string]int)
	)

	// 逐个主机分析
	for i, host := range hosts {
		fmt.Printf("[%d/%d] 分析主机 %s (告警数: %d)...\n", i+1, len(hosts), host.HostIP, host.Count)

		// 直接获取该主机的所有告警（不限时间）
		alerts, err := fetchAllAlertsByHost(database, host.HostIP)
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

		// 调用Ollama分析（增大超时时间）
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
			AnalysisType:    "host",
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

		fmt.Printf("  风险等级: %s | 攻击模式: %s\n", report.RiskLevel, report.AttackPattern)
	}

	// 输出汇总报告
	fmt.Println("\n==========================================")
	fmt.Println("分析汇总报告")
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

	// 按风险等级统计
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
		fmt.Printf("攻击模式: %s\n", report.AttackPattern)
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

	fmt.Println("\n==========================================")
	fmt.Println("风险等级分布")
	fmt.Println("==========================================")
	fmt.Printf("Critical: %d\n", riskStats["critical"])
	fmt.Printf("High: %d\n", riskStats["high"])
	fmt.Printf("Medium: %d\n", riskStats["medium"])
	fmt.Printf("Low: %d\n", riskStats["low"])

	// 保存汇总报告
	summary := map[string]interface{}{
		"total_hosts": len(hosts),
		"analyzed":    len(reports),
		"skipped":     skipCount,
		"failed":      failCount,
		"risk_stats":  riskStats,
		"model":       cfg.Analysis.OllamaModel,
		"analyzed_at": time.Now().Format(time.RFC3339),
		"reports":     reports,
	}

	summaryData, _ := json.MarshalIndent(summary, "", "  ")
	os.WriteFile("/tmp/server/analysis_reports/summary_all.json", summaryData, 0644)
	fmt.Println("\n汇总报告已保存到: /tmp/server/analysis_reports/summary_all.json")
}

type HostAlertCount struct {
	HostIP string
	Count  int64
}

func getHostsWithAlerts(ctx context.Context) ([]HostAlertCount, error) {
	var hosts []HostAlertCount

	query := `
		SELECT host_ip, COUNT(*) as count
		FROM v_alert_unified
		GROUP BY host_ip
		ORDER BY count DESC
	`

	database := db.GetDB()
	if err := database.WithContext(ctx).Raw(query).Scan(&hosts).Error; err != nil {
		return nil, err
	}

	return hosts, nil
}

// fetchAllAlertsByHost 获取主机的所有告警（不限时间）
func fetchAllAlertsByHost(db *gorm.DB, hostIP string) ([]analysis.AlertContext, error) {
	var alerts []analysis.AlertContext

	query := `
		SELECT alert_type, id, agent_id, host_id, host_name, host_ip,
		       status, alert_time, created_at, updated_at, details
		FROM v_alert_unified
		WHERE host_ip = ?
		ORDER BY alert_time DESC
		LIMIT 20
	`

	if err := db.Raw(query, hostIP).Scan(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}
