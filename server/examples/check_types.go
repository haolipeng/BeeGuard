//go:build ignore

package main

import (
	"fmt"
	"github.com/haolipeng/BeeGuard/server/internal/config"
	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/log"
)

func main() {
	cfg, _ := config.Load("./conf/server.yaml")
	log.Init(&cfg.Log)
	db.Init(&cfg.Database)
	defer db.Close()

	database := db.GetDB()

	// 按告警类型统计
	var typeStats []struct {
		AlertType string
		Count     int64
	}
	database.Raw(`
		SELECT alert_type, COUNT(*) as count
		FROM v_alert_unified
		GROUP BY alert_type
		ORDER BY count DESC
	`).Scan(&typeStats)

	fmt.Println("告警类型统计:")
	for _, t := range typeStats {
		fmt.Printf("  %s: %d\n", t.AlertType, t.Count)
	}

	// 按主机统计
	var hostStats []struct {
		HostIP string
		Count  int64
	}
	database.Raw(`
		SELECT host_ip, COUNT(*) as count
		FROM v_alert_unified
		GROUP BY host_ip
		ORDER BY count DESC
	`).Scan(&hostStats)

	fmt.Println("\n主机统计:")
	for _, h := range hostStats {
		fmt.Printf("  %s: %d 条告警\n", h.HostIP, h.Count)
	}

	// 总数
	var total int64
	database.Raw("SELECT COUNT(*) FROM v_alert_unified").Scan(&total)
	fmt.Printf("\n告警总数: %d\n", total)
}
