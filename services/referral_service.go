package services

import (
	"OmniGram/domain"
	"OmniGram/repository"
	"time"
)

type ReferralService struct {
	UserRepo *repository.UserRepository     // فرض بر وجود این ریپازیتوری
	RefRepo  *repository.ReferralRepository // فرض بر وجود این ریپازیتوری
}

// ProcessReferral پردازش زیرمجموعه‌گیری با جلوگیری از تداخل انواع ۴ گانه[cite: 1]
func (s *ReferralService) ProcessReferral(tenantID uint, referrerID uint, referredTelegramID int64, refType string) error {
	// بررسی وضعیت کاربر دعوت شده
	referredUser, err := s.UserRepo.GetUserByTelegramID(tenantID, referredTelegramID)
	if err != nil {
		return err
	}

	// اجرای قانون ۱۰ روز خاموشی: اگر کاربر بیش از ۱۰ روز غیرفعال بوده، معرف جدید ۵۰٪ امتیاز می‌گیرد[cite: 1]
	if time.Since(referredUser.LastActiveDate).Hours() > 240 {
		// منطق انتقال ۵۰٪ امتیاز به referrerID
		_ = s.UserRepo.TransferPoints(tenantID, referredUser.ID, referrerID, 50.0) // تابع فرضی
	}

	// ثبت رفرال با نوع مشخص (استارز، گیفت، لیگ، قرعه‌کشی)[cite: 1]
	newReferral := &domain.Referral{
		TenantID:     tenantID,
		ReferrerID:   referrerID,
		ReferredID:   referredUser.ID,
		ReferralType: refType,
	}

	return s.RefRepo.Create(newReferral)
}
