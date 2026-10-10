package repository

import (
	"OmniGram/domain"

	"gorm.io/gorm"
)

type TicketRepository struct {
	DB *gorm.DB
}

func (r *TicketRepository) CreateTicket(ticket *domain.SupportTicket) error {
	return r.DB.Create(ticket).Error
}

func (r *TicketRepository) GetOpenTicket(tenantID uint, userID uint) (*domain.SupportTicket, error) {
	var ticket domain.SupportTicket
	err := r.DB.Where("tenant_id = ? AND user_id = ? AND status = ?", tenantID, userID, "OPEN").First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *TicketRepository) SaveMessage(msg *domain.SupportMessage) error {
	return r.DB.Create(msg).Error
}
