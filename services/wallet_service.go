package services

import (
	"OmniGram/core"
	"OmniGram/domain"
	"OmniGram/repository"
	"errors"
)

type WalletService struct {
	UserRepo repository.UserRepository
	TxRepo   *repository.TransactionRepository
	Logger   *core.CentralLogger
}

func NewWalletService(
	userRepo repository.UserRepository,
	txRepo *repository.TransactionRepository,
	logger *core.CentralLogger) *WalletService {
	return &WalletService{
		UserRepo: userRepo,
		TxRepo:   txRepo,
		Logger:   logger,
	}
}

// TransferBalance انتقال موجودی بین دو کاربر با احتساب کارمزد
func (s *WalletService) TransferBalance(tenantID uint, senderID uint, destTelegramID int64, amount float64, currency string, feePercent float64) error {
	if amount <= 0 {
		return errors.New("مبلغ وارد شده نامعتبر است")
	}

	// 1. پیدا کردن فرستنده و گیرنده
	sender, err := s.UserRepo.GetUserByID(tenantID, senderID)
	if err != nil {
		return errors.New("کاربر مبدأ یافت نشد")
	}

	receiver, err := s.UserRepo.GetUserByTelegramID(tenantID, destTelegramID)
	if err != nil {
		return errors.New("کاربر مقصد در سیستم یافت نشد")
	}

	if sender.ID == receiver.ID {
		return errors.New("امکان انتقال موجودی به خودتان وجود ندارد")
	}

	// 2. بررسی موجودی فرستنده بر اساس نوع ارز
	switch currency {
	case "POINT":
		if sender.PointBalance < amount {
			return errors.New("موجودی امتیازی شما کافی نیست")
		}
	case "STARS":
		if sender.StarsBalance < amount {
			return errors.New("موجودی استارز شما کافی نیست")
		}
	case "TOMAN":
		if sender.TomanBalance < amount {
			return errors.New("موجودی تومانی شما کافی نیست")
		}
	default:
		return errors.New("نوع کیف پول نامعتبر است")
	}

	// 3. محاسبه مبلغ خالص دریافتی (پس از کسر کارمزد)
	fee := amount * (feePercent / 100.0)
	receiveAmount := amount - fee

	// 4. تعیین مقادیر برای آپدیت دیتابیس
	var sToman, sPoint, sStars float64
	var rToman, rPoint, rStars float64

	switch currency {
	case "POINT":
		sPoint = -amount
		rPoint = receiveAmount
	case "STARS":
		sStars = -amount
		rStars = receiveAmount
	case "TOMAN":
		sToman = -amount
		rToman = receiveAmount
	}

	// 5. کسر موجودی از فرستنده
	err = s.UserRepo.AddBalances(tenantID, sender.ID, sToman, sPoint, sStars)
	if err != nil {
		s.Logger.Error("WalletService - AddBalances (Decrease Sender)", err)
		return errors.New("خطا در کسر موجودی از حساب شما")
	}

	// 6. واریز موجودی به گیرنده
	err = s.UserRepo.AddBalances(tenantID, receiver.ID, rToman, rPoint, rStars)
	if err != nil {
		s.Logger.Error("WalletService - AddBalances (Increase Receiver)", err)
		return errors.New("خطا در واریز به حساب مقصد")
	}

	// 7. ثبت تراکنش در تاریخچه برای فرستنده (منفی)
	_ = s.TxRepo.Create(&domain.Transaction{
		TenantID: tenantID,
		UserID:   sender.ID,
		Amount:   -amount,
		Type:     "TRANSFER",
		Currency: currency,
	})

	// 8. ثبت تراکنش در تاریخچه برای گیرنده (مثبت)
	_ = s.TxRepo.Create(&domain.Transaction{
		TenantID: tenantID,
		UserID:   receiver.ID,
		Amount:   receiveAmount,
		Type:     "DEPOSIT",
		Currency: currency,
	})

	return nil
}
