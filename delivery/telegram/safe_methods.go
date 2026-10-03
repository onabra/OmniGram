package telegram

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SafeBot مدیریت امن درخواست‌های تلگرام و لاگ‌گیری در کانال
type SafeBot struct {
	Bot          *tgbotapi.BotAPI
	LogChannelID int64 // آیدی عددی کانال لاگ (مثلاً -100123456789)
}

// sendLogToChannel ارسال پیام خطای فرمت‌بندی شده به کانال لاگ
func (s *SafeBot) sendLogToChannel(errMsg string) {
	if s.LogChannelID != 0 {
		msg := tgbotapi.NewMessage(s.LogChannelID, "⚠️ *خطای سیستمی ربات:*\n\n`"+errMsg+"`")
		msg.ParseMode = "Markdown"
		_, _ = s.Bot.Send(msg)
	}
}

// SafeEditText ویرایش امن پیام برای جلوگیری از کرش هنگام کلیک‌های تکراری
func (s *SafeBot) SafeEditText(chatID int64, messageID int, text string, markup *tgbotapi.InlineKeyboardMarkup) bool {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = "Markdown"
	if markup != nil {
		msg.ReplyMarkup = markup
	}

	_, err := s.Bot.Send(msg)
	if err != nil {
		errText := strings.ToLower(err.Error())
		// اگر دیتای جدید دقیقاً مثل دیتای قبلی بود، با موفقیت از آن عبور کن
		if strings.Contains(errText, "message is not modified") {
			return true
		}

		log.Printf("Failed to edit message: %v", err)
		s.sendLogToChannel(fmt.Sprintf("SafeEditText Error [Chat: %d]: %v", chatID, err))
		return false
	}
	return true
}

// SafeAnswerCallback پاسخ امن به کال‌بک‌های شیشه‌ای
func (s *SafeBot) SafeAnswerCallback(queryID string, text string, showAlert bool) bool {
	cfg := tgbotapi.NewCallback(queryID, text)
	cfg.ShowAlert = showAlert

	_, err := s.Bot.Request(cfg)
	if err != nil {
		errText := strings.ToLower(err.Error())
		if strings.Contains(errText, "query is too old") || strings.Contains(errText, "query id is invalid") {
			log.Printf("Stale callback ignored: %v", err)
			return false
		}

		log.Printf("Failed to answer callback: %v", err)
		s.sendLogToChannel(fmt.Sprintf("SafeAnswerCallback Error: %v", err))
		return false
	}
	return true
}

// SafeDeleteMessage پاک کردن امن پیام با نادیده گرفتن خطاهای رایج
func (s *SafeBot) SafeDeleteMessage(chatID int64, messageID int) bool {
	msg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := s.Bot.Request(msg)
	if err != nil {
		log.Printf("Failed to delete message (safely ignored): %v", err)
		return false
	}
	return true
}

// SendMessage ارسال پیام متنی ساده به کاربر (مورد نیاز برای اطلاع‌رسانی رفرال)
func (s *SafeBot) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := s.Bot.Send(msg)
	if err != nil {
		s.sendLogToChannel(fmt.Sprintf("SendMessage Error [Chat: %d]: %v", chatID, err))
	}
	return err
}
