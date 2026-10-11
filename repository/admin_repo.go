// [repository/admin_repo.go]
package repository

import (
	"OmniGram/domain"
	"time"

	"gorm.io/gorm"
)

type AdminRepository struct {
	DB *gorm.DB
}

type SystemStats struct {
	TotalUsers   int64
	ActiveUsers  int64
	TotalOrders  int64
	TotalRevenue float64
	TotalPoints  float64
	TotalStars   float64
}

// GetStats محاسبه آمار زنده سیستم برای پنل ادمین
func (r *AdminRepository) GetStats(tenantID uint) (*SystemStats, error) {
	stats := &SystemStats{}

	r.DB.Model(&domain.User{}).Where("tenant_id = ?", tenantID).Count(&stats.TotalUsers)

	tenDaysAgo := time.Now().AddDate(0, 0, -10)
	r.DB.Model(&domain.User{}).Where("tenant_id = ? AND last_active_date >= ?", tenantID, tenDaysAgo).Count(&stats.ActiveUsers)

	r.DB.Model(&domain.Order{}).Where("tenant_id = ?", tenantID).Count(&stats.TotalOrders)

	r.DB.Model(&domain.Order{}).Where("tenant_id = ? AND status = ?", tenantID, "COMPLETED").Select("COALESCE(SUM(total_amount), 0)").Scan(&stats.TotalRevenue)
	r.DB.Model(&domain.User{}).Where("tenant_id = ?", tenantID).Select("COALESCE(SUM(point_balance), 0)").Scan(&stats.TotalPoints)
	r.DB.Model(&domain.User{}).Where("tenant_id = ?", tenantID).Select("COALESCE(SUM(stars_balance), 0)").Scan(&stats.TotalStars)

	return stats, nil
}

// UpdateTenantSetting بروزرسانی تنظیمات اصلی کارفرما
func (r *AdminRepository) UpdateTenantSetting(tenant *domain.Tenant) error {
	return r.DB.Save(tenant).Error
}
