package repository

import (
	"OmniGram/domain" // نام ماژول خود را جایگزین کنید

	"gorm.io/gorm"
)

// UserRepository اینترفیسی برای مدیریت کاربران
type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByTelegramID(tenantID uint, telegramID int64) (*domain.User, error)
	UpdateBalances(tenantID uint, userID uint, toman, point, stars float64) error

	GetUserByID(tenantID uint, id uint) (*domain.User, error)
	AddBalances(tenantID uint, userID uint, toman, point, stars float64) error
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
