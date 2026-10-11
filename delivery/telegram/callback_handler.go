package telegram

import (
	"OmniGram/domain"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleCallback هندل کردن منوهای شیشه‌ای و جایگزینی صفحات
func (h *PollingHandler) HandleCallback(tenant *domain.Tenant, update tgbotapi.Update, lang string) {
	callback := update.CallbackQuery
	data := callback.Data
	chatID := callback.Message.Chat.ID
	messageID := callback.Message.MessageID

	// غیرفعال کردن آیکون ساعت شنی روی دکمه بعد از کلیک
	h.SafeBot.SafeAnswerCallback(callback.ID, "", false)

	// ساخت موقت BotHandler برای استفاده از متد CheckForceJoin
	botHandlerTemp := &BotHandler{
		ForceJoinRepo:   h.ForceJoinRepo,
		SafeBot:         h.SafeBot,
		KeyboardBuilder: h.KeyboardBuilder,
		SuperAdminIDs:   h.SuperAdminIDs,
		AdminIDs:        h.AdminIDs,
	}

	// هندل کردن دکمه‌ی اختصاصی بررسی جوین اجباری
	if data == "verify_force_join" {
		if botHandlerTemp.CheckForceJoin(tenant, chatID) {
			h.SafeBot.SafeDeleteMessage(chatID, messageID)
			h.SafeBot.SafeAnswerCallback(callback.ID, "✅ عضویت شما تایید شد! ربات فعال گردید.", true)

			// ارسال مجدد منوی اصلی بعد از تأیید
			mainMenu, _ := h.KeyboardBuilder.BuildMainMenu(tenant.ID, lang)
			msg := tgbotapi.NewMessage(chatID, "🔓 قفل ربات باز شد. لطفاً یکی از گزینه‌های زیر را انتخاب کنید:")
			msg.ReplyMarkup = mainMenu
			_, _ = h.SafeBot.Bot.Send(msg)
		} else {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ شما هنوز در تمام کانال‌ها عضو نشده‌اید!", true)
		}
		return
	}

	// قفل کردن سایر دکمه‌های شیشه‌ای در صورت عدم عضویت (به جز ادمین)
	if !botHandlerTemp.CheckForceJoin(tenant, chatID) {
		h.SafeBot.SafeAnswerCallback(callback.ID, "🛑 ابتدا در کانال‌های اجباری عضو شوید.", true)
		return
	}

	// --- سیستم فروشگاه داینامیک ---
	if data == "shop_main" {
		botHandler := &BotHandler{CategoryRepo: h.CategoryRepo}
		categories, _ := botHandler.CategoryRepo.GetMainCategories(tenant.ID)
		if len(categories) == 0 {
			h.SafeBot.SafeAnswerCallback(callback.ID, "درحال حاضر دسته‌بندی فعالی وجود ندارد.", true)
			return
		}
		markup := h.KeyboardBuilder.BuildDynamicCategoryMenu(categories, false)
		h.SafeBot.SafeEditText(chatID, messageID, "🛍 *فروشگاه خدمات مجازی*\n\nلطفاً دسته‌بندی مورد نظر خود را انتخاب کنید:", &markup)
		return
	}

	if strings.HasPrefix(data, "shop_cat_") {
		catIDStr := strings.TrimPrefix(data, "shop_cat_")
		catID, _ := strconv.Atoi(catIDStr)

		botHandler := &BotHandler{CategoryRepo: h.CategoryRepo, ServiceRepo: h.ServiceRepo}

		// بررسی وجود زیردسته
		subCategories, _ := botHandler.CategoryRepo.GetSubCategories(tenant.ID, uint(catID))
		if len(subCategories) > 0 {
			markup := h.KeyboardBuilder.BuildDynamicCategoryMenu(subCategories, true)
			h.SafeBot.SafeEditText(chatID, messageID, "📂 لطفاً زیرمجموعه مورد نظر را انتخاب کنید:", &markup)
			return
		}

		// اگر زیردسته نداشت، مستقیما لیست سرویس‌های این دسته را بیاور
		services, _ := botHandler.ServiceRepo.GetServicesByCategory(tenant.ID, uint(catID))
		if len(services) == 0 {
			h.SafeBot.SafeAnswerCallback(callback.ID, "سرویسی در این دسته یافت نشد.", true)
			return
		}
		markup := h.KeyboardBuilder.BuildDynamicServiceMenu(services, uint(catID))
		h.SafeBot.SafeEditText(chatID, messageID, "📦 لیست سرویس‌های فعال:\nلطفاً سرویس مورد نظر را جهت خرید انتخاب کنید:", &markup)
		return
	}

	if strings.HasPrefix(data, "shop_srv_") {
		srvIDStr := strings.TrimPrefix(data, "shop_srv_")
		srvID, _ := strconv.Atoi(srvIDStr)

		botHandler := &BotHandler{
			ServiceRepo: h.ServiceRepo,
			UserRepo:    h.UserRepo,
			OrderRepo:   h.OrderRepo,
		}
		user, _ := botHandler.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		service, _ := botHandler.ServiceRepo.GetServiceByID(tenant.ID, uint(srvID))

		if user != nil && service != nil {
			// سیستم هوشمند VIP: بررسی میزان خرید ماه گذشته و اعمال تخفیف
			totalSpent, _ := botHandler.OrderRepo.GetUserTotalSpentLast30Days(tenant.ID, user.ID)
			discountPercent := 0.0
			if totalSpent > 500000 {
				discountPercent = 5.0
			} else if totalSpent > 200000 {
				discountPercent = 3.0
			} else if totalSpent > 100000 {
				discountPercent = 1.0
			}

			finalPrice := service.PriceToman * (1.0 - (discountPercent / 100.0))

			// ساخت پیش‌نویس سفارش خدمات برای کاربر
			trackingCode := fmt.Sprintf("SRV_%d_%v", user.ID, callback.Message.MessageID)
			draftOrder := &domain.Order{
				TenantID:      tenant.ID,
				UserID:        user.ID,
				ServiceItemID: service.ID,
				TrackingCode:  trackingCode,
				Status:        "DRAFT",
				TotalAmount:   finalPrice, // موقتاً قیمت واحد، در مرحله بعد در تعداد ضرب خواهد شد
			}
			_ = botHandler.OrderRepo.CreateOrder(draftOrder)

			_ = botHandler.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateWaitServiceLink)

			msgText := fmt.Sprintf("✅ *سرویس انتخاب شد:*\n`%s`\n\n💰 قیمت پایه (واحد): %.0f تومان\n💎 درصد تخفیف VIP شما: %.0f%%\n\n🔗 لطفاً لینک هدف (لینک پست، کانال یا آیدی پروفایل) را ارسال کنید:", service.Title, service.PriceToman, discountPercent)
			h.SafeBot.SafeEditText(chatID, messageID, msgText, nil)
		}
		return
	}

	// کاربر روی یکی از کیفیت‌های اسپانسر کلیک کرده
	if strings.HasPrefix(data, "sponsor_quality_") {
		qualityType := strings.TrimPrefix(data, "sponsor_quality_")

		// پیدا کردن کاربر
		user, err := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if err != nil || user == nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "کاربر یافت نشد.", true)
			return
		}

		// واکشی سرویس مربوطه از دیتابیس بر اساس کلید (مثلا quality_shop)
		serviceKey := "quality_" + qualityType
		serviceItem, err := h.ServiceRepo.GetServiceByKey(tenant.ID, serviceKey)
		if err != nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "این کیفیت در حال حاضر موجود نیست.", true)
			return
		}

		// ایجاد سفارش پیش‌نویس (DRAFT) برای ذخیره موقت اطلاعات
		trackingCode := fmt.Sprintf("SP_%d_%v", user.ID, callback.Message.MessageID)
		draftOrder := &domain.Order{
			TenantID:      tenant.ID,
			UserID:        user.ID,
			ServiceItemID: serviceItem.ID,
			TrackingCode:  trackingCode,
			Status:        "DRAFT", // وضعیت سفارش تا قبل از پرداخت نهایی، پیش‌نویس است
		}
		_ = h.OrderRepo.CreateOrder(draftOrder)

		// تغییر وضعیت کاربر به "در انتظار لینک"
		user.State = domain.StateWaitSponsorLink

		_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateWaitSponsorLink)

		// پیام درخواست لینک
		msgText := "🔗 لطفاً لینک کانال، گروه یا لینک رفرال خود را جهت ثبت در سیستم ارسال کنید:"
		h.SafeBot.SafeEditText(chatID, messageID, msgText, nil)
		return

	}

	// کاربر روی یکی از روش‌های پرداخت کلیک کرده است
	if strings.HasPrefix(data, "pay_") {
		parts := strings.Split(data, "_")
		if len(parts) < 3 {
			h.SafeBot.SafeAnswerCallback(callback.ID, "دیتای نامعتبر", true)
			return
		}

		payMethod := parts[1] // toman, point, stars, gate
		orderID, _ := strconv.Atoi(parts[2])

		// ۱. واکشی سفارش
		order, err := h.OrderRepo.GetOrderByID(tenant.ID, uint(orderID))
		if err != nil || order.Status != "DRAFT" {
			h.SafeBot.SafeAnswerCallback(callback.ID, "سفارش یافت نشد یا قبلاً پردازش شده است.", true)
			return
		}

		// ۲. واکشی کاربر
		user, err := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if err != nil || user == nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "کاربر یافت نشد.", true)
			return
		}

		switch payMethod {

		//case "toman":
		//	if user.TomanBalance < order.TotalAmount {
		//		h.SafeBot.SafeAnswerCallback(callback.ID, "❌ موجودی کیف پول تومانی شما کافی نیست.", true)
		//		return
		//	}
		//
		//	// کسر از کیف پول و آپدیت دیتابیس
		//	user.TomanBalance -= order.TotalAmount
		//	_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)
		//
		//	// آپدیت وضعیت سفارش
		//	order.PaymentMethod = "WALLET_TOMAN"
		//	order.PaymentStatus = "PAID"
		//	order.Status = "COMPLETED" // تغییر وضعیت از DRAFT به تکمیل شده
		//	_ = h.OrderRepo.UpdateOrder(order)
		//
		//	// TODO: ثبت خودکار در بخش تسک (نیازمند فایل task_repo.go)
		//
		//	successMsg := fmt.Sprintf("✅ پرداخت %.0f تومان از کیف پول انجام شد.\n\n🔗 لینک شما در سیستم ثبت شد و به زودی وارد بخش وظایف می‌شود.", order.TotalAmount)
		//	h.SafeBot.SafeEditText(chatID, messageID, successMsg, nil)
		//
		//	if h.Logger != nil {
		//		h.Logger.Info(fmt.Sprintf("Order %d PAID via Toman by User %d", order.ID, user.ID))
		//	}
		//
		//case "point":
		//	// توجه: در اینجا فرض شده امتیاز معادل تومان محاسبه می‌شود. در صورت نیاز به ضریب تبدیل، متغیر order.TotalAmount را ضرب/تقسیم کنید.
		//	if user.PointBalance < order.TotalAmount {
		//		h.SafeBot.SafeAnswerCallback(callback.ID, "❌ موجودی کیف پول امتیازی شما کافی نیست.", true)
		//		return
		//	}
		//
		//	user.PointBalance -= order.TotalAmount
		//	_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)
		//
		//	order.PaymentMethod = "WALLET_POINT"
		//	order.PaymentStatus = "PAID"
		//	order.Status = "COMPLETED"
		//	_ = h.OrderRepo.UpdateOrder(order)
		//
		//	// TODO: ثبت خودکار در بخش تسک (نیازمند فایل task_repo.go)
		//
		//	successMsg := fmt.Sprintf("✅ پرداخت %.0f امتیاز از کیف پول انجام شد.\n\n🔗 کانال شما جهت عضوگیری در سیستم ثبت شد.", order.TotalAmount)
		//	h.SafeBot.SafeEditText(chatID, messageID, successMsg, nil)

		case "toman":
			if user.TomanBalance < order.TotalAmount {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ موجودی کیف پول تومانی شما کافی نیست.", true)
				return
			}

			// ۱. کسر از کیف پول و آپدیت دیتابیس
			user.TomanBalance -= order.TotalAmount
			_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)

			// ۲. آپدیت وضعیت سفارش
			order.PaymentMethod = "WALLET_TOMAN"
			order.PaymentStatus = "PAID"
			order.Status = "COMPLETED" // تغییر وضعیت از DRAFT به تکمیل شده
			_ = h.OrderRepo.UpdateOrder(order)

			// ۳. ثبت خودکار سفارش به عنوان تسک جدید
			newTask := &domain.Task{
				TenantID:      tenant.ID,
				OrderID:       order.ID, // ارتباط با فاکتور اسپانسر
				Title:         "📢 عضویت در کانال اسپانسر",
				TargetLink:    order.TargetLink,
				Capacity:      order.Quantity, // تعداد ممبری که اسپانسر پولش را داده
				TaskType:      "auto_join",
				RewardType:    "point",
				RewardAmount:  10.0, // مقدار پاداش (می‌تواند بعداً از جدول تنظیمات کارفرما خوانده شود)
				StayDuration:  24,   // ۲۴ ساعت ماندگاری اجباری در کانال
				PenaltyAmount: 50.0, // جریمه لفت دادن
				Status:        "ACTIVE",
			}
			_ = h.TaskRepo.CreateTask(newTask)

			successMsg := fmt.Sprintf("✅ پرداخت %.0f تومان از کیف پول انجام شد.\n\n🔗 لینک شما با موفقیت در سیستم ثبت شد و هم‌اکنون در بخش «وظایف» برای کاربران فعال است.", order.TotalAmount)
			h.SafeBot.SafeEditText(chatID, messageID, successMsg, nil)

			if h.Logger != nil {
				h.Logger.Info(fmt.Sprintf("Order %d PAID via Toman & Task Created for User %d", order.ID, user.ID))
			}

		case "point":
			if user.PointBalance < order.TotalAmount {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ موجودی کیف پول امتیازی شما کافی نیست.", true)
				return
			}

			// ۱. کسر از کیف پول و آپدیت دیتابیس
			user.PointBalance -= order.TotalAmount
			_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)

			// ۲. آپدیت وضعیت سفارش
			order.PaymentMethod = "WALLET_POINT"
			order.PaymentStatus = "PAID"
			order.Status = "COMPLETED"
			_ = h.OrderRepo.UpdateOrder(order)

			// ۳. ثبت خودکار سفارش به عنوان تسک جدید
			newTask := &domain.Task{
				TenantID:      tenant.ID,
				OrderID:       order.ID,
				Title:         "📢 عضویت در کانال اسپانسر",
				TargetLink:    order.TargetLink,
				Capacity:      order.Quantity,
				TaskType:      "auto_join",
				RewardType:    "point",
				RewardAmount:  10.0,
				StayDuration:  24,
				PenaltyAmount: 50.0,
				Status:        "ACTIVE",
			}
			_ = h.TaskRepo.CreateTask(newTask)

			successMsg := fmt.Sprintf("✅ پرداخت %.0f امتیاز از کیف پول انجام شد.\n\n🔗 کانال شما جهت عضوگیری در بخش وظایف ثبت شد.", order.TotalAmount)
			h.SafeBot.SafeEditText(chatID, messageID, successMsg, nil)

		case "gate":
			h.SafeBot.SafeAnswerCallback(callback.ID, "⏳ در حال ایجاد فاکتور درگاه...", false)
			// منطق تولید لینک پرداخت زرین‌پال یا نکست‌‌پی در آینده اینجا قرار می‌گیرد
			gateMsg := fmt.Sprintf("💳 برای پرداخت فاکتور %.0f تومانی روی لینک زیر کلیک کنید:\n\n[🔗 ورود به درگاه پرداخت](https://example.com/pay/%d)", order.TotalAmount, order.ID)
			h.SafeBot.SafeEditText(chatID, messageID, gateMsg, nil)

		case "stars":
			h.SafeBot.SafeAnswerCallback(callback.ID, "⏳ سیستم پرداخت استارز تلگرام به زودی متصل می‌شود.", true)
			// ارسال Invoice اختصاصی تلگرام (SendInvoice) برای استارز

		case "card":
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateWaitPaymentReceipt)
			// شماره کارت را ادمین می‌تواند در آینده از طریق پنل در دیتابیس ذخیره کند، فعلاً هاردکد موقت نمایشی:
			cardMsg := fmt.Sprintf("💳 لطفاً مبلغ %.0f تومان را به شماره کارت زیر واریز کنید:\n\n`1234-5678-9012-3456`\nبه نام: مدیریت\n\n📸 سپس عکس رسید پرداخت را همینجا ارسال کنید:", order.TotalAmount)
			h.SafeBot.SafeEditText(chatID, messageID, cardMsg, nil)
		}

		return
	}

	// کاربر روی دکمه ثبت تیکتِ یک سفارش کلیک کرده است
	if strings.HasPrefix(data, "new_ticket_") {
		orderIDStr := strings.TrimPrefix(data, "new_ticket_")
		orderID, _ := strconv.Atoi(orderIDStr)

		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)

		// ساخت تیکت و متصل کردن به سفارش
		ticket := &domain.SupportTicket{
			TenantID: tenant.ID,
			UserID:   user.ID,
			OrderID:  uint(orderID),
			Status:   "OPEN",
		}
		_ = h.TicketRepo.CreateTicket(ticket)
		_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateWaitSupportMsg)

		h.SafeBot.SafeEditText(chatID, messageID, "✅ سفارش یافت شد.\n\n📝 لطفاً مشکل یا گزارش خود را در مورد این سفارش کامل بنویسید:", nil)
		return
	}

	// [delivery/telegram/callback_handler.go]
	// feat(callback): route admin panel button clicks

	// [delivery/telegram/callback_handler.go]
	// feat(callback): verify channel membership and reward user for tasks

	// [delivery/telegram/callback_handler.go]
	// کاربر روی دکمه پیوستن به بازی گروهی کلیک کرده است
	if strings.HasPrefix(data, "join_game_") {
		gameIDStr := strings.TrimPrefix(data, "join_game_")
		gameID, _ := strconv.Atoi(gameIDStr)

		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if user == nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ ابتدا ربات را در پی‌وی استارت بزنید.", true)
			return
		}

		game, err := h.GameRepo.GetGameByID(tenant.ID, uint(gameID))
		if err != nil || game.Status != "WAITING" {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ این بازی لغو شده یا نفر دوم قبلاً عضو شده است.", true)
			return
		}

		if game.CreatorID == user.ID {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ شما سازنده این میز هستید و نمی‌توانید با خودتان بازی کنید!", true)
			return
		}

		if user.PointBalance < game.BetAmount {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ موجودی امتیاز شما برای پیوستن به این شرط کافی نیست.", true)
			return
		}

		activeGame, _ := h.GameRepo.GetActiveGameForUser(tenant.ID, user.ID)
		if activeGame != nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ شما یک بازی در جریان دارید. ابتدا آن را تمام کنید.", true)
			return
		}

		user.PointBalance -= game.BetAmount
		_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)

		game.JoinerID = &user.ID
		game.Status = "IN_PROGRESS"
		_ = h.GameRepo.UpdateGame(game)

		h.SafeBot.SafeAnswerCallback(callback.ID, "✅ با موفقیت به بازی پیوستید!", false)

		emoji := "🎲"
		if game.GameType == "CASINO" {
			emoji = "🎰"
		}
		startMsg := fmt.Sprintf("🔥 ظرفیت بازی تکمیل شد!\n\nهر دو شرکت‌کننده لطفاً ایموجی %s را در گروه ارسال کنید.", emoji)
		h.SafeBot.SafeEditText(chatID, messageID, startMsg, nil)
		return
	}

	// کاربر روی بررسی عضویت تسک کلیک کرده است
	if strings.HasPrefix(data, "verify_task_") {
		taskIDStr := strings.TrimPrefix(data, "verify_task_")
		taskID, _ := strconv.Atoi(taskIDStr)

		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if user == nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "کاربر یافت نشد.", true)
			return
		}

		task, err := h.TaskRepo.GetTaskByID(tenant.ID, uint(taskID))
		if err != nil || task.Status != "ACTIVE" || task.Capacity <= 0 {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ این تسک منقضی شده یا ظرفیت آن به پایان رسیده است.", true)
			h.SafeBot.SafeDeleteMessage(chatID, messageID)
			return
		}

		// جلوگیری مضاعف از پاداش تکراری

		if h.TaskRepo.HasUserCompletedTask(tenant.ID, user.ID, task.ID) {
			h.SafeBot.SafeAnswerCallback(callback.ID, "✅ شما قبلاً پاداش این تسک را دریافت کرده‌اید.", true)
			h.SafeBot.SafeDeleteMessage(chatID, messageID)
			return
		}

		if task.TaskType == "manual_proof" {
			_ = h.UserRepo.UpdateUserTaskID(tenant.ID, user.ID, task.ID)
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateWaitTask)

			h.SafeBot.SafeAnswerCallback(callback.ID, "📸 لطفاً عکس (اسکرین‌شات) انجام این تسک را ارسال کنید.", true)
			h.SafeBot.SafeDeleteMessage(chatID, messageID) // پیام دکمه‌ها را پاک می‌کنیم
			msg := tgbotapi.NewMessage(chatID, "📸 لطفاً **اسکرین‌شات** اثبات انجام این تسک را دقیقاً در همینجا ارسال کنید:\n\n*(فقط یک عکس ارسال کنید)*")
			msg.ParseMode = "Markdown"
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}

		// تبدیل لینک ذخیره شده به فرمت مناسب API
		targetChatID, err := ResolveChatID(task.TargetLink)
		if err != nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ ساختار لینک این تسک نامعتبر است.", true)
			return
		}

		// تشخیص اینکه آیدی عددی است یا یوزرنیم
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

		if err != nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ خطا در بررسی عضویت. ربات باید در کانال هدف ادمین باشد.", true)
			return
		}

		if member.Status == "left" || member.Status == "kicked" {
			h.SafeBot.SafeAnswerCallback(callback.ID, "⚠️ شما هنوز در کانال عضو نشده‌اید!", true)
			return
		}

		// کاربر عضو شده است -> ثبت تسک و واریز پاداش
		submission := &domain.TaskSubmission{
			TenantID: tenant.ID,
			UserID:   user.ID,
			TaskID:   task.ID,
			Status:   "approved",
		}
		_ = h.TaskRepo.SubmitTask(submission)

		// کسر از ظرفیت تسکِ اسپانسر
		_ = h.TaskRepo.DecreaseCapacity(tenant.ID, task.ID)

		// واریز پاداش
		if task.RewardType == "toman" {
			user.TomanBalance += task.RewardAmount
		} else {
			user.PointBalance += task.RewardAmount
		}
		_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)

		h.SafeBot.SafeAnswerCallback(callback.ID, "✅ عضویت شما تایید شد و پاداش دریافت کردید!", true)

		successMsg := fmt.Sprintf("🎉 پاداش انجام تسک (%.0f %s) با موفقیت به موجودی شما افزوده شد.", task.RewardAmount, task.RewardType)
		h.SafeBot.SafeEditText(chatID, messageID, successMsg, nil)

		if h.Logger != nil {
			h.Logger.Info(fmt.Sprintf("User %d completed task %d successfully", user.ID, task.ID))
		}
		return
	}

	// کاربر روی دکمه شرکت در قرعه‌کشی کلیک کرده است
	if strings.HasPrefix(data, "enter_draw_") {
		drawIDStr := strings.TrimPrefix(data, "enter_draw_")
		drawID, _ := strconv.Atoi(drawIDStr)

		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if user == nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "کاربر یافت نشد.", true)
			return
		}

		draw, err := h.DrawRepo.GetDrawByID(tenant.ID, uint(drawID))
		if err != nil || draw.Status != "ACTIVE" {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ این قرعه‌کشی پایان یافته یا لغو شده است.", true)
			return
		}

		if user.DrawTickets < draw.TicketCost {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ تیکت‌های شما برای شرکت در این قرعه‌کشی کافی نیست!", true)
			return
		}

		totalEntries, _ := h.DrawRepo.GetTotalEntries(tenant.ID, draw.ID)
		if draw.MaxParticipants > 0 && int(totalEntries) >= draw.MaxParticipants {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ ظرفیت این قرعه‌کشی تکمیل شده است.", true)
			return
		}

		// کسر تیکت و ثبت شانس
		user.DrawTickets -= draw.TicketCost
		_ = h.UserRepo.UpdateDrawTickets(tenant.ID, user.ID, user.DrawTickets)

		entry := &domain.DrawEntry{
			TenantID:    tenant.ID,
			LuckyDrawID: draw.ID,
			UserID:      user.ID,
			Entries:     1,
		}
		_ = h.DrawRepo.AddEntry(entry)

		h.SafeBot.SafeAnswerCallback(callback.ID, "✅ با موفقیت در قرعه‌کشی شرکت کردید!", true)

		successMsg := fmt.Sprintf("🎉 شما 1 شانس در قرعه‌کشی *%s* ثبت کردید.\n🎟 تیکت‌های باقی‌مانده: %d", draw.Title, user.DrawTickets)
		h.SafeBot.SafeEditText(chatID, messageID, successMsg, nil)
		return
	}

	// کاربر روی یکی از عملیات پیگیری سفارش کلیک کرده است
	if strings.HasPrefix(data, "ordact_") {
		parts := strings.Split(data, "_")
		if len(parts) < 3 {
			return
		}
		action := parts[1] // speedup, edit, refill, cancel
		orderID, _ := strconv.Atoi(parts[2])

		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		order, err := h.OrderRepo.GetOrderByID(tenant.ID, uint(orderID))
		if err != nil || user == nil || order.UserID != user.ID {
			h.SafeBot.SafeAnswerCallback(callback.ID, "سفارش یافت نشد.", true)
			return
		}

		if order.Status == "CANCELED" {
			h.SafeBot.SafeAnswerCallback(callback.ID, "این سفارش قبلاً لغو شده است.", true)
			return
		}

		switch action {
		case "speedup":
			if order.SpeedUpRequested {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ درخواست سرعت‌دهی برای این سفارش قبلاً ثبت شده است.", true)
				return
			}
			if user.PointBalance < 2.0 {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ برای سرعت‌دهی به سفارش نیاز به ۲ امتیاز دارید.", true)
				return
			}

			// کسر ۲ امتیاز و ثبت درخواست
			user.PointBalance -= 2.0
			_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)
			order.SpeedUpRequested = true
			_ = h.OrderRepo.UpdateOrder(order)

			h.SafeBot.SafeAnswerCallback(callback.ID, "✅ ۲ امتیاز کسر و درخواست سرعت‌دهی ثبت شد.", true)

			// ارسال نوتیفیکیشن به ادمین
			if tenant.ReportChannelID != "" {
				chID, _ := strconv.ParseInt(tenant.ReportChannelID, 10, 64)
				msgText := fmt.Sprintf("⚡️ #سرعت_دادن\n\nکاربر [%s](tg://user?id=%d) درخواست سرعت‌دهی برای سفارش `%s` را ثبت کرد.", user.FirstName, user.TelegramID, order.TrackingCode)
				_ = h.SafeBot.SendMessage(chID, msgText)
			}

		case "refill":
			if order.HasRefill {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ درخواست ریفیل این سفارش قبلاً در صف ادمین قرار گرفته است.", true)
				return
			}
			order.HasRefill = true
			_ = h.OrderRepo.UpdateOrder(order)
			h.SafeBot.SafeAnswerCallback(callback.ID, "✅ درخواست ریفیل و جبران ریزش برای ادمین ارسال شد.", true)

			if tenant.ReportChannelID != "" {
				chID, _ := strconv.ParseInt(tenant.ReportChannelID, 10, 64)
				msgText := fmt.Sprintf("♻️ #درخواست_ریفیل\n\nسفارش `%s` دچار ریزش شده و نیاز به جبران دارد.", order.TrackingCode)
				_ = h.SafeBot.SendMessage(chID, msgText)
			}

		case "edit":
			// بررسی محدودیت ۳۰ دقیقه‌ای
			if time.Since(order.CreatedAt).Minutes() > 30 {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ زمان مجاز برای ویرایش لینک (۳۰ دقیقه) به پایان رسیده است.", true)
				return
			}
			_ = h.UserRepo.UpdateUserOrderID(tenant.ID, user.ID, order.ID)
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateWaitNewOrderLink)

			h.SafeBot.SafeAnswerCallback(callback.ID, "✏️ لطفاً لینک جدید را پیام کنید.", false)
			msg := tgbotapi.NewMessage(chatID, "✏️ لطفاً **لینک جدید** خود را برای این سفارش ارسال کنید:\n*(شما هنوز زمان مجاز برای ویرایش را دارید)*")
			msg.ParseMode = "Markdown"
			_, _ = h.SafeBot.Bot.Send(msg)

		case "cancel":
			if order.CancelRequested {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ درخواست لغو قبلاً بررسی یا اعمال شده است.", true)
				return
			}
			// بازگردانی مبلغ با کسر ۱۰٪ کارمزد
			refundAmount := order.TotalAmount * 0.90

			if order.PaymentMethod == "WALLET_TOMAN" || order.PaymentMethod == "CARD" || order.PaymentMethod == "GATEWAY" {
				user.TomanBalance += refundAmount
			} else if order.PaymentMethod == "WALLET_POINT" {
				user.PointBalance += refundAmount
			} else if order.PaymentMethod == "STARS" {
				user.StarsBalance += refundAmount
			}
			_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)

			// لغو سفارش و تسک وابسته
			order.Status = "CANCELED"
			order.CancelRequested = true
			_ = h.OrderRepo.UpdateOrder(order)

			h.TaskRepo.DB.Model(&domain.Task{}).Where("tenant_id = ? AND order_id = ?", tenant.ID, order.ID).Update("status", "PAUSED")

			h.SafeBot.SafeAnswerCallback(callback.ID, "✅ سفارش لغو شد.", false)
			successMsg := fmt.Sprintf("✅ سفارش شما لغو گردید و مبلغ %.0f (پس از کسر ۱۰٪ کارمزد لغو) به کیف پول شما بازگشت داده شد.", refundAmount)
			h.SafeBot.SafeEditText(chatID, messageID, successMsg, nil)

			if tenant.ReportChannelID != "" {
				chID, _ := strconv.ParseInt(tenant.ReportChannelID, 10, 64)
				_ = h.SafeBot.SendMessage(chID, fmt.Sprintf("❌ #لغو_سفارش\n\nسفارش `%s` توسط کاربر لغو گردید.", order.TrackingCode))
			}
		}
		return
	}

	// کاربر روی منوهای برداشت کلیک کرده است
	if strings.HasPrefix(data, "withdraw_") {
		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if user == nil {
			return
		}

		if data == "withdraw_select_stars" {
			markup := h.KeyboardBuilder.BuildWithdrawStarsMenu()
			h.SafeBot.SafeEditText(chatID, messageID, "⭐ لطفاً تعداد استارز درخواستی (از ۱ تا ۹) را انتخاب کنید:\n\n*(هر ۱ استارز = ۳ امتیاز)*", &markup)
			return
		}

		if data == "withdraw_select_gift" {
			markup := h.KeyboardBuilder.BuildWithdrawGiftMenu()
			h.SafeBot.SafeEditText(chatID, messageID, "🎁 لطفاً گیفت مورد نظر خود را انتخاب کنید:", &markup)
			return
		}

		if strings.HasPrefix(data, "withdraw_do_") {
			parts := strings.Split(data, "_")
			if len(parts) != 4 {
				return
			}
			currency := parts[2] // STARS یا GIFT
			amount, _ := strconv.ParseFloat(parts[3], 64)

			// محاسبه هزینه به امتیاز
			cost := amount
			if currency == "STARS" {
				cost = amount * 3.0 // طبق قانون کارفرما: هر ۳ امتیاز = ۱ استارز
			}

			if user.PointBalance < cost {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ موجودی امتیاز شما کافی نیست.", true)
				return
			}

			// کسر موجودی و ثبت پیش‌نویس برداشت
			user.PointBalance -= cost
			_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)
			_ = h.UserRepo.UpdateWithdrawDraft(tenant.ID, user.ID, currency, amount)
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateWaitWithdrawWallet)

			h.SafeBot.SafeAnswerCallback(callback.ID, "✅ مبلغ کسر شد. لطفاً لینک را بفرستید.", false)

			h.SafeBot.SafeDeleteMessage(chatID, messageID)

			prompt := "🔗 لطفاً **لینک پست** (برای دریافت استارز) یا **آیدی کاربری** (برای دریافت گیفت) را ارسال کنید:"
			msg := tgbotapi.NewMessage(chatID, prompt)
			msg.ParseMode = "Markdown"
			accMenu, _ := h.KeyboardBuilder.BuildUserAccountMenu(tenant.ID, lang)
			msg.ReplyMarkup = accMenu
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}
	}

	// --- مسیریابی دکمه‌های پنل مدیریت ---
	if strings.HasPrefix(data, "admin_") {
		// بررسی امنیتی: آیا کسی که کلیک کرده واقعا ادمین است؟
		botHandler := &BotHandler{
			SuperAdminIDs: h.SuperAdminIDs,
			AdminIDs:      h.AdminIDs,
		}
		isSuper := botHandler.IsSuperAdmin(chatID)
		isAdmin := botHandler.IsAdmin(chatID)

		if !isSuper && !isAdmin {
			h.SafeBot.SafeAnswerCallback(callback.ID, "دسترسی غیرمجاز", true)
			return
		}

		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if user == nil {
			return
		}
		// بررسی بخش درخواستی
		// بررسی بخش درخواستی
		action := strings.TrimPrefix(data, "admin_")

		switch { // فیکس شد: حذف action از جلوی switch برای استفاده از شرط‌های boolean
		case action == "orders":
			// TODO: فراخوانی لیست سفارشات
			h.SafeBot.SafeAnswerCallback(callback.ID, "در حال بارگذاری سفارشات...", false)

		case strings.HasPrefix(action, "receipt_"):
			parts := strings.Split(action, "_")
			if len(parts) != 3 {
				return
			}
			receiptAction := parts[1] // approve یا reject
			receiptID, _ := strconv.Atoi(parts[2])

			receipt, err := h.ReceiptRepo.GetReceiptByID(tenant.ID, uint(receiptID))
			if err != nil || receipt.Status != "PENDING" {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ این فیش قبلاً بررسی شده است.", true)
				return
			}

			order, err := h.OrderRepo.GetOrderByID(tenant.ID, receipt.OrderID)
			if err != nil {
				return
			}

			userUser, _ := h.UserRepo.GetUserByID(tenant.ID, receipt.UserID)

			if receiptAction == "approve" {
				// تأیید فیش و قطعی کردن سفارش
				receipt.Status = "APPROVED"
				_ = h.ReceiptRepo.UpdateStatus(receipt)

				order.PaymentMethod = "CARD"
				order.PaymentStatus = "PAID"
				order.Status = "COMPLETED"
				_ = h.OrderRepo.UpdateOrder(order)

				// ثبت خودکار سفارش اسپانسر به عنوان تسک جدید
				newTask := &domain.Task{
					TenantID:      tenant.ID,
					OrderID:       order.ID,
					Title:         "📢 عضویت در کانال اسپانسر",
					TargetLink:    order.TargetLink,
					Capacity:      order.Quantity,
					TaskType:      "auto_join",
					RewardType:    "point",
					RewardAmount:  10.0,
					StayDuration:  24,
					PenaltyAmount: 50.0,
					Status:        "ACTIVE",
				}
				_ = h.TaskRepo.CreateTask(newTask)

				h.SafeBot.SafeAnswerCallback(callback.ID, "✅ فیش تأیید و تسک ساخته شد.", false)
				h.SafeBot.SafeEditText(chatID, messageID, "✅ این فیش توسط ادمین تأیید و به تسک تبدیل شد.", nil)
				if userUser != nil {
					_ = h.SafeBot.SendMessage(userUser.TelegramID, fmt.Sprintf("✅ فیش واریزی شما مبلغ %.0f تومان تأیید شد و سفارش شما فعال گردید.", receipt.Amount))
				}

			} else if receiptAction == "reject" {
				// رد فیش
				receipt.Status = "REJECTED"
				_ = h.ReceiptRepo.UpdateStatus(receipt)

				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ فیش رد شد.", false)
				h.SafeBot.SafeEditText(chatID, messageID, "❌ این فیش توسط ادمین رد شد.", nil)
				if userUser != nil {
					_ = h.SafeBot.SendMessage(userUser.TelegramID, "❌ فیش واریزی شما تأیید نشد. لطفاً در صورت بروز مشکل به پشتیبانی پیام دهید.")
				}
			}
			return

		case strings.HasPrefix(action, "taskproof_"):
			parts := strings.Split(action, "_")
			if len(parts) != 3 {
				return
			}
			proofAction := parts[1] // approve یا reject
			subID, _ := strconv.Atoi(parts[2])

			submission, err := h.TaskRepo.GetSubmissionByID(tenant.ID, uint(subID))
			if err != nil || submission.Status != "PENDING" {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ این اسکرین‌شات قبلاً بررسی شده است.", true)
				return
			}

			task, err := h.TaskRepo.GetTaskByID(tenant.ID, submission.TaskID)
			if err != nil {
				return
			}

			userUser, _ := h.UserRepo.GetUserByID(tenant.ID, submission.UserID)

			if proofAction == "approve" {
				submission.Status = "APPROVED"
				_ = h.TaskRepo.UpdateSubmissionStatus(submission)

				// کسر از ظرفیت تسک
				_ = h.TaskRepo.DecreaseCapacity(tenant.ID, task.ID)

				// واریز پاداش
				if userUser != nil {
					if task.RewardType == "toman" {
						userUser.TomanBalance += task.RewardAmount
					} else {
						userUser.PointBalance += task.RewardAmount
					}
					_ = h.UserRepo.UpdateBalances(tenant.ID, userUser.ID, userUser.TomanBalance, userUser.PointBalance, userUser.StarsBalance)

					_ = h.SafeBot.SendMessage(userUser.TelegramID, fmt.Sprintf("✅ اسکرین‌شات شما تأیید شد!\n🎉 پاداش %.0f %s به کیف پول شما اضافه شد.", task.RewardAmount, task.RewardType))
				}

				h.SafeBot.SafeAnswerCallback(callback.ID, "✅ تسک تأیید و پاداش واریز شد.", false)
				h.SafeBot.SafeEditText(chatID, messageID, "✅ اسکرین‌شات تأیید و پاداش به کاربر داده شد.", nil)

			} else if proofAction == "reject" {
				submission.Status = "REJECTED"
				_ = h.TaskRepo.UpdateSubmissionStatus(submission)

				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ اسکرین‌شات رد شد.", false)
				h.SafeBot.SafeEditText(chatID, messageID, "❌ این اسکرین‌شات رد شد.", nil)

				if userUser != nil {
					_ = h.SafeBot.SendMessage(userUser.TelegramID, "❌ اسکرین‌شات ارسالی شما برای تسک تأیید نشد. لطفاً مجدداً با دقت انجام دهید.")
				}
			}
			return

		case action == "tickets":
			// TODO: فراخوانی تیکت‌های پشتیبانی
			h.SafeBot.SafeAnswerCallback(callback.ID, "در حال بارگذاری تیکت‌ها...", false)

		case action == "users":
			if isSuper {
				_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitUserID)
				h.SafeBot.SafeAnswerCallback(callback.ID, "ارسال آیدی", false)
				msg := tgbotapi.NewMessage(chatID, "👤 لطفاً **آیدی عددی تلگرام** کاربر مورد نظر را برای مدیریت ارسال کنید:")
				msg.ParseMode = "Markdown"
				_, _ = h.SafeBot.Bot.Send(msg)
			}
			return

		case action == "stats":
			if isSuper {
				stats, _ := h.AdminRepo.GetStats(tenant.ID)
				h.SafeBot.SafeAnswerCallback(callback.ID, "درحال محاسبه...", false)
				msgText := fmt.Sprintf("📊 *آمار کلی سیستم (زنده)*\n\n👥 کل کاربران: %d\n🟢 کاربران فعال (۱۰ روز اخیر): %d\n🛒 کل سفارشات: %d\n\n💰 مجموع درآمد خالص (تومان): %.0f\n🎁 موجودی در گردش امتیازات: %.0f\n⭐ موجودی در گردش استارز: %.0f",
					stats.TotalUsers, stats.ActiveUsers, stats.TotalOrders, stats.TotalRevenue, stats.TotalPoints, stats.TotalStars)
				h.SafeBot.SafeEditText(chatID, messageID, msgText, nil)
			}
			return

		case action == "settings":
			if isSuper {
				markup := h.KeyboardBuilder.BuildAdminSettingsMenu()
				msgText := fmt.Sprintf("⚙️ *تنظیمات مالی سیستم*\n\nدرصد کارمزد لغو: %.1f%%\nدرصد کارمزد انتقال گروه: %.1f%%\nپاداش بازگشت غیرفعال: %.0f\n\nجهت ویرایش روی گزینه مورد نظر کلیک کنید:",
					tenant.CancelFeePercent, tenant.TransferFeePercent, tenant.InactiveBonusAmount)
				h.SafeBot.SafeEditText(chatID, messageID, msgText, &markup)
			}
			return
		case strings.HasPrefix(action, "withdraw_"):
			parts := strings.Split(action, "_")
			if len(parts) != 3 {
				return
			}
			wdAction := parts[1] // approve یا reject
			wdID, _ := strconv.Atoi(parts[2])

			withdrawal, err := h.WithdrawalRepo.GetByID(tenant.ID, uint(wdID))
			if err != nil || withdrawal.Status != "PENDING" {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ این درخواست قبلاً بررسی شده است.", true)
				return
			}

			userUser, _ := h.UserRepo.GetUserByID(tenant.ID, withdrawal.UserID)

			if wdAction == "approve" {
				withdrawal.Status = "APPROVED"
				_ = h.WithdrawalRepo.UpdateStatus(withdrawal)

				h.SafeBot.SafeAnswerCallback(callback.ID, "✅ درخواست تأیید شد.", false)
				h.SafeBot.SafeEditText(chatID, messageID, "✅ این درخواست برداشت انجام و واریز شد.", nil)
				if userUser != nil {
					_ = h.SafeBot.SendMessage(userUser.TelegramID, fmt.Sprintf("✅ درخواست برداشت %.0f %s شما تأیید و واریز شد.", withdrawal.Amount, withdrawal.Currency))
				}
			} else if wdAction == "reject" {
				withdrawal.Status = "REJECTED"
				_ = h.WithdrawalRepo.UpdateStatus(withdrawal)

				// عودت مبلغ به کاربر
				if userUser != nil {
					refundCost := withdrawal.Amount
					if withdrawal.Currency == "STARS" {
						refundCost = withdrawal.Amount * 3.0
					}
					userUser.PointBalance += refundCost
					_ = h.UserRepo.UpdateBalances(tenant.ID, userUser.ID, userUser.TomanBalance, userUser.PointBalance, userUser.StarsBalance)

					_ = h.SafeBot.SendMessage(userUser.TelegramID, "❌ درخواست برداشت شما تأیید نشد و مبلغ به حساب شما بازگشت داده شد.")
				}

				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ درخواست رد شد.", false)
				h.SafeBot.SafeEditText(chatID, messageID, "❌ این درخواست رد و مبلغ عودت داده شد.", nil)
			}
			return

		case action == "services":
			if isSuper { // فقط سوپر ادمین دسترسی دارد
				markup := h.KeyboardBuilder.BuildAdminServicesMenu()
				h.SafeBot.SafeEditText(chatID, messageID, "🔧 *مدیریت خدمات مجازی*\n\nاز این بخش می‌توانید دسته‌بندی‌ها و سرویس‌های جدید ایجاد کنید:", &markup)
			} else {
				h.SafeBot.SafeAnswerCallback(callback.ID, "⛔️ این بخش فقط برای توسعه‌دهنده (سوپرادمین) در دسترس است.", true)
			}
			return

		case action == "prices":
			if isSuper { // فقط سوپر ادمین دسترسی دارد
				markup := h.KeyboardBuilder.BuildAdminPricesMenu()
				msgText := fmt.Sprintf("💰 *مدیریت هوشمند قیمت‌ها*\n\n💵 قیمت فعلی دلار در سیستم: **%.0f تومان**\n\nبا تغییر قیمت دلار، قیمت تمامی سرویس‌ها طبق فرمول سود به صورت خودکار بروزرسانی می‌شود.", tenant.DollarPrice)
				h.SafeBot.SafeEditText(chatID, messageID, msgText, &markup)
			} else {
				h.SafeBot.SafeAnswerCallback(callback.ID, "⛔️ دسترسی غیرمجاز.", true)
			}
			return

		// ... سایر کیس‌ها به مرور تکمیل می‌شوند ...
		default:
			h.SafeBot.SafeAnswerCallback(callback.ID, "این بخش به زودی فعال می‌شود.", false)
		}

		return
	}

	// مدیریت کلیک‌های بخش خدمات و قیمت‌ها (ادمین)
	if strings.HasPrefix(data, "admsrv_") || strings.HasPrefix(data, "admprc_") {
		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if !botHandlerTemp.IsSuperAdmin(chatID) { // 👈 فیکس شد: استفاده از botHandlerTemp
			h.SafeBot.SafeAnswerCallback(callback.ID, "⛔️ دسترسی غیرمجاز.", true)
			return
		}

		if data == "admprc_setdollar" {
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitDollarPrice)
			h.SafeBot.SafeAnswerCallback(callback.ID, "لطفاً قیمت جدید دلار را ارسال کنید.", false)
			_, _ = h.SafeBot.Bot.Send(tgbotapi.NewMessage(chatID, "💵 لطفاً **قیمت جدید دلار** را به تومان (فقط عدد) ارسال کنید:"))
			return
		}

		if data == "admsrv_addcat" {
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitCategoryName)
			h.SafeBot.SafeAnswerCallback(callback.ID, "لطفاً نام دسته‌بندی را ارسال کنید.", false)
			_, _ = h.SafeBot.Bot.Send(tgbotapi.NewMessage(chatID, "📂 لطفاً **نام دسته‌بندی جدید** را ارسال کنید (مثل: ممبر تلگرام):"))
			return
		}

		if data == "admsrv_addsrv" {
			// برای سادگی در این نسخه، فقط از ادمین می‌خواهیم مستقیم اطلاعات را بفرستد
			// (در نسخه پیشرفته‌تر می‌توان ابتدا دسته‌ها را لیست کرد)
			_ = h.UserRepo.UpdateAdminTargets(tenant.ID, user.ID, 0, "", "") // ریسِت تارگت‌ها
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitServiceName)
			h.SafeBot.SafeAnswerCallback(callback.ID, "منتظر دریافت اطلاعات...", false)

			msg := tgbotapi.NewMessage(chatID, "➕ لطفاً **نام سرویس جدید** را ارسال کنید:")
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}
	}

	// مدیریت کلیک‌های زیرمنوی اکشن کاربر (ادمین)
	if strings.HasPrefix(data, "admuser_") {
		parts := strings.Split(data, "_")
		if len(parts) != 3 {
			return
		}
		admAction := parts[1] // addtoman, subtoman, addpoint, subpoint, refs
		targetID, _ := strconv.Atoi(parts[2])

		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if !botHandlerTemp.IsSuperAdmin(chatID) && !botHandlerTemp.IsAdmin(chatID) { // 👈 فیکس شد
			return
		}

		if admAction == "refs" {
			stats, _ := h.RefRepo.GetReferralStats(tenant.ID, uint(targetID)) // 👈 فیکس شد: استفاده از RefRepo
			h.SafeBot.SafeAnswerCallback(callback.ID, "آمار رفرال کاربر استخراج شد", false)
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📊 *آمار رفرال این کاربر*\n\nلیگ: %d\nقرعه‌کشی: %d\nامتیاز: %d\nکیف‌پول: %d\nغیرفعال‌ها: %d",
				stats.LeagueTotal, stats.DrawTotal, stats.PointTotal, stats.WalletTotal, stats.InactiveUsers))
			msg.ParseMode = "Markdown"
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}

		// برای تراکنش‌ها
		_ = h.UserRepo.UpdateAdminTargets(tenant.ID, user.ID, uint(targetID), "", admAction)
		_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitBalanceAmount)

		h.SafeBot.SafeAnswerCallback(callback.ID, "لطفاً مبلغ را ارسال کنید.", false)
		_, _ = h.SafeBot.Bot.Send(tgbotapi.NewMessage(chatID, "💰 لطفاً مبلغ مورد نظر برای شارژ/کسر را به صورت **عدد** ارسال کنید:"))
		return
	}

	// مدیریت کلیک‌های تنظیمات سیستمی (ادمین)
	if strings.HasPrefix(data, "admset_") {
		settingType := strings.TrimPrefix(data, "admset_")
		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if !botHandlerTemp.IsSuperAdmin(chatID) { // 👈 فیکس شد
			return
		}

		_ = h.UserRepo.UpdateAdminTargets(tenant.ID, user.ID, 0, settingType, "")
		_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitSettingValue)

		h.SafeBot.SafeAnswerCallback(callback.ID, "لطفاً مقدار جدید را ارسال کنید.", false)
		_, _ = h.SafeBot.Bot.Send(tgbotapi.NewMessage(chatID, "⚙️ لطفاً مقدار جدید (عدد یا درصد) را پیام کنید:"))
		return
	}

}

