package repository

import (
	"OmniGram/domain"

	"gorm.io/gorm"
)

type ReferralRepository struct {
	db *gorm.DB
}

func NewReferralRepository(db *gorm.DB) *ReferralRepository {
	return &ReferralRepository{db: db}
}

// Create ثبت رفرال جدید
func (r *ReferralRepository) Create(referral *domain.Referral) error {
	return r.db.Create(referral).Error
}

// GetPendingReferral دریافت رفرال تایید نشده کاربر
func (r *ReferralRepository) GetPendingReferral(tenantID uint, referredID uint) (*domain.Referral, error) {
	var referral domain.Referral
	err := r.db.Where("tenant_id = ? AND referred_id = ? AND status = ?", tenantID, referredID, "PENDING").First(&referral).Error
	if err != nil {
		return nil, err
	}
	return &referral, nil
}

// UpdateStatus آپدیت وضعیت رفرال (مثلاً از PENDING به APPROVED)
func (r *ReferralRepository) UpdateStatus(referral *domain.Referral) error {
	return r.db.Model(referral).Update("status", referral.Status).Error
}

// GetReferrer پیدا کردن معرف یک کاربر (برای کسر امتیاز در زمان لفت دادن)
func (r *ReferralRepository) GetReferrer(tenantID uint, referredID uint) (*domain.Referral, error) {
	var referral domain.Referral
	// فقط رفرال‌های تایید شده (APPROVED) را می‌گیریم چون امتیاز فقط برای آن‌ها واریز شده است
	err := r.db.Where("tenant_id = ? AND referred_id = ? AND status = ?", tenantID, referredID, "APPROVED").First(&referral).Error
	if err != nil {
		return nil, err
	}
	return &referral, nil
}
