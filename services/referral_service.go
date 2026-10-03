package services

import (
	"OmniGram/domain"
	"OmniGram/repository"
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

	var rewardPoint float64 = 0
	var rewardStars float64 = 0

	switch referral.ReferralType {
	case "stars":
		rewardStars = 1.0
	case "league", "lottery":
		rewardPoint = 1.0
	case "gift":
		rewardPoint = 2.0
	}

	err = s.UserRepo.AddBalances(tenantID, referral.ReferrerID, 0, rewardPoint, rewardStars)
	if err != nil {
		return err
	}

	referrerUser, err := s.UserRepo.GetUserByID(tenantID, referral.ReferrerID)
	if err == nil && s.BotClient != nil {
		successMsg := "تسک‌های زیرمجموعه شما تایید شد و پاداش بخش " + referral.ReferralType + " واریز گردید."
		_ = s.BotClient.SendMessage(referrerUser.TelegramID, successMsg)
	}

	return nil
}
