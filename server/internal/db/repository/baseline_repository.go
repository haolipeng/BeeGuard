package repository

import (
	"context"

	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/back"
	"github.com/haolipeng/BeeGuard/server/internal/models/baseline"
)

// BaselineRepository 基线数据仓库
type BaselineRepository struct{}

// NewBaselineRepository 创建基线仓库实例
func NewBaselineRepository() *BaselineRepository {
	return &BaselineRepository{}
}

// GetTemplate 获取基线模板
func (r *BaselineRepository) GetTemplate(ctx context.Context, id int64) (*back.BaselineTemplate, error) {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[BaselineRepository] 数据库未初始化，跳过查询")
		return nil, nil
	}

	var template back.BaselineTemplate
	err := database.WithContext(ctx).Where("id = ? AND is_enabled = 1", id).First(&template).Error
	if err != nil {
		log.Errorf("[BaselineRepository] 查询基线模板失败: %v", err)
		return nil, err
	}
	return &template, nil
}

// ListCheckItemsByTemplateID 获取基线模板下的所有检查项
func (r *BaselineRepository) ListCheckItemsByTemplateID(ctx context.Context, templateID int64) ([]*back.BaselineCheckItem, error) {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[BaselineRepository] 数据库未初始化，跳过查询")
		return nil, nil
	}

	var items []*back.BaselineCheckItem
	err := database.WithContext(ctx).Where("template_id = ?", templateID).Find(&items).Error
	if err != nil {
		log.Errorf("[BaselineRepository] 查询基线检查项失败: %v", err)
		return nil, err
	}
	return items, nil
}

// CreateCheckResult 创建基线检查结果记录
func (r *BaselineRepository) CreateCheckResult(ctx context.Context, result *baseline.CheckResult) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[BaselineRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(result).Error
	if err != nil {
		log.Errorf("[BaselineRepository] 基线检查结果写入失败: %v", err)
	}
	return err
}

// BatchCreateCheckDetails 批量创建基线检查明细记录
func (r *BaselineRepository) BatchCreateCheckDetails(ctx context.Context, details []*baseline.BaselineCheckDetail) error {
	database := db.GetDB()
	if database == nil {
		log.Warnf("[BaselineRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(&details).Error
	if err != nil {
		log.Errorf("[BaselineRepository] 基线检查明细批量写入失败: %v", err)
	}
	return err
}
