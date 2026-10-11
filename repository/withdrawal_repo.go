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
