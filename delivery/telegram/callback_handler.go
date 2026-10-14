package telegram

import (
	"OmniGram/domain"
	"fmt"
	"math/rand"
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
	userID := callback.From.ID
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
	//if data == "verify_force_join" {
	//	if botHandlerTemp.CheckForceJoin(tenant, chatID) {
	//		h.SafeBot.SafeDeleteMessage(chatID, messageID)
	//		h.SafeBot.SafeAnswerCallback(callback.ID, "✅ عضویت شما تایید شد! ربات فعال گردید.", true)
	//
	//		// ارسال مجدد منوی اصلی بعد از تأیید
	//		isAdmin := botHandlerTemp.IsSuperAdmin(chatID) || botHandlerTemp.IsAdmin(chatID)
	//		mainMenu, _ := h.KeyboardBuilder.BuildMainMenu(tenant.ID, lang, isAdmin)
	//		msg := tgbotapi.NewMessage(chatID, "🔓 قفل ربات باز شد. لطفاً یکی از گزینه‌های زیر را انتخاب کنید:")
	//		msg.ReplyMarkup = mainMenu
	//		_, _ = h.SafeBot.Bot.Send(msg)
	//	} else {
	//		h.SafeBot.SafeAnswerCallback(callback.ID, "❌ شما هنوز در تمام کانال‌ها عضو نشده‌اید!", true)
	//	}
	//	return
	//}

	//if data == "verify_force_join" {
	//	if botHandlerTemp.CheckForceJoin(tenant, userID) {
	//		h.SafeBot.SafeDeleteMessage(chatID, messageID)
	//		h.SafeBot.SafeAnswerCallback(callback.ID, "✅ عضویت شما تایید شد! ربات فعال گردید.", true)
	//
	//		// ارسال مجدد منوی اصلی بعد از تأیید
	//		isAdmin := botHandlerTemp.IsSuperAdmin(userID) || botHandlerTemp.IsAdmin(userID)
	//		mainMenu, _ := h.KeyboardBuilder.BuildMainMenu(tenant.ID, lang, isAdmin)
	//		msg := tgbotapi.NewMessage(chatID, "🔓 قفل ربات باز شد. لطفاً یکی از گزینه‌های زیر را انتخاب کنید:")
	//		msg.ReplyMarkup = mainMenu
	//		_, _ = h.SafeBot.Bot.Send(msg)
	//	} else {
	//		h.SafeBot.SafeAnswerCallback(callback.ID, "❌ شما هنوز در تمام کانال‌ها عضو نشده‌اید!", true)
	//	}
	//	return
	//}

	// هندل کردن دکمه‌ی اختصاصی بررسی جوین اجباری
	if data == "verify_force_join" {
		if botHandlerTemp.CheckForceJoin(tenant, userID) {
			h.SafeBot.SafeDeleteMessage(chatID, messageID)
			h.SafeBot.SafeAnswerCallback(callback.ID, "✅ عضویت شما تایید شد! ربات فعال گردید.", true)

			// تأیید رفرال معلق کاربر و واریز پاداش به معرف پس از عضویت در کانال‌ها
			if u, err := h.UserRepo.GetUserByTelegramID(tenant.ID, userID); err == nil && u != nil && h.ReferralService != nil {
				_ = h.ReferralService.ApproveReferral(tenant, u.ID)
			}

			// ارسال مجدد منوی اصلی بعد از تأیید
			isAdmin := botHandlerTemp.IsSuperAdmin(userID) || botHandlerTemp.IsAdmin(userID)
			mainMenu, _ := h.KeyboardBuilder.BuildMainMenu(tenant.ID, lang, isAdmin)
			msg := tgbotapi.NewMessage(chatID, "🔓 قفل ربات باز شد. لطفاً یکی از گزینه‌های زیر را انتخاب کنید:")
			msg.ReplyMarkup = mainMenu
			_, _ = h.SafeBot.Bot.Send(msg)
		} else {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ شما هنوز در تمام کانال‌ها عضو نشده‌اید!", true)
		}
		return
	}

	// قفل کردن سایر دکمه‌های شیشه‌ای در صورت عدم عضویت (به جز ادمین)
	//if !botHandlerTemp.CheckForceJoin(tenant, chatID) {
	//	h.SafeBot.SafeAnswerCallback(callback.ID, "🛑 ابتدا در کانال‌های اجباری عضو شوید.", true)
	//	return
	//}

	// قفل کردن سایر دکمه‌های شیشه‌ای در صورت عدم عضویت (بررسی با userID به جای chatID)
	if !botHandlerTemp.CheckForceJoin(tenant, userID) {
		h.SafeBot.SafeAnswerCallback(callback.ID, "🛑 ابتدا در کانال‌های اجباری عضو شوید (پیام ربات در پی‌وی را بررسی کنید).", true)
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
		//	// ۱. کسر از کیف پول و آپدیت دیتابیس
		//	user.TomanBalance -= order.TotalAmount
		//	_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)
		//
		//	// ۲. آپدیت وضعیت سفارش
		//	order.PaymentMethod = "WALLET_TOMAN"
		//	order.PaymentStatus = "PAID"
		//	order.Status = "COMPLETED" // تغییر وضعیت از DRAFT به تکمیل شده
		//	_ = h.OrderRepo.UpdateOrder(order)
		//
		//	// ۳. ثبت خودکار سفارش به عنوان تسک جدید
		//	newTask := &domain.Task{
		//		TenantID:      tenant.ID,
		//		OrderID:       order.ID, // ارتباط با فاکتور اسپانسر
		//		Title:         "📢 عضویت در کانال اسپانسر",
		//		TargetLink:    order.TargetLink,
		//		Capacity:      order.Quantity, // تعداد ممبری که اسپانسر پولش را داده
		//		TaskType:      "auto_join",
		//		RewardType:    "point",
		//		RewardAmount:  10.0, // مقدار پاداش (می‌تواند بعداً از جدول تنظیمات کارفرما خوانده شود)
		//		StayDuration:  24,   // ۲۴ ساعت ماندگاری اجباری در کانال
		//		PenaltyAmount: 50.0, // جریمه لفت دادن
		//		Status:        "ACTIVE",
		//	}
		//	_ = h.TaskRepo.CreateTask(newTask)
		//
		//	successMsg := fmt.Sprintf("✅ پرداخت %.0f تومان از کیف پول انجام شد.\n\n🔗 لینک شما با موفقیت در سیستم ثبت شد و هم‌اکنون در بخش «وظایف» برای کاربران فعال است.", order.TotalAmount)
		//	h.SafeBot.SafeEditText(chatID, messageID, successMsg, nil)
		//
		//	if h.Logger != nil {
		//		h.Logger.Info(fmt.Sprintf("Order %d PAID via Toman & Task Created for User %d", order.ID, user.ID))
		//	}
		//
		//case "point":
		//	if user.PointBalance < order.TotalAmount {
		//		h.SafeBot.SafeAnswerCallback(callback.ID, "❌ موجودی کیف پول امتیازی شما کافی نیست.", true)
		//		return
		//	}
		//
		//	// ۱. کسر از کیف پول و آپدیت دیتابیس
		//	user.PointBalance -= order.TotalAmount
		//	_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)
		//
		//	// ۲. آپدیت وضعیت سفارش
		//	order.PaymentMethod = "WALLET_POINT"
		//	order.PaymentStatus = "PAID"
		//	order.Status = "COMPLETED"
		//	_ = h.OrderRepo.UpdateOrder(order)
		//
		//	// ۳. ثبت خودکار سفارش به عنوان تسک جدید
		//	newTask := &domain.Task{
		//		TenantID:      tenant.ID,
		//		OrderID:       order.ID,
		//		Title:         "📢 عضویت در کانال اسپانسر",
		//		TargetLink:    order.TargetLink,
		//		Capacity:      order.Quantity,
		//		TaskType:      "auto_join",
		//		RewardType:    "point",
		//		RewardAmount:  10.0,
		//		StayDuration:  24,
		//		PenaltyAmount: 50.0,
		//		Status:        "ACTIVE",
		//	}
		//	_ = h.TaskRepo.CreateTask(newTask)
		//
		//	successMsg := fmt.Sprintf("✅ پرداخت %.0f امتیاز از کیف پول انجام شد.\n\n🔗 کانال شما جهت عضوگیری در بخش وظایف ثبت شد.", order.TotalAmount)
		//	h.SafeBot.SafeEditText(chatID, messageID, successMsg, nil)

		case "toman":
			if user.TomanBalance < order.TotalAmount {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ موجودی کیف پول تومانی شما کافی نیست.", true)
				return
			}

			user.TomanBalance -= order.TotalAmount
			_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)

			order.PaymentMethod = "WALLET_TOMAN"
			order.PaymentStatus = "PAID"
			order.Status = "COMPLETED"
			_ = h.OrderRepo.UpdateOrder(order)

			var successMsg string
			if strings.HasPrefix(order.TrackingCode, "SP_") {
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
				successMsg = fmt.Sprintf("✅ پرداخت %.0f تومان انجام شد.\n🔖 کد رهگیری: `%s`\n\n🔗 لینک شما در بخش «وظایف» فعال گردید.", order.TotalAmount, order.TrackingCode)
			} else {
				successMsg = fmt.Sprintf("✅ سفارش شما با موفقیت ثبت و پرداخت شد!\n\n🔖 کد رهگیری: `%s`\n💰 مبلغ: %.0f تومان", order.TrackingCode, order.TotalAmount)
			}
			h.SafeBot.SafeEditText(chatID, messageID, successMsg, nil)

		case "point":
			if user.PointBalance < order.TotalAmount {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ موجودی کیف پول امتیازی شما کافی نیست.", true)
				return
			}

			user.PointBalance -= order.TotalAmount
			_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)

			order.PaymentMethod = "WALLET_POINT"
			order.PaymentStatus = "PAID"
			order.Status = "COMPLETED"
			_ = h.OrderRepo.UpdateOrder(order)

			var successMsg string
			if strings.HasPrefix(order.TrackingCode, "SP_") {
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
				successMsg = fmt.Sprintf("✅ پرداخت %.0f امتیاز انجام شد.\n🔖 کد رهگیری: `%s`\n\n🔗 کانال شما در بخش «وظایف» ثبت شد.", order.TotalAmount, order.TrackingCode)
			} else {
				successMsg = fmt.Sprintf("✅ سفارش شما با موفقیت ثبت و پرداخت شد!\n\n🔖 کد رهگیری: `%s`", order.TrackingCode)
			}
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
	//if strings.HasPrefix(data, "join_game_") {
	//	gameIDStr := strings.TrimPrefix(data, "join_game_")
	//	gameID, _ := strconv.Atoi(gameIDStr)
	//
	//	user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
	//	if user == nil {
	//		h.SafeBot.SafeAnswerCallback(callback.ID, "❌ ابتدا ربات را در پی‌وی استارت بزنید.", true)
	//		return
	//	}
	//
	//	game, err := h.GameRepo.GetGameByID(tenant.ID, uint(gameID))
	//	if err != nil || game.Status != "WAITING" {
	//		h.SafeBot.SafeAnswerCallback(callback.ID, "❌ این بازی لغو شده یا نفر دوم قبلاً عضو شده است.", true)
	//		return
	//	}
	//
	//	if game.CreatorID == user.ID {
	//		h.SafeBot.SafeAnswerCallback(callback.ID, "❌ شما سازنده این میز هستید و نمی‌توانید با خودتان بازی کنید!", true)
	//		return
	//	}
	//
	//	if user.PointBalance < game.BetAmount {
	//		h.SafeBot.SafeAnswerCallback(callback.ID, "❌ موجودی امتیاز شما برای پیوستن به این شرط کافی نیست.", true)
	//		return
	//	}
	//
	//	activeGame, _ := h.GameRepo.GetActiveGameForUser(tenant.ID, user.ID)
	//	if activeGame != nil {
	//		h.SafeBot.SafeAnswerCallback(callback.ID, "❌ شما یک بازی در جریان دارید. ابتدا آن را تمام کنید.", true)
	//		return
	//	}
	//
	//	user.PointBalance -= game.BetAmount
	//	_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)
	//
	//	game.JoinerID = &user.ID
	//	game.Status = "IN_PROGRESS"
	//	_ = h.GameRepo.UpdateGame(game)
	//
	//	h.SafeBot.SafeAnswerCallback(callback.ID, "✅ با موفقیت به بازی پیوستید!", false)
	//
	//	emoji := "🎲"
	//	if game.GameType == "CASINO" {
	//		emoji = "🎰"
	//	}
	//	startMsg := fmt.Sprintf("🔥 ظرفیت بازی تکمیل شد!\n\nهر دو شرکت‌کننده لطفاً ایموجی %s را در گروه ارسال کنید.", emoji)
	//	h.SafeBot.SafeEditText(chatID, messageID, startMsg, nil)
	//	return
	//}

	// کاربر روی دکمه پیوستن به بازی گروهی کلیک کرده است
	if strings.HasPrefix(data, "join_game_") {
		gameIDStr := strings.TrimPrefix(data, "join_game_")
		gameID, _ := strconv.Atoi(gameIDStr)

		// اصلاح شد: استفاده از userID (آیدی کلیک‌کننده) به جای chatID (آیدی گروه)
		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, userID)
		if user == nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ ابتدا ربات را در پی‌وی استارت بزنید.", true)
			return
		}

		game, err := h.GameRepo.GetGameByID(tenant.ID, uint(gameID))
		if err != nil || game == nil || game.Status != "WAITING" {
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

		// کسر امن مبلغ شرط از نفر دوم
		_ = h.UserRepo.AddBalances(tenant.ID, user.ID, 0, -game.BetAmount, 0)

		game.JoinerID = &user.ID
		game.Status = "IN_PROGRESS"
		_ = h.GameRepo.UpdateGame(game)

		h.SafeBot.SafeAnswerCallback(callback.ID, "✅ با موفقیت به بازی پیوستید!", false)

		emoji := "🎲"
		if game.GameType == "CASINO" {
			emoji = "🎰"
		}
		creatorUser, _ := h.UserRepo.GetUserByID(tenant.ID, game.CreatorID)
		creatorName := "نفر اول"
		if creatorUser != nil {
			creatorName = creatorUser.FirstName
		}

		startMsg := fmt.Sprintf("🔥 *ظرفیت بازی تکمیل شد!*\n\n👤 بازیکن اول: %s\n👤 بازیکن دوم: %s\n💰 مبلغ شرط هر نفر: %.0f امتیاز\n\n👇 هر دو شرکت‌کننده لطفاً استیکر/ایموجی %s را در گروه ارسال کنید:",
			creatorName, user.FirstName, game.BetAmount, emoji)
		h.SafeBot.SafeEditText(chatID, messageID, startMsg, nil)
		return
	}

	// سازنده روی دکمه لغو میز بازی کلیک کرده است
	if strings.HasPrefix(data, "cancel_game_") {
		gameIDStr := strings.TrimPrefix(data, "cancel_game_")
		gameID, _ := strconv.Atoi(gameIDStr)

		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, userID)
		if user == nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ کاربر یافت نشد.", true)
			return
		}

		game, err := h.GameRepo.GetGameByID(tenant.ID, uint(gameID))
		if err != nil || game == nil || game.Status != "WAITING" {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ امکان لغو وجود ندارد (بازی شروع شده یا قبلاً لغو شده است).", true)
			return
		}

		if game.CreatorID != user.ID {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ فقط سازنده میز می‌تواند بازی را لغو کند!", true)
			return
		}

		game.Status = "CANCELED"
		_ = h.GameRepo.UpdateGame(game)

		// بازگرداندن مبلغ شرط به حساب سازنده
		_ = h.UserRepo.AddBalances(tenant.ID, user.ID, 0, game.BetAmount, 0)

		h.SafeBot.SafeAnswerCallback(callback.ID, "✅ بازی لغو شد و امتیاز شما به حسابتان برگشت.", true)
		h.SafeBot.SafeEditText(chatID, messageID, "❌ این میز بازی توسط سازنده لغو شد و مبلغ شرط بازگشت داده شد.", nil)
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

		//if task.TaskType == "manual_proof" {
		//	_ = h.UserRepo.UpdateUserTaskID(tenant.ID, user.ID, task.ID)
		//	_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateWaitTask)
		//
		//	h.SafeBot.SafeAnswerCallback(callback.ID, "📸 لطفاً عکس (اسکرین‌شات) انجام این تسک را ارسال کنید.", true)
		//	h.SafeBot.SafeDeleteMessage(chatID, messageID) // پیام دکمه‌ها را پاک می‌کنیم
		//	msg := tgbotapi.NewMessage(chatID, "📸 لطفاً **اسکرین‌شات** اثبات انجام این تسک را دقیقاً در همینجا ارسال کنید:\n\n*(فقط یک عکس ارسال کنید)*")
		//	msg.ParseMode = "Markdown"
		//	_, _ = h.SafeBot.Bot.Send(msg)
		//	return
		//}

		if task.TaskType == "manual_proof" {
			_ = h.UserRepo.UpdateUserTaskID(tenant.ID, user.ID, task.ID)
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateWaitTask)

			h.SafeBot.SafeAnswerCallback(callback.ID, "📸 لطفاً عکس (اسکرین‌شات) انجام این تسک را ارسال کنید.", true)
			h.SafeBot.SafeDeleteMessage(chatID, messageID)
			msg := tgbotapi.NewMessage(chatID, "📸 لطفاً **اسکرین‌شات** اثبات انجام این تسک را دقیقاً در همینجا ارسال کنید:\n\n*(فقط یک عکس ارسال کنید)*")
			msg.ParseMode = "Markdown"
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}

		if task.TaskType == "reaction" {
			// محاسبه تأخیر (۲ ثانیه به ازای هر ری‌اکشن؛ تعداد ری‌اکشن در StayDuration ذخیره می‌شود، پیش‌فرض ۱ ری‌اکشن = ۲ ثانیه)
			reactionCount := task.StayDuration
			if reactionCount <= 0 {
				reactionCount = 1
			}
			delaySeconds := reactionCount * 2
			h.SafeBot.SafeAnswerCallback(callback.ID, fmt.Sprintf("⏳ در حال بررسی ثبت ری‌اکشن (%d ثانیه)...", delaySeconds), false)
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		} else {
			// تسک عضویت در کانال (auto_join)
			targetChatID, err := ResolveChatID(task.TargetLink)
			if err != nil {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ ساختار لینک این تسک نامعتبر است.", true)
				return
			}

			var chatConfig tgbotapi.ChatConfigWithUser
			if strings.HasPrefix(targetChatID, "-100") {
				numericID, _ := strconv.ParseInt(targetChatID, 10, 64)
				chatConfig = tgbotapi.ChatConfigWithUser{
					ChatID: numericID,
					UserID: userID,
				}
			} else {
				chatConfig = tgbotapi.ChatConfigWithUser{
					SuperGroupUsername: targetChatID,
					UserID:             userID,
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
		}

		// ثبت تسک با حروف بزرگ APPROVED و واریز پاداش
		submission := &domain.TaskSubmission{
			TenantID: tenant.ID,
			UserID:   user.ID,
			TaskID:   task.ID,
			Status:   "APPROVED",
		}
		_ = h.TaskRepo.SubmitTask(submission)

		// تبدیل لینک ذخیره شده به فرمت مناسب API
		//targetChatID, err := ResolveChatID(task.TargetLink)
		//if err != nil {
		//	h.SafeBot.SafeAnswerCallback(callback.ID, "❌ ساختار لینک این تسک نامعتبر است.", true)
		//	return
		//}
		//
		//// تشخیص اینکه آیدی عددی است یا یوزرنیم
		//var chatConfig tgbotapi.ChatConfigWithUser
		//if strings.HasPrefix(targetChatID, "-100") {
		//	numericID, _ := strconv.ParseInt(targetChatID, 10, 64)
		//	chatConfig = tgbotapi.ChatConfigWithUser{
		//		ChatID: numericID,
		//		UserID: chatID,
		//	}
		//} else {
		//	chatConfig = tgbotapi.ChatConfigWithUser{
		//		SuperGroupUsername: targetChatID,
		//		UserID:             chatID,
		//	}
		//}
		//
		//member, err := h.SafeBot.Bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		//	ChatConfigWithUser: chatConfig,
		//})
		//
		//if err != nil {
		//	h.SafeBot.SafeAnswerCallback(callback.ID, "❌ خطا در بررسی عضویت. ربات باید در کانال هدف ادمین باشد.", true)
		//	return
		//}
		//
		//if member.Status == "left" || member.Status == "kicked" {
		//	h.SafeBot.SafeAnswerCallback(callback.ID, "⚠️ شما هنوز در کانال عضو نشده‌اید!", true)
		//	return
		//}

		// کاربر عضو شده است -> ثبت تسک و واریز پاداش
		//submission := &domain.TaskSubmission{
		//	TenantID: tenant.ID,
		//	UserID:   user.ID,
		//	TaskID:   task.ID,
		//	Status:   "approved",
		//}
		//_ = h.TaskRepo.SubmitTask(submission)

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

	// [delivery/telegram/callback_handler.go]
	// --- سیستم پردازش منوهای لیگ و مسابقات ---
	if strings.HasPrefix(data, "league_") {
		action := strings.TrimPrefix(data, "league_")
		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if user == nil {
			return
		}

		switch action {
		case "main":
			markup := h.KeyboardBuilder.BuildLeagueMainMenu()
			msgText := "🏆 *بخش مسابقات و برترین‌ها*\n\nبه سیستم رقابتی ربات خوش آمدید! لطفاً یکی از بخش‌های زیر را انتخاب کنید:"
			h.SafeBot.SafeEditText(chatID, messageID, msgText, &markup)

		case "customers":
			rankings, _ := h.OrderRepo.GetCustomerRankingsLast30Days(tenant.ID)

			var msgText strings.Builder
			msgText.WriteString("🏆 *برترین مشتریان 30 روز گذشته*\n\n")

			userRank := 0
			var userSpent float64 = 0

			for i, rank := range rankings {
				if rank.UserID == user.ID {
					userRank = i + 1
					userSpent = rank.TotalSpent
				}

				if i < 10 { // نمایش فقط ۱۰ نفر برتر
					emoji := "🎖"
					switch i {
					case 0:
						emoji = "🥇"
					case 1:
						emoji = "🥈"
					case 2:
						emoji = "🥉"
					case 3:
						emoji = "🎗"
					case 4:
						emoji = "⭐"
					default:
						emoji = fmt.Sprintf("%d️⃣", i+1)
					}
					msgText.WriteString(fmt.Sprintf("%s %s — %.0f تومان\n", emoji, rank.FirstName, rank.TotalSpent))
				}
			}

			msgText.WriteString("\n")
			if userRank > 0 {
				msgText.WriteString(fmt.Sprintf("👤 رتبهٔ شما: #%d — مجموع خرید: %.0f\n\n", userRank, userSpent))
			} else {
				msgText.WriteString("👤 رتبهٔ شما: خارج از رتبه‌بندی — مجموع خرید: 0\n\n")
			}

			msgText.WriteString("💸 *تخفیف هوشمند بر اساس رتبه:*\n🥇 رتبهٔ 1: 5%\n🥈 رتبهٔ 2: 4%\n🥉 رتبهٔ 3: 3%\n🎗 رتبهٔ 4: 2%\n⭐ رتبهٔ 5: 1%")

			markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "league_main")))
			h.SafeBot.SafeEditText(chatID, messageID, msgText.String(), &markup)

		case "referrals":
			rankings, _ := h.RefRepo.GetLeagueRankings(tenant.ID)
			var msgText strings.Builder
			msgText.WriteString("🏆 *جدول رتبه‌بندی لیگ رفرال (Season 1)*\n\n")

			userRank := 0
			var userInvites int64 = 0
			var userPoints int64 = 0

			// تابع محاسبه امتیاز Milestone کارفرما
			calculatePoints := func(invites int64) int64 {
				pts := invites // 1 امتیاز پایه برای هر دعوت
				if invites >= 25 {
					pts += 50
				} else if invites >= 10 {
					pts += 15
				} else if invites >= 5 {
					pts += 5
				}
				return pts
			}

			for i, rank := range rankings {
				pts := calculatePoints(rank.InviteCount)
				if rank.UserID == user.ID {
					userRank = i + 1
					userInvites = rank.InviteCount
					userPoints = pts
				}

				if i < 10 {
					emoji := "🎖"
					switch i {
					case 0:
						emoji = "🥇"
					case 1:
						emoji = "🥈"
					case 2:
						emoji = "🥉"
					}
					msgText.WriteString(fmt.Sprintf("%s %s | 👥 %d | ⭐ %d امتیاز\n", emoji, rank.FirstName, rank.InviteCount, pts))
				}
			}

			msgText.WriteString("\n")
			if userRank > 0 {
				msgText.WriteString(fmt.Sprintf("👤 رتبهٔ شما: #%d\n👥 دعوت موفق: %d\n⭐ امتیاز لیگ: %d\n\n", userRank, userInvites, userPoints))
			} else {
				msgText.WriteString("👤 رتبهٔ شما: خارج از رتبه‌بندی\n\n")
			}
			if tenant.LeagueRules != "" {
				msgText.WriteString(fmt.Sprintf("📜 *قوانین و جوایز لیگ:*\n%s", tenant.LeagueRules))
			} else {
				msgText.WriteString("🎁 *پاداش‌های Milestone:*\nدعوت اول: ۱ امتیاز\n۵ دعوت: ۵+ امتیاز جایزه\n۱۰ دعوت: ۱۵+ امتیاز جایزه\n۲۵ دعوت: ۵۰+ امتیاز جایزه")
			}

			markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "league_main")))
			h.SafeBot.SafeEditText(chatID, messageID, msgText.String(), &markup)

			//msgText.WriteString("🎁 *پاداش‌های Milestone:*\nدعوت اول: ۱ امتیاز\n۵ دعوت: ۵+ امتیاز جایزه\n۱۰ دعوت: ۱۵+ امتیاز جایزه\n۲۵ دعوت: ۵۰+ امتیاز جایزه")
			//
			//markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت", "league_main")))
			//h.SafeBot.SafeEditText(chatID, messageID, msgText.String(), &markup)

		case "draws":
			draws, err := h.DrawRepo.GetActiveDraws(tenant.ID)
			if err != nil || len(draws) == 0 {
				h.SafeBot.SafeAnswerCallback(callback.ID, "🏆 در حال حاضر مسابقه یا قرعه‌کشی فعالی وجود ندارد.", true)
				return
			}

			h.SafeBot.SafeDeleteMessage(chatID, messageID) // پاک کردن منو برای ارسال لیست قرعه‌کشی‌ها

			// ارسال پست جداگانه برای هر قرعه‌کشی
			for _, draw := range draws {
				totalEntries, _ := h.DrawRepo.GetTotalEntries(tenant.ID, draw.ID)
				endDate := draw.EndDate.Format("2006-01-02 15:04")

				msgText := fmt.Sprintf("🎰 *%s*\n\n📝 توضیحات: %s\n\n🎟 هزینه ورود: %d تیکت\n🏆 تعداد برندگان: %d نفر\n🎁 جایزه هر برنده: %.0f %s\n👥 شانس‌های ثبت شده: %d / %d\n\n⏱ پایان: %s",
					draw.Title, draw.Description, draw.TicketCost, draw.WinnersCount, draw.RewardPerWinner, draw.RewardType, totalEntries, draw.MaxParticipants, endDate)

				markup := h.KeyboardBuilder.BuildDrawActionMenu(draw.ID, draw.TicketCost)
				msg := tgbotapi.NewMessage(chatID, msgText)
				msg.ParseMode = "Markdown"
				msg.ReplyMarkup = markup
				_, _ = h.SafeBot.Bot.Send(msg)
			}
		}
		return
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
			if isSuper || isAdmin {
				_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitOrderSearch)
				h.SafeBot.SafeAnswerCallback(callback.ID, "جستجوی سفارش", false)
				msg := tgbotapi.NewMessage(chatID, "🛒 *مدیریت سفارشات*\n\nلطفاً **کد رهگیری (Tracking Code)** سفارش مورد نظر را جهت بررسی و تغییر وضعیت ارسال کنید:")
				msg.ParseMode = "Markdown"
				_, _ = h.SafeBot.Bot.Send(msg)
			}
			return

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

			//if receiptAction == "approve" {
			//	// تأیید فیش و قطعی کردن سفارش
			//	receipt.Status = "APPROVED"
			//	_ = h.ReceiptRepo.UpdateStatus(receipt)
			//
			//	order.PaymentMethod = "CARD"
			//	order.PaymentStatus = "PAID"
			//	order.Status = "COMPLETED"
			//	_ = h.OrderRepo.UpdateOrder(order)
			//
			//	// ثبت خودکار سفارش اسپانسر به عنوان تسک جدید
			//	newTask := &domain.Task{
			//		TenantID:      tenant.ID,
			//		OrderID:       order.ID,
			//		Title:         "📢 عضویت در کانال اسپانسر",
			//		TargetLink:    order.TargetLink,
			//		Capacity:      order.Quantity,
			//		TaskType:      "auto_join",
			//		RewardType:    "point",
			//		RewardAmount:  10.0,
			//		StayDuration:  24,
			//		PenaltyAmount: 50.0,
			//		Status:        "ACTIVE",
			//	}
			//	_ = h.TaskRepo.CreateTask(newTask)
			//
			//	h.SafeBot.SafeAnswerCallback(callback.ID, "✅ فیش تأیید و تسک ساخته شد.", false)
			//	h.SafeBot.SafeEditText(chatID, messageID, "✅ این فیش توسط ادمین تأیید و به تسک تبدیل شد.", nil)
			//	if userUser != nil {
			//		_ = h.SafeBot.SendMessage(userUser.TelegramID, fmt.Sprintf("✅ فیش واریزی شما مبلغ %.0f تومان تأیید شد و سفارش شما فعال گردید.", receipt.Amount))
			//	}

			if receiptAction == "approve" {
				receipt.Status = "APPROVED"
				_ = h.ReceiptRepo.UpdateStatus(receipt)

				order.PaymentMethod = "CARD"
				order.PaymentStatus = "PAID"
				order.Status = "COMPLETED"
				_ = h.OrderRepo.UpdateOrder(order)

				if strings.HasPrefix(order.TrackingCode, "SP_") {
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
				}

				h.SafeBot.SafeAnswerCallback(callback.ID, "✅ فیش تأیید شد.", false)
				h.SafeBot.SafeEditText(chatID, messageID, fmt.Sprintf("✅ این فیش توسط ادمین تأیید شد (سفارش `%s`).", order.TrackingCode), nil)
				if userUser != nil {
					_ = h.SafeBot.SendMessage(userUser.TelegramID, fmt.Sprintf("✅ فیش واریزی شما به مبلغ %.0f تومان تأیید شد و سفارش `%s` فعال گردید.", receipt.Amount, order.TrackingCode))
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
			if isSuper || isAdmin {
				tickets, _ := h.TicketRepo.GetOpenTickets(tenant.ID)
				markup := h.KeyboardBuilder.BuildAdminTicketsListMenu(tickets)
				h.SafeBot.SafeEditText(chatID, messageID, "🎫 *پشتیبانی و تیکت‌ها*\n\nلیست تیکت‌های در انتظار پاسخ در زیر نمایش داده شده است. جهت بررسی روی آن‌ها کلیک کنید:", &markup)
			}
			return

		case action == "draws":
			if isSuper || isAdmin {
				draws, _ := h.DrawRepo.GetActiveDraws(tenant.ID)
				markup := h.KeyboardBuilder.BuildAdminDrawListMenu(draws)
				h.SafeBot.SafeEditText(chatID, messageID, "🎰 *مدیریت قرعه‌کشی‌ها*\n\nلیست مسابقات فعال جهت پایان و قرعه‌کشی:", &markup)
			}
			return

		case action == "broadcast":
			if isSuper {
				_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitBroadcastMsg)
				h.SafeBot.SafeAnswerCallback(callback.ID, "ارسال پیام همگانی...", false)
				msg := tgbotapi.NewMessage(chatID, "📢 *ارسال پیام همگانی*\n\nلطفاً پیام خود را (متن، عکس، ویدیو، یا فوروارد) دقیقاً در همین‌جا ارسال کنید.\nربات با استفاده از سیستم صف‌بندی (Worker) پیام را برای تمام کاربران ارسال خواهد کرد.")
				msg.ParseMode = "Markdown"
				_, _ = h.SafeBot.Bot.Send(msg)
			}
			return

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

		//case action == "settings":
		//	if isSuper {
		//		markup := h.KeyboardBuilder.BuildAdminSettingsMenu()
		//		msgText := fmt.Sprintf("⚙️ *تنظیمات مالی سیستم*\n\nدرصد کارمزد لغو: %.1f%%\nدرصد کارمزد انتقال گروه: %.1f%%\nپاداش بازگشت غیرفعال: %.0f\n\nجهت ویرایش روی گزینه مورد نظر کلیک کنید:",
		//			tenant.CancelFeePercent, tenant.TransferFeePercent, tenant.InactiveBonusAmount)
		//		h.SafeBot.SafeEditText(chatID, messageID, msgText, &markup)
		//	}
		//	return

		case action == "settings":
			if isSuper {
				markup := h.KeyboardBuilder.BuildAdminSettingsMenu()
				msgText := fmt.Sprintf("⚙️ *تنظیمات مالی سیستم*\n\nدرصد کارمزد لغو: %.1f%%\nدرصد کارمزد انتقال گروه: %.1f%%\nپاداش بازگشت غیرفعال: %.0f\nنرخ تبدیل به تومان (۱ امتیاز =): %.0f\nنرخ تبدیل به استارز (؟ امتیاز = ۱ استارز): %.0f\n\nجهت ویرایش روی گزینه مورد نظر کلیک کنید:",
					tenant.CancelFeePercent, tenant.TransferFeePercent, tenant.InactiveBonusAmount, tenant.PointToTomanRate, tenant.PointToStarRate)
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

		case action == "forcejoin":
			if isSuper {
				channels, _ := h.ForceJoinRepo.GetActiveForceJoins(tenant.ID)
				markup := h.KeyboardBuilder.BuildAdminForceJoinListMenu(channels)
				h.SafeBot.SafeEditText(chatID, messageID, "🛑 *مدیریت کانال‌های جوین اجباری*\n\nلیست کانال‌های فعال در سیستم در زیر نمایش داده شده است. برای حذف، روی آن‌ها کلیک کنید:", &markup)
			}
			return

		//case action == "tasks":
		//	if isSuper || isAdmin {
		//		tasks, _ := h.TaskRepo.GetAllTasks(tenant.ID)
		//		markup := h.KeyboardBuilder.BuildAdminTasksListMenu(tasks)
		//		h.SafeBot.SafeEditText(chatID, messageID, "📝 *مدیریت تسک‌ها و اسپانسرها*\n\nبرای تغییر ظرفیت، تعیین جریمه یا متوقف کردن هر تسک روی آن کلیک کنید:", &markup)
		//	}
		//	return

		case action == "tasks":
			if isSuper || isAdmin {
				tasks, _ := h.TaskRepo.GetAllTasks(tenant.ID)
				markup := h.KeyboardBuilder.BuildAdminTasksListMenu(tasks)
				h.SafeBot.SafeEditText(chatID, messageID, "📝 *مدیریت تسک‌ها*\n\nبرای ویرایش هر تسک روی آن کلیک کنید یا تسک جدید بسازید:", &markup)
			}
			return

		case action == "sponsors":
			if isSuper || isAdmin {
				tasks, _ := h.TaskRepo.GetSponsorTasks(tenant.ID)
				markup := h.KeyboardBuilder.BuildAdminTasksListMenu(tasks)
				h.SafeBot.SafeEditText(chatID, messageID, "📢 *مدیریت تسک‌های اسپانسرها*\n\nلیست کانال‌های ثبت شده توسط اسپانسرها:", &markup)
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

	// مدیریت کانال‌های جوین اجباری (ادمین)
	if strings.HasPrefix(data, "admfj_") {
		if !botHandlerTemp.IsSuperAdmin(chatID) {
			return
		}
		action := strings.TrimPrefix(data, "admfj_")

		if action == "add" {
			user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitForceJoinData)
			h.SafeBot.SafeAnswerCallback(callback.ID, "ارسال اطلاعات...", false)
			msg := tgbotapi.NewMessage(chatID, "🛑 لطفاً **آیدی کانال** و **لینک عضویت** را با یک فاصله (Space) ارسال کنید.\n\nمثال کانال خصوصی:\n`-100123456789 https://t.me/joinchat/XXXX`\nمثال کانال عمومی:\n`@mychannel https://t.me/mychannel`")
			msg.ParseMode = "Markdown"
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}

		if strings.HasPrefix(action, "del_") {
			fjID, _ := strconv.Atoi(strings.TrimPrefix(action, "del_"))
			_ = h.ForceJoinRepo.Delete(tenant.ID, uint(fjID))
			h.SafeBot.SafeAnswerCallback(callback.ID, "✅ کانال حذف شد.", false)

			// رفرش منو
			channels, _ := h.ForceJoinRepo.GetActiveForceJoins(tenant.ID)
			markup := h.KeyboardBuilder.BuildAdminForceJoinListMenu(channels)
			h.SafeBot.SafeEditText(chatID, messageID, "🛑 *مدیریت کانال‌های جوین اجباری*\n\n✅ لیست بروزرسانی شد:", &markup)
			return
		}
	}

	// مدیریت تسک‌ها (ادمین)
	//if strings.HasPrefix(data, "admtask_") {
	//	if !botHandlerTemp.IsSuperAdmin(chatID) && !botHandlerTemp.IsAdmin(chatID) {
	//		return
	//	}
	//	parts := strings.Split(data, "_")
	//	if len(parts) < 3 {
	//		return
	//	}
	//	action := parts[1] // edit, cap, pen, toggle
	//	taskID, _ := strconv.Atoi(parts[2])
	//
	//	task, err := h.TaskRepo.GetTaskByID(tenant.ID, uint(taskID))
	//	if err != nil {
	//		return
	//	}
	//	user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)

	// مدیریت تسک‌ها (ادمین)
	if strings.HasPrefix(data, "admtask_") {
		if !botHandlerTemp.IsSuperAdmin(userID) && !botHandlerTemp.IsAdmin(userID) {
			return
		}
		parts := strings.Split(data, "_")
		if len(parts) < 3 {
			return
		}
		action := parts[1] // add, edit, cap, pen, toggle
		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, userID)

		if action == "add" {
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitNewTaskData)
			h.SafeBot.SafeAnswerCallback(callback.ID, "فرم ثبت تسک جدید", false)
			guideMsg := "➕ *افزودن تسک جدید*\n\n" +
				"لطفاً اطلاعات تسک را در یک خط و با جداکننده `|` به فرمت زیر ارسال کنید:\n\n" +
				"`عنوان | لینک هدف | نوع | ظرفیت | پاداش | جریمه`\n\n" +
				"📌 *انواع مجاز:*\n" +
				"• `auto_join` (عضویت کانال با بررسی خودکار)\n" +
				"• `manual_proof` (تسک دستی با ارسال اسکرین‌شات)\n" +
				"• `reaction` (تسک ری‌اکشن با تأخیر هوشمند)\n\n" +
				"💡 *مثال تسک عضویت:*\n" +
				"`عضویت در کانال حامی | @mychannel | auto_join | 500 | 10 | 20`\n\n" +
				"💡 *مثال تسک اسکرین‌شات:*\n" +
				"`ثبت نام در سایت | https://example.com | manual_proof | 100 | 50 | 0`"
			msg := tgbotapi.NewMessage(chatID, guideMsg)
			msg.ParseMode = "Markdown"
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}

		taskID, _ := strconv.Atoi(parts[2])
		task, err := h.TaskRepo.GetTaskByID(tenant.ID, uint(taskID))
		if err != nil {
			return
		}
		switch action {
		case "edit":
			markup := h.KeyboardBuilder.BuildAdminTaskEditMenu(task.ID, task.Status)
			msgText := fmt.Sprintf("📝 *ویرایش تسک*\n\n📌 عنوان: %s\n👥 ظرفیت فعلی: %d\n💰 پاداش عضویت: %.0f %s\n⚖️ جریمه لفت دادن: %.0f\nوضعیت: **%s**\n\nعملیات مورد نظر را انتخاب کنید:",
				task.Title, task.Capacity, task.RewardAmount, task.RewardType, task.PenaltyAmount, task.Status)
			h.SafeBot.SafeEditText(chatID, messageID, msgText, &markup)
		case "cap":
			_ = h.UserRepo.UpdateAdminTaskTarget(tenant.ID, user.ID, task.ID)
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitTaskCapacity)
			h.SafeBot.SafeAnswerCallback(callback.ID, "ارسال ظرفیت", false)
			_, _ = h.SafeBot.Bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("👥 لطفاً **ظرفیت جدید** را برای تسک «%s» به صورت عدد ارسال کنید:", task.Title)))
		case "pen":
			_ = h.UserRepo.UpdateAdminTaskTarget(tenant.ID, user.ID, task.ID)
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitTaskPenalty)
			h.SafeBot.SafeAnswerCallback(callback.ID, "ارسال جریمه", false)
			_, _ = h.SafeBot.Bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("⚖️ لطفاً **مبلغ جریمه (لفت دادن)** را برای تسک «%s» به صورت عدد ارسال کنید:", task.Title)))
		case "toggle":
			if task.Status == "ACTIVE" {
				task.Status = "PAUSED"
			} else {
				task.Status = "ACTIVE"
			}
			_ = h.TaskRepo.UpdateTask(task)
			h.SafeBot.SafeAnswerCallback(callback.ID, "وضعیت تغییر کرد.", false)
			// رفرش منو
			markup := h.KeyboardBuilder.BuildAdminTaskEditMenu(task.ID, task.Status)
			msgText := fmt.Sprintf("📝 *ویرایش تسک*\n\n📌 عنوان: %s\n👥 ظرفیت فعلی: %d\n💰 پاداش عضویت: %.0f %s\n⚖️ جریمه لفت دادن: %.0f\nوضعیت: **%s**\n\nعملیات مورد نظر را انتخاب کنید:",
				task.Title, task.Capacity, task.RewardAmount, task.RewardType, task.PenaltyAmount, task.Status)
			h.SafeBot.SafeEditText(chatID, messageID, msgText, &markup)
		}
		return
	}

	// مدیریت عملیات‌های پشتیبانی و تیکت (ادمین)
	if strings.HasPrefix(data, "admtick_") {
		if !botHandlerTemp.IsSuperAdmin(chatID) && !botHandlerTemp.IsAdmin(chatID) {
			return
		}
		parts := strings.Split(data, "_")
		if len(parts) < 3 {
			return
		}
		action := parts[1] // view, reply, close
		ticketID, _ := strconv.Atoi(parts[2])

		ticket, err := h.TicketRepo.GetTicketByID(tenant.ID, uint(ticketID))
		if err != nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ تیکت یافت نشد.", true)
			return
		}
		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)

		switch action {
		case "view":
			msgs, _ := h.TicketRepo.GetTicketMessages(ticket.ID)
			userText := "متنی ثبت نشده است."
			if len(msgs) > 0 {
				userText = msgs[0].Text // گرفتن اولین پیام کاربر
			}

			order, _ := h.OrderRepo.GetOrderByID(tenant.ID, ticket.OrderID)
			msgText := fmt.Sprintf("🎫 *بررسی تیکت*\n\nشماره سفارش: `%s`\nوضعیت سفارش: %s\n\n📝 *متن مشکل کاربر:*\n`%s`\n\nجهت پاسخگویی از منوی زیر استفاده کنید:", order.TrackingCode, order.Status, userText)
			markup := h.KeyboardBuilder.BuildAdminTicketActionMenu(ticket.ID)
			h.SafeBot.SafeEditText(chatID, messageID, msgText, &markup)

		case "reply":
			_ = h.UserRepo.UpdateAdminTicketTarget(tenant.ID, user.ID, ticket.ID)
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateAdminWaitTicketReply)
			h.SafeBot.SafeAnswerCallback(callback.ID, "لطفاً پاسخ را تایپ کنید.", false)
			_, _ = h.SafeBot.Bot.Send(tgbotapi.NewMessage(chatID, "💬 لطفاً پاسخ خود را به این کاربر تایپ کرده و ارسال کنید:"))

		case "close":
			ticket.Status = "CLOSED"
			_ = h.TicketRepo.UpdateStatus(ticket)
			h.SafeBot.SafeAnswerCallback(callback.ID, "تیکت بسته شد.", false)

			// بازگشت خودکار به لیست تیکت‌ها
			tickets, _ := h.TicketRepo.GetOpenTickets(tenant.ID)
			markup := h.KeyboardBuilder.BuildAdminTicketsListMenu(tickets)
			h.SafeBot.SafeEditText(chatID, messageID, "✅ تیکت بسته شد.\n\n🎫 *تیکت‌های در انتظار پاسخ:*", &markup)
		}
		return
	}
	if strings.HasPrefix(data, "admleague_") {
		if !botHandlerTemp.IsSuperAdmin(userID) && !botHandlerTemp.IsAdmin(userID) {
			return
		}
		adminUser, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, userID)
		if data == "admleague_rules" {
			_ = h.UserRepo.UpdateUserState(tenant.ID, adminUser.ID, domain.StateAdminWaitLeagueRules)
			h.SafeBot.SafeAnswerCallback(callback.ID, "ویرایش قوانین لیگ", false)
			_, _ = h.SafeBot.Bot.Send(tgbotapi.NewMessage(chatID, "📝 لطفاً **متن جدید قوانین و جوایز لیگ** را ارسال کنید:"))
			return
		}
		if data == "admleague_reset" {
			_ = h.RefRepo.ResetLeagueReferrals(tenant.ID)
			h.SafeBot.SafeAnswerCallback(callback.ID, "✅ فصل جدید لیگ آغاز شد و جدول امتیازات ریست گردید.", true)
			return
		}
	}

	// مدیریت قرعه‌کشی‌ها (ادمین)
	if strings.HasPrefix(data, "admdraw_") {
		if !botHandlerTemp.IsSuperAdmin(userID) && !botHandlerTemp.IsAdmin(userID) {
			return
		}
		parts := strings.Split(data, "_")
		if len(parts) < 3 {
			return
		}
		action := parts[1] // add, view, roll

		if action == "add" {
			adminUser, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, userID)
			_ = h.UserRepo.UpdateUserState(tenant.ID, adminUser.ID, domain.StateAdminWaitNewDrawData)
			h.SafeBot.SafeAnswerCallback(callback.ID, "ایجاد قرعه‌کشی جدید", false)
			guide := "🎰 *ایجاد قرعه‌کشی جدید*\n\n" +
				"لطفاً اطلاعات قرعه‌کشی را با جداکننده `|` به فرمت زیر ارسال کنید:\n\n" +
				"`عنوان | توضیحات | هزینه تیکت | ظرفیت کل | تعداد برنده | مبلغ جایزه | نوع جایزه (STARS/POINT/TOMAN) | مهلت به روز`\n\n" +
				"💡 *مثال:*\n" +
				"`قرعه‌کشی بزرگ پاییزه | ویژه اعضا | 1 | 500 | 3 | 50 | STARS | 7`"
			msg := tgbotapi.NewMessage(chatID, guide)
			msg.ParseMode = "Markdown"
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}

		drawID, _ := strconv.Atoi(parts[2])
		draw, err := h.DrawRepo.GetDrawByID(tenant.ID, uint(drawID))
		if err != nil || draw.Status != "ACTIVE" {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ مسابقه یافت نشد یا قبلاً تمام شده است.", true)
			return
		}
		//if strings.HasPrefix(data, "admdraw_") {
		//	if !botHandlerTemp.IsSuperAdmin(chatID) && !botHandlerTemp.IsAdmin(chatID) {
		//		return
		//	}
		//	parts := strings.Split(data, "_")
		//	if len(parts) < 3 {
		//		return
		//	}
		//	action := parts[1] // view, roll
		//	drawID, _ := strconv.Atoi(parts[2])
		//
		//	draw, err := h.DrawRepo.GetDrawByID(tenant.ID, uint(drawID))
		//	if err != nil || draw.Status != "ACTIVE" {
		//		h.SafeBot.SafeAnswerCallback(callback.ID, "❌ مسابقه یافت نشد یا قبلاً تمام شده است.", true)
		//		return
		//	}

		if action == "view" {
			totalEntries, _ := h.DrawRepo.GetTotalEntries(tenant.ID, draw.ID)
			msgText := fmt.Sprintf("🎰 *جزئیات مسابقه*\n\n📌 عنوان: %s\n👥 ظرفیت فعلی: %d / %d\n🎁 جایزه هر برنده: %.0f %s\n🏆 تعداد برنده‌ها: %d نفر\n\nبرای انتخاب سیستمی برندگان روی دکمه زیر کلیک کنید:",
				draw.Title, totalEntries, draw.MaxParticipants, draw.RewardPerWinner, draw.RewardType, draw.WinnersCount)

			markup := h.KeyboardBuilder.BuildAdminDrawActionMenu(draw.ID)
			h.SafeBot.SafeEditText(chatID, messageID, msgText, &markup)
			return
		}

		if action == "roll" {
			entries, _ := h.DrawRepo.GetDrawEntries(tenant.ID, draw.ID)
			totalEntries, _ := h.DrawRepo.GetTotalEntries(tenant.ID, draw.ID)

			if len(entries) == 0 {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ هیچکس در این قرعه‌کشی شرکت نکرده است!", true)
				return
			}

			// ساخت استخر شانس‌ها (هر کاربر به تعداد Entries در استخر تکرار می‌شود تا شانسش بیشتر باشد)
			var pool []uint
			for _, entry := range entries {
				for i := 0; i < entry.Entries; i++ {
					pool = append(pool, entry.UserID)
				}
			}

			// بر هم زدن استخر با متد تصادفی
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

			// استخراج برندگان (بدون تکراری شدن یک شخص)
			var winners []uint
			wonMap := make(map[uint]bool)

			for _, uid := range pool {
				if !wonMap[uid] {
					winners = append(winners, uid)
					wonMap[uid] = true
				}
				if len(winners) == draw.WinnersCount {
					break
				}
			}

			// واریز پاداش به کیف پول برنده‌ها و ساخت پیام نتایج
			var resultMsg strings.Builder
			resultMsg.WriteString(fmt.Sprintf("🎰 *نتیجه مسابقه: %s*\n\n━━━━━━━━━━━━\n\n👥 تعداد شرکت‌کنندگان: %d نفر\n\n🎁 *برندگان:*\n\n", draw.Title, totalEntries))

			for i, winnerID := range winners {
				wUser, _ := h.UserRepo.GetUserByID(tenant.ID, winnerID)
				if wUser == nil {
					continue
				}

				// واریز بر اساس نوع جایزه
				var addToman, addPoint, addStars float64
				switch draw.RewardType {
				case "STARS":
					addStars = draw.RewardPerWinner
				case "POINT":
					addPoint = draw.RewardPerWinner
				case "TOMAN":
					addToman = draw.RewardPerWinner
				}
				_ = h.UserRepo.AddBalances(tenant.ID, wUser.ID, addToman, addPoint, addStars)

				// اطلاع رسانی پی‌وی به برنده
				_ = h.SafeBot.SendMessage(wUser.TelegramID, fmt.Sprintf("🎉 *تبریک!*\n\nشما در قرعه‌کشی «%s» برنده شدید و مبلغ %.0f %s به کیف پول شما واریز شد.", draw.Title, draw.RewardPerWinner, draw.RewardType))

				// اضافه کردن به لیست
				emoji := "🏅"
				if i == 0 {
					emoji = "🥇"
				} else if i == 1 {
					emoji = "🥈"
				} else if i == 2 {
					emoji = "🥉"
				}

				mention := fmt.Sprintf("[%s](tg://user?id=%d)", wUser.FirstName, wUser.TelegramID)
				resultMsg.WriteString(fmt.Sprintf("%s %s — %.0f %s\n", emoji, mention, draw.RewardPerWinner, draw.RewardType))
			}

			resultMsg.WriteString("\n🎉 تبریک به برندگان!\n━━━━━━━━━━━━")

			// تغییر وضعیت قرعه‌کشی
			draw.Status = "COMPLETED"
			_ = h.DrawRepo.UpdateDraw(draw)

			// ارسال پیام به کانال نتایج (اگر ثبت نشده بود به کانال گزارش یا پی‌وی ادمین)
			targetChannel := tenant.WinnerChannelID
			if targetChannel == "" {
				targetChannel = tenant.ReportChannelID
			}

			if targetChannel != "" {
				chID, _ := strconv.ParseInt(targetChannel, 10, 64)
				msg := tgbotapi.NewMessage(chID, resultMsg.String())
				msg.ParseMode = "Markdown"
				_, _ = h.SafeBot.Bot.Send(msg)
			}

			// گزارش پایان به ادمین
			h.SafeBot.SafeAnswerCallback(callback.ID, "✅ قرعه‌کشی انجام شد و پاداش‌ها واریز گردید.", false)
			h.SafeBot.SafeEditText(chatID, messageID, "✅ قرعه‌کشی با موفقیت به پایان رسید و نتایج در کانال منتشر شد.\n\n"+resultMsg.String(), nil)
			return
		}
	}

	// مدیریت دکمه‌های تبدیل موجودی (کاربر)
	if strings.HasPrefix(data, "convert_to_") {
		user, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if user == nil {
			return
		}

		targetType := strings.TrimPrefix(data, "convert_to_")

		if targetType == "toman" {
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateWaitConvertAmountToman)
			h.SafeBot.SafeAnswerCallback(callback.ID, "ارسال مقدار امتیاز...", false)

			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("🔄 لطفاً مشخص کنید **چه مقدار امتیاز** می‌خواهید به تومان تبدیل کنید؟ (ارسال فقط عدد)\n\n💵 نرخ تبدیل: هر ۱ امتیاز = %.0f تومان", tenant.PointToTomanRate))
			msg.ParseMode = "Markdown"
			_, _ = h.SafeBot.Bot.Send(msg)
		} else if targetType == "stars" {
			_ = h.UserRepo.UpdateUserState(tenant.ID, user.ID, domain.StateWaitConvertAmountStars)
			h.SafeBot.SafeAnswerCallback(callback.ID, "ارسال تعداد استارز...", false)

			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⭐ لطفاً مشخص کنید **چه تعداد استارز** می‌خواهید دریافت کنید؟ (ارسال فقط عدد)\n\n⭐ نرخ تبدیل: برای دریافت هر ۱ استارز، %.0f امتیاز کسر می‌شود.", tenant.PointToStarRate))
			msg.ParseMode = "Markdown"
			_, _ = h.SafeBot.Bot.Send(msg)
		}
		return
	}

	// مدیریت سفارشات توسط ادمین (تکمیل و لغو)
	if strings.HasPrefix(data, "admord_") {
		if !botHandlerTemp.IsSuperAdmin(chatID) && !botHandlerTemp.IsAdmin(chatID) {
			return
		}
		parts := strings.Split(data, "_")
		if len(parts) < 3 {
			return
		}
		action := parts[1] // complete, cancel
		orderID, _ := strconv.Atoi(parts[2])

		order, err := h.OrderRepo.GetOrderByID(tenant.ID, uint(orderID))
		if err != nil {
			h.SafeBot.SafeAnswerCallback(callback.ID, "❌ سفارش یافت نشد.", true)
			return
		}

		orderUser, _ := h.UserRepo.GetUserByID(tenant.ID, order.UserID)

		if action == "complete" {
			order.Status = "COMPLETED"
			_ = h.OrderRepo.UpdateOrder(order)
			h.SafeBot.SafeAnswerCallback(callback.ID, "✅ سفارش تکمیل شد.", false)
			h.SafeBot.SafeEditText(chatID, messageID, fmt.Sprintf("✅ سفارش `%s` با موفقیت **تکمیل** شد.", order.TrackingCode), nil)

			if orderUser != nil {
				_ = h.SafeBot.SendMessage(orderUser.TelegramID, fmt.Sprintf("✅ سفارش شما با کد رهگیری `%s` توسط ادمین انجام و تکمیل شد.", order.TrackingCode))
			}
		} else if action == "cancel" {
			if order.Status == "CANCELED" {
				h.SafeBot.SafeAnswerCallback(callback.ID, "❌ قبلاً لغو شده است.", true)
				return
			}
			order.Status = "CANCELED"
			_ = h.OrderRepo.UpdateOrder(order)

			// عودت وجه بدون کارمزد (چون لغو از سمت ادمین بوده نه کاربر)
			if orderUser != nil {
				if order.PaymentMethod == "WALLET_TOMAN" || order.PaymentMethod == "CARD" {
					orderUser.TomanBalance += order.TotalAmount
				} else if order.PaymentMethod == "WALLET_POINT" {
					orderUser.PointBalance += order.TotalAmount
				} else if order.PaymentMethod == "STARS" {
					orderUser.StarsBalance += order.TotalAmount
				}
				_ = h.UserRepo.UpdateBalances(tenant.ID, orderUser.ID, orderUser.TomanBalance, orderUser.PointBalance, orderUser.StarsBalance)
				_ = h.SafeBot.SendMessage(orderUser.TelegramID, fmt.Sprintf("❌ سفارش شما با کد `%s` توسط پشتیبانی لغو شد و مبلغ به کیف پول شما بازگشت داده شد.", order.TrackingCode))
			}

			h.SafeBot.SafeAnswerCallback(callback.ID, "✅ سفارش لغو و مبلغ عودت داده شد.", false)
			h.SafeBot.SafeEditText(chatID, messageID, fmt.Sprintf("❌ سفارش `%s` لغو شد و مبلغ به کیف پول کاربر برگشت داده شد.", order.TrackingCode), nil)
		}
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
