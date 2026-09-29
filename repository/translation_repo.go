package repository

import (
	"OmniGram/domain"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type TranslationRepo struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// GetText دریافت متن از کش ردیس، در صورت نبودن دریافت از دیتابیس و کش مجدد
func (r *TranslationRepo) GetText(tenantID uint, lang string, key string) (string, error) {
	cacheKey := fmt.Sprintf("tenant:%d:lang:%s:key:%s", tenantID, lang, key)

	// تلاش برای خواندن از Redis[cite: 1]
	val, err := r.Redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		return val, nil
	}

	// خواندن از PostgreSQL با در نظر گرفتن ایزوله‌سازی کارفرما[cite: 1]
	var trans domain.Translation
	if err := r.DB.Where("tenant_id = ? AND key = ?", tenantID, key).First(&trans).Error; err != nil {
		return "", err
	}

	var result string
	switch lang {
	case "fa":
		result = trans.Fa
	case "en":
		result = trans.En
	case "ru":
		result = trans.Ru
	case "ar":
		result = trans.Ar
	} // پشتیبانی از ۴ زبان[cite: 1]

	// ذخیره در Redis برای کوئری‌های بعدی[cite: 1]
	r.Redis.Set(context.Background(), cacheKey, result, 24*time.Hour)

	return result, nil
}
