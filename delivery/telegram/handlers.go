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
	UserRepo        repository.UserRepository
	ReferralRepo    *repository.ReferralRepository
	TxRepo          *repository.TransactionRepository
	WalletService   *services.WalletService
	OrderRepo       *repository.OrderRepository
	ServiceRepo     *repository.ServiceRepository
	TaskRepo        *repository.TaskRepository
	TicketRepo      *repository.TicketRepository
	GameRepo        *repository.GameRepository
	ReceiptRepo     *repository.ReceiptRepository
	DrawRepo        *repository.DrawRepositor

	ForceJoinRepo *repository.ForceJoinRepository

	CategoryRepo *repository.CategoryRepository

	SuperAdminIDs []int64
	AdminIDs      []int64
}

// HandleStart مدیریت دستور /start
func (h *BotHandler) HandleStart(tenant *domain.Tenant, chatID int64, lang string, text string) {

	if !h.CheckForceJoin(tenant, chatID) {
		return
	}
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
func (h *BotHandler) HandleMessage(tenant *domain.Tenant, chatID int64, text string, lang string) {
	// قفل جوین اجباری
	if !h.CheckForceJoin(tenant, chatID) {
		return
	}

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

	btnShop := getText("btn_shop", "فروشگاه")

	taskMenuBtnText := getText("btn_task", "وظایف")
	sponsorBtnText := getText("btn_sponsor", "📢 اسپانسر شدن (جذب ممبر)")

	supportBtnText := getText("btn_support", "ارتباط با پشتیبانی")

	btnDoTasks := getText("btn_do_tasks", "📝 انجام وظایف (کسب درآمد)")

	btnRefLeague := getText("btn_ref_league", "لیگ")

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

	if currentUser.State == domain.StateWaitSponsorLink {
		// اگر کاربر دکمه بازگشت را زد، وضعیتش را ریست می‌کنیم
		if text == backMainBtnText {
			_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateNormal)
		} else {
			draftOrder, err := h.OrderRepo.GetDraftOrder(tenant.ID, currentUser.ID)
			if err != nil {
				_ = h.SafeBot.SendMessage(chatID, "سفارش پیش‌نویسی یافت نشد. لطفاً دوباره از منو اقدام کنید.")
				_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateNormal)
				return
			}

			// آپدیت لینک و تغییر وضعیت به گرفتن تعداد
			// استفاده از فایل Utils برای بررسی و استانداردسازی لینک
			resolvedLink, err := ResolveChatID(text)
			if err != nil {
				if err.Error() == "PRIVATE_LINK" {
					_ = h.SafeBot.SendMessage(chatID, "❌ ربات نمی‌تواند عضویت در لینک‌های خصوصی (پرایوت) را بررسی کند.\n\nلطفاً برای کانال‌های خصوصی، فقط **آیدی عددی** کانال (شروع با 100-) را ارسال کنید.")
				} else {
					_ = h.SafeBot.SendMessage(chatID, "❌ لینک یا آیدی نامعتبر است. لطفاً یک لینک عمومی (t.me/id) یا آیدی (id@) یا آیدی عددی معتبر بفرستید.")
				}
				return
			}

			// آپدیت لینک با فرمت استاندارد شده
			draftOrder.TargetLink = resolvedLink
			_ = h.OrderRepo.UpdateOrder(draftOrder)
			_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateWaitSponsorCount)

			_ = h.SafeBot.SendMessage(chatID, "✅ لینک با موفقیت ثبت شد.\n\n🔢 لطفاً تعداد ممبر درخواستی خود را فقط به صورت عدد ارسال کنید (مثلاً 1000):")
			return
		}
	}

	if currentUser.State == domain.StateWaitSponsorCount {
		if text == backMainBtnText {
			_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateNormal)
		} else {
			quantity, err := strconv.Atoi(text)
			if err != nil || quantity <= 0 {
				_ = h.SafeBot.SendMessage(chatID, "❌ لطفاً یک عدد معتبر بزرگتر از صفر ارسال کنید.")
				return
			}

			draftOrder, err := h.OrderRepo.GetDraftOrder(tenant.ID, currentUser.ID)
			if err != nil {
				_ = h.SafeBot.SendMessage(chatID, "سفارش پیش‌نویسی یافت نشد. لطفاً دوباره تلاش کنید.")
				_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateNormal)
				return
			}

			// واکشی قیمت از سرویس
			serviceItem, err := h.ServiceRepo.GetServiceByID(tenant.ID, draftOrder.ServiceItemID)
			if err != nil {
				_ = h.SafeBot.SendMessage(chatID, "خطا در دریافت قیمت سرویس.")
				return
			}

			// محاسبه قیمت (با فرض اینکه قیمت پایه به تومان است)
			totalAmount := float64(quantity) * serviceItem.PriceToman
			draftOrder.Quantity = quantity
			draftOrder.TotalAmount = totalAmount
			_ = h.OrderRepo.UpdateOrder(draftOrder)

			// خروج کاربر از وضعیت اسپانسرینگ
			_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateNormal)

			// ارسال فاکتور و کیبورد پرداخت با سیستم ایزوله‌سازی
			msgText := fmt.Sprintf("✅ فاکتور سفارش شما:\n\n🔗 لینک: %s\n👥 تعداد درخواستی: %d\n💰 مبلغ کل: %.0f تومان\n\n👇 لطفاً روش پرداخت را انتخاب کنید:", draftOrder.TargetLink, quantity, totalAmount)
			paymentMenu, _ := h.KeyboardBuilder.BuildPaymentMethodMenu(tenant.ID, lang, draftOrder.ID)

			msg := tgbotapi.NewMessage(chatID, msgText)
			msg.ReplyMarkup = paymentMenu
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}
	}

	// --- سیستم پشتیبانی ---
	if currentUser.State == domain.StateWaitSupportOrderCode {
		if text == backMainBtnText {
			_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateNormal)
		} else {
			// جستجوی سفارش با کد رهگیری
			order, err := h.OrderRepo.GetOrder(tenant.ID, text)
			if err != nil || order.UserID != currentUser.ID {
				_ = h.SafeBot.SendMessage(chatID, "❌ سفارش یافت نشد یا متعلق به شما نیست. لطفاً کد رهگیری صحیح را بفرستید:")
				return
			}

			// ساخت تیکت
			ticket := &domain.SupportTicket{
				TenantID: tenant.ID,
				UserID:   currentUser.ID,
				OrderID:  order.ID,
				Status:   "OPEN",
			}
			_ = h.TicketRepo.CreateTicket(ticket)
			_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateWaitSupportMsg)

			_ = h.SafeBot.SendMessage(chatID, "✅ سفارش تأیید شد.\n\n📝 لطفاً مشکل خود را در مورد این سفارش کامل بنویسید:")
			return
		}
	}

	if currentUser.State == domain.StateWaitSupportMsg {
		if text == backMainBtnText {
			_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateNormal)
		} else {
			ticket, err := h.TicketRepo.GetOpenTicket(tenant.ID, currentUser.ID)
			if err != nil {
				_ = h.SafeBot.SendMessage(chatID, "تیکت بازی یافت نشد. لطفاً مجدداً اقدام کنید.")
				_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateNormal)
				return
			}

			// ذخیره پیام
			msg := &domain.SupportMessage{
				TicketID: ticket.ID,
				SenderID: chatID,
				IsAdmin:  false,
				Text:     text,
			}
			_ = h.TicketRepo.SaveMessage(msg)

			// واکشی اطلاعات سفارش برای گزارش
			order, _ := h.OrderRepo.GetOrderByID(tenant.ID, ticket.OrderID)

			// فرمت‌بندی و ارسال گزارش به کانال/ادمین
			reportText := fmt.Sprintf("🎫 *تیکت جدید (سفارش)*\n\n👤 کاربر: [%s](tg://user?id=%d)\n🛒 شماره سفارش: `%s`\n🔢 تعداد: %d\n💰 مبلغ: %.0f\n\n📝 متن مشکل:\n%s", currentUser.FirstName, currentUser.TelegramID, order.TrackingCode, order.Quantity, order.TotalAmount, text)

			targetChannel := tenant.ReportChannelID
			if targetChannel != "" {
				chID, _ := strconv.ParseInt(targetChannel, 10, 64)
				_ = h.SafeBot.SendMessage(chID, reportText)
			} else {
				// در صورت ست نبودن کانال گزارش، به لاگر بفرست
				h.SafeBot.SendLogToChannel(reportText)
			}

			_ = h.SafeBot.SendMessage(chatID, "✅ پیام شما با موفقیت ثبت و برای تیم پشتیبانی ارسال شد.")
			_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateNormal)
			return
		}
	}

	switch text {

	case btnShop:
		categories, err := h.CategoryRepo.GetMainCategories(tenant.ID)
		if err != nil || len(categories) == 0 {
			_ = h.SafeBot.SendMessage(chatID, "🛍 فروشگاه در حال حاضر خالی است.")
		} else {
			msgText := "🛍 *به فروشگاه خدمات مجازی خوش آمدید!*\n\nلطفاً دسته‌بندی مورد نظر خود را انتخاب کنید:"
			shopMenu := h.KeyboardBuilder.BuildDynamicCategoryMenu(categories, false)
			msg := tgbotapi.NewMessage(chatID, msgText)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = shopMenu
			_, _ = h.SafeBot.Bot.Send(msg)
		}

	case btnRefLeague:
		// واکشی قرعه‌کشی‌های فعال
		draws, err := h.DrawRepo.GetActiveDraws(tenant.ID)
		if err != nil || len(draws) == 0 {
			_ = h.SafeBot.SendMessage(chatID, "🏆 در حال حاضر مسابقه یا قرعه‌کشی فعالی وجود ندارد. لطفاً بعداً سر بزنید.")
			return
		}

		// ارسال پست جداگانه برای هر قرعه‌کشی
		for _, draw := range draws {
			totalEntries, _ := h.DrawRepo.GetTotalEntries(tenant.ID, draw.ID)

			// تبدیل تاریخ میلادی به فرمت خواناتر (در نسخه‌های بعدی می‌توان شمسی کرد)
			endDate := draw.EndDate.Format("2006-01-02 15:04")

			msgText := fmt.Sprintf("🎰 *%s*\n\n📝 توضیحات: %s\n\n🎟 هزینه ورود: %d تیکت\n🏆 تعداد برندگان: %d نفر\n🎁 جایزه هر برنده: %.0f %s\n👥 تعداد شانس‌های ثبت شده: %d / %d\n\n⏱ پایان: %s",
				draw.Title, draw.Description, draw.TicketCost, draw.WinnersCount, draw.RewardPerWinner, draw.RewardType, totalEntries, draw.MaxParticipants, endDate)

			markup := h.KeyboardBuilder.BuildDrawActionMenu(draw.ID, draw.TicketCost)
			msg := tgbotapi.NewMessage(chatID, msgText)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = markup
			_, _ = h.SafeBot.Bot.Send(msg)
		}

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
				botUsername := h.SafeBot.Bot.Self.UserName
				dashboardMsg := fmt.Sprintf("👥 *داشبورد رفرال‌های شما*\n\n"+
					"🏆 *رفرال لیگ*: %d نفر\n"+
					"🔗 لینک: `https://t.me/%s?start=league_%d`\n\n"+
					"🎰 *رفرال قرعه‌کشی*: %d نفر\n"+
					"🔗 لینک: `https://t.me/%s?start=draw_%d`\n\n"+
					"⭐ *رفرال امتیاز*: %d نفر\n"+
					"🔗 لینک: `https://t.me/%s?start=point_%d`\n\n"+
					"💰 *رفرال کیف پول*: %d نفر\n"+
					"🔗 لینک: `https://t.me/%s?start=wallet_%d`\n\n"+
					"➖➖➖➖➖➖➖\n"+
					"🎟 *تیکت‌های شما*: %d\n"+
					"💤 زیرمجموعه‌های غیرفعال: %d",
					stats.LeagueTotal, botUsername, currentUser.ID,
					stats.DrawTotal, botUsername, currentUser.ID,
					stats.PointTotal, botUsername, currentUser.ID,
					stats.WalletTotal, botUsername, currentUser.ID,
					currentUser.DrawTickets, stats.InactiveUsers)

				msg := tgbotapi.NewMessage(chatID, dashboardMsg)
				msg.ParseMode = "Markdown"
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

	case taskMenuBtnText:
		// باز کردن زیرمنوی وظایف
		msgText := getText("msg_task_section", "به بخش وظایف خوش آمدید. می‌خواهید تسک انجام دهید یا اسپانسر شوید؟")
		taskMenu, _ := h.KeyboardBuilder.BuildTasksSubMenu(tenant.ID, lang)

		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ReplyMarkup = taskMenu
		_, _ = h.SafeBot.Bot.Send(msg)

	// [delivery/telegram/handlers.go]
	//feat(handler): serve next available active task to user
	case btnDoTasks:
		// واکشی تمام تسک‌های فعال کارفرما
		tasks, err := h.TaskRepo.GetActiveTasks(tenant.ID)
		if err != nil || len(tasks) == 0 {
			_ = h.SafeBot.SendMessage(chatID, "🎉 در حال حاضر تسک جدیدی برای انجام وجود ندارد. لطفاً بعداً سر بزنید.")
			return
		}

		// پیدا کردن اولین تسکی که کاربر هنوز انجام نداده است
		var availableTask *domain.Task
		for _, t := range tasks {
			if !h.TaskRepo.HasUserCompletedTask(tenant.ID, currentUser.ID, t.ID) {
				// به دلیل ساختار متغیر حلقه در نسخه‌های گو، باید یک کپی از آن بگیریم
				taskCopy := t
				availableTask = &taskCopy
				break
			}
		}

		if availableTask == nil {
			_ = h.SafeBot.SendMessage(chatID, "🎉 شما تمام تسک‌های موجود را انجام داده‌اید! لطفاً بعداً سر بزنید.")
			return
		}

		msgText := fmt.Sprintf("📝 *تسک جدید*\n\n📌 عنوان: %s\n🎁 پاداش: %.0f %s\n\n👇 برای دریافت پاداش، ابتدا از طریق دکمه زیر در کانال عضو شوید و سپس روی بررسی کلیک کنید:", availableTask.Title, availableTask.RewardAmount, availableTask.RewardType)
		taskMenu := h.KeyboardBuilder.BuildTaskActionMenu(availableTask.ID, availableTask.TargetLink)

		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = taskMenu
		_, _ = h.SafeBot.Bot.Send(msg)

	case sponsorBtnText:
		// استارت فرآیند اسپانسر شدن (نمایش دکمه‌های شیشه‌ای انتخاب کیفیت)
		msgText := getText("msg_choose_sponsor_quality", "لطفاً کیفیت ممبر درخواستی خود را برای تسک انتخاب کنید:")
		sponsorMenu, _ := h.KeyboardBuilder.BuildSponsorQualityMenu(tenant.ID, lang)

		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ReplyMarkup = sponsorMenu
		_, _ = h.SafeBot.Bot.Send(msg)

	case supportBtnText:
		_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateWaitSupportOrderCode)
		msgText := "پشتیبانی ربات 🎧\n\nجهت ثبت تیکت، لطفاً **کد رهگیری** سفارش خود را ارسال کنید:"

		// نمایش دکمه بازگشت
		accMenu, _ := h.KeyboardBuilder.BuildUserAccountMenu(tenant.ID, lang)
		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = accMenu // موقتاً از کیبورد حاوی بازگشت استفاده می‌کنیم
		_, _ = h.SafeBot.Bot.Send(msg)

	default:
		if strings.HasPrefix(text, "انتقال") && currentUser != nil && h.WalletService != nil {
			// پردازش لینک‌های رفرال ۴ گانه مستقل (در صورت وجود)
			parts := strings.Split(text, " ")
			if len(parts) > 1 {
				payload := parts[1] // نمونه: league_123456
				payloadParts := strings.Split(payload, "_")

				if len(payloadParts) == 2 {
					refType := payloadParts[0] // باید یکی از این ۴ مقدار باشد: league, draw, point, wallet
					referrerID, err := strconv.ParseUint(payloadParts[1], 10, 32)

					if err == nil && (refType == "league" || refType == "draw" || refType == "point" || refType == "wallet") {
						if h.ReferralService != nil {
							// ثبت رفرال به صورت معلق (PENDING)
							_ = h.ReferralService.ProcessReferral(tenant.ID, uint(referrerID), chatID, refType)
						}
					}
				}
			}
			//parts := strings.Split(text, " ")
			//if len(parts) != 4 {
			//	_ = h.SafeBot.SendMessage(chatID, "فرمت دستور اشتباه است. روی دکمه انتقال کلیک کنید تا راهنما را ببینید.")
			//	return
			//}

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

// IsSuperAdmin بررسی اینکه آیا کاربر سوپرادمین است
func (h *BotHandler) IsSuperAdmin(telegramID int64) bool {
	for _, id := range h.SuperAdminIDs {
		if id == telegramID {
			return true
		}
	}
	return false
}

// IsAdmin بررسی اینکه آیا کاربر مدیر عادی است
func (h *BotHandler) IsAdmin(telegramID int64) bool {
	for _, id := range h.AdminIDs {
		if id == telegramID {
			return true
		}
	}
	return false
}

// HandlePhoto مدیریت دریافت تصاویر (مثل فیش واریزی)
func (h *BotHandler) HandlePhoto(tenant *domain.Tenant, message *tgbotapi.Message, lang string) {
	chatID := message.Chat.ID
	currentUser, err := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
	if err != nil || currentUser == nil {
		return
	}

	if currentUser.State == domain.StateWaitPaymentReceipt {
		// گرفتن باکیفیت‌ترین سایز عکس ارسالی
		photo := message.Photo[len(message.Photo)-1]
		fileID := photo.FileID

		draftOrder, err := h.OrderRepo.GetDraftOrder(tenant.ID, currentUser.ID)
		if err != nil {
			_ = h.SafeBot.SendMessage(chatID, "❌ سفارش معلقی یافت نشد. لطفاً مجدداً اقدام کنید.")
			_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateNormal)
			return
		}

		// ثبت فیش در دیتابیس
		receipt := &domain.Receipt{
			TenantID: tenant.ID,
			UserID:   currentUser.ID,
			OrderID:  draftOrder.ID,
			FileID:   fileID,
			Amount:   draftOrder.TotalAmount,
			Status:   "PENDING",
		}
		_ = h.ReceiptRepo.CreateReceipt(receipt)

		_ = h.SafeBot.SendMessage(chatID, "✅ فیش پرداختی شما با موفقیت ثبت شد و در صف بررسی ادمین قرار گرفت. پس از تأیید، پیام اطلاع‌رسانی برای شما ارسال خواهد شد.")
		_ = h.UserRepo.UpdateUserState(tenant.ID, currentUser.ID, domain.StateNormal)

		// ارسال به کانال گزارش برای ادمین‌ها
		if tenant.ReportChannelID != "" {
			chID, _ := strconv.ParseInt(tenant.ReportChannelID, 10, 64)
			caption := fmt.Sprintf("🧾 *فیش واریزی جدید*\n\n👤 کاربر: [%s](tg://user?id=%d)\n🛒 شماره سفارش: `%s`\n💰 مبلغ پرداخت: %.0f تومان\n\nلطفاً فیش را بررسی و عملیات را انتخاب کنید:",
				currentUser.FirstName, currentUser.TelegramID, draftOrder.TrackingCode, draftOrder.TotalAmount)

			photoMsg := tgbotapi.NewPhoto(chID, tgbotapi.FileID(fileID))
			photoMsg.Caption = caption
			photoMsg.ParseMode = "Markdown"
			photoMsg.ReplyMarkup = h.KeyboardBuilder.BuildAdminReceiptMenu(receipt.ID)
			_, _ = h.SafeBot.Bot.Send(photoMsg)
		}
	}
}

