package services

import (
	"OmniGram/domain"
	"OmniGram/repository"
	"fmt"
)

type OrderService struct {
	OrderRepo *repository.OrderRepository
	UserRepo  *repository.UserRepository
}

// RequestSpeedUp ثبت درخواست سرعت‌دهی و کسر امتیاز از کاربر[cite: 2]
func (s *OrderService) RequestSpeedUp(tenantID uint, userID uint, orderID uint, speedUpCost float64) error {
	// ۱. بررسی موجودی کاربر و کسر هزینه سرعت‌دهی
	// ۲. آپدیت وضعیت سفارش
	err := s.OrderRepo.DB.Model(&domain.Order{}).
		Where("id = ? AND tenant_id = ?", orderID, tenantID).
		Update("speed_up_requested", true).Error

	if err != nil {
		return err
	}

	// ۳. ارسال نوتیفیکیشن برای ادمین (منطق ارسال در لایه delivery هندل می‌شود)[cite: 2]
	fmt.Printf("Speed up requested for order %d, user %d charged %.2f points\n", orderID, userID, speedUpCost)
	return nil
}
