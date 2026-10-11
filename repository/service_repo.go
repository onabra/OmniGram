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

// GetServicesByCategory واکشی سرویس‌های یک دسته‌بندی خاص برای فروشگاه
func (r *ServiceRepository) GetServicesByCategory(tenantID uint, categoryID uint) ([]domain.ServiceItem, error) {
	var services []domain.ServiceItem
	err := r.DB.Where("tenant_id = ? AND category_id = ?", tenantID, categoryID).Find(&services).Error
	return services, err
}

// CreateService ساخت سرویس جدید
func (r *ServiceRepository) CreateService(service *domain.ServiceItem) error {
	return r.DB.Create(service).Error
}

// UpdateAllPricesByDollar آپدیت قیمت تمامی سرویس‌ها بر اساس فرمول و قیمت جدید دلار
// فرمول: قیمت تومانی = قیمت دلاری × قیمت دلار کارفرما × (1 + درصد سود/100)
func (r *ServiceRepository) UpdateAllPricesByDollar(tenantID uint, dollarPrice float64) error {
	// استفاده از کوئری مستقیم دیتابیس برای آپدیت دسته‌جمعی و افزایش سرعت
	return r.DB.Model(&domain.ServiceItem{}).
		Where("tenant_id = ?", tenantID).
		UpdateColumn("price_toman", gorm.Expr("price_usd * ? * (1.0 + (profit_percent / 100.0))", dollarPrice)).Error
}
