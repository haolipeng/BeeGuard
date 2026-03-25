package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/system"
)

// AgentInfoRepository Agent信息数据仓库
type AgentInfoRepository struct{}

// NewAgentInfoRepository 创建AgentInfo仓库实例
func NewAgentInfoRepository() *AgentInfoRepository {
	return &AgentInfoRepository{}
}

// getDB 获取数据库连接
func (r *AgentInfoRepository) getDB() *gorm.DB {
	return db.GetDB()
}

// RegisterAgent Agent连接时插入或更新记录（UPSERT）
// 冲突键为 agent_id，更新 version/status/hostname/ip/os_type/os_version/last_connected_at
// 不覆盖 registered_at
func (r *AgentInfoRepository) RegisterAgent(ctx context.Context, agentInfo *system.AgentInfo) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[AgentInfoRepo] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "agent_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"agent_version", "connection_status", "host_name", "host_ip",
			"os_type", "os_version", "last_connected_at", "updated_at",
		}),
	}).Create(agentInfo).Error

	if err != nil {
		log.Errorf("[AgentInfoRepo] Agent注册写入失败: %v", err)
	}
	return err
}

// DisconnectAgent Agent断开时将连接状态设为断开
func (r *AgentInfoRepository) DisconnectAgent(ctx context.Context, agentID string) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[AgentInfoRepo] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).
		Model(&system.AgentInfo{}).
		Where("agent_id = ?", agentID).
		Updates(map[string]interface{}{
			"connection_status": system.ConnectionStatusDisconnected,
			"updated_at":        time.Now(),
		}).Error

	if err != nil {
		log.Errorf("[AgentInfoRepo] Agent断开更新失败 agent_id=%s: %v", agentID, err)
	}
	return err
}

// UpdateLastConnected 心跳节流更新最后连接时间
func (r *AgentInfoRepository) UpdateLastConnected(ctx context.Context, agentID string) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[AgentInfoRepo] 数据库未初始化，跳过写入")
		return nil
	}

	now := time.Now()
	err := database.WithContext(ctx).
		Model(&system.AgentInfo{}).
		Where("agent_id = ?", agentID).
		Updates(map[string]interface{}{
			"last_connected_at": now,
			"updated_at":        now,
		}).Error

	if err != nil {
		log.Errorf("[AgentInfoRepo] Agent心跳更新失败 agent_id=%s: %v", agentID, err)
	}
	return err
}

// ResetAllConnectionStatus 服务启动时重置所有在线状态为断开
func (r *AgentInfoRepository) ResetAllConnectionStatus(ctx context.Context) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[AgentInfoRepo] 数据库未初始化，跳过重置连接状态")
		return nil
	}

	result := database.WithContext(ctx).
		Model(&system.AgentInfo{}).
		Where("connection_status = ?", system.ConnectionStatusConnected).
		Update("connection_status", system.ConnectionStatusDisconnected)

	if result.Error != nil {
		log.Errorf("[AgentInfoRepo] 重置连接状态失败: %v", result.Error)
		return result.Error
	}

	log.Infof("[AgentInfoRepo] 服务启动重置连接状态，影响行数: %d", result.RowsAffected)
	return nil
}
