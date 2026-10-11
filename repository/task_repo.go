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

// [repository/task_repo.go]
// feat(repository): add methods for user task validation and submission

// GetTaskByID واکشی اطلاعات یک تسک خاص
func (r *TaskRepository) GetTaskByID(tenantID uint, taskID uint) (*domain.Task, error) {
	var task domain.Task
	err := r.DB.Where("tenant_id = ? AND id = ?", tenantID, taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// HasUserCompletedTask بررسی اینکه آیا کاربر قبلاً این تسک را انجام داده یا خیر (ضد پاداش تکراری)
func (r *TaskRepository) HasUserCompletedTask(tenantID uint, userID uint, taskID uint) bool {
	var count int64
	r.DB.Model(&domain.TaskSubmission{}).Where("tenant_id = ? AND user_id = ? AND task_id = ?", tenantID, userID, taskID).Count(&count)
	return count > 0
}

// SubmitTask ثبت موفقیت‌آمیز انجام تسک توسط کاربر
func (r *TaskRepository) SubmitTask(submission *domain.TaskSubmission) error {
	return r.DB.Create(submission).Error
}

// [repository/task_repo.go]

// GetSubmissionByID دریافت اطلاعات یک سابمیشن (اثبات تسک)
func (r *TaskRepository) GetSubmissionByID(tenantID uint, subID uint) (*domain.TaskSubmission, error) {
	var sub domain.TaskSubmission
	err := r.DB.Where("tenant_id = ? AND id = ?", tenantID, subID).First(&sub).Error
	return &sub, err
}

// UpdateSubmissionStatus آپدیت وضعیت اسکرین‌شات ارسالی کاربر
func (r *TaskRepository) UpdateSubmissionStatus(sub *domain.TaskSubmission) error {
	return r.DB.Model(sub).Update("status", sub.Status).Error
}
