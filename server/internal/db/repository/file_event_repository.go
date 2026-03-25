package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"
)

// FileEventRepository 文件事件数据仓库
type FileEventRepository struct{}

// NewFileEventRepository 创建文件事件仓库实例
func NewFileEventRepository() *FileEventRepository {
	return &FileEventRepository{}
}

// getDB 获取数据库连接
func (r *FileEventRepository) getDB() *gorm.DB {
	return db.GetDB()
}

// Create 插入文件事件记录（事件流数据，只插入不更新）
func (r *FileEventRepository) Create(ctx context.Context, fileEvent *host.FileEvent) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[Repository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(fileEvent).Error
	if err != nil {
		log.Errorf("[Repository] 文件事件写入失败: %v", err)
	}
	return err
}
