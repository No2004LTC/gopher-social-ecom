package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/pkg/auth"
)

type authUsecase struct {
	userRepo domain.UserRepository
	cfg      *config.Config
}

// NewAuthUsecase khởi tạo tầng nghiệp vụ Authentication
func NewAuthUsecase(repo domain.UserRepository, cfg *config.Config) domain.UserUsecase {
	return &authUsecase{
		userRepo: repo,
		cfg:      cfg,
	}
}

// Register xử lý logic Đăng ký tài khoản
func (u *authUsecase) Register(ctx context.Context, username, email, password string) error {
	// 1. Kiểm tra xem email đã tồn tại chưa
	existingUser, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("email đã được sử dụng")
	}

	// 2. Hash mật khẩu bằng Argon2 (Sử dụng công cụ ở Task 2)
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	// 3. Tạo thực thể User mới
	newUser := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 4. Gọi Repository để lưu vào DB (Sử dụng công cụ ở Task 3)
	return u.userRepo.Create(ctx, newUser)
}

// Login xử lý logic Đăng nhập và trả về JWT Token
func (u *authUsecase) Login(ctx context.Context, email, password string) (string, error) {
	// 1. Tìm user theo email
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("thông tin đăng nhập không chính xác")
	}

	// 2. So sánh mật khẩu (Sử dụng công cụ ở Task 2)
	match, err := auth.ComparePassword(password, user.PasswordHash)
	if err != nil || !match {
		return "", errors.New("thông tin đăng nhập không chính xác")
	}

	// 3. Tạo JWT Token (Sử dụng công cụ ở Task 2)
	expiry, _ := time.ParseDuration(u.cfg.JWTExpiry)
	token, err := auth.GenerateToken(user.ID, u.cfg.JWTSecret, expiry)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *authUsecase) UpdateAvatar(ctx context.Context, userID int64, url string) error {
	// 1. Validate input
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	if url == "" {
		return errors.New("avatar URL cannot be empty")
	}

	// 2. Check if user exists
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 3. Update avatar in database
	return u.userRepo.UpdateAvatar(ctx, userID, url)
}

func (u *authUsecase) GetProfile(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

func (u *authUsecase) UpdateProfile(ctx context.Context, userID int64, username string) error {
	user := &domain.User{
		ID:       userID,
		Username: username,
	}

	return u.userRepo.Update(ctx, user)
}
