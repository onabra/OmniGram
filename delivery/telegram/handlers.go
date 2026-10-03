package telegram

import (
	"OmniGram/domain"
	"OmniGram/repository"
	"OmniGram/services"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	TenantRepo      *repository.TenantRepository
	TranslationRepo *repository.TranslationRepo
	KeyboardBuilder *KeyboardBuilder
	SafeBot         *SafeBot
	ReferralService *services.ReferralService
}

// HandleStart مدیریت دستور /start
func (h *BotHandler) HandleStart(tenant *domain.Tenant, chatID int64, lang string, text string) {

	// پردازش لینک رفرال (در صورت وجود)
	parts := strings.Split(text, " ")
	if len(parts) > 1 && strings.HasPrefix(parts[1], "ref_") {
		refPayload := strings.TrimPrefix(parts[1], "ref_")
		payloadParts := strings.Split(refPayload, "_")

		if len(payloadParts) == 2 {
			refType := payloadParts[0] // نوع رفرال (stars, gift, league, lottery)
			referrerID, err := strconv.ParseUint(payloadParts[1], 10, 32)

			if err == nil && h.ReferralService != nil {
				// ثبت رفرال به صورت معلق (PENDING)
				_ = h.ReferralService.ProcessReferral(tenant.ID, uint(referrerID), chatID, refType)
			}
		}
	}
	// گرفتن پیام خوش‌آمدگویی بدون هاردکد[cite: 4]
	welcomeMsg, err := h.TranslationRepo.GetText(tenant.ID, lang, "msg_welcome")
	if err != nil {
		log.Println("Error getting translation:", err)
		welcomeMsg = "Welcome!" // فال‌بک ضروری
	}

	// ساخت منوی اصلی
	mainMenu, err := h.KeyboardBuilder.BuildMainMenu(tenant.ID, lang)
	if err != nil {
		log.Println("Error building menu:", err)
	}

	// ارسال امن پیام به همراه کیبورد
	msg := tgbotapi.NewMessage(chatID, welcomeMsg)
	msg.ReplyMarkup = mainMenu

	if _, err := h.SafeBot.Bot.Send(msg); err != nil {
		h.SafeBot.sendLogToChannel(err.Error())
	}
}

// HandleMessage مدیریت پیام‌های متنی ارسالی از سمت کاربر
func (h *BotHandler) HandleMessage(tenant *domain.Tenant, chatID int64, text string, lang string) {
	// فراخوانی عنوان دکمه بدون هاردکد[cite: 4]
	devBtnText, _ := h.TranslationRepo.GetText(tenant.ID, lang, "btn_developer")

	switch text {
	case devBtnText:
		// فراخوانی متن معرفی از کش/دیتابیس[cite: 4]
		devDesc, _ := h.TranslationRepo.GetText(tenant.ID, lang, "msg_dev_description")

		// خواندن لینک دایرکت توسعه‌دهنده از دیتابیس به جای هاردکد کردن در کد[cite: 4]
		developerURL, err := h.TranslationRepo.GetText(tenant.ID, lang, "url_developer")
		if err != nil || developerURL == "" {
			developerURL = "https://t.me/default_admin" // فال‌بک نهایی در صورت خالی بودن دیتابیس
		}

		inlineMenu, _ := h.KeyboardBuilder.BuildDeveloperInlineMenu(tenant.ID, lang, developerURL)

		msg := tgbotapi.NewMessage(chatID, devDesc)
		msg.ReplyMarkup = inlineMenu

		_, _ = h.SafeBot.Bot.Send(msg)

	default:
		// سایر پردازش‌ها
	}
}

//func (h *BotHandler) HandleMessage(tenant *domain.Tenant, chatID int64, text string, lang string) {
//	// فراخوانی عنوان دکمه توسعه‌دهنده برای مقایسه با پیام ارسالی
//	devBtnText, _ := h.TranslationRepo.GetText(tenant.ID, lang, "btn_developer")
//
//	switch text {
//	case devBtnText:
//		// فراخوانی متن معرفی توسعه‌دهنده از دیتابیس
//		devDesc, _ := h.TranslationRepo.GetText(tenant.ID, lang, "msg_dev_description")
//
//		// لینک دایرکت تلگرام شما (می‌تواند از دیتابیس خوانده شود یا در تنظیمات تنظیم شود)
//		developerURL := "https://t.me/YourDirectUsername"
//
//		inlineMenu, _ := h.KeyboardBuilder.BuildDeveloperInlineMenu(tenant.ID, lang, developerURL)
//
//		msg := tgbotapi.NewMessage(chatID, devDesc)
//		msg.ReplyMarkup = inlineMenu
//
//		_, _ = h.SafeBot.Bot.Send(msg)
//
//	// سایر دکمه‌ها در اینجا هندل می‌شوند...
//	default:
//		// ...
//	}
//}
