// [repository/game_repo.go]
package repository

import (
	"OmniGram/domain"

	"gorm.io/gorm"
)

type GameRepository struct {
	DB *gorm.DB
}

func (r *GameRepository) CreateGame(game *domain.GameSession) error {
	return r.DB.Create(game).Error
}

func (r *GameRepository) GetGameByID(tenantID uint, id uint) (*domain.GameSession, error) {
	var game domain.GameSession
	err := r.DB.Where("tenant_id = ? AND id = ?", tenantID, id).First(&game).Error
	return &game, err
}

func (r *GameRepository) UpdateGame(game *domain.GameSession) error {
	return r.DB.Save(game).Error
}

func (r *GameRepository) GetActiveGameForUser(tenantID uint, userID uint) (*domain.GameSession, error) {
	var game domain.GameSession
	err := r.DB.Where("tenant_id = ? AND status IN ('WAITING', 'IN_PROGRESS') AND (creator_id = ? OR joiner_id = ?)", tenantID, userID, userID).First(&game).Error
	return &game, err
}
