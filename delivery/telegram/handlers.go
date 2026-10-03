package telegram

import (
	"OmniGram/domain"
	"OmniGram/repository"
	"OmniGram/services"
	"fmt"
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

	UserRepo     repository.UserRepository
	ReferralRepo *repository.ReferralRepository

	TxRepo        *repository.TransactionRepository
	WalletService *services.WalletService
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
		h.SafeBot.SendLogToChannel(err.Error())
	}
}

// HandleMessage مدیریت پیام‌های متنی ارسالی از سمت کاربر
// HandleMessage مدیریت پیام‌های متنی ارسالی از سمت کاربر
func (h *BotHandler) HandleMessage(tenant *domain.Tenant, chatID int64, text string, lang string) {
	// تابع کمکی برای گرفتن سریع متن‌ها
	getText := func(key, fallback string) string {
		val, err := h.TranslationRepo.GetText(tenant.ID, lang, key)
		if err != nil || val == "" {
			return fallback
		}
		return val
	}

	// گرفتن نام دکمه‌ها از دیتابیس برای مقایسه
	devBtnText := getText("btn_developer", "توسعه‌دهنده")
	accountBtnText := getText("btn_account", "حساب کاربری")
	balanceBtnText := getText("btn_balance", "موجودی")
	refStatsBtnText := getText("btn_ref_stats", "رفرال ها")
	backMainBtnText := getText("btn_back_main", "بازگشت به منوی اصلی")

	recentTxsBtnText := getText("btn_recent_txs", "تراکنشات اخیر")
	transferBtnText := getText("btn_transfer", "انتقال")

	// پیدا کردن کاربر از دیتابیس برای عملیاتی که نیاز به اطلاعات کاربر دارند
	// پیدا کردن کاربر از دیتابیس
	currentUser, err := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)

	// اگر کاربر در دیتابیس نبود، او را ثبت می‌کنیم
	if err != nil || currentUser == nil {
		currentUser = &domain.User{
			TenantID:   tenant.ID,
			TelegramID: chatID,
			FirstName:  "کاربر", // در آینده می‌توانید نام اصلی را از update.Message بگیرید
			Language:   lang,
			State:      domain.StateNormal,
		}
		_ = h.UserRepo.CreateUser(currentUser)
	}

	switch text {
	case devBtnText:
		devDesc := getText("msg_dev_description", "متن معرفی توسعه دهنده")
		developerURL := getText("url_developer", "https://t.me/default_admin")
		inlineMenu, _ := h.KeyboardBuilder.BuildDeveloperInlineMenu(tenant.ID, lang, developerURL)

		msg := tgbotapi.NewMessage(chatID, devDesc)
		msg.ReplyMarkup = inlineMenu
		_, _ = h.SafeBot.Bot.Send(msg)

	case accountBtnText:
		// باز کردن منوی حساب کاربری
		msgText := getText("msg_account_section", "به بخش حساب کاربری خوش آمدید. لطفا یک گزینه را انتخاب کنید:")
		accMenu, _ := h.KeyboardBuilder.BuildUserAccountMenu(tenant.ID, lang)

		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ReplyMarkup = accMenu
		_, _ = h.SafeBot.Bot.Send(msg)

	case balanceBtnText:
		if currentUser != nil {
			// دریافت قالب نمایش موجودی از دیتابیس با قابلیت PlaceHolder
			balanceTemplate := getText("msg_balance_template", "💰 موجودی شما:\n\nکیف پول (تومان): %v\nامتیاز رفرال: %v\nاستارز: %v")

			// اگر fmt در بالا نیست در ایمپورت‌های فایل اضافه کنید
			msgText := fmt.Sprintf(balanceTemplate, currentUser.TomanBalance, currentUser.PointBalance, currentUser.StarsBalance)

			msg := tgbotapi.NewMessage(chatID, msgText)
			_, _ = h.SafeBot.Bot.Send(msg)
		}

	case refStatsBtnText:
		if currentUser != nil {
			stats, err := h.ReferralRepo.GetReferralStats(tenant.ID, currentUser.ID)
			if err == nil && stats != nil {
				statsTemplate := getText("msg_ref_stats_template", "📊 آمار زیرمجموعه‌های شما:\n\n👥 کل دعوت شده‌ها: %v\n✅ تایید شده‌ها: %v\n💤 غیرفعال (بیش از ۱۰ روز): %v")

				msgText := fmt.Sprintf(statsTemplate, stats.TotalInvited, stats.Approved, stats.InactiveUsers)

				msg := tgbotapi.NewMessage(chatID, msgText)
				_, _ = h.SafeBot.Bot.Send(msg)
			}
		}

	case recentTxsBtnText: // دکمه تراکنشات اخیر
		if currentUser != nil && h.TxRepo != nil {
			txs, err := h.TxRepo.GetRecentTransactions(tenant.ID, currentUser.ID, 5)
			if err != nil || len(txs) == 0 {
				_ = h.SafeBot.SendMessage(chatID, "هیچ تراکنشی برای شما ثبت نشده است.")
			} else {
				msgText := "🧾 ۵ تراکنش اخیر شما:\n\n"
				for i, tx := range txs {
					// نمایش مبلغ با علامت مثبت/منفی و فرمت مناسب
					sign := "+"
					if tx.Amount < 0 {
						sign = "" // منفی خودش در عدد هست
					}
					msgText += fmt.Sprintf("%d. نوع: %s | مبلغ: %s%.2f | ارز: %s\n", i+1, tx.Type, sign, tx.Amount, tx.Currency)
				}
				_ = h.SafeBot.SendMessage(chatID, msgText)
			}
		}

	case transferBtnText: // دکمه راهنمای انتقال
		msgText := getText("msg_transfer_guide", "جهت انتقال موجودی، پیام خود را دقیقاً با فرمت زیر ارسال کنید:\n\n`انتقال [آیدی_عددی_مقصد] [مبلغ] [نوع_ارز]`\n\nانواع ارز: `TOMAN`, `POINT`, `STARS`\n\nمثال:\nانتقال 123456789 10 POINT")
		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"
		_, _ = h.SafeBot.Bot.Send(msg)

	case backMainBtnText:
		// بازگشت به منوی اصلی
		msgText := getText("msg_main_menu", "به منوی اصلی برگشتیم:")
		mainMenu, _ := h.KeyboardBuilder.BuildMainMenu(tenant.ID, lang)

		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ReplyMarkup = mainMenu
		_, _ = h.SafeBot.Bot.Send(msg)

	default:
		if strings.HasPrefix(text, "انتقال") && currentUser != nil && h.WalletService != nil {
			parts := strings.Split(text, " ")
			if len(parts) != 4 {
				_ = h.SafeBot.SendMessage(chatID, "فرمت دستور اشتباه است. روی دکمه انتقال کلیک کنید تا راهنما را ببینید.")
				return
			}

			destTelegramID, errID := strconv.ParseInt(parts[1], 10, 64)
			amount, errAmount := strconv.ParseFloat(parts[2], 64)
			currency := strings.ToUpper(parts[3])

			if errID != nil || errAmount != nil || amount <= 0 {
				_ = h.SafeBot.SendMessage(chatID, "مقادیر وارد شده معتبر نیست. آیدی و مبلغ باید عدد باشند.")
				return
			}

			// اگر کارمزد در دیتابیس تنظیم نشده بود، پیش‌فرض ۵٪ لحاظ شود
			feePercent := tenant.TransferFeePercent
			if feePercent <= 0 {
				feePercent = 5.0
			}

			err := h.WalletService.TransferBalance(tenant.ID, currentUser.ID, destTelegramID, amount, currency, feePercent)
			if err != nil {
				_ = h.SafeBot.SendMessage(chatID, "❌ خطا: "+err.Error())
			} else {
				successMsg := fmt.Sprintf("✅ مبلغ %.2f %s با موفقیت پس از کسر کارمزد انتقال یافت.", amount, currency)
				_ = h.SafeBot.SendMessage(chatID, successMsg)
			}
			return
		}
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
