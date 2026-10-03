package core

import (
	"fmt"
	"log"
)

// TelegramReporter اینترفیسی برای جلوگیری از وابستگی مستقیم لایه Core به لایه Telegram
type TelegramReporter interface {
	SendLogToChannel(msg string)
}

type CentralLogger struct {
	reporter TelegramReporter
}

// NewCentralLogger ایجاد نمونه اولیه لاگر (مستقل از تلگرام در زمان بوت)
func NewCentralLogger() *CentralLogger {
	return &CentralLogger{}
}

// AttachReporter اتصال کلاینت تلگرام به لاگر پس از راه‌اندازی ربات
func (l *CentralLogger) AttachReporter(r TelegramReporter) {
	l.reporter = r
}

// Error ثبت خطاهای عادی (err != nil) و ارسال خودکار به کانال لاگ
func (l *CentralLogger) Error(context string, err error) {
	if err == nil {
		return
	}
	// ثبت در کنسول سرور
	log.Printf("❌ [ERROR] %s: %v", context, err)

	// ارسال به کانال در صورت اتصال ربات
	if l.reporter != nil {
		msg := fmt.Sprintf("⚠️ *System_Alert (Non-Fatal):*\n\nSECTION: `%s`\nBUG_TEXT: `%v`", context, err)
		// استفاده از Goroutine تا ارسال پیام به کانال، فرآیند اصلی را کند نکند
		go l.reporter.SendLogToChannel(msg)
	}
}

// Info لاگ‌های اطلاعاتی (فقط نمایش در کنسول)
func (l *CentralLogger) Info(msg string) {
	log.Printf("✅ [INFO] %s", msg)
}
