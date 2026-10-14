package telegram

import (
	"OmniGram/domain"
	"OmniGram/repository"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type KeyboardBuilder struct {
	TranslationRepo *repository.TranslationRepo
}

// BuildMainMenu ساخت منوی اصلی بر اساس چیدمان درخواستی و بدون هاردکد[cite: 4]
func (kb *KeyboardBuilder) BuildMainMenu(tenantID uint, lang string, isAdmin bool) (tgbotapi.ReplyKeyboardMarkup, error) {

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
	btnAdmin := getText("btn_admin_panel", "⚙️ پنل مدیریت")

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

	if isAdmin {
		keyboard.Keyboard = append([][]tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnAdmin)),
		}, keyboard.Keyboard...)
	}

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

// BuildUserAccountMenu ساخت منوی حساب کاربری به همراه پیگیری سفارش و نرخ لحظه‌ای ارزها
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
	btnConvert := getText("btn_convert", "🔄 تبدیل موجودی")
	btnRecentTxs := getText("btn_recent_txs", "تراکنشات اخیر")
	btnTrackOrder := getText("btn_track_order", "🔍 پیگیری سفارش")
	btnCryptoPrices := getText("btn_crypto_prices", "💎 قیمت لحظه‌ای ارزها")
	btnRefStats := getText("btn_ref_stats", "رفرال ها")
	btnSupport := getText("btn_support", "ارتباط با پشتیبانی")
	btnBack := getText("btn_back_main", "بازگشت به منوی اصلی")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnBalance), tgbotapi.NewKeyboardButton(btnTransfer)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnConvert), tgbotapi.NewKeyboardButton(btnRecentTxs)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnTrackOrder), tgbotapi.NewKeyboardButton(btnCryptoPrices)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnRefStats), tgbotapi.NewKeyboardButton(btnSupport)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnBack)),
	)

	keyboard.ResizeKeyboard = true
	return keyboard, nil
}

//// BuildUserAccountMenu ساخت منوی حساب کاربری با چیدمان درخواستی
//func (kb *KeyboardBuilder) BuildUserAccountMenu(tenantID uint, lang string) (tgbotapi.ReplyKeyboardMarkup, error) {
//	getText := func(key, fallback string) string {
//		text, err := kb.TranslationRepo.GetText(tenantID, lang, key)
//		if err != nil || text == "" {
//			return fallback
//		}
//		return text
//	}
//
//	btnBalance := getText("btn_balance", "موجودی")
//	btnTransfer := getText("btn_transfer", "انتقال")
//	btnConvert := getText("btn_convert", "🔄 تبدیل موجودی")
//	btnRecentTxs := getText("btn_recent_txs", "تراکنشات اخیر")
//	btnRefStats := getText("btn_ref_stats", "رفرال ها")
//	btnSupport := getText("btn_support", "ارتباط با پشتیبانی")
//	btnBack := getText("btn_back_main", "بازگشت به منوی اصلی")
//
//	keyboard := tgbotapi.NewReplyKeyboard(
//		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnBalance), tgbotapi.NewKeyboardButton(btnTransfer)),
//		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnConvert), tgbotapi.NewKeyboardButton(btnRecentTxs)),
//		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnRefStats), tgbotapi.NewKeyboardButton(btnSupport)),
//		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnBack)),
//	)
//
//	keyboard.ResizeKeyboard = true
//	return keyboard, nil
//}

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
	btnCard := getText("btn_pay_card", "🏦 کارت به کارت (ارسال فیش)")

	idStr := fmt.Sprintf("%d", orderID) // تزریق آیدی سفارش به دیتای دکمه جهت رهگیری

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnCard, "pay_card_"+idStr)),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnWalletToman, "pay_toman_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData(btnWalletPoint, "pay_point_"+idStr),
		),
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

// [delivery/telegram/keyboards.go]
// feat(keyboard): build structured inline keyboards for admin panel

