package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"
)

// DNSRepository DNS事件数据仓库
type DNSRepository struct{}

// NewDNSRepository 创建DNS仓库实例
func NewDNSRepository() *DNSRepository {
	return &DNSRepository{}
}

// getDB 获取数据库连接
func (r *DNSRepository) getDB() *gorm.DB {
	return db.GetDB()
}

// Create 插入DNS事件记录（事件流数据，只插入不更新）
func (r *DNSRepository) Create(ctx context.Context, dns *host.DNS) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[Repository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(dns).Error

	if err != nil {
		log.Errorf("[Repository] DNS事件写入失败: %v", err)
	}
	return err
}
