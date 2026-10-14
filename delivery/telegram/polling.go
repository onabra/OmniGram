package telegram

import (
	"OmniGram/core"
	"OmniGram/domain"
	"fmt"
	"log"
	"runtime/debug"
	"time"

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
	OrderRepo       *repository.OrderRepository
	ServiceRepo     *repository.ServiceRepository
	TaskRepo        *repository.TaskRepository
	GameRepo        *repository.GameRepository
	TicketRepo      *repository.TicketRepository
	ReceiptRepo     *repository.ReceiptRepository
	DrawRepo        *repository.DrawRepository
	CategoryRepo    *repository.CategoryRepository
	ForceJoinRepo   *repository.ForceJoinRepository
	WithdrawalRepo  *repository.WithdrawalRepository

	ReferralService *services.ReferralService

	AdminRepo *repository.AdminRepository

	SuperAdminIDs []int64
	AdminIDs      []int64
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
	h.ReferralService = &services.ReferralService{
		UserRepo:  h.UserRepo,
		RefRepo:   h.RefRepo,
		BotClient: h.SafeBot,
	}
	if h.Logger != nil {
		h.Logger.AttachReporter(h.SafeBot)
	}

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "شروع ربات و نمایش منوی اصلی"},
		{Command: "myid", Description: "دریافت آیدی عددی شما"},
	}
	_, _ = bot.Request(tgbotapi.NewSetMyCommands(commands...))

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

	// --- منطق پردازش روتینگ ---
	tenant, err := h.TenantRepo.GetByToken(h.MasterBotToken)
	if err != nil || !tenant.IsActive {
		return // رد کردن آپدیت
	}

	lang := "fa"

	if update.CallbackQuery != nil {
		h.HandleCallback(tenant, update, lang) // 👈 ارور نوع داده در اینجا فیکس شد
		return
	}

	if update.ChatMember != nil {
		h.HandleChatMemberUpdate(tenant, update.ChatMember) // 👈 از کامنت خارج شد
		return
	}

	if update.Message != nil && (update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup()) {
		h.HandleGroupMessage(tenant, update.Message) // 👈 از کامنت خارج شد
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
			OrderRepo:       h.OrderRepo,
			ServiceRepo:     h.ServiceRepo,
			TaskRepo:        h.TaskRepo,
			SuperAdminIDs:   h.SuperAdminIDs,
			AdminIDs:        h.AdminIDs,
			TicketRepo:      h.TicketRepo,
			GameRepo:        h.GameRepo,
			ReceiptRepo:     h.ReceiptRepo,
			ForceJoinRepo:   h.ForceJoinRepo,
			WithdrawalRepo:  h.WithdrawalRepo,
			AdminRepo:       h.AdminRepo,
			DrawRepo:        h.DrawRepo,     // تزریق شده از بلوک‌های تکراری
			CategoryRepo:    h.CategoryRepo, // تزریق شده از بلوک‌های تکراری
			ReferralService: h.ReferralService,
		}

		adminUser, _ := botHandler.UserRepo.GetUserByTelegramID(tenant.ID, chatID)
		if adminUser != nil && adminUser.State == domain.StateAdminWaitBroadcastMsg {
			if update.Message.Text == "بازگشت به منوی اصلی" {
				_ = botHandler.UserRepo.UpdateUserState(tenant.ID, adminUser.ID, domain.StateNormal)
				botHandler.HandleMessage(tenant, chatID, update.Message.Text, lang)
				return
			}

			// خروج از وضعیت
			_ = botHandler.UserRepo.UpdateUserState(tenant.ID, adminUser.ID, domain.StateNormal)

			// اجرای Goroutine مستقل برای ارسال همگانی
			go botHandler.RunBroadcast(tenant.ID, chatID, update.Message)
			return
		}

		//if update.Message.Command() == "start" {
		//	botHandler.HandleStart(tenant, chatID, lang, update.Message.Text)
		//} else if update.Message.Photo != nil {
		//	botHandler.HandlePhoto(tenant, update.Message, lang)
		//} else if update.Message.Text != "" {
		//	botHandler.HandleMessage(tenant, chatID, update.Message.Text, lang)
		//}

		if update.Message.Command() == "start" {
			botHandler.HandleStart(tenant, chatID, lang, update.Message.Text)
		} else if update.Message.Command() == "myid" {
			// 👈 دریافت آیدی عددی با دستور /myid
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("🆔 آیدی عددی تلگرام شما:\n\n`%d`", update.Message.From.ID))
			msg.ParseMode = "Markdown"
			msg.ReplyToMessageID = update.Message.MessageID
			_, _ = h.SafeBot.Bot.Send(msg)
		} else if update.Message.Photo != nil {
			botHandler.HandlePhoto(tenant, update.Message, lang)
		} else if update.Message.Text != "" {
			botHandler.HandleMessage(tenant, chatID, update.Message.Text, lang)
		}
	}
}

// RunBroadcast صف‌بندی ایمن و اجرای ارسال همگانی در یک Goroutine کاملاً مستقل
func (h *BotHandler) RunBroadcast(tenantID uint, adminChatID int64, msgToBroadcast *tgbotapi.Message) {
	userIDs, err := h.UserRepo.GetAllUserTelegramIDs(tenantID)
	if err != nil || len(userIDs) == 0 {
		_ = h.SafeBot.SendMessage(adminChatID, "❌ کاربری در دیتابیس برای ارسال پیام یافت نشد.")
		return
	}

	_ = h.SafeBot.SendMessage(adminChatID, fmt.Sprintf("⏳ عملیات پیام همگانی برای %d کاربر آغاز شد...\n\nاین فرآیند به صورت پس‌زمینه (Background) و با رعایت محدودیت‌های تلگرام انجام می‌شود تا ربات بلاک نشود. لطفاً صبور باشید.", len(userIDs)))

	success := 0
	failed := 0

	for _, uid := range userIDs {
		var req tgbotapi.Chattable

		// اگر پیام فوروارد شده بود، با همون استایل فوروارد می‌فرستیم
		if msgToBroadcast.ForwardFrom != nil || msgToBroadcast.ForwardFromChat != nil {
			req = tgbotapi.NewForward(uid, msgToBroadcast.Chat.ID, msgToBroadcast.MessageID)
		} else {
			// اگر محتوای تولیدی خودش بود (متن، عکس، ویدیو، دکمه و...)، از CopyMessage استفاده می‌کنیم تا دقیقاً مشابهش بدون تگ فوروارد ارسال شود
			req = tgbotapi.NewCopyMessage(uid, msgToBroadcast.Chat.ID, msgToBroadcast.MessageID)
		}

		_, err := h.SafeBot.Bot.Request(req)
		if err != nil {
			failed++
		} else {
			success++
		}

		// مهم‌ترین بخش: جلوگیری از ارور Flood Wait (محدودیت تلگرام 30 پیام در ثانیه است. با 40 میلی‌ثانیه روی 25 تنظیم می‌کنیم)
		time.Sleep(40 * time.Millisecond)
	}

	// گزارش پایان کار به ادمین
	resultMsg := fmt.Sprintf("✅ *پایان پیام همگانی*\n\n👥 کل مخاطبین: %d\n🟢 ارسال موفق: %d\n🔴 ناموفق (بلاک کرده‌ها): %d", len(userIDs), success, failed)
	_ = h.SafeBot.SendMessage(adminChatID, resultMsg)
}