// BuildAdminInlineMenu ساخت منوی شیشه‌ای اصلی ادمین بر اساس سطح دسترسی
func (kb *KeyboardBuilder) BuildAdminInlineMenu(isSuperAdmin bool) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// بخش اول: مدیریت‌های مشترک (سفارشات، پشتیبانی و مالی)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🛒 سفارش‌ها", "admin_orders"),
		tgbotapi.NewInlineKeyboardButtonData("🎫 پشتیبانی", "admin_tickets"),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("💳 بررسی برداشت‌ها", "admin_withdraws"),
		tgbotapi.NewInlineKeyboardButtonData("💼 تراکنش‌ها", "admin_transactions"),
	))

	// بخش دوم: مدیریت سیستم (تسک، اسپانسر، رفرال)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📝 تسک‌ها", "admin_tasks"),
		tgbotapi.NewInlineKeyboardButtonData("📢 اسپانسرها", "admin_sponsors"),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("👥 رفرال‌ها", "admin_referrals"),
		tgbotapi.NewInlineKeyboardButtonData("🎁 گیفت‌ها", "admin_gifts"),
	))

	// بخش سوم: امکانات ویژه (فقط سوپر ادمین)
	if isSuperAdmin {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔧 خدمات", "admin_services"),
			tgbotapi.NewInlineKeyboardButtonData("💰 قیمت‌ها", "admin_prices"),
		))
		//rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		//	tgbotapi.NewInlineKeyboardButtonData("🛑 جوین اجباری", "admin_forcejoin"),
		//	tgbotapi.NewInlineKeyboardButtonData("🏆 لیگ", "admin_league"),
		//))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛑 جوین اجباری", "admin_forcejoin"),
			tgbotapi.NewInlineKeyboardButtonData("🏆 لیگ و قرعه‌کشی", "admin_draws"),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 کاربران", "admin_users"),
			tgbotapi.NewInlineKeyboardButtonData("📊 آمار ربات", "admin_stats"),
		))
		// مدیریت محتوای داینامیک
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔤 متون و دکمه‌ها", "admin_texts"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ تنظیمات مالی", "admin_settings"),
		))

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📢 پیام و فوروارد همگانی", "admin_broadcast"),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BuildOrderTicketMenu دکمه شیشه‌ای ثبت تیکت برای یک سفارش خاص
func (kb *KeyboardBuilder) BuildOrderTicketMenu(orderID uint) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", orderID)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎫 ثبت تیکت پشتیبانی", "new_ticket_"+idStr),
		),
	)
}

// BuildTaskActionMenu دکمه‌های شیشه‌ای برای انجام وظیفه (عضویت، ری‌اکشن یا ارسال اسکرین‌شات)
func (kb *KeyboardBuilder) BuildTaskActionMenu(taskID uint, targetLink string, taskType string) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", taskID)

	// تبدیل یوزرنیم به لینک برای دکمه شیشه‌ای
	var btnURL string
	if strings.HasPrefix(targetLink, "@") {
		btnURL = "https://t.me/" + strings.TrimPrefix(targetLink, "@")
	} else if strings.HasPrefix(targetLink, "http://") || strings.HasPrefix(targetLink, "https://") {
		btnURL = targetLink
	} else if strings.Contains(targetLink, "t.me/") {
		btnURL = "https://" + strings.TrimPrefix(targetLink, "https://")
	} else {
		btnURL = "https://t.me"
	}

	verifyBtnText := "✅ بررسی عضویت و دریافت پاداش"
	if taskType == "manual_proof" {
		verifyBtnText = "📸 ارسال اسکرین‌شات (اثبات انجام)"
	} else if taskType == "reaction" {
		verifyBtnText = "👍 ثبت ری‌اکشن و دریافت پاداش"
	}

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🔗 ورود به لینک / کانال", btnURL),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(verifyBtnText, "verify_task_"+idStr),
		),
	)
}

