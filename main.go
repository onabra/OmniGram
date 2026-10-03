package main

import (
	"OmniGram/services"
	"log"

	"OmniGram/core"
	"OmniGram/delivery/telegram"
	"OmniGram/domain"
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
	appLogger := core.NewCentralLogger()
	// --- جایگزینی: مایگریشن کامل تمام جداول برای حل ارور relation does not exist ---
	log.Println("در حال ساخت و آپدیت جداول دیتابیس...")
	err = db.AutoMigrate(
		&domain.Tenant{},
		&domain.Category{},
		&domain.ServiceItem{},
		&domain.Translation{}, // حل مشکل ارور نبودن جدول ترجمه‌ها
		&domain.User{},
		&domain.Referral{},
		&domain.Task{},
		&domain.TaskSubmission{},
		&domain.GameSession{},
		&domain.VerifiedCard{},
		&domain.Order{},
		&domain.Transaction{},
	)
	if err != nil {
		log.Println("⚠️ خطا در مایگریشن جداول:", err)
	}

	// --- جایگزینی: فراخوانی تابع مجزای Seed Data ---
	repository.RunSeeder(db, cfg.MasterBotToken)

	redisClient := core.SetupRedis(cfg.RedisAddr, cfg.RedisPassword)

	// ۳. مقداردهی ریپازیتوری‌ها (لایه Repository)
	tenantRepo := &repository.TenantRepository{DB: db}
	translationRepo := &repository.TranslationRepo{DB: db, Redis: redisClient}

	userRepo := repository.NewUserRepository(db)
	refRepo := repository.NewReferralRepository(db)

	txRepo := repository.NewTransactionRepository(db)
	walletService := services.NewWalletService(userRepo, txRepo, appLogger)

	orderRepo := &repository.OrderRepository{DB: db}
	serviceRepo := &repository.ServiceRepository{DB: db}

	taskRepo := &repository.TaskRepository{DB: db}

	// ۴. مقداردهی ماژول کیبوردساز داینامیک
	keyboardBuilder := &telegram.KeyboardBuilder{
		TranslationRepo: translationRepo,
	}

	// ۵. مقداردهی هسته جدید Long Polling
	pollingHandler := &telegram.PollingHandler{
		TenantRepo:      tenantRepo,
		TranslationRepo: translationRepo,
		KeyboardBuilder: keyboardBuilder,
		LogChannelID:    cfg.LogChannelID,
		MasterBotToken:  cfg.MasterBotToken,

		UserRepo: userRepo,
		RefRepo:  refRepo,

		TxRepo:        txRepo,
		WalletService: walletService,

		Logger: appLogger,

		OrderRepo:   orderRepo,
		ServiceRepo: serviceRepo,

		TaskRepo: taskRepo,
	}

	// ۶. روشن کردن ربات و اجرای حلقه دریافت پیام‌ها
	log.Println("Starting OmniGram via Long Polling...")
	pollingHandler.StartPolling()
}
