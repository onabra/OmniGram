package core

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Ctx = context.Background()

// SetupDatabase اتصال به دیتابیس PostgreSQL برای ذخیره اطلاعات ایزوله شده
func SetupDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// SetupRedis اتصال به ردیس برای کش کردن ترجمه‌ها و دیکشنری داینامیک
func SetupRedis(addr string, password string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return rdb
}
