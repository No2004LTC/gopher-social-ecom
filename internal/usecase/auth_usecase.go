package usecase

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"github.com/No2004LTC/gopher-social-ecom/pkg/auth"
	"github.com/redis/go-redis/v9"
)

// Struct dùng các hàm trong UserRepository và load config
type authUsecase struct {
	userRepo    domain.UserRepository
	cfg         *config.Config
	redisClient *redis.Client
}

// NewAuthUsecase khởi tạo tầng nghiệp vụ Authentication
func NewAuthUsecase(repo domain.UserRepository, cfg *config.Config, rdb *redis.Client) domain.UserUsecase {
	return &authUsecase{
		userRepo:    repo,
		cfg:         cfg,
		redisClient: rdb,
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
func (u *authUsecase) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	// Chỉ dùng Email
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, err
	}

	if user == nil || user.ID == 0 {
		return "", nil, errors.New("email hoặc mật khẩu không chính xác")
	}

	match, err := auth.ComparePassword(password, user.PasswordHash)
	if err != nil || !match {
		return "", nil, errors.New("email hoặc mật khẩu không chính xác")
	}

	expiry, _ := time.ParseDuration(u.cfg.JWTExpiry)
	token, err := auth.GenerateToken(user.ID, u.cfg.JWTSecret, expiry)
	if err != nil {
		return "", nil, err
	}

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

func (uc *authUsecase) GetFriendSuggestions(ctx context.Context, userID int64) ([]domain.SuggestedUser, error) {
	// Nghiệp vụ: Lấy tối đa 10 người gợi ý để tránh nặng server
	limit := 10

	suggestions, err := uc.userRepo.GetSuggestedUsers(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	return suggestions, nil
}

func (uc *authUsecase) GetOnlineContacts(ctx context.Context, userID int64) ([]dto.UserCompact, error) {
	// 1. Lấy danh sách những người đang Follow mình từ Database
	// Chúng ta tận dụng hàm GetFollowers có sẵn trong Repo của cậu
	// Limit 50 người, Offset 0 (Cậu có thể tùy chỉnh nếu muốn phân trang)
	contacts, err := uc.userRepo.GetFollowers(ctx, userID, 50, 0)
	log.Printf("🔍 [DEBUG] User %d có %d followers trong DB", userID, len(contacts))
	if err != nil {
		return nil, err
	}

	// 2. Lấy toàn bộ danh sách ID đang online từ Redis (Hash Map)
	// Key "system:online_users" lưu trữ [userID]:[connection_count]
	onlineMap, err := uc.redisClient.HGetAll(ctx, "system:online_users").Result()
	log.Printf("🔍 [DEBUG] Redis Online Map: %v", onlineMap)
	if err != nil {
		// Nếu Redis gặp sự cố, ta vẫn trả về danh sách nhưng IsOnline mặc định là false
		// để tránh làm sập cả trang web của người dùng.
		return contacts, nil
	}

	// 3. Khớp dữ liệu: Duyệt qua danh sách từ DB và kiểm tra trạng thái trong Redis
	var onlineList []dto.UserCompact
	for _, u := range contacts {
		userIDStr := strconv.FormatInt(u.ID, 10)
		countStr, exists := onlineMap[userIDStr]
		log.Printf("🔍 [DEBUG] Kiểm tra User %s: Exists=%v, Count=%s", userIDStr, exists, countStr)

		if exists {
			count, _ := strconv.Atoi(countStr)
			if count > 0 {
				u.IsOnline = true
				onlineList = append(onlineList, u)
			}
		}
	}
	return onlineList, nil
}
