package telegram

import (
	"OmniGram/core"
	"fmt"
	"log"
	"runtime/debug"

	"OmniGram/repository"
	"OmniGram/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type PollingHandler struct {
	TenantRepo      *repository.TenantRepository
	TranslationRepo *repository.TranslationRepo
	KeyboardBuilder *KeyboardBuilder
	LogChannelID    int64
	MasterBotToken  string
	UserRepo        repository.UserRepository
	RefRepo         *repository.ReferralRepository
	TxRepo          *repository.TransactionRepository
	WalletService   *services.WalletService
	SafeBot         *SafeBot
	Logger          *core.CentralLogger
}

func (h *PollingHandler) StartPolling() {
	bot, err := tgbotapi.NewBotAPI(h.MasterBotToken)
	if err != nil {
		log.Fatal("توکن نامعتبر است:", err)
	}

	h.SafeBot = &SafeBot{
		Bot:          bot,
		LogChannelID: h.LogChannelID,
	}
	if h.Logger != nil {
		h.Logger.AttachReporter(h.SafeBot)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Printf("🤖 OmniGram Bot is running locally on account @%s via Long Polling...", bot.Self.UserName)

	for update := range updates {
		// هر آپدیت در یک Goroutine جداگانه پردازش می‌شود تا در صورت کرش، کل ربات متوقف نشود
		go h.processUpdateWithRecovery(update)
	}
}

// processUpdateWithRecovery این متد نقش میدلور محافظ را دارد
func (h *PollingHandler) processUpdateWithRecovery(update tgbotapi.Update) {
	// 🛡 سیستم شکارچی اتوماتیک خطاها (Panic Recovery)
	defer func() {
		if r := recover(); r != nil {
			// دریافت مسیر دقیق کدی که باعث خطا شده است
			stackTrace := string(debug.Stack())

			// فرمت‌بندی خطای مهلک برای ارسال به کانال
			errorMsg := fmt.Sprintf("🚨 *خطای بحرانی (Crash Prevented):*\n\n`نوع خطا: %v`\n\n🔍 *ردیابی:*\n`%s`", r, stackTrace[:1000]) // محدودیت 1000 کاراکتر برای تلگرام

			log.Printf("Recovered from panic: %v", r)
			h.SafeBot.SendLogToChannel(errorMsg)
		}
	}()

	// --- منطق پردازش روتینگ قبلی شما ---
	tenant, err := h.TenantRepo.GetByToken(h.MasterBotToken)
	if err != nil || !tenant.IsActive {
		return // رد کردن آپدیت
	}

	lang := "fa"

	if update.CallbackQuery != nil {
		// h.HandleCallback(tenant, update, lang) // در صورت وجود متد آن را از کامنت خارج کنید
		return
	}

	if update.ChatMember != nil {
		// h.HandleChatMemberUpdate(tenant, update.ChatMember) // در صورت وجود متد آن را از کامنت خارج کنید
		return
	}

	if update.Message != nil && (update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup()) {
		// h.HandleGroupMessage(tenant, update.Message) // در صورت وجود متد آن را از کامنت خارج کنید
		return
	}

	if update.Message != nil {
		chatID := update.Message.Chat.ID

		botHandler := &BotHandler{
			TenantRepo:      h.TenantRepo,
			TranslationRepo: h.TranslationRepo,
			KeyboardBuilder: h.KeyboardBuilder,
			SafeBot:         h.SafeBot,
			UserRepo:        h.UserRepo,
			ReferralRepo:    h.RefRepo,
			TxRepo:          h.TxRepo,
			WalletService:   h.WalletService,
		}

		if update.Message.Command() == "start" {
			botHandler.HandleStart(tenant, chatID, lang, update.Message.Text)
		} else if update.Message.Text != "" {
			botHandler.HandleMessage(tenant, chatID, update.Message.Text, lang)
		}
	}
}