//// buildMainCategoryMenu نمایش دسته‌های اصلی (شبیه‌ساز دیتابیس)
//func (h *PollingHandler) buildMainCategoryMenu() (*tgbotapi.InlineKeyboardMarkup, string) {
//	keyboard := tgbotapi.NewInlineKeyboardMarkup(
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("〔ممبر〕", "cat_1"),
//			tgbotapi.NewInlineKeyboardButtonData("〔ویو〕", "cat_2"),
//		),
//		tgbotapi.NewInlineKeyboardRow(
//			tgbotapi.NewInlineKeyboardButtonData("〔ری‌اکشن〕", "cat_3"),
//			tgbotapi.NewInlineKeyboardButtonData("〔بوست〕", "cat_4"),
//		),
//	)
//	return &keyboard, "لطفاً بخش مورد نظر را انتخاب کنید:"
//}
//
//// buildSubCategoryMenu نمایش زیرمجموعه‌های یک دسته همراه با دکمه بازگشت
//func (h *PollingHandler) buildSubCategoryMenu(parentID uint) (*tgbotapi.InlineKeyboardMarkup, string) {
//	var keyboard tgbotapi.InlineKeyboardMarkup
//
//	if parentID == 1 { // فرض کنیم 1 آیدی ممبر است
//		keyboard = tgbotapi.NewInlineKeyboardMarkup(
//			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("‐ ممبر ایرانی", "item_1")),
//			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("‐ ممبر خارجی", "item_2")),
//			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت به منوی قبل", "menu_services")),
//		)
//	} else {
//		// برای سایر دسته‌ها
//		keyboard = tgbotapi.NewInlineKeyboardMarkup(
//			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت به منوی قبل", "menu_services")),
//		)
//	}
//
//	return &keyboard, "نوع سرویس را انتخاب کنید:"
//}
