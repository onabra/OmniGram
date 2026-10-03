package repository

import (
	"OmniGram/domain"
	"time"

	"gorm.io/gorm"
)

// ReferralStats ساختار برای خروجی آمار رفرال
type ReferralStats struct {
	TotalInvited  int64
	Approved      int64
	InactiveUsers int64
}

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

// GetReferralStats دریافت آمار سه‌گانه رفرال برای یک کاربر
func (r *ReferralRepository) GetReferralStats(tenantID uint, referrerID uint) (*ReferralStats, error) {
	stats := &ReferralStats{}

	// 1. تعداد کل دعوت شده‌ها
	r.db.Model(&domain.Referral{}).
		Where("tenant_id = ? AND referrer_id = ?", tenantID, referrerID).
		Count(&stats.TotalInvited)

	// 2. تعداد تایید شده‌ها
	r.db.Model(&domain.Referral{}).
		Where("tenant_id = ? AND referrer_id = ? AND status = ?", tenantID, referrerID, "APPROVED").
		Count(&stats.Approved)

	// 3. تعداد غیرفعال‌ها (بیشتر از 10 روز)
	tenDaysAgo := time.Now().AddDate(0, 0, -10)
	r.db.Table("referrals").
		Joins("JOIN users ON users.id = referrals.referred_id").
		Where("referrals.tenant_id = ? AND referrals.referrer_id = ?", tenantID, referrerID).
		Where("users.last_active_date < ?", tenDaysAgo).
		Count(&stats.InactiveUsers)

	return stats, nil
}
