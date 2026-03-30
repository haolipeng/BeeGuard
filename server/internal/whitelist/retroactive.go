package whitelist

import (
	"fmt"

	"github.com/haolipeng/BeeGuard/server/internal/log"
	wlModel "github.com/haolipeng/BeeGuard/server/internal/models/whitelist"
)

// RetroactiveCheck 规则变更时异步追溯检查已有告警
// 当新增或修改白名单规则时调用
func (c *Checker) RetroactiveCheck(alertType string, rule *wlModel.WhitelistRule) {
	go c.doRetroactiveCheck(alertType, rule)
}

// doRetroactiveCheck 执行追溯检查（在 goroutine 中运行）
func (c *Checker) doRetroactiveCheck(alertType string, rule *wlModel.WhitelistRule) {
	alertTable, ok := wlModel.AlertTypeToAlertTable[alertType]
	if !ok {
		log.Errorf("[Whitelist] retroactive: unknown alert type %s", alertType)
		return
	}

	// 容器告警需要检查多张表
	var tables []string
	if alertType == "container_alert" {
		tables = []string{
			"alert_container_dangerous_command",
			"alert_container_reverse_shell",
			"alert_container_sensitive_file",
		}
	} else {
		tables = []string{alertTable}
	}

	for _, table := range tables {
		c.retroactiveCheckTable(table, rule)
	}

	log.Infof("[Whitelist] retroactive check completed for rule %d (%s)", rule.ID, alertType)
}

// retroactiveCheckTable 对单张告警表执行追溯检查
func (c *Checker) retroactiveCheckTable(tableName string, rule *wlModel.WhitelistRule) {
	const batchSize = 500
	var offset int

	for {
		// 查询未命中白名单的告警记录
		var results []map[string]interface{}
		err := c.db.Table(tableName).
			Select("id, agent_id, *").
			Where("whitelist_hit = ?", false).
			Order("id ASC").
			Limit(batchSize).
			Offset(offset).
			Find(&results).Error

		if err != nil {
			log.Errorf("[Whitelist] retroactive query %s failed: %v", tableName, err)
			return
		}

		if len(results) == 0 {
			break
		}

		// 检查每条记录是否匹配规则
		var matchedIDs []int64
		for _, row := range results {
			fields := c.rowToFields(row)
			agentID := fmt.Sprintf("%v", row["agent_id"])

			// 检查 scope
			if rule.Scope == wlModel.ScopeAgent {
				if !c.matchAgentScope(rule.AgentIDs, agentID) {
					continue
				}
			}

			if c.evaluateConditions(&rule.Conditions, fields) {
				if id, ok := row["id"].(int64); ok {
					matchedIDs = append(matchedIDs, id)
				}
			}
		}

		// 批量更新匹配的记录
		if len(matchedIDs) > 0 {
			err := c.db.Table(tableName).
				Where("id IN ?", matchedIDs).
				Updates(map[string]interface{}{
					"whitelist_hit":     true,
					"whitelist_rule_id": rule.ID,
				}).Error
			if err != nil {
				log.Errorf("[Whitelist] retroactive update %s failed: %v", tableName, err)
			}

			// 更新命中计数
			c.db.Table(wlModel.AlertTypeToTable[c.getAlertTypeByTable(tableName)]).
				Where("id = ?", rule.ID).
				UpdateColumn("hit_count", c.db.Raw("hit_count + ?", len(matchedIDs)))
		}

		if len(results) < batchSize {
			break
		}
		offset += batchSize
	}
}

// RestoreOnDelete 删除白名单规则时，恢复对应告警的白名单状态
func (c *Checker) RestoreOnDelete(alertType string, ruleID int64) {
	go c.doRestoreOnDelete(alertType, ruleID)
}

func (c *Checker) doRestoreOnDelete(alertType string, ruleID int64) {
	alertTable, ok := wlModel.AlertTypeToAlertTable[alertType]
	if !ok {
		return
	}

	var tables []string
	if alertType == "container_alert" {
		tables = []string{
			"alert_container_dangerous_command",
			"alert_container_reverse_shell",
			"alert_container_sensitive_file",
		}
	} else {
		tables = []string{alertTable}
	}

	for _, table := range tables {
		err := c.db.Table(table).
			Where("whitelist_rule_id = ?", ruleID).
			Updates(map[string]interface{}{
				"whitelist_hit":     false,
				"whitelist_rule_id": nil,
			}).Error
		if err != nil {
			log.Errorf("[Whitelist] restore on delete failed for %s rule %d: %v", table, ruleID, err)
		}
	}

	log.Infof("[Whitelist] restored alerts for deleted rule %d (%s)", ruleID, alertType)
}

// rowToFields 将数据库行转换为 string map 用于规则匹配
func (c *Checker) rowToFields(row map[string]interface{}) map[string]string {
	fields := make(map[string]string, len(row))
	for k, v := range row {
		if v != nil {
			fields[k] = fmt.Sprintf("%v", v)
		}
	}
	return fields
}

// getAlertTypeByTable 根据告警表名反查告警类型
func (c *Checker) getAlertTypeByTable(tableName string) string {
	for alertType, table := range wlModel.AlertTypeToAlertTable {
		if table == tableName {
			return alertType
		}
	}
	// 容器告警表映射
	switch tableName {
	case "alert_container_dangerous_command",
		"alert_container_reverse_shell",
		"alert_container_sensitive_file":
		return "container_alert"
	}
	return ""
}
