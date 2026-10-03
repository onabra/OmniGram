// [repository/task_repo.go]
// feat(repository): create task repository for handling sponsor tasks

package repository

import (
	"OmniGram/domain"

	"gorm.io/gorm"
)

type TaskRepository struct {
	DB *gorm.DB
}

// CreateTask ایجاد تسک جدید در دیتابیس (تبدیل سفارش به تسک)
func (r *TaskRepository) CreateTask(task *domain.Task) error {
	return r.DB.Create(task).Error
}

// GetActiveTasks واکشی تمام تسک‌های فعال و دارای ظرفیت برای یک کارفرما
func (r *TaskRepository) GetActiveTasks(tenantID uint) ([]domain.Task, error) {
	var tasks []domain.Task
	err := r.DB.Where("tenant_id = ? AND status = ? AND capacity > 0", tenantID, "ACTIVE").Find(&tasks).Error
	return tasks, err
}

// DecreaseCapacity کسر ظرفیت تسک پس از انجام موفقیت‌آمیز توسط کاربر
func (r *TaskRepository) DecreaseCapacity(tenantID uint, taskID uint) error {
	return r.DB.Model(&domain.Task{}).
		Where("id = ? AND tenant_id = ? AND capacity > 0", taskID, tenantID).
		UpdateColumn("capacity", gorm.Expr("capacity - 1")).Error
}
