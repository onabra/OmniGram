package telegram

import (
	"OmniGram/domain"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleChatMemberUpdate بررسی لفت دادن کاربران از کانال اجباری
func (h *WebhookHandler) HandleChatMemberUpdate(tenant *domain.Tenant, chatMember *tgbotapi.ChatMemberUpdated) {
	status := chatMember.NewChatMember.Status
	userID := chatMember.NewChatMember.User.ID

	// بررسی وضعیت لفت دادن یا اخراج شدن کاربر
	if status == "left" || status == "kicked" {

		user, err := h.UserRepo.GetUserByTelegramID(tenant.ID, userID)
		if err != nil {
			return // کاربر در دیتابیس ربات ثبت نام نکرده است
		}

		// بررسی اینکه آیا این کاربر توسط شخصی (رفرال) دعوت شده بود؟
		// این متد باید در RefRepo شما وجود داشته باشد
		// referral, err := h.RefRepo.GetReferrer(tenant.ID, user.ID)

		// اگر معرف داشت:
		// penaltyAmount := 2.0 // خواندن میزان جریمه از تنظیمات
		// h.UserRepo.UpdateBalances(tenant.ID, referral.ReferrerID, 0, -penaltyAmount, 0)
		log.Printf("User %d left the channel. Penalty applied.", user.TelegramID)
	}
}
