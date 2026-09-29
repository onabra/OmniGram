package telegram

import (
	"OmniGram/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type KeyboardBuilder struct {
	TranslationRepo *repository.TranslationRepo
}

// BuildMainMenu ساخت منوی اصلی بر اساس چیدمان درخواستی و بدون هاردکد[cite: 4]
func (kb *KeyboardBuilder) BuildMainMenu(tenantID uint, lang string) (tgbotapi.ReplyKeyboardMarkup, error) {
	// فراخوانی متون از دیتابیس/کش برای جلوگیری از هاردکد[cite: 4]
	btnWithdraw, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_withdraw")
	btnAccount, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_account")

	btnShop, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_shop")

	// دکمه جدید توسعه‌دهنده
	btnDeveloper, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_developer")

	btnTask, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_task")
	btnReferral, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_referral")

	btnFreeGift, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_free_gift")
	btnRefLeague, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_ref_league")

	// ساختار کیبورد ریپلای با چیدمان دقیق
	keyboard := tgbotapi.NewReplyKeyboard(
		// ردیف اول
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnWithdraw),
			tgbotapi.NewKeyboardButton(btnAccount),
		),
		// ردیف دوم
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnShop),
		),
		// ردیف سوم (دکمه بزرگ زرد رنگ توسعه‌دهنده - تمام عرض)
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnDeveloper),
		),
		// ردیف چهارم
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnTask),
			tgbotapi.NewKeyboardButton(btnReferral),
		),
		// ردیف پنجم
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnFreeGift),
			tgbotapi.NewKeyboardButton(btnRefLeague),
		),
	)

	keyboard.ResizeKeyboard = true
	return keyboard, nil
}

// BuildDeveloperInlineMenu ساخت دکمه شیشه‌ای حاوی لینک دایرکت شما
func (kb *KeyboardBuilder) BuildDeveloperInlineMenu(tenantID uint, lang string, developerURL string) (tgbotapi.InlineKeyboardMarkup, error) {
	btnText, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_contact_dev")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(btnText, developerURL),
		),
	)
	return keyboard, nil
}