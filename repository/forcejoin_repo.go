package repository

import (
	"OmniGram/domain"

	"gorm.io/gorm"
)

type ForceJoinRepository struct {
	DB *gorm.DB
}

// GetActiveForceJoins واکشی تمام کانال‌های جوین اجباری فعال
func (r *ForceJoinRepository) GetActiveForceJoins(tenantID uint) ([]domain.ForceJoin, error) {
	var channels []domain.ForceJoin
	err := r.DB.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&channels).Error
	return channels, err
}

// Create افزودن کانال اجباری جدید
func (r *ForceJoinRepository) Create(fj *domain.ForceJoin) error {
	return r.DB.Create(fj).Error
}

// Delete حذف کانال اجباری
func (r *ForceJoinRepository) Delete(tenantID uint, id uint) error {
	return r.DB.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&domain.ForceJoin{}).Error
}
