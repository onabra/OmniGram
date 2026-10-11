package repository

import (
	"OmniGram/domain"
	"time"

	"gorm.io/gorm"
)

type OrderRepository struct {
	DB *gorm.DB
}

// TopCustomerResult ساختار خروجی مشتریان برتر
type TopCustomerResult struct {
	UserID     uint
	FirstName  string
	TotalSpent float64
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

// GetCustomerRankingsLast30Days واکشی و مرتب‌سازی تمام مشتریان بر اساس خرید ۳۰ روز گذشته
func (r *OrderRepository) GetCustomerRankingsLast30Days(tenantID uint) ([]TopCustomerResult, error) {
	var results []TopCustomerResult
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err := r.DB.Table("orders").
		Select("users.id as user_id, users.first_name, SUM(orders.total_amount) as total_spent").
		Joins("JOIN users ON users.id = orders.user_id").
		Where("orders.tenant_id = ? AND orders.status = ? AND orders.created_at >= ?", tenantID, "COMPLETED", thirtyDaysAgo).
		Group("users.id, users.first_name").
		Order("total_spent DESC").
		Scan(&results).Error
	return results, err
}