//// BuildTaskActionMenu دکمه‌های شیشه‌ای برای انجام وظیفه (عضویت و بررسی یا ارسال اسکرین‌شات)
//func (kb *KeyboardBuilder) BuildTaskActionMenu(taskID uint, targetLink string, taskType string) tgbotapi.InlineKeyboardMarkup {
//	idStr := fmt.Sprintf("%d", taskID)
//
//	// تبدیل یوزرنیم به لینک برای دکمه شیشه‌ای
//	var btnURL string
//	if strings.HasPrefix(targetLink, "@") {
//		btnURL = "https://t.me/" + strings.TrimPrefix(targetLink, "@")
//	} else if strings.HasPrefix(targetLink, "-100") {
//		btnURL = "https://t.me"
//	} else {
//		btnURL = targetLink
//	}
//
//	verifyBtnText := "✅ بررسی عضویت و دریافت پاداش"
//	if taskType == "manual_proof" {
//		verifyBtnText = "📸 ارسال اسکرین‌شات (اثبات انجام)"
//	}
//
//	return tgbotapi.NewInlineKeyboardMarkup(
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonURL("🔗 ورود به لینک / کانال", btnURL),
//		),
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData(verifyBtnText, "verify_task_"+idStr),
//		),
//	)
//}

// BuildGameJoinMenu دکمه شیشه‌ای برای پیوستن به میز بازی و لغو توسط سازنده
func (kb *KeyboardBuilder) BuildGameJoinMenu(gameID uint, amount float64) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", gameID)
	btnText := fmt.Sprintf("🎮 پیوستن به بازی (%.0f امتیاز)", amount)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, "join_game_"+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ لغو بازی (مخصوص سازنده)", "cancel_game_"+idStr),
		),
	)
}

// BuildGameJoinMenu دکمه شیشه‌ای برای پیوستن به میز بازی
//func (kb *KeyboardBuilder) BuildGameJoinMenu(gameID uint, amount float64) tgbotapi.InlineKeyboardMarkup {
//	idStr := fmt.Sprintf("%d", gameID)
//	btnText := fmt.Sprintf("🎮 پیوستن به بازی (%.0f امتیاز)", amount)
//	return tgbotapi.NewInlineKeyboardMarkup(
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData(btnText, "join_game_"+idStr),
//		),
//	)
//}

// BuildAdminReceiptMenu دکمه‌های تأیید فیش برای ادمین
func (kb *KeyboardBuilder) BuildAdminReceiptMenu(receiptID uint) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", receiptID)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ تأیید فیش و ثبت سفارش", "admin_receipt_approve_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData("❌ رد فیش", "admin_receipt_reject_"+idStr),
		),
	)
}

// BuildDynamicCategoryMenu ساخت منوی شیشه‌ای داینامیک برای دسته‌ها و زیردسته‌ها
func (kb *KeyboardBuilder) BuildDynamicCategoryMenu(categories []domain.Category, isSub bool) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for i, cat := range categories {
		btnText := "▪️ " + cat.NameKey // در دیتابیس به جای کلید، مستقیما نام فارسی را ثبت می‌کنیم یا بعدا هندل می‌شود
		btnData := fmt.Sprintf("shop_cat_%d", cat.ID)

		currentRow = append(currentRow, tgbotapi.NewInlineKeyboardButtonData(btnText, btnData))
		if len(currentRow) == 2 || i == len(categories)-1 {
			rows = append(rows, currentRow)
			currentRow = nil
		}
	}

	if isSub {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت به دسته‌ها", "shop_main"),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BuildDynamicServiceMenu ساخت منوی شیشه‌ای سرویس‌های یک دسته
func (kb *KeyboardBuilder) BuildDynamicServiceMenu(services []domain.ServiceItem, catID uint) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, srv := range services {
		btnText := fmt.Sprintf("🏷 %s (%.0f تومان)", srv.Title, srv.PriceToman)
		btnData := fmt.Sprintf("shop_srv_%d", srv.ID)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, btnData),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "shop_main"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BuildDrawActionMenu دکمه شیشه‌ای برای شرکت در قرعه‌کشی
func (kb *KeyboardBuilder) BuildDrawActionMenu(drawID uint, cost int) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", drawID)
	btnText := fmt.Sprintf("🎟 شرکت در قرعه‌کشی (%d تیکت)", cost)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, "enter_draw_"+idStr),
		),
	)
}

