package telegram

import (
	"OmniGram/domain"
	"fmt"
	"strconv"
	"strings"

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

	// کاربر روی دسته اصلی (مثل خدمات) کلیک کرده
	if data == "menu_services" {
		markup, text := h.buildMainCategoryMenu()
		h.SafeBot.SafeEditText(chatID, messageID, text, markup)
		return
	}

	// کاربر روی یکی از دسته‌ها (مثل ممبر یا ویو) کلیک کرده
	if strings.HasPrefix(data, "cat_") {
		catIDStr := strings.TrimPrefix(data, "cat_")
		catID, _ := strconv.Atoi(catIDStr)

		markup, text := h.buildSubCategoryMenu(uint(catID))
		h.SafeBot.SafeEditText(chatID, messageID, text, markup)
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
		}
		return
	}

}

// buildMainCategoryMenu نمایش دسته‌های اصلی (شبیه‌ساز دیتابیس)
func (h *PollingHandler) buildMainCategoryMenu() (*tgbotapi.InlineKeyboardMarkup, string) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("〔ممبر〕", "cat_1"),
			tgbotapi.NewInlineKeyboardButtonData("〔ویو〕", "cat_2"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("〔ری‌اکشن〕", "cat_3"),
			tgbotapi.NewInlineKeyboardButtonData("〔بوست〕", "cat_4"),
		),
	)
	return &keyboard, "لطفاً بخش مورد نظر را انتخاب کنید:"
}

// buildSubCategoryMenu نمایش زیرمجموعه‌های یک دسته همراه با دکمه بازگشت
func (h *PollingHandler) buildSubCategoryMenu(parentID uint) (*tgbotapi.InlineKeyboardMarkup, string) {
	var keyboard tgbotapi.InlineKeyboardMarkup

	if parentID == 1 { // فرض کنیم 1 آیدی ممبر است
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("‐ ممبر ایرانی", "item_1")),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("‐ ممبر خارجی", "item_2")),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت به منوی قبل", "menu_services")),
		)
	} else {
		// برای سایر دسته‌ها
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 بازگشت به منوی قبل", "menu_services")),
		)
	}

	return &keyboard, "نوع سرویس را انتخاب کنید:"
}
