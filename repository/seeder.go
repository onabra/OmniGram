package repository

import (
	"log"

	"OmniGram/domain"

	"gorm.io/gorm"
)

// RunSeeder داده‌های اولیه سیستم SaaS را بررسی و در صورت نیاز در دیتابیس ثبت می‌کند
func RunSeeder(db *gorm.DB, masterBotToken string) {
	// ۱. بررسی و ثبت Tenant (کارفرمای اصلی)
	var count int64
	db.Model(&domain.Tenant{}).Where("bot_token = ?", masterBotToken).Count(&count)

	var tenantID uint = 1
	if count == 0 {
		newTenant := domain.Tenant{
			BotToken: masterBotToken,
			IsActive: true,
		}
		db.Create(&newTenant)
		tenantID = newTenant.ID
		log.Println("✅ توکن جدید کارفرما در دیتابیس ثبت شد.")
	} else {
		var existingTenant domain.Tenant
		db.Where("bot_token = ?", masterBotToken).First(&existingTenant)
		tenantID = existingTenant.ID
	}

	// ۲. بررسی و ثبت متون اولیه منوها (دیکشنری داینامیک بدون هاردکد)
	var transCount int64
	db.Model(&domain.Translation{}).Where("tenant_id = ?", tenantID).Count(&transCount)

	if transCount == 0 {
		defaultTranslations := []domain.Translation{
			{TenantID: tenantID, Key: "msg_welcome", Fa: "به ربات خوش آمدید!", En: "Welcome to the bot!", Ru: "Добро пожаловать!", Ar: "أهلاً بك!"},
			{TenantID: tenantID, Key: "btn_withdraw", Fa: "💳 برداشت", En: "Withdraw", Ru: "Снять", Ar: "سحب"},
			{TenantID: tenantID, Key: "btn_account", Fa: "👤 حساب کاربری", En: "Account", Ru: "Аккаунт", Ar: "حساب"},
			{TenantID: tenantID, Key: "btn_shop", Fa: "🛒 فروشگاه", En: "Shop", Ru: "Магазин", Ar: "متجر"},
			{TenantID: tenantID, Key: "btn_developer", Fa: "👨‍💻 توسعه‌دهنده", En: "Developer", Ru: "Разработчик", Ar: "مطور"},
			{TenantID: tenantID, Key: "btn_task", Fa: "📋 وظایف", En: "Tasks", Ru: "Задачи", Ar: "مهام"},
			{TenantID: tenantID, Key: "btn_referral", Fa: "🔗 زیرمجموعه‌گیری", En: "Referrals", Ru: "Рефералы", Ar: "إحالة"},
			{TenantID: tenantID, Key: "btn_free_gift", Fa: "🎁 هدیه رایگان", En: "Free Gift", Ru: "Бесплатный подарок", Ar: "هدية مجانية"},
			{TenantID: tenantID, Key: "btn_ref_league", Fa: "🏆 لیگ", En: "League", Ru: "Лига", Ar: "دوري"},
		}
		db.Create(&defaultTranslations)
		log.Println("✅ داده‌های اولیه (دکمه‌ها و متون) در جدول Translations ثبت شد.")
	}
}