// BuildForceJoinMenu دکمه‌های شیشه‌ای کانال‌های اجباری
func (kb *KeyboardBuilder) BuildForceJoinMenu(channels []domain.ForceJoin) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i, ch := range channels {
		btnText := fmt.Sprintf("📢 عضویت در کانال %d", i+1)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(btnText, ch.Link),
		))
	}

	// دکمه بررسی عضویت در پایین همه لینک‌ها
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✅ بررسی عضویت", "verify_force_join"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BuildAdminTaskProofMenu دکمه‌های تأیید اسکرین‌شات تسک دستی برای ادمین
func (kb *KeyboardBuilder) BuildAdminTaskProofMenu(subID uint) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", subID)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ تأیید و واریز پاداش", "admin_taskproof_approve_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData("❌ رد اسکرین‌شات", "admin_taskproof_reject_"+idStr),
		),
	)
}

// BuildOrderActionsMenu دکمه‌های عملیات پیشرفته برای یک سفارش خاص
func (kb *KeyboardBuilder) BuildOrderActionsMenu(orderID uint) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", orderID)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚡ سرعت‌دهی (کسر ۲ امتیاز)", "ordact_speedup_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData("✏️ ویرایش لینک", "ordact_edit_"+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("♻️ درخواست ریفیل (جبران ریزش)", "ordact_refill_"+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ لغو سفارش (کسر ۱۰٪ کارمزد)", "ordact_cancel_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData("🎫 ثبت تیکت", "new_ticket_"+idStr),
		),
	)
}

// BuildWithdrawTypeMenu انتخاب نوع برداشت
func (kb *KeyboardBuilder) BuildWithdrawTypeMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐ برداشت استارز", "withdraw_select_stars"),
			tgbotapi.NewInlineKeyboardButtonData("🎁 برداشت گیفت", "withdraw_select_gift"),
		),
	)
}

// BuildWithdrawStarsMenu انتخاب تعداد استارز (۱ تا ۹)
func (kb *KeyboardBuilder) BuildWithdrawStarsMenu() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for i := 1; i <= 9; i++ {
		btnText := fmt.Sprintf("%d ⭐", i)
		btnData := fmt.Sprintf("withdraw_do_STARS_%d", i)
		currentRow = append(currentRow, tgbotapi.NewInlineKeyboardButtonData(btnText, btnData))
		if len(currentRow) == 3 || i == 9 {
			rows = append(rows, currentRow)
			currentRow = nil
		}
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BuildWithdrawGiftMenu ساخت منوی داینامیک انتخاب گیفت بر اساس دیتابیس
func (kb *KeyboardBuilder) BuildWithdrawGiftMenu(gifts []domain.GiftItem) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, g := range gifts {
		btnText := fmt.Sprintf("%s (%.0f امتیاز)", g.Title, g.CostPoints)
		btnData := fmt.Sprintf("withdraw_gift_%d", g.ID)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, btnData),
		))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BuildAdminGiftsListMenu منوی مدیریت گیفت‌ها در پنل ادمین
func (kb *KeyboardBuilder) BuildAdminGiftsListMenu(gifts []domain.GiftItem) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, g := range gifts {
		btnText := fmt.Sprintf("❌ حذف %s (%.0f امتیاز)", g.Title, g.CostPoints)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, fmt.Sprintf("admgift_del_%d", g.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➕ افزودن گیفت جدید", "admgift_add"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BuildAdminWithdrawMenu دکمه‌های تأیید برداشت برای ادمین
func (kb *KeyboardBuilder) BuildAdminWithdrawMenu(wID uint) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", wID)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ تأیید واریز", "admin_withdraw_approve_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData("❌ رد درخواست", "admin_withdraw_reject_"+idStr),
		),
	)
}

