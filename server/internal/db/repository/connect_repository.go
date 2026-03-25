package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"
)

// ConnectRepository connect事件数据仓库
type ConnectRepository struct{}

// NewConnectRepository 创建connect仓库实例
func NewConnectRepository() *ConnectRepository {
	return &ConnectRepository{}
}

// getDB 获取数据库连接
func (r *ConnectRepository) getDB() *gorm.DB {
	return db.GetDB()
}

// Create 插入connect事件记录（事件流数据，只插入不更新）
func (r *ConnectRepository) Create(ctx context.Context, connect *host.Connect) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[Repository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(connect).Error

	if err != nil {
		log.Errorf("[Repository] connect事件写入失败: %v", err)
	}
	return err
}
