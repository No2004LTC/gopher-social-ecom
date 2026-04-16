package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"github.com/No2004LTC/gopher-social-ecom/pkg/auth"
	"github.com/No2004LTC/gopher-social-ecom/pkg/mail"
	"github.com/redis/go-redis/v9"
)

// Struct dùng các hàm trong UserRepository và load config
type authUsecase struct {
	userRepo    domain.UserRepository
	cfg         *config.Config
	redisClient *redis.Client
	emailSender mail.EmailSender
}

// NewAuthUsecase khởi tạo tầng nghiệp vụ Authentication
func NewAuthUsecase(repo domain.UserRepository, cfg *config.Config, rdb *redis.Client, mailSender mail.EmailSender) domain.UserUsecase {
	return &authUsecase{
		userRepo:    repo,
		cfg:         cfg,
		redisClient: rdb,
		emailSender: mailSender,
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

func (u *authUsecase) UpdateCover(ctx context.Context, userID int64, url string) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	if url == "" {
		return errors.New("cover URL cannot be empty")
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	return u.userRepo.UpdateCover(ctx, userID, url)
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

func (uc *authUsecase) GetUserProfileByUsername(ctx context.Context, currentUserID int64, username string) (*domain.User, error) {
	user, err := uc.userRepo.GetUserProfileByUsername(ctx, currentUserID, username)
	if err != nil {
		return nil, errors.New("lỗi hệ thống khi lấy thông tin người dùng")
	}

	if user == nil {
		return nil, errors.New("không tìm thấy người dùng này")
	}

	// Bảo mật: Ẩn mật khẩu trước khi trả về Frontend
	user.PasswordHash = ""
	return user, nil
}

// Cập nhật username
func (uc *authUsecase) UpdateProfile(ctx context.Context, userID int64, input dto.UpdateProfileInput) error {
	updates := make(map[string]interface{})

	// Chỉ nhét vào map những trường nào THỰC SỰ được Frontend gửi lên
	if input.Username != nil {
		updates["username"] = *input.Username
	}
	if input.Bio != nil {
		updates["bio"] = *input.Bio
	}
	if input.AvatarURL != nil {
		updates["avatar_url"] = *input.AvatarURL
	}

	// Nếu không có gì thay đổi thì bỏ qua luôn, đỡ tốn query DB
	if len(updates) == 0 {
		return nil
	}

	return uc.userRepo.UpdateProfile(ctx, userID, updates)
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

func (uc *authUsecase) SendPasswordOTP(ctx context.Context, email string) error {
	// 1. Kiểm tra xem email có tồn tại không
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("email không tồn tại trong hệ thống")
	}

	// 2. Tạo OTP ngẫu nhiên 6 số
	// *Lưu ý: Để rand.Intn không lặp, ta lấy UnixNano làm seed
	rand.NewSource(time.Now().UnixNano())
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))

	// 3. Lưu vào Redis, tự hủy sau 5 phút
	redisKey := fmt.Sprintf("otp:password_reset:%s", email)
	err = uc.redisClient.Set(ctx, redisKey, otp, 5*time.Minute).Err()
	if err != nil {
		log.Printf("Lỗi lưu OTP vào Redis: %v", err)
		return errors.New("lỗi hệ thống khi tạo mã")
	}

	// 4. Soạn thư và Gửi (Chạy ngầm)
	subject := "🔒 Mã xác thực đổi mật khẩu - Gopher Social"
	content := fmt.Sprintf(`
		<h3>Xin chào %s,</h3>
		<p>Bạn vừa yêu cầu đổi mật khẩu cho tài khoản Gopher Social.</p>
		<p>Mã OTP của bạn là: <b style="font-size:24px; color:blue;">%s</b></p>
		<p><i>Mã này sẽ hết hạn sau 5 phút. Vui lòng không chia sẻ mã này!</i></p>
	`, user.Username, otp)

	go func() {
		// uc.emailSender sẽ tự động lấy cấu hình hệ thống từ struct Config
		err := uc.emailSender.SendEmail(email, subject, content)
		if err != nil {
			log.Printf("Lỗi gửi email cho %s: %v", email, err)
		}
	}()

	return nil
}

// ChangePasswordWithOTP kiểm tra OTP và cập nhật mật khẩu mới
func (uc *authUsecase) ChangePasswordWithOTP(ctx context.Context, email, otp, newPassword string) error {
	redisKey := fmt.Sprintf("otp:password_reset:%s", email)

	// 1. Lấy OTP từ Redis
	storedOTP, err := uc.redisClient.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return errors.New("mã OTP đã hết hạn hoặc không tồn tại")
	} else if err != nil {
		return errors.New("lỗi hệ thống khi xác thực OTP")
	}

	// 2. Đối chiếu
	if storedOTP != otp {
		return errors.New("mã OTP không chính xác")
	}

	// 3. Lấy thông tin User để lấy ID
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New("không tìm thấy người dùng")
	}

	// 4. Băm mật khẩu (Sử dụng hàm HashPassword của cậu đã viết trong package auth)
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return errors.New("lỗi mã hóa mật khẩu")
	}

	// 5. Cập nhật vào DB
	err = uc.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword))
	if err != nil {
		return err
	}

	// 6. Xóa OTP khỏi Redis sau khi đổi xong
	_ = uc.redisClient.Del(ctx, redisKey).Err()

	return nil
}
