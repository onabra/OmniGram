package repository

import (
	"OmniGram/domain"

	"gorm.io/gorm"
)

type DrawRepository struct {
	DB *gorm.DB
}

// GetActiveDraws واکشی تمام قرعه‌کشی‌های فعال یک کارفرما
func (r *DrawRepository) GetActiveDraws(tenantID uint) ([]domain.LuckyDraw, error) {
	var draws []domain.LuckyDraw
	err := r.DB.Where("tenant_id = ? AND status = ?", tenantID, "ACTIVE").Find(&draws).Error
	return draws, err
}

// GetDrawByID گرفتن اطلاعات یک قرعه‌کشی خاص
func (r *DrawRepository) GetDrawByID(tenantID uint, drawID uint) (*domain.LuckyDraw, error) {
	var draw domain.LuckyDraw
	err := r.DB.Where("tenant_id = ? AND id = ?", tenantID, drawID).First(&draw).Error
	return &draw, err
}

// GetTotalEntries مجموع شانس‌های ثبت شده در یک قرعه‌کشی
func (r *DrawRepository) GetTotalEntries(tenantID uint, drawID uint) (int64, error) {
	var total int64
	err := r.DB.Model(&domain.DrawEntry{}).Where("tenant_id = ? AND lucky_draw_id = ?", tenantID, drawID).Select("COALESCE(SUM(entries), 0)").Scan(&total).Error
	return total, err
}

// AddEntry ثبت شرکت‌کننده در قرعه‌کشی
func (r *DrawRepository) AddEntry(entry *domain.DrawEntry) error {
	var existing domain.DrawEntry
	// اگر قبلاً شرکت کرده، شانس‌هایش را افزایش بده
	err := r.DB.Where("tenant_id = ? AND lucky_draw_id = ? AND user_id = ?", entry.TenantID, entry.LuckyDrawID, entry.UserID).First(&existing).Error
	if err == nil {
		return r.DB.Model(&existing).UpdateColumn("entries", gorm.Expr("entries + ?", entry.Entries)).Error
	}
	// در غیر این صورت یک رکورد جدید بساز
	return r.DB.Create(entry).Error
}

// [repository/draw_repo.go]

// UpdateDraw ذخیره وضعیت جدید قرعه‌کشی
func (r *DrawRepository) UpdateDraw(draw *domain.LuckyDraw) error {
	return r.DB.Save(draw).Error
}

// GetDrawEntries واکشی تمام شانس‌های ثبت شده برای یک قرعه‌کشی خاص
func (r *DrawRepository) GetDrawEntries(tenantID uint, drawID uint) ([]domain.DrawEntry, error) {
	var entries []domain.DrawEntry
	err := r.DB.Where("tenant_id = ? AND lucky_draw_id = ?", tenantID, drawID).Find(&entries).Error
	return entries, err
}

// CreateDraw ایجاد یک قرعه‌کشی جدید توسط ادمین
func (r *DrawRepository) CreateDraw(draw *domain.LuckyDraw) error {
	return r.DB.Create(draw).Error
}
