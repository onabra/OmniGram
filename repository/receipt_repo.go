package repository

import (
	"OmniGram/domain"

	"gorm.io/gorm"
)

type ReceiptRepository struct {
	DB *gorm.DB
}

func (r *ReceiptRepository) CreateReceipt(receipt *domain.Receipt) error {
	return r.DB.Create(receipt).Error
}

func (r *ReceiptRepository) GetReceiptByID(tenantID uint, id uint) (*domain.Receipt, error) {
	var receipt domain.Receipt
	err := r.DB.Where("tenant_id = ? AND id = ?", tenantID, id).First(&receipt).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (r *ReceiptRepository) UpdateStatus(receipt *domain.Receipt) error {
	return r.DB.Model(receipt).Update("status", receipt.Status).Error
}

// GetVerifiedCard بررسی اینکه آیا کاربر قبلاً احراز هویت کارت انجام داده است یا خیر
func (r *ReceiptRepository) GetVerifiedCard(tenantID uint, userID uint) (*domain.VerifiedCard, error) {
	var card domain.VerifiedCard
	err := r.DB.Where("tenant_id = ? AND user_id = ? AND is_verified = ?", tenantID, userID, true).First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// ApproveUserCard ثبت و تایید دائم کارت کاربر پس از اولین خرید موفق جهت تسریع در خریدهای بعدی
func (r *ReceiptRepository) ApproveUserCard(tenantID uint, userID uint, cardNumber string) error {
	var existing domain.VerifiedCard
	err := r.DB.Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&existing).Error
	if err == nil {
		existing.IsVerified = true
		if cardNumber != "" {
			existing.CardNumber = cardNumber
		}
		return r.DB.Save(&existing).Error
	}
	return r.DB.Create(&domain.VerifiedCard{
		TenantID:   tenantID,
		UserID:     userID,
		CardNumber: cardNumber,
		IsVerified: true,
	}).Error
}
