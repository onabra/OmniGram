package telegram

import (
	"OmniGram/domain"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleGroupMessage پردازش پیام‌ها و بازی‌ها در گروه‌ها
func (h *PollingHandler) HandleGroupMessage(tenant *domain.Tenant, message *tgbotapi.Message) {

	// تشخیص ارسال پیام متنی (مثل کلمه 'بازی 2')
	if message.Text != "" {
		if strings.HasPrefix(message.Text, "بازی") {
			// ساخت میز بازی جدید در دیتابیس و نمایش دکمه شیشه‌ای برای نفر دوم
			msg := tgbotapi.NewMessage(message.Chat.ID, "یک بازی جدید ایجاد شد. نفر دوم شرکت کند.")
			h.SafeBot.Bot.Send(msg)
		}
		return
	}

	// تشخیص ارسال ایموجی تاس یا کازینو
	if message.Dice != nil {
		// emoji := message.Dice.Emoji // "🎲" یا "🎰"
		value := message.Dice.Value // نتیجه تصادفی که تلگرام تولید کرده

		// اینجا کاربر فرستنده (message.From.ID) را در دیتابیس پیدا می‌کنید
		// چک می‌کنید اگر در بازی بود، مقدار Value را برایش ثبت می‌کنید

		// پیغام نتیجه
		resultMsg := fmt.Sprintf("کاربر عزیز، عدد شانس شما: %d", value)
		msg := tgbotapi.NewMessage(message.Chat.ID, resultMsg)
		msg.ReplyToMessageID = message.MessageID
		h.SafeBot.Bot.Send(msg)
	}
}
