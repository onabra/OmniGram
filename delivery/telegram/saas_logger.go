package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"OmniGram/domain"
)

// SendSaaSErrorLog ارسال امن خطاها به کانال مانیتورینگ با تفکیک پروژه
func (h *WebhookHandler) SendSaaSErrorLog(tenant *domain.Tenant, section string, errText string) {
	if h.LogChannelID == 0 || h.MasterBotToken == "" {
		return // اگر تنظیمات لاگ وارد نشده بود، عملیات متوقف می‌شود
	}

	// استخراج اطلاعات کارفرما برای تشخیص اینکه کدام پروژه مشکل خورده است
	projectInfo := "🔴 پروژه نامشخص (خطای هسته/روتینگ)"
	if tenant != nil {
		// نمایش آیدی کارفرما و بخشی از توکن برای امنیت
		maskedToken := "مخفی"
		if len(tenant.BotToken) > 15 {
			maskedToken = tenant.BotToken[:10] + "..."
		}
		projectInfo = fmt.Sprintf("🏢 Tenant ID: %d\n🤖 ربات: `%s`", tenant.ID, maskedToken)
	}

	// فرمت‌بندی متن لاگ
	logMsg := fmt.Sprintf("⚠️ *هشدار سیستم SaaS*\n\n%s\n\n📍 *بخش درگیر:* `%s`\n❌ *شرح خطا:*\n`%s`",
		projectInfo, section, errText)

	// استفاده از API مستقیم تلگرام برای اطمینان از ارسال پیام (بدون نیاز به پکیج جانبی)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", h.MasterBotToken)
	payload := map[string]interface{}{
		"chat_id":    h.LogChannelID,
		"text":       logMsg,
		"parse_mode": "Markdown",
	}

	jsonPayload, _ := json.Marshal(payload)
	_, _ = http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
}