//// BuildAdminUserActionMenu منوی عملیات دستی ادمین روی یک کاربر خاص
//func (kb *KeyboardBuilder) BuildAdminUserActionMenu(userID uint) tgbotapi.InlineKeyboardMarkup {
//	idStr := fmt.Sprintf("%d", userID)
//	return tgbotapi.NewInlineKeyboardMarkup(
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("➕ افزایش تومان", "admuser_addtoman_"+idStr),
//			tgbotapi.NewInlineKeyboardButtonData("➖ کسر تومان", "admuser_subtoman_"+idStr),
//		),
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("➕ افزایش امتیاز", "admuser_addpoint_"+idStr),
//			tgbotapi.NewInlineKeyboardButtonData("➖ کسر امتیاز", "admuser_subpoint_"+idStr),
//		),
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("📊 نمایش رفرال‌های کاربر", "admuser_refs_"+idStr),
//		),
//	)
//}
//
//// BuildAdminSettingsMenu منوی تنظیمات متغیرهای عددی سیستم
//func (kb *KeyboardBuilder) BuildAdminSettingsMenu() tgbotapi.InlineKeyboardMarkup {
//	return tgbotapi.NewInlineKeyboardMarkup(
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("⚙️ کارمزد لغو سفارش", "admset_cancelfee"),
//			tgbotapi.NewInlineKeyboardButtonData("⚙️ کارمزد انتقال گروه", "admset_transferfee"),
//		),
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("⚙️ نرخ تبدیل: ۱ امتیاز = ؟ تومان", "admset_ratetoman"),
//		),
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("⚙️ نرخ تبدیل: ؟ امتیاز = ۱ استارز", "admset_ratestar"),
//		),
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("⚙️ پاداش دعوت از غیرفعال‌ها", "admset_inactivebonus"),
//		),
//	)
//}
//
//// BuildAdminServicesMenu منوی مدیریت خدمات
//func (kb *KeyboardBuilder) BuildAdminServicesMenu() tgbotapi.InlineKeyboardMarkup {
//	return tgbotapi.NewInlineKeyboardMarkup(
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("➕ افزودن دسته‌بندی اصلی", "admsrv_addcat"),
//		),
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("➕ افزودن سرویس به دسته‌ها", "admsrv_addsrv"),
//		),
//	)
//}

// BuildAdminPricesMenu منوی مدیریت قیمت‌ها
func (kb *KeyboardBuilder) BuildAdminPricesMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💵 تنظیم قیمت روز دلار", "admprc_setdollar"),
		),
	)
}

