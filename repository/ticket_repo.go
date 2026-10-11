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

// GetOpenTickets واکشی لیست تیکت‌های در انتظار پاسخ
func (r *TicketRepository) GetOpenTickets(tenantID uint) ([]domain.SupportTicket, error) {
	var tickets []domain.SupportTicket
	err := r.DB.Where("tenant_id = ? AND status = ?", tenantID, "OPEN").Order("id asc").Find(&tickets).Error
	return tickets, err
}

// GetTicketByID واکشی اطلاعات یک تیکت خاص
func (r *TicketRepository) GetTicketByID(tenantID uint, ticketID uint) (*domain.SupportTicket, error) {
	var ticket domain.SupportTicket
	err := r.DB.Where("tenant_id = ? AND id = ?", tenantID, ticketID).First(&ticket).Error
	return &ticket, err
}

// UpdateStatus تغییر وضعیت تیکت (مثلا از OPEN به CLOSED یا ANSWERED)
func (r *TicketRepository) UpdateStatus(ticket *domain.SupportTicket) error {
	return r.DB.Model(ticket).Update("status", ticket.Status).Error
}

// GetTicketMessages واکشی پیام‌های یک تیکت
func (r *TicketRepository) GetTicketMessages(ticketID uint) ([]domain.SupportMessage, error) {
	var msgs []domain.SupportMessage
	err := r.DB.Where("ticket_id = ?", ticketID).Order("created_at asc").Find(&msgs).Error
	return msgs, err
}
