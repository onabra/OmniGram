package telegram

import (
	"OmniGram/repository"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type KeyboardBuilder struct {
	TranslationRepo *repository.TranslationRepo
}

// BuildMainMenu ساخت منوی اصلی بر اساس چیدمان درخواستی و بدون هاردکد[cite: 4]
func (kb *KeyboardBuilder) BuildMainMenu(tenantID uint, lang string) (tgbotapi.ReplyKeyboardMarkup, error) {

	getText := func(key, fallback string) string {
		text, err := kb.TranslationRepo.GetText(tenantID, lang, key)
		if err != nil || text == "" {
			return fallback
		}
		return text
	}
	// فراخوانی متون از دیتابیس/کش برای جلوگیری از هاردکد[cite: 4]
	//btnWithdraw, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_withdraw")
	//btnAccount, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_account")
	//
	//btnShop, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_shop")
	//
	//// دکمه جدید توسعه‌دهنده
	//btnDeveloper, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_developer")
	//
	//btnTask, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_task")
	//btnReferral, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_referral")
	//
	//btnFreeGift, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_free_gift")
	//btnRefLeague, _ := kb.TranslationRepo.GetText(tenantID, lang, "btn_ref_league")

	btnWithdraw := getText("btn_withdraw", "برداشت")
	btnAccount := getText("btn_account", "حساب کاربری")
	btnShop := getText("btn_shop", "فروشگاه")
	btnDeveloper := getText("btn_developer", "توسعه‌دهنده")
	btnTask := getText("btn_task", "وظایف")
	btnReferral := getText("btn_referral", "زیرمجموعه‌گیری")
	btnFreeGift := getText("btn_free_gift", "هدیه رایگان")
	btnRefLeague := getText("btn_ref_league", "لیگ")

	// ساختار کیبورد ریپلای با چیدمان دقیق
	//keyboard := tgbotapi.NewReplyKeyboard(
	//	// ردیف اول
	//	tgbotapi.NewKeyboardButtonRow(
	//		tgbotapi.NewKeyboardButton(btnWithdraw),
	//		tgbotapi.NewKeyboardButton(btnAccount),
	//	),
	//	// ردیف دوم
	//	tgbotapi.NewKeyboardButtonRow(
	//		tgbotapi.NewKeyboardButton(btnShop),
	//	),
	//	// ردیف سوم (دکمه بزرگ زرد رنگ توسعه‌دهنده - تمام عرض)
	//	tgbotapi.NewKeyboardButtonRow(
	//		tgbotapi.NewKeyboardButton(btnDeveloper),
	//	),
	//	// ردیف چهارم
	//	tgbotapi.NewKeyboardButtonRow(
	//		tgbotapi.NewKeyboardButton(btnTask),
	//		tgbotapi.NewKeyboardButton(btnReferral),
	//	),
	//	// ردیف پنجم
	//	tgbotapi.NewKeyboardButtonRow(
	//		tgbotapi.NewKeyboardButton(btnFreeGift),
	//		tgbotapi.NewKeyboardButton(btnRefLeague),
	//	),
	//)

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnWithdraw), tgbotapi.NewKeyboardButton(btnAccount)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnShop)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnDeveloper)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnTask), tgbotapi.NewKeyboardButton(btnReferral)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnFreeGift), tgbotapi.NewKeyboardButton(btnRefLeague)),
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

// BuildUserAccountMenu ساخت منوی حساب کاربری با چیدمان درخواستی
func (kb *KeyboardBuilder) BuildUserAccountMenu(tenantID uint, lang string) (tgbotapi.ReplyKeyboardMarkup, error) {
	getText := func(key, fallback string) string {
		text, err := kb.TranslationRepo.GetText(tenantID, lang, key)
		if err != nil || text == "" {
			return fallback
		}
		return text
	}

	btnBalance := getText("btn_balance", "موجودی")
	btnTransfer := getText("btn_transfer", "انتقال")
	btnRecentTxs := getText("btn_recent_txs", "تراکنشات اخیر")
	btnRefStats := getText("btn_ref_stats", "رفرال ها")
	btnSupport := getText("btn_support", "ارتباط با پشتیبانی")
	btnBack := getText("btn_back_main", "بازگشت به منوی اصلی")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnBalance), tgbotapi.NewKeyboardButton(btnTransfer)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnRecentTxs), tgbotapi.NewKeyboardButton(btnRefStats)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnSupport)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnBack)),
	)

	keyboard.ResizeKeyboard = true
	return keyboard, nil
}

// feat(keyboard): add sponsor quality and payment method inline keyboards

// BuildSponsorQualityMenu ساخت منوی انتخاب کیفیت ممبر برای اسپانسر
func (kb *KeyboardBuilder) BuildSponsorQualityMenu(tenantID uint, lang string) (tgbotapi.InlineKeyboardMarkup, error) {
	getText := func(key, fallback string) string {
		text, err := kb.TranslationRepo.GetText(tenantID, lang, key)
		if err != nil || text == "" {
			return fallback
		}
		return text
	}

	btnShop := getText("btn_quality_shop", "🌟 کیفیت بالا (شاپ)")
	btnNormal := getText("btn_quality_normal", "⭐ کیفیت متوسط")
	btnEco := getText("btn_quality_eco", "⚡ کیفیت اقتصادی")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnShop, "sponsor_quality_shop")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnNormal, "sponsor_quality_normal")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnEco, "sponsor_quality_eco")),
	)
	return keyboard, nil
}

// BuildPaymentMethodMenu ساخت منوی انتخاب روش پرداخت برای اسپانسر
func (kb *KeyboardBuilder) BuildPaymentMethodMenu(tenantID uint, lang string, orderID uint) (tgbotapi.InlineKeyboardMarkup, error) {
	getText := func(key, fallback string) string {
		text, err := kb.TranslationRepo.GetText(tenantID, lang, key)
		if err != nil || text == "" {
			return fallback
		}
		return text
	}

	btnWalletToman := getText("btn_pay_toman", "💳 کیف پول (تومان)")
	btnWalletPoint := getText("btn_pay_point", "🎁 کیف پول (امتیاز)")
	btnStars := getText("btn_pay_stars", "⭐ پرداخت استارز")
	btnGateway := getText("btn_pay_gateway", "🔗 درگاه پرداخت")

	idStr := fmt.Sprintf("%d", orderID) // تزریق آیدی سفارش به دیتای دکمه جهت رهگیری

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnWalletToman, "pay_toman_"+idStr)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnWalletPoint, "pay_point_"+idStr)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnStars, "pay_stars_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData(btnGateway, "pay_gate_"+idStr),
		),
	)
	return keyboard, nil
}

// [delivery/telegram/keyboards.go]
// feat(keyboard): add Tasks sub-menu containing Sponsor button

// BuildTasksSubMenu ساخت زیرمنوی وظایف شامل دکمه اسپانسر شدن
func (kb *KeyboardBuilder) BuildTasksSubMenu(tenantID uint, lang string) (tgbotapi.ReplyKeyboardMarkup, error) {
	getText := func(key, fallback string) string {
		text, err := kb.TranslationRepo.GetText(tenantID, lang, key)
		if err != nil || text == "" {
			return fallback
		}
		return text
	}

	btnDoTasks := getText("btn_do_tasks", "📝 انجام وظایف (کسب درآمد)")
	btnSponsor := getText("btn_sponsor", "📢 اسپانسر شدن (جذب ممبر)")
	btnBack := getText("btn_back_main", "بازگشت به منوی اصلی")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnDoTasks)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnSponsor)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnBack)),
	)

	keyboard.ResizeKeyboard = true
	return keyboard, nil
}
