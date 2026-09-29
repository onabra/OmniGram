package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"OmniGram/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WebhookHandler ساختار مدیریت وب‌هوک‌ها با وابستگی‌های لازم برای معماری SaaS
type WebhookHandler struct {
	TenantRepo      *repository.TenantRepository
	TranslationRepo *repository.TranslationRepo
	KeyboardBuilder *KeyboardBuilder
	LogChannelID    int64  // آیدی کانال لاگ که از .env خوانده شده
	MasterBotToken  string // توکن ربات ارشد برای ارسال خطاها

	// این ۳ مورد جدید برای هندل کردن دکمه‌های شیشه‌ای، جریمه و بازی‌ها اضافه شد
	UserRepo repository.UserRepository
	RefRepo  *repository.ReferralRepository
	SafeBot  *SafeBot
}

// HandleUpdate نقطه ورود وب‌هوک‌های تلگرام و روتینگ درخواست‌ها به رباتِ مربوطه بر اساس Bot_Token.
func (h *WebhookHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	// ۱. استخراج توکن بات از URL
	pathParts := strings.Split(r.URL.Path, "/")
	botToken := pathParts[len(pathParts)-1]

	// ۲. پیدا کردن Tenant بر اساس توکن
	tenant, err := h.TenantRepo.GetByToken(botToken)
	if err != nil || !tenant.IsActive {
		h.SendSaaSErrorLog(nil, "Webhook Routing", "درخواست وب‌هوک برای توکن نامعتبر یا کارفرمای منقضی شده: "+botToken)
		http.Error(w, "Tenant not found or inactive", http.StatusUnauthorized)
		return
	}

	// ۳. خواندن اطلاعات آپدیت تلگرام
	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		h.SendSaaSErrorLog(tenant, "JSON Decode", "خطا در پردازش دیتای تلگرام: "+err.Error())
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	bot, botErr := tgbotapi.NewBotAPI(botToken)
	if botErr != nil {
		h.SendSaaSErrorLog(tenant, "Bot API", "توکن نامعتبر است: "+botErr.Error())
		w.WriteHeader(http.StatusOK)
		return
	}

	// --- ترفند کلیدی SaaS ---
	// ساخت یک کپی از هندلر که SafeBot مخصوص همین یک کارفرما را دارد
	// این کار از قاطی شدن پیام‌های ربات‌ها جلوگیری می‌کند.
	reqHandler := *h
	reqHandler.SafeBot = &telegram.SafeBot{
		Bot:          bot,
		LogChannelID: h.LogChannelID,
	}

	lang := "fa"

	// روتینگ ۱: پردازش کلیک روی دکمه‌های شیشه‌ای
	if update.CallbackQuery != nil {
		reqHandler.HandleCallback(tenant, update, lang) // اصلاح شد: h تبدیل به reqHandler شد
		w.WriteHeader(http.StatusOK)
		return
	}

	// روتینگ ۲: پردازش عضویت یا لفت دادن از کانال (برای جریمه لفت دادن)
	if update.ChatMember != nil {
		reqHandler.HandleChatMemberUpdate(tenant, update.ChatMember) // اصلاح شد: h تبدیل به reqHandler شد
		w.WriteHeader(http.StatusOK)
		return
	}

	// روتینگ ۳: پردازش پیام‌های متنی
	if update.Message != nil {
		chatID := update.Message.Chat.ID

		// اگر پیام داخل گروه است (برای سیستم بازی تاس و کازینو)
		if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
			reqHandler.HandleGroupMessage(tenant, update.Message) // اصلاح شد: h تبدیل به reqHandler شد
			w.WriteHeader(http.StatusOK)
			return
		}

		// پردازش پیام‌های داخل پی‌وی
		welcomeMsg, err := reqHandler.TranslationRepo.GetText(tenant.ID, lang, "msg_welcome")
		if err != nil {
			reqHandler.SendSaaSErrorLog(tenant, "Translation DB", fmt.Sprintf("متن 'msg_welcome' یافت نشد: %v", err))
			welcomeMsg = "خوش آمدید!" // فال‌بک ضروری
		}

		// کدهای اضافه و تکراریِ مربوط به NewBotAPI از اینجا حذف شد

		msg := tgbotapi.NewMessage(chatID, welcomeMsg)
		_, sendErr := reqHandler.SafeBot.Bot.Send(msg) // اصلاح شد: استفاده از رباتِ مخصوصِ همین کارفرما
		if sendErr != nil {
			reqHandler.SendSaaSErrorLog(tenant, "Send Message", fmt.Sprintf("خطا در ارسال پیام به کاربر %d: %v", chatID, sendErr))
		}
	}

	w.WriteHeader(http.StatusOK)
}

//func (h *WebhookHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
//	// ۱. استخراج توکن بات از URL
//	pathParts := strings.Split(r.URL.Path, "/")
//	botToken := pathParts[len(pathParts)-1]
//
//	// ۲. پیدا کردن Tenant بر اساس توکن
//	tenant, err := h.TenantRepo.GetByToken(botToken)
//	if err != nil || !tenant.IsActive {
//		h.SendSaaSErrorLog(nil, "Webhook Routing", "درخواست وب‌هوک برای توکن نامعتبر یا کارفرمای منقضی شده: "+botToken)
//		http.Error(w, "Tenant not found or inactive", http.StatusUnauthorized)
//		return
//	}
//
//	// ۳. خواندن اطلاعات آپدیت تلگرام
//	var update tgbotapi.Update
//	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
//		h.SendSaaSErrorLog(tenant, "JSON Decode", "خطا در پردازش دیتای تلگرام: "+err.Error())
//		http.Error(w, "Bad Request", http.StatusBadRequest)
//		return
//	}
//
//	// استخراج ChatID
//	var chatID int64
//	if update.Message != nil {
//		chatID = update.Message.Chat.ID
//	} else if update.CallbackQuery != nil {
//		chatID = update.CallbackQuery.Message.Chat.ID
//	}
//
//	if chatID == 0 {
//		w.WriteHeader(http.StatusOK)
//		return
//	}
//
//	// ۴. سیستم مدیریت متون بدون هاردکد از دیتابیس/کش[cite: 4].
//	lang := "fa"
//	welcomeMsg, err := h.TranslationRepo.GetText(tenant.ID, lang, "msg_welcome")
//	if err != nil {
//		h.SendSaaSErrorLog(tenant, "Translation DB", fmt.Sprintf("متن 'msg_welcome' یافت نشد: %v", err))
//		welcomeMsg = "خوش آمدید!" // فال‌بک ضروری
//	}
//
//	// ۵. ارسال پیام به کاربر
//	bot, botErr := tgbotapi.NewBotAPI(botToken)
//	if botErr == nil {
//		msg := tgbotapi.NewMessage(chatID, welcomeMsg)
//		_, sendErr := bot.Send(msg)
//		if sendErr != nil {
//			h.SendSaaSErrorLog(tenant, "Send Message", fmt.Sprintf("خطا در ارسال پیام به کاربر %d: %v", chatID, sendErr))
//		}
//	}
//
//	w.WriteHeader(http.StatusOK)
//}
