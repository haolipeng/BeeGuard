//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

	// 创建分析引擎（不使用缓存）
	engine := analysis.NewEngine(analysis.EngineConfig{
		OllamaURL:   cfg.Analysis.OllamaURL,
		OllamaModel: cfg.Analysis.OllamaModel,
		CacheDir:    "/tmp/server/test_cache",
		ReportDir:   "/tmp/server/analysis_reports",
	})

	// 分析主机
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	hostIP := "10.107.12.99"
	fmt.Printf("=== 分析主机 %s ===\n", hostIP)

	report, err := engine.AnalyzeByHost(ctx, hostIP)
	if err != nil {
		fmt.Printf("分析失败: %v\n", err)
		return
	}

	if report == nil {
		fmt.Println("无待分析告警")
		return
	}

	// 输出报告
	fmt.Println("\n=== 分析报告 ===")
	data, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(data))

	// 检查报告文件
	reports, _ := engine.GetReports()
	fmt.Printf("\n=== 已保存报告: %d 个 ===\n", len(reports))
	for _, r := range reports {
		fmt.Println(r)
	}
}