// CheckForceJoin بررسی عضویت کاربر در کانال‌های اجباری (میدل‌ور)
func (h *BotHandler) CheckForceJoin(tenant *domain.Tenant, chatID int64) bool {
	// استثنا برای سوپر ادمین‌ها و مدیران
	if h.IsSuperAdmin(chatID) || h.IsAdmin(chatID) {
		return true
	}

	channels, err := h.ForceJoinRepo.GetActiveForceJoins(tenant.ID)
	if err != nil || len(channels) == 0 {
		return true
	}

	var missingChannels []domain.ForceJoin

	for _, ch := range channels {
		// استفاده از مبدل چت آیدی برای استانداردسازی
		targetChatID, err := ResolveChatID(ch.ChatID)
		if err != nil {
			continue // اگر آیدی کانال نامعتبر بود، از آن رد می‌شود
		}

		var chatConfig tgbotapi.ChatConfigWithUser
		if strings.HasPrefix(targetChatID, "-100") {
			numericID, _ := strconv.ParseInt(targetChatID, 10, 64)
			chatConfig = tgbotapi.ChatConfigWithUser{
				ChatID: numericID,
				UserID: chatID,
			}
		} else {
			chatConfig = tgbotapi.ChatConfigWithUser{
				SuperGroupUsername: targetChatID,
				UserID:             chatID,
			}
		}

		member, err := h.SafeBot.Bot.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: chatConfig,
		})

		// اگر خطای API داشت یا عضو نبود
		if err != nil || member.Status == "left" || member.Status == "kicked" {
			missingChannels = append(missingChannels, ch)
		}
	}

	if len(missingChannels) > 0 {
		msgText := "🛑 *جهت استفاده از ربات، ابتدا باید در کانال‌های حامی زیر عضو شوید:*\n\n👇 پس از عضویت در تمامی کانال‌ها، روی دکمه «بررسی عضویت» کلیک کنید."
		markup := h.KeyboardBuilder.BuildForceJoinMenu(missingChannels)

		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = markup
		_, _ = h.SafeBot.Bot.Send(msg)
		return false
	}

	return true
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