// BuildAdminForceJoinListMenu منوی مدیریت کانال‌های جوین اجباری
func (kb *KeyboardBuilder) BuildAdminForceJoinListMenu(channels []domain.ForceJoin) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, ch := range channels {
		btnText := fmt.Sprintf("❌ حذف %s", ch.ChatID)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, fmt.Sprintf("admfj_del_%d", ch.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➕ افزودن کانال جدید", "admfj_add"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

//// BuildAdminTasksListMenu لیست تسک‌ها و اسپانسرها برای ادمین
//func (kb *KeyboardBuilder) BuildAdminTasksListMenu(tasks []domain.Task) tgbotapi.InlineKeyboardMarkup {
//	var rows [][]tgbotapi.InlineKeyboardButton
//	for _, t := range tasks {
//		statusIcon := "🟢"
//		if t.Status != "ACTIVE" {
//			statusIcon = "🔴"
//		}
//		btnText := fmt.Sprintf("%s %s (ظرفیت: %d)", statusIcon, t.Title, t.Capacity)
//		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData(btnText, fmt.Sprintf("admtask_edit_%d", t.ID)),
//		))
//	}
//	return tgbotapi.NewInlineKeyboardMarkup(rows...)
//}

// BuildAdminTasksListMenu لیست تسک‌ها و اسپانسرها برای ادمین به همراه دکمه ساخت تسک جدید
func (kb *KeyboardBuilder) BuildAdminTasksListMenu(tasks []domain.Task) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range tasks {
		statusIcon := "🟢"
		if t.Status != "ACTIVE" {
			statusIcon = "🔴"
		}
		btnText := fmt.Sprintf("%s %s (ظرفیت: %d)", statusIcon, t.Title, t.Capacity)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, fmt.Sprintf("admtask_edit_%d", t.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➕ افزودن تسک جدید", "admtask_add_0"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BuildAdminTaskEditMenu منوی ویرایش یک تسک خاص
func (kb *KeyboardBuilder) BuildAdminTaskEditMenu(taskID uint, status string) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", taskID)
	statusToggle := "🔴 توقف موقت تسک"
	if status != "ACTIVE" {
		statusToggle = "🟢 فعال‌سازی مجدد تسک"
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ تغییر ظرفیت", "admtask_cap_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData("⚖️ تعیین جریمه لفت", "admtask_pen_"+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(statusToggle, "admtask_toggle_"+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت به لیست تسک‌ها", "admin_tasks"),
		),
	)
}

// BuildAdminTicketsListMenu منوی لیست تیکت‌های باز
func (kb *KeyboardBuilder) BuildAdminTicketsListMenu(tickets []domain.SupportTicket) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range tickets {
		btnText := fmt.Sprintf("🎫 بررسی تیکت سفارش #%d", t.OrderID)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, fmt.Sprintf("admtick_view_%d", t.ID)),
		))
	}
	if len(tickets) == 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ هیچ تیکت بازی وجود ندارد", "ignore"),
		))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BuildAdminTicketActionMenu منوی عملیات روی یک تیکت
func (kb *KeyboardBuilder) BuildAdminTicketActionMenu(ticketID uint) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", ticketID)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 پاسخ به کاربر", "admtick_reply_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData("❌ بستن تیکت", "admtick_close_"+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت به لیست", "admin_tickets"),
		),
	)
}

// BuildLeagueMainMenu منوی شیشه‌ای اصلی بخش مسابقات و برترین‌ها
func (kb *KeyboardBuilder) BuildLeagueMainMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏆 رفرال لیگ (لیدربورد)", "league_referrals"),
			tgbotapi.NewInlineKeyboardButtonData("🌟 برترین مشتریان (VIP)", "league_customers"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎰 قرعه‌کشی‌های فعال", "league_draws"),
		),
	)
}

// BuildAdminDrawListMenu لیست قرعه‌کشی‌ها و تنظیمات لیگ برای ادمین
func (kb *KeyboardBuilder) BuildAdminDrawListMenu(draws []domain.LuckyDraw) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, d := range draws {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("🎰 %s", d.Title), fmt.Sprintf("admdraw_view_%d", d.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➕ ایجاد قرعه‌کشی جدید", "admdraw_add_0"),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📝 ویرایش قوانین لیگ", "admleague_rules"),
		tgbotapi.NewInlineKeyboardButtonData("🔄 شروع فصل جدید لیگ", "admleague_reset"),
	))
	if len(draws) == 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ هیچ قرعه‌کشی فعالی وجود ندارد", "ignore"),
		))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BuildAdminDrawListMenu لیست قرعه‌کشی‌ها برای ادمین
//func (kb *KeyboardBuilder) BuildAdminDrawListMenu(draws []domain.LuckyDraw) tgbotapi.InlineKeyboardMarkup {
//	var rows [][]tgbotapi.InlineKeyboardButton
//	for _, d := range draws {
//		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("🎰 %s", d.Title), fmt.Sprintf("admdraw_view_%d", d.ID)),
//		))
//	}
//	if len(draws) == 0 {
//		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("✅ هیچ قرعه‌کشی فعالی وجود ندارد", "ignore"),
//		))
//	}
//	return tgbotapi.NewInlineKeyboardMarkup(rows...)
//}

// BuildAdminDrawActionMenu منوی عملیات پایان قرعه‌کشی
func (kb *KeyboardBuilder) BuildAdminDrawActionMenu(drawID uint) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", drawID)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏁 پایان قرعه‌کشی و انتخاب برندگان", "admdraw_roll_"+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت به لیست", "admin_draws"),
		),
	)
}

