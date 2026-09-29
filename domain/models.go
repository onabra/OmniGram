package domain

import (
	"time"
)

// اضافه کردن وضعیت کاربر برای جلوگیری از تداخل عملیات‌ها
type UserState string

const (
	StateNormal   UserState = "NORMAL"
	StateInGame   UserState = "IN_GAME"
	StateWaitTask UserState = "WAITING_FOR_TASK_PROOF"
)

// Tenant (موجودیت اصلی برای معماری SaaS)
type Tenant struct {
	ID              uint   `gorm:"primaryKey"`
	BotToken        string // توکن اختصاصی ربات برای این کارفرما[cite: 1]
	IsActive        bool   // وضعیت اشتراک (فعال/منقضی)[cite: 1]
	ReportChannelID string // آیدی کانال‌های گزارش مربوط به همین کارفرما[cite: 1]
	OrderChannelID  string // آیدی کانال‌های سفارشات مربوط به همین کارفرما[cite: 1]
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Category (مثلاً ممبر، ویو، بوست)
type Category struct {
	ID       uint   `gorm:"primaryKey"`
	TenantID uint   `gorm:"index"`
	ParentID *uint  // برای زیرمجموعه‌سازی (مثلاً ممبر -> ایرانی -> گارانتی‌دار)
	NameKey  string // فراخوانی از Translation بر اساس کلید
	IsActive bool
}

// ServiceItem (مثلاً ممبر ایرانی گارانتی 30 روزه)
type ServiceItem struct {
	ID         uint `gorm:"primaryKey"`
	CategoryID uint `gorm:"index"`
	TenantID   uint `gorm:"index"`
	Title      string
	NameKey    string // کلید ترجمه نام سرویس
	PriceToman float64
	PriceStars float64
	Capacity   int
}

// Translation (سیستم چندزبانه و بدون هاردکد)
type Translation struct {
	ID       uint   `gorm:"primaryKey"`
	TenantID uint   `gorm:"index"` // ایزوله‌سازی داده‌های هر کارفرما[cite: 1]
	Key      string `gorm:"index"` // کلید واژه برای فراخوانی در کد (مثل btn_stars)[cite: 1]
	Fa       string // مقدار به زبان فارسی[cite: 1]
	En       string // مقدار به زبان انگلیسی[cite: 1]
	Ru       string // مقدار به زبان روسی[cite: 1]
	Ar       string // مقدار به زبان عربی[cite: 1]
}

// User (اطلاعات کاربران و کیف پول‌های سه‌گانه)
type User struct {
	ID             uint      `gorm:"primaryKey"`
	TenantID       uint      `gorm:"index"` // فیلد کلیدی برای ایزوله ماندن داده‌های کارفرما[cite: 1]
	TelegramID     int64     `gorm:"index"` // شناسه کاربری در تلگرام
	TomanBalance   float64   // موجودی تومانی کاربر[cite: 1]
	PointBalance   float64   // موجودی امتیازی کاربر[cite: 1]
	StarsBalance   float64   // موجودی استارز کاربر[cite: 1]
	Language       string    // زبان انتخابی کاربر برای نمایش متون[cite: 1]
	LastActiveDate time.Time // تاریخ آخرین فعالیت برای محاسبه قانون ۱۰ روز خاموشی و تخصیص ۵۰٪ امتیاز به معرف جدید[cite: 1]
	State          UserState `gorm:"default:'NORMAL'"` // برای جلوگیری از باز شدن چند بخش همزمان
}

// Referral (سیستم رفرال‌های ۴ گانه)
type Referral struct {
	ID           uint   `gorm:"primaryKey"`
	TenantID     uint   `gorm:"index"`
	ReferrerID   uint   `gorm:"index"` // شناسه کاربری که دعوت کرده است (معرف)
	ReferredID   uint   `gorm:"index"` // شناسه کاربر دعوت شده (زیرمجموعه)
	ReferralType string // نوع لینک: استارز، گیفت، لیگ، یا قرعه‌کشی[cite: 1]
	// با استفاده از تگ‌های gorm باید یکتایی (Unique) روی ترکیب ReferrerID, ReferredID و ReferralType ایجاد شود تا از تداخل جلوگیری شود[cite: 1]
}

// Task (وظایف، ظرفیت‌ها و جریمه‌ها)
type Task struct {
	ID            uint    `gorm:"primaryKey"`
	TenantID      uint    `gorm:"index"`
	Title         string  // عنوان وظیفه
	Capacity      int     // ظرفیت تعیین شده برای انجام تسک[cite: 1]
	StayDuration  int     // زمان ماندگاری در کانال/گروه (بر حسب ساعت یا روز)[cite: 1]
	PenaltyAmount float64 // مبلغ جریمه در صورت لفت دادن از کانال‌های جوین اجباری[cite: 1]
	RewardAmount  float64 // پاداشی که پس از تایید تسک داده می‌شود
	TaskType      string  // "auto_join" یا "manual_proof"
	RewardType    string  // "point" یا "toman"
	RequiredDesc  string  // توضیحات ثبت نام آبان تتر
	Description   string  // توضیحات برای تسک‌های دستی
}

type TaskSubmission struct {
	ID       uint
	TenantID uint
	UserID   uint
	TaskID   uint
	Status   string // "pending", "approved", "rejected"
	FileID   string // آیدی عکس اسکرین شات ارسالی
}

type GameSession struct {
	ID          uint `gorm:"primaryKey"`
	TenantID    uint `gorm:"index"`
	CreatorID   uint
	JoinerID    *uint
	BetAmount   float64
	GameType    string // "DICE" یا "CASINO"
	Status      string // "WAITING", "IN_PROGRESS", "FINISHED"
	CreatorRoll *int
	JoinerRoll  *int
}

// VerifiedCard (سیستم KYC و احراز هویت)
type VerifiedCard struct {
	ID         uint   `gorm:"primaryKey"`
	TenantID   uint   `gorm:"index"`
	UserID     uint   `gorm:"index"` // ارتباط با کاربری که کارت را ثبت کرده
	CardNumber string // ذخیره دائم شماره کارت تأیید شده[cite: 1]
	IsVerified bool   // وضعیت تایید کارت جهت تسریع در خریدهای بعدی بدون نیاز به ارسال مجدد عکس[cite: 1]
}

// Order (رهگیری و مدیریت سفارشات)
type Order struct {
	ID               uint   `gorm:"primaryKey"`
	TenantID         uint   `gorm:"index"`
	UserID           uint   `gorm:"index"`
	TrackingCode     string `gorm:"uniqueIndex"` // کد یکتای رهگیری سفارشات[cite: 1]
	Status           string // وضعیت فعلی سفارش (در انتظار، تکمیل، لغو)[cite: 1]
	CancelRequested  bool   // ثبت درخواست لغو سفارش با محاسبه ۱۰٪ کارمزد[cite: 1]
	HasRefill        bool   // وضعیت درخواست ریفیل (ریزش)[cite: 1]
	SpeedUpRequested bool   // درخواست سرعت‌دهی به سفارش[cite: 1]
}
