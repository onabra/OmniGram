package telegram

import (
	"OmniGram/domain"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleCallback هندل کردن منوهای شیشه‌ای و جایگزینی صفحات
func (h *WebhookHandler) HandleCallback(tenant *domain.Tenant, update tgbotapi.Update, lang string) {
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
}

// buildMainCategoryMenu نمایش دسته‌های اصلی (شبیه‌ساز دیتابیس)
func (h *WebhookHandler) buildMainCategoryMenu() (*tgbotapi.InlineKeyboardMarkup, string) {
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
func (h *WebhookHandler) buildSubCategoryMenu(parentID uint) (*tgbotapi.InlineKeyboardMarkup, string) {
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
