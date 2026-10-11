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