// BuildConvertMenu منوی انتخاب نوع تبدیل موجودی
func (kb *KeyboardBuilder) BuildConvertMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 تبدیل امتیاز به تومان", "convert_to_toman"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐ تبدیل امتیاز به استارز", "convert_to_stars"),
		),
	)
}

// BuildAdminOrderManageMenu منوی مدیریت یک سفارش توسط ادمین
func (kb *KeyboardBuilder) BuildAdminOrderManageMenu(orderID uint) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", orderID)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ تکمیل سفارش", "admord_complete_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData("❌ لغو و عودت وجه", "admord_cancel_"+idStr),
		),
	)
}

// BuildAdminUserActionMenu منوی عملیات دستی ادمین روی موجودی سه‌گانه و رفرال کاربر
func (kb *KeyboardBuilder) BuildAdminUserActionMenu(userID uint) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", userID)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ افزایش تومان", "admuser_addtoman_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData("➖ کسر تومان", "admuser_subtoman_"+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ افزایش امتیاز", "admuser_addpoint_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData("➖ کسر امتیاز", "admuser_subpoint_"+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ افزایش استارز ⭐", "admuser_addstars_"+idStr),
			tgbotapi.NewInlineKeyboardButtonData("➖ کسر استارز ⭐", "admuser_substars_"+idStr),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 نمایش رفرال‌های کاربر", "admuser_refs_"+idStr),
		),
	)
}

// BuildAdminSettingsMenu منوی جامع تنظیمات مالی، کانال‌ها، کارت بانکی و پشتیبانی
func (kb *KeyboardBuilder) BuildAdminSettingsMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ کارمزد لغو سفارش", "admset_cancelfee"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ کارمزد انتقال", "admset_transferfee"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💵 نرخ: ۱ امتیاز = ؟ تومان", "admset_ratetoman"),
			tgbotapi.NewInlineKeyboardButtonData("⭐ نرخ: ؟ امتیاز = ۱ استارز", "admset_ratestar"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎁 پاداش پیش‌فرض تسک", "admset_taskreward"),
			tgbotapi.NewInlineKeyboardButtonData("⚖️ جریمه پیش‌فرض تسک", "admset_taskpenalty"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("♻️ پاداش دعوت غیرفعال‌ها", "admset_inactivebonus"),
			tgbotapi.NewInlineKeyboardButtonData("💳 تنظیم کارت بانکی", "admset_cardinfo"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📢 آیدی کانال گزارش", "admset_reportchan"),
			tgbotapi.NewInlineKeyboardButtonData("🛒 آیدی کانال سفارشات", "admset_orderchan"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏆 آیدی کانال برندگان", "admset_winnerchan"),
			tgbotapi.NewInlineKeyboardButtonData("🎧 لینک گروه پشتیبانی", "admset_supportlink"),
		),
	)
}

// BuildAdminServicesMenu منوی مدیریت دسته‌ها، زیردسته‌ها و سرویس‌های فروشگاه
func (kb *KeyboardBuilder) BuildAdminServicesMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ افزودن دسته اصلی / زیردسته", "admsrv_addcat"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ افزودن سرویس به دسته‌ها", "admsrv_addsrv"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 مشاهده لیست دسته‌ها و آیدی‌ها", "admsrv_listcats"),
		),
	)
}

// BuildAdminWithdrawsListMenu منوی لیست درخواست‌های برداشت در انتظار تایید
func (kb *KeyboardBuilder) BuildAdminWithdrawsListMenu(withdrawals []domain.Withdrawal) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, w := range withdrawals {
		btnText := fmt.Sprintf("💳 برداشت #%d | %.0f (%s)", w.ID, w.Amount, w.Currency)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, fmt.Sprintf("admin_wdview_%d", w.ID)),
		))
	}
	if len(withdrawals) == 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ هیچ درخواست برداشت معلقی وجود ندارد", "ignore"),
		))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
