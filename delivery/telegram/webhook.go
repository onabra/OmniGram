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
}

// HandleUpdate نقطه ورود وب‌هوک‌های تلگرام و روتینگ درخواست‌ها به رباتِ مربوطه بر اساس Bot_Token[cite: 4].
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

	// استخراج ChatID
	var chatID int64
	if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
	}

	if chatID == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	// ۴. سیستم مدیریت متون بدون هاردکد از دیتابیس/کش[cite: 4].
	lang := "fa"
	welcomeMsg, err := h.TranslationRepo.GetText(tenant.ID, lang, "msg_welcome")
	if err != nil {
		h.SendSaaSErrorLog(tenant, "Translation DB", fmt.Sprintf("متن 'msg_welcome' یافت نشد: %v", err))
		welcomeMsg = "خوش آمدید!" // فال‌بک ضروری
	}

	// ۵. ارسال پیام به کاربر
	bot, botErr := tgbotapi.NewBotAPI(botToken)
	if botErr == nil {
		msg := tgbotapi.NewMessage(chatID, welcomeMsg)
		_, sendErr := bot.Send(msg)
		if sendErr != nil {
			h.SendSaaSErrorLog(tenant, "Send Message", fmt.Sprintf("خطا در ارسال پیام به کاربر %d: %v", chatID, sendErr))
		}
	}

	w.WriteHeader(http.StatusOK)
}
