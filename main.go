package main

import (
	"log"
	"net/http"

	"OmniGram/core"
	"OmniGram/delivery/telegram"
	"OmniGram/repository"
)

func main() {
	// ۱. بارگذاری کامل تنظیمات استاندارد به جای هاردکد کردن مقادیر
	cfg := core.LoadConfig()

	// ۲. اتصال به دیتابیس و کَش با مقادیر خوانده شده از کانفیگ
	// در معماری SaaS، این دیتابیس تمامی اطلاعات کارفرماها را با شرط tenant_id ایزوله نگه می‌دارد
	db, err := core.SetupDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}

	// استفاده از ردیس برای جلوگیری از افت سرعت ربات در فراخوانی متون ترجمه شده[cite: 4]
	redisClient := core.SetupRedis(cfg.RedisAddr, cfg.RedisPassword)

	// ۳. مقداردهی ریپازیتوری‌ها (لایه Repository)
	tenantRepo := &repository.TenantRepository{DB: db}
	translationRepo := &repository.TranslationRepo{DB: db, Redis: redisClient}

	// ۴. مقداردهی ماژول کیبوردساز داینامیک
	keyboardBuilder := &telegram.KeyboardBuilder{
		TranslationRepo: translationRepo,
	}

	// ۵. مقداردهی هندلرهای تلگرام (لایه Delivery)
	// تمامی درخواست‌های تلگرام به این وب‌هوک می‌رسند و بر اساس Bot_Token به ربات مربوطه روت می‌شوند[cite: 4]
	webhookHandler := &telegram.WebhookHandler{
		TenantRepo:      tenantRepo,
		TranslationRepo: translationRepo,
		KeyboardBuilder: keyboardBuilder,
		LogChannelID:    cfg.LogChannelID, // آیدی کانال لاگ ارشد که از .env خوانده شده است
		MasterBotToken:  cfg.MasterBotToken,
	}

	// ۶. اجرای سرور برای دریافت Webhookها
	http.HandleFunc("/webhook/", webhookHandler.HandleUpdate)

	log.Printf("SaaS Bot Engine is running on port %s...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
}
