package repository

import (
	"OmniGram/domain" // نام ماژول خود را جایگزین کنید

	"gorm.io/gorm"
)

// UserRepository اینترفیسی برای مدیریت کاربران
type UserRepository interface {
	UpdateUserState(tenantID uint, userID uint, state domain.UserState) error
	CreateUser(user *domain.User) error
	GetUserByTelegramID(tenantID uint, telegramID int64) (*domain.User, error)
	UpdateBalances(tenantID uint, userID uint, toman, point, stars float64) error
	GetUserByID(tenantID uint, id uint) (*domain.User, error)
	AddBalances(tenantID uint, userID uint, toman, point, stars float64) error
	UpdateDrawTickets(tenantID uint, userID uint, tickets int) error
	UpdateUserTaskID(tenantID uint, userID uint, taskID uint) error
	UpdateUserOrderID(tenantID uint, userID uint, orderID uint) error
	UpdateWithdrawDraft(tenantID uint, userID uint, currency string, amount float64) error
	UpdateAdminTargets(tenantID uint, adminID uint, targetUserID uint, setting string, action string) error
	UpdateAdminTaskTarget(tenantID uint, adminID uint, taskID uint) error
	UpdateAdminTicketTarget(tenantID uint, adminID uint, ticketID uint) error
	GetAllUserTelegramIDs(tenantID uint) ([]int64, error)
}

type userRepository struct {
	db *gorm.DB
}

// GetUserByID جستجوی کاربر با آیدی دیتابیس
func (r *userRepository) GetUserByID(tenantID uint, id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// AddBalances افزایش موجودی کاربر
func (r *userRepository) AddBalances(tenantID uint, userID uint, toman, point, stars float64) error {
	return r.db.Model(&domain.User{}).
		Where("tenant_id = ? AND id = ?", tenantID, userID).
		Updates(map[string]interface{}{
			"toman_balance": gorm.Expr("toman_balance + ?", toman),
			"point_balance": gorm.Expr("point_balance + ?", point),
			"stars_balance": gorm.Expr("stars_balance + ?", stars),
		}).Error
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser ایجاد کاربر جدید
func (r *userRepository) CreateUser(user *domain.User) error {
	return r.db.Create(user).Error
}

// GetUserByTelegramID دریافت کاربر با رعایت ایزوله‌سازی کارفرما
func (r *userRepository) GetUserByTelegramID(tenantID uint, telegramID int64) (*domain.User, error) {
	var user domain.User
	// در معماری SaaS همیشه باید tenant_id در شرط‌ها حضور داشته باشد
	err := r.db.Where("tenant_id = ? AND telegram_id = ?", tenantID, telegramID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateBalances بروزرسانی موجودی سه‌گانه کاربر
func (r *userRepository) UpdateBalances(tenantID uint, userID uint, toman, point, stars float64) error {
	return r.db.Model(&domain.User{}).
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		Updates(map[string]interface{}{
			"toman_balance": toman,
			"point_balance": point,
			"stars_balance": stars,
		}).Error
}

// UpdateUserState تغییر وضعیت کاربر برای مدیریت عملیات‌های چند مرحله‌ای
func (r *userRepository) UpdateUserState(tenantID uint, userID uint, state domain.UserState) error {
	return r.db.Model(&domain.User{}).
		Where("tenant_id = ? AND id = ?", tenantID, userID).
		Update("state", state).Error
}

// UpdateDrawTickets بروزرسانی تعداد تیکت‌های کاربر
func (r *userRepository) UpdateDrawTickets(tenantID uint, userID uint, tickets int) error {
	return r.db.Model(&domain.User{}).
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		Update("draw_tickets", tickets).Error
}

// UpdateUserTaskID ذخیره آیدی تسک در حال انجام کاربر
func (r *userRepository) UpdateUserTaskID(tenantID uint, userID uint, taskID uint) error {
	return r.db.Model(&domain.User{}).
		Where("tenant_id = ? AND id = ?", tenantID, userID).
		Update("current_task_id", taskID).Error
}

// UpdateUserOrderID ذخیره آیدی سفارشی که در حال ویرایش است
func (r *userRepository) UpdateUserOrderID(tenantID uint, userID uint, orderID uint) error {
	return r.db.Model(&domain.User{}).
		Where("tenant_id = ? AND id = ?", tenantID, userID).
		Update("current_order_id", orderID).Error
}

// UpdateWithdrawDraft ذخیره اطلاعات پیش‌نویس برداشت کاربر
func (r *userRepository) UpdateWithdrawDraft(tenantID uint, userID uint, currency string, amount float64) error {
	return r.db.Model(&domain.User{}).
		Where("tenant_id = ? AND id = ?", tenantID, userID).
		Updates(map[string]interface{}{
			"current_withdraw_currency": currency,
			"current_withdraw_amount":   amount,
		}).Error
}

// UpdateAdminTargets ذخیره اطلاعات کاربری که ادمین قصد ویرایش آن را دارد
func (r *userRepository) UpdateAdminTargets(tenantID uint, adminID uint, targetUserID uint, setting string, action string) error {
	return r.db.Model(&domain.User{}).
		Where("tenant_id = ? AND id = ?", tenantID, adminID).
		Updates(map[string]interface{}{
			"admin_target_user_id": targetUserID,
			"admin_target_setting": setting,
			"admin_target_action":  action,
		}).Error
}

// UpdateAdminTaskTarget ذخیره آیدی تسکی که ادمین قصد ویرایش آن را دارد
func (r *userRepository) UpdateAdminTaskTarget(tenantID uint, adminID uint, taskID uint) error {
	return r.db.Model(&domain.User{}).
		Where("tenant_id = ? AND id = ?", tenantID, adminID).
		Update("admin_target_task_id", taskID).Error
}

// UpdateAdminTicketTarget ذخیره آیدی تیکتی که ادمین قصد پاسخ به آن را دارد
func (r *userRepository) UpdateAdminTicketTarget(tenantID uint, adminID uint, ticketID uint) error {
	return r.db.Model(&domain.User{}).
		Where("tenant_id = ? AND id = ?", tenantID, adminID).
		Update("admin_target_ticket_id", ticketID).Error
}

// GetAllUserTelegramIDs واکشی آیدی تمام کاربران یک کارفرما برای پیام همگانی
func (r *userRepository) GetAllUserTelegramIDs(tenantID uint) ([]int64, error) {
	var ids []int64
	err := r.db.Model(&domain.User{}).Where("tenant_id = ?", tenantID).Pluck("telegram_id", &ids).Error
	return ids, err
}
