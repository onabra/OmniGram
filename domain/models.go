package domain

import (
	"time"
)

// اضافه کردن وضعیت کاربر برای جلوگیری از تداخل عملیات‌ها
type UserState string

const (
	StateNormal             UserState = "NORMAL"
	StateInGame             UserState = "IN_GAME"
	StateWaitTask           UserState = "WAITING_FOR_TASK_PROOF"
	StateWaitTransferDest   UserState = "WAIT_TRANSFER_DEST"   // در انتظار ارسال آیدی مقصد
	StateWaitTransferAmount UserState = "WAIT_TRANSFER_AMOUNT" // در انتظار ارسال مبلغ انتقال

	// Sponsorship States added
	StateWaitSponsorLink  UserState = "WAIT_SPONSOR_LINK"  // Waiting for sponsor to send their channel/group link
	StateWaitSponsorCount UserState = "WAIT_SPONSOR_COUNT" // Waiting for sponsor to send the requested member count

	// Support & Withdrawal States
	StateWaitSupportMsg       UserState = "WAIT_SUPPORT_MSG"        // در انتظار تایپ پیام برای پشتیبانی
	StateWaitWithdrawAmount   UserState = "WAIT_WITHDRAW_AMOUNT"    // در انتظار مبلغ برداشت
	StateWaitWithdrawWallet   UserState = "WAIT_WITHDRAW_WALLET"    // در انتظار آدرس ولت/آیدی مقصد
	StateWaitSupportOrderCode UserState = "WAIT_SUPPORT_ORDER_CODE" // در انتظار کد رهگیری سفارش
	StateWaitPaymentReceipt   UserState = "WAIT_PAYMENT_RECEIPT"    // در انتظار ارسال عکس فیش واریزی

	// Shop States
	StateWaitServiceLink  UserState = "WAIT_SERVICE_LINK"  // انتظار لینک برای انجام سفارش فروشگاه
	StateWaitServiceCount UserState = "WAIT_SERVICE_COUNT" // انتظار تعداد سفارش فروشگاه

	StateWaitTrackingCode UserState = "WAIT_TRACKING_CODE"  // در انتظار کد رهگیری سفارش
	StateWaitNewOrderLink UserState = "WAIT_NEW_ORDER_LINK" // در انتظار لینک جدید برای ویرایش سفارش

	StateAdminWaitUserID        UserState = "ADMIN_WAIT_USER_ID"        // ادمین در انتظار ارسال آیدی عددی کاربر
	StateAdminWaitBalanceAmount UserState = "ADMIN_WAIT_BALANCE_AMOUNT" // ادمین در انتظار ارسال مبلغ برای شارژ/کسر
	StateAdminWaitSettingValue  UserState = "ADMIN_WAIT_SETTING_VALUE"  // ادمین در انتظار ارسال مقدار جدید برای تنظیمات

	StateAdminWaitCategoryName    UserState = "ADMIN_WAIT_CAT_NAME"      // انتظار نام دسته‌بندی
	StateAdminWaitServiceName     UserState = "ADMIN_WAIT_SRV_NAME"      // انتظار نام سرویس
	StateAdminWaitServicePriceUSD UserState = "ADMIN_WAIT_SRV_PRICE_USD" // انتظار قیمت دلاری سرویس
	StateAdminWaitDollarPrice     UserState = "ADMIN_WAIT_DOLLAR_PRICE"  // انتظار قیمت روز دلار

	StateAdminWaitForceJoinData UserState = "ADMIN_WAIT_FORCEJOIN_DATA" // انتظار لینک و آیدی کانال اجباری
	StateAdminWaitTaskPenalty   UserState = "ADMIN_WAIT_TASK_PENALTY"   // انتظار مبلغ جریمه تسک
	StateAdminWaitTaskCapacity  UserState = "ADMIN_WAIT_TASK_CAPACITY"  // انتظار ظرفیت جدید تسک

	StateAdminWaitTicketReply  UserState = "ADMIN_WAIT_TICKET_REPLY"  // انتظار پاسخ ادمین به تیکت کاربر
	StateAdminWaitBroadcastMsg UserState = "ADMIN_WAIT_BROADCAST_MSG" // انتظار ارسال پیام یا فوروارد برای ارسال همگانی

	StateWaitConvertAmountStars UserState = "WAIT_CONVERT_AMOUNT_STARS" // انتظار مقدار استارز برای تبدیل
	StateWaitConvertAmountToman UserState = "WAIT_CONVERT_AMOUNT_TOMAN" // انتظار مقدار امتیاز برای تبدیل به تومان

	StateAdminWaitOrderSearch UserState = "ADMIN_WAIT_ORDER_SEARCH" // ادمین در انتظار ارسال کد رهگیری سفارش

)

