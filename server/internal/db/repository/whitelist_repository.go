package repository

import (
	"context"
	"fmt"

	"github.com/haolipeng/BeeGuard/server/internal/log"
	wlModel "github.com/haolipeng/BeeGuard/server/internal/models/whitelist"
	"gorm.io/gorm"
)

type WhitelistRepository struct {
	DB *gorm.DB
}

func NewWhitelistRepository(db *gorm.DB) *WhitelistRepository {
	return &WhitelistRepository{DB: db}
}

// Create inserts a new whitelist rule
func (r *WhitelistRepository) Create(ctx context.Context, alertType string, rule *wlModel.WhitelistRule) error {
	tableName, err := wlModel.GetWhitelistTableName(alertType)
	if err != nil {
		return err
	}
	result := r.DB.WithContext(ctx).Table(tableName).Create(rule)
	if result.Error != nil {
		log.Errorf("[WhitelistRepo] create rule in %s failed: %v", tableName, result.Error)
		return result.Error
	}
	return nil
}

// GetByID retrieves a whitelist rule by ID
func (r *WhitelistRepository) GetByID(ctx context.Context, alertType string, id int64) (*wlModel.WhitelistRule, error) {
	tableName, err := wlModel.GetWhitelistTableName(alertType)
	if err != nil {
		return nil, err
	}
	var rule wlModel.WhitelistRule
	result := r.DB.WithContext(ctx).Table(tableName).Where("id = ?", id).First(&rule)
	if result.Error != nil {
		return nil, result.Error
	}
	return &rule, nil
}

// List retrieves whitelist rules with pagination
func (r *WhitelistRepository) List(ctx context.Context, alertType string, page, limit int) ([]wlModel.WhitelistRule, int64, error) {
	tableName, err := wlModel.GetWhitelistTableName(alertType)
	if err != nil {
		return nil, 0, err
	}
	var rules []wlModel.WhitelistRule
	var total int64

	query := r.DB.WithContext(ctx).Table(tableName)
	if err := query.Count(&total).Error; err != nil {
		log.Errorf("[WhitelistRepo] count rules in %s failed: %v", tableName, err)
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&rules).Error; err != nil {
		log.Errorf("[WhitelistRepo] list rules in %s failed: %v", tableName, err)
		return nil, 0, err
	}
	return rules, total, nil
}

// Update updates a whitelist rule
func (r *WhitelistRepository) Update(ctx context.Context, alertType string, id int64, updates map[string]interface{}) error {
	tableName, err := wlModel.GetWhitelistTableName(alertType)
	if err != nil {
		return err
	}
	result := r.DB.WithContext(ctx).Table(tableName).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		log.Errorf("[WhitelistRepo] update rule %d in %s failed: %v", id, tableName, result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("rule %d not found in %s", id, tableName)
	}
	return nil
}

// Delete deletes a whitelist rule
func (r *WhitelistRepository) Delete(ctx context.Context, alertType string, id int64) error {
	tableName, err := wlModel.GetWhitelistTableName(alertType)
	if err != nil {
		return err
	}
	result := r.DB.WithContext(ctx).Table(tableName).Where("id = ?", id).Delete(&wlModel.WhitelistRule{})
	if result.Error != nil {
		log.Errorf("[WhitelistRepo] delete rule %d in %s failed: %v", id, tableName, result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("rule %d not found in %s", id, tableName)
	}
	return nil
}
