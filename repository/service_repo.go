package repository

import (
	"OmniGram/domain"

	"gorm.io/gorm"
)

type ServiceRepository struct {
	DB *gorm.DB
}

// GetServiceByKey واکشی اطلاعات سرویس بر اساس کلید نام (مثلا کیفیت شاپ) و آیدی کارفرما
func (r *ServiceRepository) GetServiceByKey(tenantID uint, nameKey string) (*domain.ServiceItem, error) {
	var service domain.ServiceItem
	err := r.DB.Where("tenant_id = ? AND name_key = ?", tenantID, nameKey).First(&service).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}

// GetServiceByID واکشی سرویس بر اساس آیدی جهت محاسبه قیمت نهایی
func (r *ServiceRepository) GetServiceByID(tenantID uint, id uint) (*domain.ServiceItem, error) {
	var service domain.ServiceItem
	err := r.DB.Where("tenant_id = ? AND id = ?", tenantID, id).First(&service).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}
