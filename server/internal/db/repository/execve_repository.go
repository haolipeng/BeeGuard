package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"
)

// ExecveRepository execve事件数据仓库
type ExecveRepository struct{}

// NewExecveRepository 创建execve仓库实例
func NewExecveRepository() *ExecveRepository {
	return &ExecveRepository{}
}

// getDB 获取数据库连接
func (r *ExecveRepository) getDB() *gorm.DB {
	return db.GetDB()
}

// Create 插入execve事件记录（事件流数据，只插入不更新）
func (r *ExecveRepository) Create(ctx context.Context, execve *host.Execve) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[Repository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(execve).Error

	if err != nil {
		log.Errorf("[Repository] execve事件写入失败: %v", err)
	}
	return err
}

// List 查询execve事件列表（按时间倒序）
func (r *ExecveRepository) List(ctx context.Context, agentID string, limit int) ([]*host.Execve, error) {
	database := r.getDB()
	if database == nil {
		return nil, nil
	}

	var list []*host.Execve
	query := database.WithContext(ctx).
		Order("event_time DESC").
		Limit(limit)

	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}

	err := query.Find(&list).Error
	if err != nil {
		log.Errorf("[Repository] 查询execve事件失败: %v", err)
		return nil, err
	}

	return list, nil
}

// CountByAgent 统计指定Agent的execve事件数量
func (r *ExecveRepository) CountByAgent(ctx context.Context, agentID string) (int64, error) {
	database := r.getDB()
	if database == nil {
		return 0, nil
	}

	var count int64
	err := database.WithContext(ctx).
		Model(&host.Execve{}).
		Where("agent_id = ?", agentID).
		Count(&count).Error

	if err != nil {
		log.Errorf("[Repository] 统计execve事件失败: %v", err)
		return 0, err
	}

	return count, nil
}

// DeleteOldRecords 删除指定天数之前的旧记录（用于数据清理）
func (r *ExecveRepository) DeleteOldRecords(ctx context.Context, daysAgo int) (int64, error) {
	database := r.getDB()
	if database == nil {
		return 0, nil
	}

	result := database.WithContext(ctx).
		Where("event_time < NOW() - INTERVAL '1 day' * ?", daysAgo).
		Delete(&host.Execve{})

	if result.Error != nil {
		log.Errorf("[Repository] 删除旧execve记录失败: %v", result.Error)
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
