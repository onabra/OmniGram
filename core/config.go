package core

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// AppConfig ساختار جامع تنظیمات برنامه
type AppConfig struct {
	ServerPort     string
	DatabaseURL    string
	RedisAddr      string
	RedisPassword  string
	LogChannelID   int64
	MasterBotToken string // فیلد اضافه شده برای توکن ربات ارشد

	SuperAdminIDs []int64 // اضافه شده برای آیدی سوپرادمین‌ها
	AdminIDs      []int64 // اضافه شده برای آیدی مدیران
}

// LoadConfig مدیریت مرکزی متغیرهای محیطی و اتصال به دیتابیس.
func LoadConfig() *AppConfig {
	// تلاش برای بارگذاری فایل .env (در صورت وجود)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from OS environment variables")
	}

	logChannelID, _ := strconv.ParseInt(os.Getenv("MASTER_LOG_CHANNEL_ID"), 10, 64)

	return &AppConfig{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "host=localhost user=postgres dbname=saas_bot port=5432"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		LogChannelID:   logChannelID,
		MasterBotToken: getEnv("MASTER_BOT_TOKEN", ""),

		SuperAdminIDs: getEnvAsIntSlice("SUPER_ADMIN_IDS"),
		AdminIDs:      getEnvAsIntSlice("ADMIN_IDS"),
	}
}

// [core/config.go]
// getEnvAsIntSlice تابع کمکی برای تبدیل لیست آیدی‌های کاما سپریت شده به آرایه عددی
func getEnvAsIntSlice(key string) []int64 {
	valStr := os.Getenv(key)
	if valStr == "" {
		return nil
	}

	strSlice := strings.Split(valStr, ",")
	var intSlice []int64
	for _, s := range strSlice {
		s = strings.TrimSpace(s)
		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			intSlice = append(intSlice, id)
		}
	}
	return intSlice
}

// getEnv تابع کمکی برای خواندن متغیر محیطی با مقدار پیش‌فرض
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
