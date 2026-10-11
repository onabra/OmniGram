package repository

import (
	"OmniGram/domain"
	"time"

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

func (r *OrderRepository) UpdateOrder(order *domain.Order) error {
	return r.DB.Save(order).Error
}

// GetDraftOrder دریافت سفارش پیش‌نویس کاربر با رعایت ایزوله‌سازی
func (r *OrderRepository) GetDraftOrder(tenantID uint, userID uint) (*domain.Order, error) {
	var order domain.Order
	err := r.DB.Where("tenant_id = ? AND user_id = ? AND status = ?", tenantID, userID, "DRAFT").First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetOrderByID(tenantID uint, id uint) (*domain.Order, error) {
	var order domain.Order
	err := r.DB.Where("tenant_id = ? AND id = ?", tenantID, id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetUserTotalSpentLast30Days محاسبه مجموع خرید کاربر در ۳۰ روز گذشته برای سیستم VIP تخفیف
func (r *OrderRepository) GetUserTotalSpentLast30Days(tenantID uint, userID uint) (float64, error) {
	var total float64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err := r.DB.Model(&domain.Order{}).
		Where("tenant_id = ? AND user_id = ? AND status = ? AND created_at >= ?", tenantID, userID, "COMPLETED", thirtyDaysAgo).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&total).Error
	return total, err
}
