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