// Tenant (موجودیت اصلی برای معماری SaaS)
type Tenant struct {
	ID                  uint    `gorm:"primaryKey"`
	BotToken            string  // توکن اختصاصی ربات برای این کارفرما[cite: 1]
	IsActive            bool    // وضعیت اشتراک (فعال/منقضی)[cite: 1]
	ReportChannelID     string  // آیدی کانال‌های گزارش مربوط به همین کارفرما[cite: 1]
	OrderChannelID      string  // آیدی کانال‌های سفارشات مربوط به همین کارفرما[cite: 1]
	WinnerChannelID     string  //  آیدی کانال مخصوص نتایج مسابقات و قرعه‌کشی‌ها
	TransferFeePercent  float64 // درصد کارمزد انتقال (مثلا 5.0 برای ۵ درصد)
	CancelFeePercent    float64 // درصد کارمزد لغو سفارش
	InactiveBonusAmount float64 // مقدار امتیازی که اگر کاربر ۱۰ روز غیرفعال بود به معرف جدید میرسه
	DollarPrice         float64 `gorm:"default:60000"` //  قیمت روز دلار (تومان)
	PointToTomanRate    float64 `gorm:"default:800"`   //  هر ۱ امتیاز معادل این مقدار تومان است
	PointToStarRate     float64 `gorm:"default:3.0"`   //  هر ۱ استارز نیازمند این مقدار امتیاز است
	CreatedAt           time.Time
	UpdatedAt           time.Time
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
	ID            uint `gorm:"primaryKey"`
	CategoryID    uint `gorm:"index"`
	TenantID      uint `gorm:"index"`
	Title         string
	NameKey       string // کلید ترجمه نام سرویس
	PriceToman    float64
	PriceStars    float64
	Capacity      int
	PriceUSD      float64 //  قیمت پایه دلاری
	ProfitPercent float64 //  درصد سود این سرویس
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

type Transaction struct {
	ID        uint    `gorm:"primaryKey"`
	TenantID  uint    `gorm:"index"`
	UserID    uint    `gorm:"index"`
	Amount    float64 // مقدار (مثبت برای واریز/دریافت، منفی برای خرید/انتقال)
	Type      string  // "DEPOSIT", "PURCHASE", "TRANSFER"
	Currency  string  // "TOMAN", "POINT", "STARS"
	CreatedAt time.Time
}

// User (اطلاعات کاربران و کیف پول‌های سه‌گانه)
type User struct {
	ID                      uint  `gorm:"primaryKey"`
	TenantID                uint  `gorm:"index"` // فیلد کلیدی برای ایزوله ماندن داده‌های کارفرما[cite: 1]
	TelegramID              int64 `gorm:"index"` // شناسه کاربری در تلگرام
	FirstName               string
	TomanBalance            float64 // موجودی تومانی کاربر[cite: 1]
	PointBalance            float64 // موجودی امتیازی کاربر[cite: 1]
	StarsBalance            float64 // موجودی استارز کاربر[cite: 1]
	DrawTickets             int
	CurrentTaskID           uint    //  ذخیره آیدی تسک دستی که کاربر در حال ارسال عکس برای آن است
	CurrentOrderID          uint    // ذخیره آیدی سفارشی که کاربر در حال ویرایش لینک آن است
	Language                string  // زبان انتخابی کاربر برای نمایش متون[cite: 1]
	CurrentWithdrawCurrency string  // ذخیره نوع ارز درخواستی برای برداشت (STARS یا GIFT)
	CurrentWithdrawAmount   float64 // ذخیره مقدار درخواستی برای برداشت
	AdminTargetTaskID       uint
	AdminTargetTicketID     uint
	AdminTargetUserID       uint   // آیدی کاربری که ادمین در حال ویرایش اوست
	AdminTargetSetting      string // نام تنظیمی که ادمین در حال ویرایش آن است
	AdminTargetAction       string // نوع عملیات ادمین (مثل کسر یا افزایش)
	AdminTargetCategoryID   uint
	LastActiveDate          time.Time // تاریخ آخرین فعالیت برای محاسبه قانون ۱۰ روز خاموشی و تخصیص ۵۰٪ امتیاز به معرف جدید[cite: 1]
	LastFreeGiftDate        time.Time
	State                   UserState `gorm:"default:'NORMAL'"` // برای جلوگیری از باز شدن چند بخش همزمان
}

// Referral (سیستم رفرال‌های ۴ گانه)
type Referral struct {
	ID           uint   `gorm:"primaryKey"`
	TenantID     uint   `gorm:"index"`
	ReferrerID   uint   `gorm:"index"` // شناسه کاربری که دعوت کرده است (معرف)
	ReferredID   uint   `gorm:"index"` // شناسه کاربر دعوت شده (زیرمجموعه)
	ReferralType string // نوع لینک: استارز، گیفت، لیگ، یا قرعه‌کشی[cite: 1]
	// با استفاده از تگ‌های gorm باید یکتایی (Unique) روی ترکیب ReferrerID, ReferredID و ReferralType ایجاد شود تا از تداخل جلوگیری شود[cite: 1]
	Status string `gorm:"default:'PENDING'"` // مقادیر: PENDING, APPROVED, REJECTED
}

// Task (وظایف، ظرفیت‌ها و جریمه‌ها)
type Task struct {
	ID            uint    `gorm:"primaryKey"`
	TenantID      uint    `gorm:"index"`
	OrderID       uint    `gorm:"index"` // اضافه شده: ارتباط این تسک با سفارش اسپانسر
	Title         string  // عنوان وظیفه
	TargetLink    string  // اضافه شده: لینک کانال/گروه اسپانسر برای عضویت کاربران
	Capacity      int     // ظرفیت تعیین شده برای انجام تسک (همان تعداد ممبر درخواستی)
	StayDuration  int     // زمان ماندگاری در کانال/گروه (بر حسب ساعت یا روز)
	PenaltyAmount float64 // مبلغ جریمه در صورت لفت دادن از کانال‌های جوین اجباری
	RewardAmount  float64 // پاداشی که پس از تایید تسک داده می‌شود
	TaskType      string  // "auto_join" یا "manual_proof"
	RewardType    string  // "point" یا "toman"
	Status        string  // اضافه شده: "ACTIVE", "COMPLETED", "PAUSED"
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

	CreatedAt time.Time //  برای محاسبه تعداد بازی‌ های امروز
	UpdatedAt time.Time
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
	ID           uint   `gorm:"primaryKey"`
	TenantID     uint   `gorm:"index"`
	UserID       uint   `gorm:"index"`
	TrackingCode string `gorm:"uniqueIndex"` // کد یکتای رهگیری سفارشات[cite: 1]

	// Sponsorship & Service Fields added
	ServiceItemID uint   `gorm:"index"` // ارتباط با ServiceItem (مثلا ممبر کیفیت بالا)
	TargetLink    string // لینک کانال/گروه اسپانسر برای انجام تسک
	Quantity      int    // تعداد ممبر درخواستی

	// Payment Fields added
	TotalAmount   float64 // مبلغ کل محاسبه شده سفارش
	PaymentMethod string  // روش پرداخت: WALLET_TOMAN, WALLET_POINT, STARS, GATEWAY, CARD
	PaymentStatus string  // وضعیت پرداخت: PENDING, PAID, REJECTED

	Status           string // وضعیت فعلی سفارش (در انتظار، تکمیل، لغو)
	CancelRequested  bool   // ثبت درخواست لغو سفارش با محاسبه ۱۰٪ کارمزد
	HasRefill        bool   // وضعیت درخواست ریفیل (ریزش)
	SpeedUpRequested bool   // درخواست سرعت‌دهی به سفارش

	CreatedAt time.Time //  زمان ثبت سفارش (برای محاسبه محدودیت ویرایش ۳۰ دقیقه‌ای)
	UpdatedAt time.Time //  زمان آخرین آپدیت وضعیت سفارش
}

// [domain/models.go]
// feat(models): add models for force join, withdrawal, support, and league

// ForceJoin (کانال‌های جوین اجباری)
type ForceJoin struct {
	ID       uint   `gorm:"primaryKey"`
	TenantID uint   `gorm:"index"`
	ChatID   string // آیدی عددی یا یوزرنیم کانال
	Link     string // لینک جوین پرایوت یا عمومی
	IsActive bool   `gorm:"default:true"`
}

// Withdrawal (درخواست‌های برداشت استارز و گیفت)
type Withdrawal struct {
	ID            uint    `gorm:"primaryKey"`
	TenantID      uint    `gorm:"index"`
	UserID        uint    `gorm:"index"`
	Amount        float64 // مقدار درخواستی
	Currency      string  // نوع ارز: STARS یا GIFT
	WalletAddress string  // آدرس کیف پول یا آیدی مقصد برای دریافت
	Status        string  `gorm:"default:'PENDING'"` // وضعیت: PENDING, APPROVED, REJECTED
	CreatedAt     time.Time
}

// SupportTicket (تیکت‌های پشتیبانی)
type SupportTicket struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"index"`
	UserID    uint   `gorm:"index"`
	OrderID   uint   `gorm:"index"`          // ارتباط این تیکت با یک سفارش خاص
	Status    string `gorm:"default:'OPEN'"` // وضعیت: OPEN, ANSWERED, CLOSED
	CreatedAt time.Time
}

// SupportMessage (پیام‌های درون هر تیکت پشتیبانی)
type SupportMessage struct {
	ID        uint   `gorm:"primaryKey"`
	TicketID  uint   `gorm:"index"`
	SenderID  int64  // آیدی تلگرام فرستنده (کاربر یا ادمین)
	IsAdmin   bool   // مشخص میکند آیا این پاسخ ادمین است یا پیام کاربر
	Text      string // متن پیام
	CreatedAt time.Time
}

// League (تنظیمات دوره‌های رفرال لیگ)
type League struct {
	ID        uint   `gorm:"primaryKey"`
	TenantID  uint   `gorm:"index"`
	Title     string // عنوان لیگ (مثلا لیگ پاییزه)
	StartDate time.Time
	EndDate   time.Time
	IsActive  bool `gorm:"default:true"`
}

// Receipt (فیش‌های واریزی)
type Receipt struct {
	ID        uint    `gorm:"primaryKey"`
	TenantID  uint    `gorm:"index"`
	UserID    uint    `gorm:"index"`
	OrderID   uint    `gorm:"index"` // سفارشی که این فیش برای آن پرداخت شده
	FileID    string  // آیدی عکس در تلگرام
	Amount    float64 // مبلغ فاکتور
	Status    string  `gorm:"default:'PENDING'"` // وضعیت: PENDING, APPROVED, REJECTED
	CreatedAt time.Time
}

// LuckyDraw (قرعه‌کشی‌ها)
type LuckyDraw struct {
	ID              uint      `gorm:"primaryKey"`
	TenantID        uint      `gorm:"index"`
	Title           string    // نام قرعه‌کشی
	Description     string    // توضیحات
	TicketCost      int       // هزینه ورود (تعداد تیکت)
	MinParticipants int       // حداقل شرکت‌کننده
	MaxParticipants int       // حداکثر شرکت‌کننده
	WinnersCount    int       // تعداد برندگان
	RewardPerWinner float64   // جایزه هر برنده (استارز/امتیاز)
	RewardType      string    // نوع جایزه: STARS, POINT, TOMAN
	Status          string    `gorm:"default:'ACTIVE'"` // ACTIVE, COMPLETED, CANCELED
	EndDate         time.Time // زمان پایان
}

// DrawEntry (شرکت‌کنندگان قرعه‌کشی)
type DrawEntry struct {
	ID          uint `gorm:"primaryKey"`
	TenantID    uint `gorm:"index"`
	LuckyDrawID uint `gorm:"index"`
	UserID      uint `gorm:"index"`
	Entries     int  // تعداد شانس‌های ثبت شده (کاربر می‌تواند چند تیکت خرج کند)
}
