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

	var count int64
	database.Raw("SELECT COUNT(*) FROM v_alert_unified").Scan(&count)
	fmt.Printf("告警总数: %d\n", count)

	var recentCount int64
	database.Raw("SELECT COUNT(*) FROM v_alert_unified WHERE alert_time >= NOW() - INTERVAL '2 hours'").Scan(&recentCount)
	fmt.Printf("最近2小时告警: %d\n", recentCount)

	var hosts []struct {
		HostIP string
		Count  int64
	}
	database.Raw(`
		SELECT host_ip, COUNT(*) as count
		FROM v_alert_unified
		WHERE alert_time >= NOW() - INTERVAL '2 hours'
		GROUP BY host_ip
	`).Scan(&hosts)

	for _, h := range hosts {
		fmt.Printf("主机 %s: %d 条告警\n", h.HostIP, h.Count)
	}
}
