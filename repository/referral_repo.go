package repository

import (
	"OmniGram/domain"
	"time"

	"gorm.io/gorm"
)

// ReferralStats ساختار برای خروجی آمار رفرال
type ReferralStats struct {
	LeagueTotal   int64
	DrawTotal     int64
	PointTotal    int64
	WalletTotal   int64
	TotalInvited  int64
	Approved      int64
	InactiveUsers int64
	StarsTotal    int64
	GiftTotal     int64
}

// LeagueRankResult ساختار خروجی جدول رتبه‌بندی لیگ رفرال
type LeagueRankResult struct {
	UserID      uint
	FirstName   string
	InviteCount int64
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

// GetReferralStats دریافت آمار تفکیک شده رفرال‌های ۴ گانه برای یک کاربر
func (r *ReferralRepository) GetReferralStats(tenantID uint, referrerID uint) (*ReferralStats, error) {
	stats := &ReferralStats{}

	r.db.Model(&domain.Referral{}).Where("tenant_id = ? AND referrer_id = ? AND status IN ('APPROVED', 'ARCHIVED') AND referral_type = ?", tenantID, referrerID, "league").Count(&stats.LeagueTotal)
	r.db.Model(&domain.Referral{}).Where("tenant_id = ? AND referrer_id = ? AND status = ? AND referral_type IN ('draw', 'lottery')", tenantID, referrerID, "APPROVED").Count(&stats.DrawTotal)
	r.db.Model(&domain.Referral{}).Where("tenant_id = ? AND referrer_id = ? AND status = ? AND referral_type = ?", tenantID, referrerID, "APPROVED", "stars").Count(&stats.StarsTotal)
	r.db.Model(&domain.Referral{}).Where("tenant_id = ? AND referrer_id = ? AND status = ? AND referral_type = ?", tenantID, referrerID, "APPROVED", "gift").Count(&stats.GiftTotal)
	r.db.Model(&domain.Referral{}).Where("tenant_id = ? AND referrer_id = ? AND status = ? AND referral_type = ?", tenantID, referrerID, "APPROVED", "point").Count(&stats.PointTotal)
	r.db.Model(&domain.Referral{}).Where("tenant_id = ? AND referrer_id = ? AND status = ? AND referral_type = ?", tenantID, referrerID, "APPROVED", "wallet").Count(&stats.WalletTotal)

	tenDaysAgo := time.Now().AddDate(0, 0, -10)
	r.db.Table("referrals").
		Joins("JOIN users ON users.id = referrals.referred_id").
		Where("referrals.tenant_id = ? AND referrals.referrer_id = ?", tenantID, referrerID).
		Where("users.last_active_date < ?", tenDaysAgo).
		Count(&stats.InactiveUsers)

	return stats, nil
}

// GetLeagueRankings واکشی لیست تمام شرکت‌کنندگان لیگ بر اساس تعداد دعوت
func (r *ReferralRepository) GetLeagueRankings(tenantID uint) ([]LeagueRankResult, error) {
	var results []LeagueRankResult
	err := r.db.Table("referrals").
		Select("users.id as user_id, users.first_name, COUNT(referrals.id) as invite_count").
		Joins("JOIN users ON users.id = referrals.referrer_id").
		Where("referrals.tenant_id = ? AND referrals.status = ? AND referrals.referral_type = ?", tenantID, "APPROVED", "league").
		Group("users.id, users.first_name").
		Order("invite_count DESC").
		Scan(&results).Error
	return results, err
}

// HasEverBeenReferred بررسی اینکه آیا کاربر قبلاً توسط کسی دعوت شده است یا خیر
func (r *ReferralRepository) HasEverBeenReferred(tenantID uint, referredID uint) bool {
	var count int64
	r.db.Model(&domain.Referral{}).
		Where("tenant_id = ? AND referred_id = ? AND status IN ('PENDING', 'APPROVED', 'ARCHIVED')", tenantID, referredID).
		Count(&count)
	return count > 0
}

// ResetLeagueReferrals آرشیو کردن رفرال‌های فصل قبلی لیگ برای شروع فصل جدید
func (r *ReferralRepository) ResetLeagueReferrals(tenantID uint) error {
	return r.db.Model(&domain.Referral{}).
		Where("tenant_id = ? AND referral_type = ? AND status = ?", tenantID, "league", "APPROVED").
		Update("status", "ARCHIVED").Error
}

// GetTenantTotalReferralStats دریافت آمار کلی رفرال‌های ثبت شده در کل ربات برای پنل ادمین
func (r *ReferralRepository) GetTenantTotalReferralStats(tenantID uint) (*ReferralStats, error) {
	stats := &ReferralStats{}
	r.db.Model(&domain.Referral{}).Where("tenant_id = ?", tenantID).Count(&stats.TotalInvited)
	r.db.Model(&domain.Referral{}).Where("tenant_id = ? AND status IN ('APPROVED', 'ARCHIVED')", tenantID).Count(&stats.Approved)
	r.db.Model(&domain.Referral{}).Where("tenant_id = ? AND status IN ('APPROVED', 'ARCHIVED') AND referral_type = ?", tenantID, "league").Count(&stats.LeagueTotal)
	r.db.Model(&domain.Referral{}).Where("tenant_id = ? AND status = ? AND referral_type IN ('draw', 'lottery')", tenantID, "APPROVED").Count(&stats.DrawTotal)
	r.db.Model(&domain.Referral{}).Where("tenant_id = ? AND status = ? AND referral_type = ?", tenantID, "APPROVED", "stars").Count(&stats.StarsTotal)
	r.db.Model(&domain.Referral{}).Where("tenant_id = ? AND status = ? AND referral_type = ?", tenantID, "APPROVED", "gift").Count(&stats.GiftTotal)
	return stats, nil
}
