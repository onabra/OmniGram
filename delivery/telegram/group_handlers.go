package telegram

import (
	"OmniGram/domain"
	"OmniGram/repository"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleGroupMessage پردازش پیام‌ها و بازی‌ها در گروه‌ها
func (h *PollingHandler) HandleGroupMessage(tenant *domain.Tenant, message *tgbotapi.Message) {
	if message.Text != "" {
		text := strings.TrimSpace(message.Text)

		// ۱. نمایش موجودی در گروه
		//if text == "موجودی" {
		//	user, err := h.UserRepo.GetUserByTelegramID(tenant.ID, message.From.ID)
		//	if err != nil || user == nil {
		//		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ ابتدا ربات را در پی‌وی استارت بزنید.")
		//		msg.ReplyToMessageID = message.MessageID
		//		_, _ = h.SafeBot.Bot.Send(msg)
		//		return
		//	}
		//
		//	msgText := fmt.Sprintf("👤 کاربر: %s\n💰 کیف پول (تومان): %.0f\n🎁 امتیاز رفرال: %.0f\n⭐ استارز: %.0f",
		//		message.From.FirstName, user.TomanBalance, user.PointBalance, user.StarsBalance)
		//	msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
		//	msg.ReplyToMessageID = message.MessageID
		//	_, _ = h.SafeBot.Bot.Send(msg)
		//	return
		//}

		// ۱. نمایش موجودی در گروه
		//if text == "موجودی" {
		//	user, err := h.UserRepo.GetUserByTelegramID(tenant.ID, message.From.ID)
		//	if err != nil || user == nil {
		//		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ ابتدا ربات را در پی‌وی استارت بزنید.")
		//		msg.ReplyToMessageID = message.MessageID
		//		_, _ = h.SafeBot.Bot.Send(msg)
		//		return
		//	}
		//
		//	// فرمت جدید درخواستی کارفرما با استفاده از تگ‌های HTML برای باکس نقل‌قول (Blockquote)
		//	msgText := fmt.Sprintf("👤 <b>موجودی شما</b>\n\n<blockquote>⭐ امتیاز: %.0f\n🌟 معادل استارز: %.0f\n💰 تومان: %.0f</blockquote>",
		//		user.PointBalance, user.StarsBalance, user.TomanBalance)
		//
		//	msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
		//	msg.ParseMode = "HTML"                   // برای پشتیبانی از تگ blockquote
		//	msg.ReplyToMessageID = message.MessageID // ریپ زدن روی پیام کاربر
		//
		//	// ساخت دکمه شیشه‌ای انتقال
		//	transferMarkup := tgbotapi.NewInlineKeyboardMarkup(
		//		tgbotapi.NewInlineKeyboardRow(
		//			tgbotapi.NewInlineKeyboardButtonData("💸 انتقال موجودی", "group_transfer_help"),
		//		),
		//	)
		//	msg.ReplyMarkup = transferMarkup
		//
		//	_, _ = h.SafeBot.Bot.Send(msg)
		//	return
		//}
		// ۱. نمایش موجودی در گروه
		if text == "موجودی" {
			user, err := h.UserRepo.GetUserByTelegramID(tenant.ID, message.From.ID)
			if err != nil || user == nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ ابتدا ربات را در پی‌وی استارت بزنید.")
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
				return
			}

			// فرمت جدید درخواستی کارفرما با استفاده از تگ‌های HTML برای باکس نقل‌قول (Blockquote)
			msgText := fmt.Sprintf("👤 <b>موجودی شما</b>\n\n<blockquote>⭐ امتیاز: %.0f\n🌟 معادل استارز: %.0f\n💰 تومان: %.0f</blockquote>",
				user.PointBalance, user.StarsBalance, user.TomanBalance)

			msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
			msg.ParseMode = "HTML"                   // برای پشتیبانی از تگ blockquote
			msg.ReplyToMessageID = message.MessageID // ریپ زدن روی پیام کاربر

			// ساخت دکمه شیشه‌ای انتقال
			transferMarkup := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("💸 انتقال موجودی", "group_transfer_help"),
				),
			)
			msg.ReplyMarkup = transferMarkup

			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}

		// ۲. انتقال امتیاز در گروه با ریپلای
		if strings.HasPrefix(text, "انتقال") {
			if message.ReplyToMessage == nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ برای انتقال، روی پیام شخص مورد نظر ریپلای کرده و بفرستید: `انتقال [مبلغ]`")
				msg.ParseMode = "Markdown"
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
				return
			}

			parts := strings.Split(text, " ")
			if len(parts) != 2 {
				return
			}

			amount, err := strconv.ParseFloat(parts[1], 64)
			if err != nil || amount <= 0 {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ مبلغ نامعتبر است.")
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
				return
			}

			sender, _ := h.UserRepo.GetUserByTelegramID(tenant.ID, message.From.ID)
			if sender == nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ ابتدا ربات را در پی‌وی استارت بزنید.")
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
				return
			}

			destTelegramID := message.ReplyToMessage.From.ID
			feePercent := tenant.TransferFeePercent
			if feePercent <= 0 {
				feePercent = 5.0
			}

			err = h.WalletService.TransferBalance(tenant.ID, sender.ID, destTelegramID, amount, "POINT", feePercent)
			if err != nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ خطا: "+err.Error())
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
			} else {
				successMsg := fmt.Sprintf("✅ مبلغ %.0f امتیاز با موفقیت به کاربر مورد نظر انتقال یافت.", amount)
				msg := tgbotapi.NewMessage(message.Chat.ID, successMsg)
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
			}
			return
		}

		// لغو بازی در انتظار (در صورتی که پیام دکمه شیشه‌ای پاک شده باشد)
		if text == "لغو بازی" {
			user, err := h.UserRepo.GetUserByTelegramID(tenant.ID, message.From.ID)
			if err != nil || user == nil {
				return
			}
			activeGame, _ := h.GameRepo.GetActiveGameForUser(tenant.ID, user.ID)
			if activeGame == nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ شما هیچ بازی فعالی ندارید.")
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
				return
			}
			if activeGame.Status != "WAITING" || activeGame.CreatorID != user.ID {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ فقط بازی‌هایی که هنوز نفر دوم به آن‌ها نپیوسته قابل لغو هستند.")
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
				return
			}

			activeGame.Status = "CANCELED"
			_ = h.GameRepo.UpdateGame(activeGame)
			_ = h.UserRepo.AddBalances(tenant.ID, user.ID, 0, activeGame.BetAmount, 0)

			msg := tgbotapi.NewMessage(message.Chat.ID, "✅ میز بازی شما لغو شد و امتیاز شرط به حسابتان برگشت.")
			msg.ReplyToMessageID = message.MessageID
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}

		// ۳. شروع بازی گروهی (تاس/کازینو)
		if strings.HasPrefix(text, "تاس ") || strings.HasPrefix(text, "کازینو ") {
			parts := strings.Split(text, " ")
			if len(parts) != 2 {
				return
			}

			amount, err := strconv.ParseFloat(parts[1], 64)
			if err != nil || amount <= 0 {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ مبلغ شرط نامعتبر است.")
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
				return
			}

			gameType := "DICE"
			if strings.HasPrefix(text, "کازینو") {
				gameType = "CASINO"
			}

			user, err := h.UserRepo.GetUserByTelegramID(tenant.ID, message.From.ID)
			if err != nil || user == nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ ابتدا ربات را در پی‌وی استارت بزنید.")
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
				return
			}

			if user.PointBalance < amount {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ موجودی امتیاز شما برای این شرط کافی نیست.")
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
				return
			}

			activeGame, _ := h.GameRepo.GetActiveGameForUser(tenant.ID, user.ID)
			if activeGame != nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ شما یک بازی در جریان دارید. ابتدا آن را تمام کنید.")
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
				return
			}

			user.PointBalance -= amount
			_ = h.UserRepo.UpdateBalances(tenant.ID, user.ID, user.TomanBalance, user.PointBalance, user.StarsBalance)

			newGame := &domain.GameSession{
				TenantID:  tenant.ID,
				CreatorID: user.ID,
				BetAmount: amount,
				GameType:  gameType,
				Status:    "WAITING",
			}
			_ = h.GameRepo.CreateGame(newGame)

			msgText := fmt.Sprintf("🎲 یک بازی %s جدید ایجاد شد!\n👤 سازنده: %s\n💰 مبلغ: %.0f امتیاز\n\nنفر دوم روی دکمه زیر کلیک کند:", gameType, message.From.FirstName, amount)
			joinMenu := h.KeyboardBuilder.BuildGameJoinMenu(newGame.ID, amount)

			msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
			msg.ReplyMarkup = joinMenu
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}

		// ۵. نمایش آمار کامل کاربر در بازی‌های گروهی
		if text == "اعتبارم" {
			user, err := h.UserRepo.GetUserByTelegramID(tenant.ID, message.From.ID)
			if err != nil || user == nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ ابتدا ربات را در پی‌وی استارت بزنید.")
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
				return
			}

			feePercent := tenant.TransferFeePercent
			if feePercent <= 0 {
				feePercent = 5.0
			}

			// واکشی آمار از دیتابیس
			stats, _ := h.GameRepo.GetUserGameStats(tenant.ID, user.ID, feePercent)
			if stats == nil {
				stats = &repository.GameStats{} // برای جلوگیری از کرش اگر آماری نبود
			}

			msgText := fmt.Sprintf("📊 *آمار و اعتبار شما*\n\n"+
				"🎲 بازی‌های برگزار شده در امروز: %d\n"+
				"🎮 بازی‌های برگزار شده: %d\n"+
				"🏆 مجموع بردها: %d\n"+
				"💔 مجموع باخت‌ها: %d\n"+
				"📈 سود امتیازی: %.0f\n"+
				"📉 ضرر امتیازی: %.0f\n\n"+
				"🎁 موجودی امتیازی: %.0f\n"+
				"💳 موجودی تومانی: %.0f",
				stats.TodayGames, stats.TotalGames, stats.TotalWins, stats.TotalLosses,
				stats.Profit, stats.Loss, user.PointBalance, user.TomanBalance)

			msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
			msg.ParseMode = "Markdown"
			msg.ReplyToMessageID = message.MessageID
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}
	}

	// ۴. هندل کردن ارسال استیکر تاس/کازینو توسط کاربر
	//if message.Dice != nil {
	//	emoji := message.Dice.Emoji
	//	value := message.Dice.Value

	// ۴. هندل کردن ارسال استیکر تاس/کازینو توسط کاربر
	if message.Dice != nil {
		// جلوگیری از تقلب: نادیده گرفتن تاس‌های فوروارد شده
		if message.ForwardDate != 0 || message.ForwardFrom != nil || message.ForwardFromChat != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ فوروارد کردن تاس مجاز نیست! لطفاً مستقیماً تاس بفرستید.")
			msg.ReplyToMessageID = message.MessageID
			_, _ = h.SafeBot.Bot.Send(msg)
			return
		}

		emoji := message.Dice.Emoji
		value := message.Dice.Value

		user, err := h.UserRepo.GetUserByTelegramID(tenant.ID, message.From.ID)
		if err != nil || user == nil {
			return
		}

		activeGame, err := h.GameRepo.GetActiveGameForUser(tenant.ID, user.ID)
		if err != nil || activeGame == nil || activeGame.Status != "IN_PROGRESS" {
			return
		}

		if activeGame.GameType == "DICE" && emoji != "🎲" {
			return
		}
		if activeGame.GameType == "CASINO" && emoji != "🎰" {
			return
		}

		if activeGame.CreatorID == user.ID {
			if activeGame.CreatorRoll != nil {
				return
			}
			activeGame.CreatorRoll = &value
		} else if activeGame.JoinerID != nil && *activeGame.JoinerID == user.ID {
			if activeGame.JoinerRoll != nil {
				return
			}
			activeGame.JoinerRoll = &value
		}

		_ = h.GameRepo.UpdateGame(activeGame)

		if activeGame.CreatorRoll != nil && activeGame.JoinerRoll != nil {
			activeGame.Status = "FINISHED"
			_ = h.GameRepo.UpdateGame(activeGame)

			creator, _ := h.UserRepo.GetUserByID(tenant.ID, activeGame.CreatorID)
			joiner, _ := h.UserRepo.GetUserByID(tenant.ID, *activeGame.JoinerID)

			cRoll := *activeGame.CreatorRoll
			jRoll := *activeGame.JoinerRoll

			totalPot := activeGame.BetAmount * 2

			feePercent := tenant.TransferFeePercent
			if feePercent <= 0 {
				feePercent = 5.0
			}
			feeAmount := totalPot * (feePercent / 100)
			winAmount := totalPot - feeAmount

			var resultMsg string
			if cRoll > jRoll {
				creator.PointBalance += winAmount
				_ = h.UserRepo.UpdateBalances(tenant.ID, creator.ID, creator.TomanBalance, creator.PointBalance, creator.StarsBalance)
				resultMsg = fmt.Sprintf("🎉 پایان بازی!\n\n%s: %d\n%s: %d\n\n🏆 برنده: %s\n💰 جایزه: %.0f امتیاز", creator.FirstName, cRoll, joiner.FirstName, jRoll, creator.FirstName, winAmount)
			} else if jRoll > cRoll {
				joiner.PointBalance += winAmount
				_ = h.UserRepo.UpdateBalances(tenant.ID, joiner.ID, joiner.TomanBalance, joiner.PointBalance, joiner.StarsBalance)
				resultMsg = fmt.Sprintf("🎉 پایان بازی!\n\n%s: %d\n%s: %d\n\n🏆 برنده: %s\n💰 جایزه: %.0f امتیاز", creator.FirstName, cRoll, joiner.FirstName, jRoll, joiner.FirstName, winAmount)
			} else {
				creator.PointBalance += activeGame.BetAmount
				joiner.PointBalance += activeGame.BetAmount
				_ = h.UserRepo.UpdateBalances(tenant.ID, creator.ID, creator.TomanBalance, creator.PointBalance, creator.StarsBalance)
				_ = h.UserRepo.UpdateBalances(tenant.ID, joiner.ID, joiner.TomanBalance, joiner.PointBalance, joiner.StarsBalance)
				resultMsg = fmt.Sprintf("⚖️ بازی مساوی شد! (%d - %d)\n\nمبلغ شرط به حساب هر دو نفر برگشت داده شد.", cRoll, jRoll)
			}

			go func() {
				time.Sleep(3 * time.Second)
				msg := tgbotapi.NewMessage(message.Chat.ID, resultMsg)
				msg.ReplyToMessageID = message.MessageID
				_, _ = h.SafeBot.Bot.Send(msg)
			}()
		}
	}
}
