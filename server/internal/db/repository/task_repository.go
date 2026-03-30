package repository

import (
	"context"

	"github.com/haolipeng/BeeGuard/server/internal/log"
	taskModel "github.com/haolipeng/BeeGuard/server/internal/models/task"
	"gorm.io/gorm"
)

type TaskRepository struct {
	DB *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{DB: db}
}

// Create inserts a new task history record
func (r *TaskRepository) Create(ctx context.Context, task *taskModel.TaskHistory) error {
	result := r.DB.WithContext(ctx).Create(task)
	if result.Error != nil {
		log.Errorf("[TaskRepo] create task failed: %v", result.Error)
		return result.Error
	}
	return nil
}

// GetByID retrieves a task history record by ID
func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*taskModel.TaskHistory, error) {
	var task taskModel.TaskHistory
	result := r.DB.WithContext(ctx).Where("id = ?", id).First(&task)
	if result.Error != nil {
		return nil, result.Error
	}
	return &task, nil
}

// GetByTaskID retrieves a task history record by task_id (UUID)
func (r *TaskRepository) GetByTaskID(ctx context.Context, taskID string) (*taskModel.TaskHistory, error) {
	var task taskModel.TaskHistory
	result := r.DB.WithContext(ctx).Where("task_id = ?", taskID).First(&task)
	if result.Error != nil {
		return nil, result.Error
	}
	return &task, nil
}

// List retrieves task history with pagination and optional filters
func (r *TaskRepository) List(ctx context.Context, page, limit int, agentID string, taskType int32, status int16) ([]taskModel.TaskHistory, int64, error) {
	var tasks []taskModel.TaskHistory
	var total int64

	query := r.DB.WithContext(ctx).Model(&taskModel.TaskHistory{})

	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}
	if taskType > 0 {
		query = query.Where("task_type = ?", taskType)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		log.Errorf("[TaskRepo] count tasks failed: %v", err)
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tasks).Error; err != nil {
		log.Errorf("[TaskRepo] list tasks failed: %v", err)
		return nil, 0, err
	}

	return tasks, total, nil
}

// UpdateStatus updates the status and result of a task
func (r *TaskRepository) UpdateStatus(ctx context.Context, taskID string, status int16, resultMessage string) error {
	result := r.DB.WithContext(ctx).Model(&taskModel.TaskHistory{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"status":         status,
			"result_message": resultMessage,
		})
	if result.Error != nil {
		log.Errorf("[TaskRepo] update task status failed: %v", result.Error)
		return result.Error
	}
	return nil
}
