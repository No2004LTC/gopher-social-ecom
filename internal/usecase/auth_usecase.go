package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/pkg/auth"
	"github.com/No2004LTC/gopher-social-ecom/pkg/mail"
	"github.com/redis/go-redis/v9"
)

type authUsecase struct {
	authRepo    domain.AuthRepository
	cfg         *config.Config
	redisClient *redis.Client
	emailSender mail.EmailSender
}

func NewAuthUsecase(repo domain.AuthRepository, cfg *config.Config, rdb *redis.Client, mailSender mail.EmailSender) domain.AuthUsecase {
	return &authUsecase{
		authRepo:    repo,
		cfg:         cfg,
		redisClient: rdb,
		emailSender: mailSender,
	}
}

// Register
func (u *authUsecase) Register(ctx context.Context, username, email, password string) error {
	existingUser, _ := u.authRepo.GetByEmail(ctx, email)
	if existingUser != nil {
		return errors.New("email đã được sử dụng")
	}

	hashedPassword, _ := auth.HashPassword(password)
	newUser := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
	}
	return u.authRepo.Create(ctx, newUser)
}

// Login
func (u *authUsecase) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	user, err := u.authRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return "", nil, errors.New("email hoặc mật khẩu không chính xác")
	}

	match, _ := auth.ComparePassword(password, user.PasswordHash)
	if !match {
		return "", nil, errors.New("email hoặc mật khẩu không chính xác")
	}

	expiry, _ := time.ParseDuration(u.cfg.JWTExpiry)
	token, _ := auth.GenerateToken(user.ID, u.cfg.JWTSecret, expiry)
	return token, user, nil
}

// SendPasswordOTP
func (u *authUsecase) SendPasswordOTP(ctx context.Context, email string) error {
	user, _ := u.authRepo.GetByEmail(ctx, email)
	if user == nil {
		return errors.New("email không tồn tại")
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	u.redisClient.Set(ctx, "otp:"+email, otp, 5*time.Minute)

	go u.emailSender.SendEmail(email, "Mã OTP", "Mã của bạn là: "+otp)
	return nil
}

// ChangePasswordWithOTP
func (u *authUsecase) ChangePasswordWithOTP(ctx context.Context, email, otp, newPass string) error {
	storedOTP, _ := u.redisClient.Get(ctx, "otp:"+email).Result()
	if storedOTP != otp {
		return errors.New("mã OTP sai")
	}

	user, _ := u.authRepo.GetByEmail(ctx, email)
	hashed, _ := auth.HashPassword(newPass)

	return u.authRepo.UpdatePassword(ctx, user.ID, hashed)
}
