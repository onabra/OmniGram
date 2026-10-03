package telegram

import (
	"log"

	"OmniGram/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// PollingHandler جایگزین کامل WebhookHandler برای سیستم‌های Single-Tenant
type PollingHandler struct {
	TenantRepo      *repository.TenantRepository
	TranslationRepo *repository.TranslationRepo
	KeyboardBuilder *KeyboardBuilder
	LogChannelID    int64
	MasterBotToken  string // توکنی که از .env خوانده می‌شود
	UserRepo        repository.UserRepository
	RefRepo         *repository.ReferralRepository
	SafeBot         *SafeBot
}

// StartPolling نقطه ورود جدید برنامه برای دریافت مداوم آپدیت‌ها از تلگرام
func (h *PollingHandler) StartPolling() {
	// ۱. اتصال مستقیم ربات
	bot, err := tgbotapi.NewBotAPI(h.MasterBotToken)
	if err != nil {
		log.Fatal("توکن نامعتبر است:", err)
	}

	h.SafeBot = &SafeBot{
		Bot:          bot,
		LogChannelID: h.LogChannelID,
	}

	// ۲. تنظیمات Long Polling (دقیقاً مشابه TelPulse)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Printf("🤖 OmniGram Bot is running locally on account @%s via Long Polling...", bot.Self.UserName)

	// ۳. حلقه بی‌نهایت پردازش پیام‌ها
	for update := range updates {

		// چون سیستم شما بر اساس Tenant دیتابیس طراحی شده، همان توکن env را به دیتابیس می‌دهیم
		tenant, err := h.TenantRepo.GetByToken(h.MasterBotToken)
		if err != nil || !tenant.IsActive {
			log.Println("کارفرمایی با این توکن در دیتابیس یافت نشد یا غیرفعال است.")
			continue // رد کردن آپدیت
		}

		lang := "fa" // زبان پیش‌فرض

		// --- روتینگ ۱: کلیک روی دکمه‌های شیشه‌ای ---
		if update.CallbackQuery != nil {
			h.HandleCallback(tenant, update, lang)
			continue
		}

		// --- روتینگ ۲: لفت دادن کاربر از کانال ---
		if update.ChatMember != nil {
			h.HandleChatMemberUpdate(tenant, update.ChatMember)
			continue
		}

		// --- روتینگ ۳: پیام‌های داخل گروه (بازی تاس و غیره) ---
		if update.Message != nil && (update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup()) {
			h.HandleGroupMessage(tenant, update.Message)
			continue
		}

		// --- روتینگ ۴: پیام‌های متنی پی‌وی (منوها و دستورات) ---
		if update.Message != nil {
			chatID := update.Message.Chat.ID

			botHandler := &BotHandler{
				TenantRepo:      h.TenantRepo,
				TranslationRepo: h.TranslationRepo,
				KeyboardBuilder: h.KeyboardBuilder,
				SafeBot:         h.SafeBot,
			}

			if update.Message.Command() == "start" {

				botHandler.HandleStart(tenant, chatID, lang, update.Message.Text)
			} else if update.Message.Text != "" {
				botHandler.HandleMessage(tenant, chatID, update.Message.Text, lang)
			}
		}
	}
}
