package repository

import (
	"OmniGram/domain"

	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create ثبت یک تراکنش جدید (واریز، برداشت، انتقال)
func (r *TransactionRepository) Create(tx *domain.Transaction) error {
	return r.db.Create(tx).Error
}

// GetRecentTransactions دریافت تراکنش‌های اخیر کاربر (محدود شده به تعداد دلخواه)
func (r *TransactionRepository) GetRecentTransactions(tenantID uint, userID uint, limit int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	// استفاده از tenant_id برای رعایت ایزوله‌سازی SaaS
	err := r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Order("created_at desc").
		Limit(limit).
		Find(&transactions).Error

	return transactions, err
}
