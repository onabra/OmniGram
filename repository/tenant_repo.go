package repository

import (
	"OmniGram/domain" // نام ماژول خود را جایگزین کنید

	"gorm.io/gorm"
)

type TenantRepository struct {
	DB *gorm.DB
}

// GetByToken دریافت اطلاعات کارفرما بر اساس توکن ربات برای شناسایی مقصد وب‌هوک
func (r *TenantRepository) GetByToken(botToken string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.DB.Where("bot_token = ?", botToken).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}
