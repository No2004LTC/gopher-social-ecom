package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"github.com/No2004LTC/gopher-social-ecom/pkg/auth"
)

// Struct dùng các hàm trong UserRepository và load config
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
	// Kiểm tra xem email đã tồn tại chưa
	existingUser, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("email đã được sử dụng")
	}

	// Hash mật khẩu bằng Argon2 (Sử dụng công cụ ở Task 2)
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

	// 4. Gọi Repository để lưu vào DB
	return u.userRepo.Create(ctx, newUser)
}

// Login xử lý logic Đăng nhập và trả về JWT Token
func (u *authUsecase) Login(ctx context.Context, identifier, password string) (string, *domain.User, error) {
	// Gọi hàm tìm kiếm linh hoạt từ Repo
	user, err := u.userRepo.GetUserByIdentifier(ctx, identifier)
	if err != nil {
		// 👉 SỬA: Thêm 'nil' vào giữa
		return "", nil, err
	}

	// Nếu ID = 0 hoặc user = nil nghĩa là không tìm thấy
	if user == nil || user.ID == 0 {
		// 👉 SỬA: Thêm 'nil' vào giữa
		return "", nil, errors.New("tài khoản hoặc mật khẩu không chính xác")
	}

	// So sánh mật khẩu
	match, err := auth.ComparePassword(password, user.PasswordHash)
	if err != nil || !match {
		// 👉 SỬA: Thêm 'nil' vào giữa
		return "", nil, errors.New("tài khoản hoặc mật khẩu không chính xác")
	}

	// Tạo JWT Token
	expiry, _ := time.ParseDuration(u.cfg.JWTExpiry)
	token, err := auth.GenerateToken(user.ID, u.cfg.JWTSecret, expiry)
	if err != nil {
		// 👉 SỬA: Thêm 'nil' vào giữa
		return "", nil, err
	}

	// Dòng này của cậu đúng chuẩn rồi!
	return token, user, nil
}

func (u *authUsecase) UpdateAvatar(ctx context.Context, userID int64, url string) error {
	// Kiêm tra đầu vào
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	if url == "" {
		return errors.New("avatar URL cannot be empty")
	}

	// Kiểm tra xem user có tồn tại không
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// lưu avatar
	return u.userRepo.UpdateAvatar(ctx, userID, url)
}

// Trả về thông tin user
func (u *authUsecase) GetProfile(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

// Cập nhật username
func (u *authUsecase) UpdateProfile(ctx context.Context, userID int64, username string) error {
	user := &domain.User{
		ID:       userID,
		Username: username,
	}

	return u.userRepo.Update(ctx, user)
}

// Tìm kiếm người dùng
func (u *authUsecase) SearchUsers(ctx context.Context, currentUserID int64, query string, limit, offset int) ([]dto.UserCompact, error) {
	// Usecase bây giờ cực kỳ nhàn, chỉ việc gọi Repo và return
	return u.userRepo.SearchUsers(ctx, currentUserID, query, limit, offset)
}

func (u *authUsecase) GetFollowing(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error) {
	// Tầng Usecase là nơi chứa Business Logic.
	// Tương lai cậu có thể thêm logic kiểm tra cache (Redis) ở đây.
	// Hiện tại, ta chỉ việc gọi thẳng xuống tầng Repository:
	return u.userRepo.GetFollowing(ctx, currentUserID, limit, offset)
}

// Lấy danh sách những người đang theo dõi mình (Followers)
func (u *authUsecase) GetFollowers(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error) {
	return u.userRepo.GetFollowers(ctx, currentUserID, limit, offset)
}
