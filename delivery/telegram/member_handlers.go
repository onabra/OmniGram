package telegram

import (
	"OmniGram/domain"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleChatMemberUpdate بررسی لفت دادن کاربران از کانال اجباری و کانال‌های تسک اسپانسر
func (h *PollingHandler) HandleChatMemberUpdate(tenant *domain.Tenant, chatMember *tgbotapi.ChatMemberUpdated) {
	status := chatMember.NewChatMember.Status
	userID := chatMember.NewChatMember.User.ID

	// بررسی وضعیت لفت دادن یا اخراج شدن کاربر
	if status == "left" || status == "kicked" {
		user, err := h.UserRepo.GetUserByTelegramID(tenant.ID, userID)
		if err != nil || user == nil {
			return // کاربر در دیتابیس ربات ثبت نام نکرده است
		}

		// ۱. بررسی جریمه لفت دادن از کانال‌های تسک (Tasks & Penalties)
		if h.TaskRepo != nil {
			var identifiers []string
			identifiers = append(identifiers, fmt.Sprintf("%d", chatMember.Chat.ID))
			if chatMember.Chat.UserName != "" {
				identifiers = append(identifiers, "@"+chatMember.Chat.UserName)
			}

			sub, task, err := h.TaskRepo.GetApprovedSubmissionByChannel(tenant.ID, user.ID, identifiers)
			if err == nil && sub != nil && task != nil && task.PenaltyAmount > 0 {
				// بررسی زمان ماندگاری (اگر StayDuration == 0 باشد همیشه شامل جریمه است، در غیر این صورت تا پایان ساعت تعیین شده)
				shouldPenalize := task.StayDuration <= 0 || time.Since(sub.CreatedAt).Hours() < float64(task.StayDuration)
				if shouldPenalize {
					var penToman, penPoint float64
					if task.RewardType == "toman" {
						penToman = -task.PenaltyAmount
					} else {
						penPoint = -task.PenaltyAmount
					}
					_ = h.UserRepo.AddBalances(tenant.ID, user.ID, penToman, penPoint, 0)

					sub.Status = "PENALIZED"
					_ = h.TaskRepo.UpdateSubmissionStatus(sub)

					if h.SafeBot != nil {
						msg := fmt.Sprintf("⚠️ شما پیش از پایان زمان ماندگاری از کانال تسک «%s» خارج شدید!\n⚖️ مبلغ **%.0f %s** به عنوان جریمه از حساب شما کسر گردید.",
							task.Title, task.PenaltyAmount, task.RewardType)
						_ = h.SafeBot.SendMessage(user.TelegramID, msg)
					}
				}
			}
		}

		// ۲. بررسی جریمه معرف در صورت لفت دادن زیرمجموعه از کانال‌های جوین اجباری
		referral, err := h.RefRepo.GetReferrer(tenant.ID, user.ID)
		if err == nil && referral != nil {
			var penaltyPoint float64 = 0
			var penaltyStars float64 = 0

			switch referral.ReferralType {
			case "stars":
				penaltyStars = -1.0
			case "league", "lottery", "draw", "point":
				penaltyPoint = -1.0
			case "gift":
				penaltyPoint = -2.0
			}

			_ = h.UserRepo.AddBalances(tenant.ID, referral.ReferrerID, 0, penaltyPoint, penaltyStars)

			referrerUser, err := h.UserRepo.GetUserByID(tenant.ID, referral.ReferrerID)
			if err == nil && h.SafeBot != nil {
				msg := "⚠️ زیرمجموعه شما از کانال‌های اجباری لفت داد. امتیاز رفرال مربوطه از حساب شما کسر شد. در صورت بازگشت کاربر، امتیاز برمی‌گردد."
				_ = h.SafeBot.SendMessage(referrerUser.TelegramID, msg)
			}

			referral.Status = "REJECTED"
			_ = h.RefRepo.UpdateStatus(referral)

			log.Printf("User %d left the channel. Penalty applied to referrer %d.", user.TelegramID, referral.ReferrerID)
		}
	}
}

//// HandleChatMemberUpdate بررسی لفت دادن کاربران از کانال اجباری
//func (h *PollingHandler) HandleChatMemberUpdate(tenant *domain.Tenant, chatMember *tgbotapi.ChatMemberUpdated) {
//	status := chatMember.NewChatMember.Status
//	userID := chatMember.NewChatMember.User.ID
//
//	// بررسی وضعیت لفت دادن یا اخراج شدن کاربر
//	if status == "left" || status == "kicked" {
//
//		user, err := h.UserRepo.GetUserByTelegramID(tenant.ID, userID)
//		if err != nil {
//			return // کاربر در دیتابیس ربات ثبت نام نکرده است
//		}
//
//		referral, err := h.RefRepo.GetReferrer(tenant.ID, user.ID)
//		if err == nil && referral != nil {
//
//			// تعیین مبلغ جریمه بر اساس نوع لینک رفرال
//			var penaltyPoint float64 = 0
//			var penaltyStars float64 = 0
//
//			switch referral.ReferralType {
//			case "stars":
//				penaltyStars = -1.0
//			case "league", "lottery":
//				penaltyPoint = -1.0
//			case "gift":
//				penaltyPoint = -2.0
//			}
//
//			// کسر موجودی از معرف (ارسال مقادیر منفی به متد AddBalances باعث کسر شدن می‌شود)
//			_ = h.UserRepo.AddBalances(tenant.ID, referral.ReferrerID, 0, penaltyPoint, penaltyStars)
//
//			// اطلاع‌رسانی به معرف که زیرمجموعه‌اش لفت داده است
//			referrerUser, err := h.UserRepo.GetUserByID(tenant.ID, referral.ReferrerID)
//			if err == nil && h.SafeBot != nil {
//				msg := "⚠️ زیرمجموعه شما از کانال‌های اجباری لفت داد. امتیاز رفرال مربوطه از حساب شما کسر شد. در صورت بازگشت کاربر، امتیاز برمی‌گردد."
//				_ = h.SafeBot.SendMessage(referrerUser.TelegramID, msg)
//			}
//
//			// تغییر وضعیت رفرال به REJECTED تا دیگر به عنوان زیرمجموعه فعال شناخته نشود
//			referral.Status = "REJECTED"
//			_ = h.RefRepo.UpdateStatus(referral)
//
//			log.Printf("User %d left the channel. Penalty applied to referrer %d.", user.TelegramID, referral.ReferrerID)
//		}
//		log.Printf("User %d left the channel. Penalty applied.", user.TelegramID)
//	}
//}
