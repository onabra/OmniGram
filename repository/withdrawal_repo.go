package repository

import (
	"OmniGram/domain"

	"gorm.io/gorm"
)

type WithdrawalRepository struct {
	DB *gorm.DB
}

func (r *WithdrawalRepository) Create(withdrawal *domain.Withdrawal) error {
	return r.DB.Create(withdrawal).Error
}

func (r *WithdrawalRepository) GetByID(tenantID uint, id uint) (*domain.Withdrawal, error) {
	var w domain.Withdrawal
	err := r.DB.Where("tenant_id = ? AND id = ?", tenantID, id).First(&w).Error
	return &w, err
}

func (r *WithdrawalRepository) UpdateStatus(withdrawal *domain.Withdrawal) error {
	return r.DB.Model(withdrawal).Update("status", withdrawal.Status).Error
}

// GetActiveGifts دریافت لیست گیفت‌های فعال برای منوی برداشت و پنل مدیریت
func (r *WithdrawalRepository) GetActiveGifts(tenantID uint) ([]domain.GiftItem, error) {
	var gifts []domain.GiftItem
	err := r.DB.Where("tenant_id = ? AND is_active = ?", tenantID, true).Order("cost_points asc").Find(&gifts).Error
	return gifts, err
}

// GetGiftByID دریافت اطلاعات یک گیفت خاص
func (r *WithdrawalRepository) GetGiftByID(tenantID uint, id uint) (*domain.GiftItem, error) {
	var gift domain.GiftItem
	err := r.DB.Where("tenant_id = ? AND id = ?", tenantID, id).First(&gift).Error
	if err != nil {
		return nil, err
	}
	return &gift, nil
}

// CreateGift افزودن گیفت جدید توسط ادمین
func (r *WithdrawalRepository) CreateGift(gift *domain.GiftItem) error {
	return r.DB.Create(gift).Error
}

// DeleteGift حذف گیفت توسط ادمین
func (r *WithdrawalRepository) DeleteGift(tenantID uint, id uint) error {
	return r.DB.Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&domain.GiftItem{}).Error
}
