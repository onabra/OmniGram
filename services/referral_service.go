package services

import (
	"OmniGram/domain"
	"OmniGram/repository"
	"errors"
	"fmt"
	"time"
)

// MessageSender یک اینترفیس برای جلوگیری از وابستگی مستقیم به SafeBot
type MessageSender interface {
	SendMessage(chatID int64, text string) error
}

type ReferralService struct {
	UserRepo  repository.UserRepository
	RefRepo   *repository.ReferralRepository
	BotClient MessageSender
}

func (s *ReferralService) ProcessReferral(tenant *domain.Tenant, referrerID uint, referredTelegramID int64, refType string, isNewUser bool) error {
	referredUser, err := s.UserRepo.GetUserByTelegramID(tenant.ID, referredTelegramID)
	if err != nil || referredUser == nil {
		return err
	}

	// ۱. جلوگیری از دعوت خود
	if referrerID == referredUser.ID {
		return errors.New("SELF_REFERRAL")
	}

	referrerUser, err := s.UserRepo.GetUserByID(tenant.ID, referrerID)
	if err != nil || referrerUser == nil {
		return err
	}

	// ۲. بررسی وضعیت کاربر (کاربر کاملاً جدید یا مشمول قانون ۱۰ روز خاموشی)
	isReactivation := false
	if !isNewUser {
		inactiveDays := time.Since(referredUser.LastActiveDate).Hours() / 24.0
		if !referredUser.LastActiveDate.IsZero() && inactiveDays >= 10.0 {
			isReactivation = true
		} else if s.RefRepo.HasEverBeenReferred(tenant.ID, referredUser.ID) {
			return errors.New("ALREADY_REFERRED")
		} else {
			// کاربری که قبلاً بدون رفرال در ربات بوده و فعال است، زیرمجموعه جدید محسوب نمی‌شود
			return errors.New("ALREADY_ACTIVE_USER")
		}
	}

	newReferral := &domain.Referral{
		TenantID:       tenant.ID,
		ReferrerID:     referrerID,
		ReferredID:     referredUser.ID,
		ReferralType:   refType,
		IsReactivation: isReactivation,
		Status:         "PENDING",
		CreatedAt:      time.Now(),
	}

	err = s.RefRepo.Create(newReferral)
	if err != nil {
		return err
	}

	if s.BotClient != nil {
		msg := fmt.Sprintf("👤 کاربر %s (آیدی: `%d`) از طریق لینک شما وارد ربات شد.\n⏳ پس از تأیید نهایی، پاداش شما واریز خواهد شد.", referredUser.FirstName, referredUser.TelegramID)
		_ = s.BotClient.SendMessage(referrerUser.TelegramID, msg)
	}

	return nil
}

func (s *ReferralService) ApproveReferral(tenant *domain.Tenant, referredUserID uint) error {
	referral, err := s.RefRepo.GetPendingReferral(tenant.ID, referredUserID)
	if err != nil || referral == nil {
		return nil
	}

	referral.Status = "APPROVED"
	err = s.RefRepo.UpdateStatus(referral)
	if err != nil {
		return err
	}

	referrerUser, err := s.UserRepo.GetUserByID(tenant.ID, referral.ReferrerID)
	if err != nil || referrerUser == nil {
		return err
	}

	// گرفتن اطلاعات کاربری که تازه عضو شده تا آیدی او را به معرف نشان دهیم
	referredUser, _ := s.UserRepo.GetUserByID(tenant.ID, referral.ReferredID)
	refInfo := "نامشخص"
	if referredUser != nil {
		refInfo = fmt.Sprintf("آیدی: `%d`", referredUser.TelegramID)
	}

	// ضریب پاداش (۱۰۰٪ برای کاربر جدید، ۵۰٪ برای کاربر غیرفعال بیش از ۱۰ روز طبق معماری)
	multiplier := 1.0
	if referral.IsReactivation {
		multiplier = 0.5
	}

	successMsg := fmt.Sprintf("✅ یک زیرمجموعه جدید (%s) برای شما تأیید شد:\n\n", refInfo)
	if referral.IsReactivation {
		successMsg = fmt.Sprintf("♻️ کاربر بازگشتی (%s) پس از ۱۰ روز خاموشی با لینک شما وارد شد (۵۰٪ پاداش):\n\n", refInfo)
	}
	// ضریب پاداش (۱۰۰٪ برای کاربر جدید، ۵۰٪ برای کاربر غیرفعال بیش از ۱۰ روز طبق معماری)
	//multiplier := 1.0
	//if referral.IsReactivation {
	//	multiplier = 0.5
	//}
	//
	//successMsg := "✅ یک زیرمجموعه جدید برای شما تایید شد:\n\n"
	//if referral.IsReactivation {
	//	successMsg = "♻️ یک کاربر غیرفعال (بیش از ۱۰ روز خاموشی) با لینک شما به ربات بازگشت (۵۰٪ پاداش):\n\n"
	//}

	switch referral.ReferralType {
	case "league":
		successMsg += "🏆 شما یک دعوت معتبر در **لیگ رفرال** کسب کردید!"
	case "draw", "lottery":
		stats, _ := s.RefRepo.GetReferralStats(tenant.ID, referral.ReferrerID)
		if stats != nil && stats.DrawTotal > 0 && stats.DrawTotal%3 == 0 {
			referrerUser.DrawTickets += 1
			_ = s.UserRepo.UpdateDrawTickets(tenant.ID, referrerUser.ID, referrerUser.DrawTickets)
			successMsg += "🎰 شما ۱ تیکت قرعه‌کشی جدید دریافت کردید! 🎟"
		} else if stats != nil {
			successMsg += fmt.Sprintf("🎰 رفرال قرعه‌کشی ثبت شد. (%d دعوت دیگر تا دریافت تیکت بعدی)", 3-(stats.DrawTotal%3))
		}
	case "point":
		reward := 1.0 * multiplier
		if referral.IsReactivation && tenant.InactiveBonusAmount > 0 {
			reward = tenant.InactiveBonusAmount
		}
		_ = s.UserRepo.AddBalances(tenant.ID, referrerUser.ID, 0, reward, 0)
		successMsg += fmt.Sprintf("⭐ مقدار %.1f امتیاز به حساب شما اضافه شد!", reward)
	case "stars":
		reward := 1.0 * multiplier
		_ = s.UserRepo.AddBalances(tenant.ID, referrerUser.ID, 0, 0, reward)
		successMsg += fmt.Sprintf("🌟 مقدار %.1f استارز به حساب شما اضافه شد!", reward)
	case "gift":
		reward := 2.0 * multiplier
		_ = s.UserRepo.AddBalances(tenant.ID, referrerUser.ID, 0, reward, 0)
		successMsg += fmt.Sprintf("🎁 مقدار %.1f امتیاز گیفت به حساب شما اضافه شد!", reward)
	case "wallet":
		reward := 2000.0 * multiplier
		_ = s.UserRepo.AddBalances(tenant.ID, referrerUser.ID, reward, 0, 0)
		successMsg += fmt.Sprintf("💰 مبلغ %.0f تومان به کیف پول شما اضافه شد!", reward)
	}

	if s.BotClient != nil {
		_ = s.BotClient.SendMessage(referrerUser.TelegramID, successMsg)
	}

	return nil
}
