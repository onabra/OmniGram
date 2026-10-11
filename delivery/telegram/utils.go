package telegram

import (
	"errors"
	"strconv"
	"strings"
)

// ResolveChatID تبدیل و اعتبارسنجی لینک تلگرام برای استفاده در API
func ResolveChatID(input string) (string, error) {
	input = strings.TrimSpace(input)

	// ۱. فرمت آیدی عددی کانال/گروه (مثل -100123456789)
	if strings.HasPrefix(input, "-100") {
		_, err := strconv.ParseInt(input, 10, 64)
		if err == nil {
			return input, nil
		}
	}

	// ۲. یوزرنیم استاندارد با @
	if strings.HasPrefix(input, "@") {
		return input, nil
	}

	// ۳. استخراج یوزرنیم از لینک‌های عمومی تلگرام
	if strings.Contains(input, "t.me/") && !strings.Contains(input, "t.me/+") && !strings.Contains(input, "joinchat") {
		parts := strings.Split(input, "t.me/")
		if len(parts) > 1 {
			username := strings.Split(parts[1], "/")[0]
			username = strings.Split(username, "?")[0] // حذف پارامترهای احتمالی
			return "@" + username, nil
		}
	}

	// ۴. لینک‌های خصوصی (پرایوت) قابل پردازش برای بررسی ممبر نیستند
	if strings.Contains(input, "t.me/+") || strings.Contains(input, "joinchat") {
		return "", errors.New("PRIVATE_LINK")
	}

	return "", errors.New("INVALID_FORMAT")
}
