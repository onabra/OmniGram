package main

import (
	"log"
	"net/http"

	"OmniGram/core"
	"OmniGram/delivery/telegram"
	"OmniGram/repository"
)

func main() {
	// ۱. بارگذاری کامل تنظیمات استاندارد
	cfg := core.LoadConfig()

	// ۲. اتصال به دیتابیس و کَش با مقادیر خوانده شده از کانفیگ
	db, err := core.SetupDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}

	redisClient := core.SetupRedis(cfg.RedisAddr, cfg.RedisPassword)

	// ۳. مقداردهی ریپازیتوری‌ها (لایه Repository)
	tenantRepo := &repository.TenantRepository{DB: db}
	translationRepo := &repository.TranslationRepo{DB: db, Redis: redisClient}

	// --- این دو خط اضافه شد تا ارور Undefined برطرف شود ---
	userRepo := repository.NewUserRepository(db)
	refRepo := &repository.ReferralRepository{DB: db}
	// ----------------------------------------------------

	// ۴. مقداردهی ماژول کیبوردساز داینامیک
	keyboardBuilder := &telegram.KeyboardBuilder{
		TranslationRepo: translationRepo,
	}

	// ۵. مقداردهی هندلرهای تلگرام (لایه Delivery)
	webhookHandler := &telegram.WebhookHandler{
		TenantRepo:      tenantRepo,
		TranslationRepo: translationRepo,
		KeyboardBuilder: keyboardBuilder,
		LogChannelID:    cfg.LogChannelID,
		MasterBotToken:  cfg.MasterBotToken,

		// پاس دادن ریپازیتوری‌های جدید
		UserRepo: userRepo,
		RefRepo:  refRepo,
		// SafeBot عمداً اینجا مقداردهی نمی‌‌شود تا برای هر کارفرما داینامیک ساخته شود
	}

	// ۶. اجرای سرور برای دریافت Webhookها
	http.HandleFunc("/webhook/", webhookHandler.HandleUpdate)

	log.Printf("SaaS Bot Engine is running on port %s...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
}
