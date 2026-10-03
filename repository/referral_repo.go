package repository

import "gorm.io/gorm"

type ReferralRepository struct {
	db *gorm.DB
}

func NewReferralRepository(db *gorm.DB) *ReferralRepository {
	return &ReferralRepository{db: db}
}
