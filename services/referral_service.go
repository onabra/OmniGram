package services

import (
	"OmniGram/domain"
	"OmniGram/repository"
	"fmt"
)

// MessageSender یک اینترفیس برای جلوگیری از وابستگی مستقیم به SafeBot
type MessageSender interface {
	SendMessage(chatID int64, text string) error
}

type ReferralService struct {
	UserRepo  repository.UserRepository
	RefRepo   *repository.ReferralRepository
	BotClient MessageSender // جایگزین TelegramClient شد
}

func (s *ReferralService) ProcessReferral(tenantID uint, referrerID uint, referredTelegramID int64, refType string) error {
	referredUser, err := s.UserRepo.GetUserByTelegramID(tenantID, referredTelegramID)
	if err != nil {
		return err
	}

	// دریافت اطلاعات معرف برای ارسال پیام
	referrerUser, err := s.UserRepo.GetUserByID(tenantID, referrerID)
	if err != nil {
		return err
	}

	newReferral := &domain.Referral{
		TenantID:     tenantID,
		ReferrerID:   referrerID,
		ReferredID:   referredUser.ID,
		ReferralType: refType,
		Status:       "PENDING",
	}

	err = s.RefRepo.Create(newReferral)
	if err != nil {
		return err
	}

	// استفاده از اینترفیس برای ارسال پیام
	if s.BotClient != nil {
		msg := "کاربر " + referredUser.FirstName + " از طریق لینک شما وارد بات شد. پس از انجام وظایف اجباری، امتیاز شما تایید خواهد شد."
		_ = s.BotClient.SendMessage(referrerUser.TelegramID, msg)
	}

	return nil
}

func (s *ReferralService) ApproveReferral(tenantID uint, referredUserID uint) error {
	referral, err := s.RefRepo.GetPendingReferral(tenantID, referredUserID)
	if err != nil || referral == nil {
		return nil
	}

	referral.Status = "APPROVED"
	err = s.RefRepo.UpdateStatus(referral)
	if err != nil {
		return err
	}

	referrerUser, err := s.UserRepo.GetUserByID(tenantID, referral.ReferrerID)
	if err != nil {
		return err
	}

	successMsg := "✅ یک زیرمجموعه جدید برای شما تایید شد:\n\n"

	switch referral.ReferralType {
	case "league":
		successMsg += "🏆 شما یک رفرال معتبر در **لیگ** کسب کردید!"
	case "draw":
		stats, _ := s.RefRepo.GetReferralStats(tenantID, referral.ReferrerID)
		if stats.DrawTotal > 0 && stats.DrawTotal%3 == 0 {
			referrerUser.DrawTickets += 1
			_ = s.UserRepo.UpdateDrawTickets(tenantID, referrerUser.ID, referrerUser.DrawTickets)
			successMsg += "🎰 شما 1 تیکت قرعه‌کشی جدید دریافت کردید! 🎟"
		} else {
			successMsg += fmt.Sprintf("🎰 رفرال قرعه‌کشی ثبت شد. (%d رفرال تا دریافت تیکت بعدی)", 3-(stats.DrawTotal%3))
		}
	case "point":
		referrerUser.PointBalance += 1.0
		_ = s.UserRepo.UpdateBalances(tenantID, referrerUser.ID, referrerUser.TomanBalance, referrerUser.PointBalance, referrerUser.StarsBalance)
		successMsg += "⭐ شما ۱ امتیاز دریافت کردید!"
	case "wallet":
		referrerUser.TomanBalance += 2000.0 // این مبلغ می‌تواند در آینده از تنظیمات ادمین خوانده شود
		_ = s.UserRepo.UpdateBalances(tenantID, referrerUser.ID, referrerUser.TomanBalance, referrerUser.PointBalance, referrerUser.StarsBalance)
		successMsg += "💰 مبلغ 2,000 تومان به کیف پول شما اضافه شد!"
	}

	if s.BotClient != nil {
		_ = s.BotClient.SendMessage(referrerUser.TelegramID, successMsg)
	}

	return nil
}
