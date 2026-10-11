// [repository/game_repo.go]
package repository

import (
	"OmniGram/domain"
	"time"

	"gorm.io/gorm"
)

type GameRepository struct {
	DB *gorm.DB
}

// GameStats ساختار خروجی آمار بازی‌های گروهی کاربر
type GameStats struct {
	TodayGames  int
	TotalGames  int
	TotalWins   int
	TotalLosses int
	Profit      float64
	Loss        float64
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

// GetUserGameStats محاسبه دقیق آمار برد و باخت و سود و زیان یک کاربر در بازی‌ها
func (r *GameRepository) GetUserGameStats(tenantID uint, userID uint, feePercent float64) (*GameStats, error) {
	var games []domain.GameSession
	// واکشی تمام بازی‌های پایان یافته‌ای که کاربر در آن حضور داشته است
	err := r.DB.Where("tenant_id = ? AND status = ? AND (creator_id = ? OR joiner_id = ?)", tenantID, "FINISHED", userID, userID).Find(&games).Error
	if err != nil {
		return nil, err
	}

	stats := &GameStats{}
	now := time.Now()

	for _, g := range games {
		stats.TotalGames++

		// بررسی اینکه آیا بازی برای امروز است
		if g.CreatedAt.Year() == now.Year() && g.CreatedAt.YearDay() == now.YearDay() {
			stats.TodayGames++
		}

		if g.CreatorRoll == nil || g.JoinerRoll == nil {
			continue
		}

		// تشخیص اینکه کاربر سازنده بازی بوده یا ملحق شونده
		isCreator := g.CreatorID == userID
		myRoll := *g.JoinerRoll
		oppRoll := *g.CreatorRoll
		if isCreator {
			myRoll = *g.CreatorRoll
			oppRoll = *g.JoinerRoll
		}

		// محاسبه سود و زیان (با کسر کارمزد میز برای بردها)
		if myRoll > oppRoll {
			stats.TotalWins++
			totalPot := g.BetAmount * 2
			feeAmount := totalPot * (feePercent / 100.0)
			winAmount := totalPot - feeAmount
			stats.Profit += (winAmount - g.BetAmount) // سود خالص (مبلغ برنده شده منهای شرط اولیه)
		} else if myRoll < oppRoll {
			stats.TotalLosses++
			stats.Loss += g.BetAmount // ضرر (مبلغی که باخته است)
		}
		// در حالت مساوی (مساوی شدن تاس‌ها) پولی از دست نرفته و سودی هم نشده است
	}
	return stats, nil
}
