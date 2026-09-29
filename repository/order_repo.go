package repository

import (
	"OmniGram/domain"

	"gorm.io/gorm"
)

type OrderRepository struct {
	DB *gorm.DB
}

// CreateOrder ثبت سفارش جدید
func (r *OrderRepository) CreateOrder(order *domain.Order) error {
	return r.DB.Create(order).Error
}

// GetOrder پیگیری سفارش با رعایت ایزوله‌سازی کارفرما
func (r *OrderRepository) GetOrder(tenantID uint, trackingCode string) (*domain.Order, error) {
	var order domain.Order
	err := r.DB.Where("tenant_id = ? AND tracking_code = ?", tenantID, trackingCode).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// RequestCancel ثبت درخواست لغو با در نظر گرفتن ۱۰٪ کارمزد[cite: 2]
func (r *OrderRepository) RequestCancel(tenantID uint, orderID uint) error {
	return r.DB.Model(&domain.Order{}).
		Where("id = ? AND tenant_id = ?", orderID, tenantID).
		Update("cancel_requested", true).Error
}
