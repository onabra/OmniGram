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
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛑 جوین اجباری", "admin_forcejoin"),
			tgbotapi.NewInlineKeyboardButtonData("🏆 لیگ", "admin_league"),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 کاربران", "admin_users"),
			tgbotapi.NewInlineKeyboardButtonData("📊 آمار ربات", "admin_stats"),
		))
		// مدیریت محتوای داینامیک
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔤 متون و دکمه‌ها", "admin_texts"),
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

// BuildTaskActionMenu دکمه‌های شیشه‌ای برای انجام وظیفه (عضویت و بررسی یا ارسال اسکرین‌شات)
func (kb *KeyboardBuilder) BuildTaskActionMenu(taskID uint, targetLink string, taskType string) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", taskID)

	// تبدیل یوزرنیم به لینک برای دکمه شیشه‌ای
	var btnURL string
	if strings.HasPrefix(targetLink, "@") {
		btnURL = "https://t.me/" + strings.TrimPrefix(targetLink, "@")
	} else if strings.HasPrefix(targetLink, "-100") {
		btnURL = "https://t.me"
	} else {
		btnURL = targetLink
	}

	verifyBtnText := "✅ بررسی عضویت و دریافت پاداش"
	if taskType == "manual_proof" {
		verifyBtnText = "📸 ارسال اسکرین‌شات (اثبات انجام)"
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

// BuildGameJoinMenu دکمه شیشه‌ای برای پیوستن به میز بازی
func (kb *KeyboardBuilder) BuildGameJoinMenu(gameID uint, amount float64) tgbotapi.InlineKeyboardMarkup {
	idStr := fmt.Sprintf("%d", gameID)
	btnText := fmt.Sprintf("🎮 پیوستن به بازی (%.0f امتیاز)", amount)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, "join_game_"+idStr),
		),
	)
}

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
