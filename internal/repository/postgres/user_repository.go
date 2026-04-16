package postgres

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"gorm.io/gorm"
)

// Tạo struct connect DB
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository khởi tạo một instance của userRepository với kết nối DB đã được thiết lập(contructor kiểu vậy)
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create thực hiện lưu một User mới vào bảng users
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	// GORM sẽ tự động mapping struct User với bảng users trong DB
	// .WithContext(ctx) cực kỳ quan trọng để xử lý timeout/cancel request
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByEmail tìm kiếm User dựa trên Email (dùng cho đăng nhập hoặc kiểm tra trùng)
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	// Sử dụng Limit(1).Find() thay vì First() sẽ không bắn ra log "record not found" nếu trống
	err := r.db.WithContext(ctx).Where("email = ?", email).Limit(1).Find(&user).Error
	if err != nil {
		return nil, err
	}

	// Kiểm tra nếu ID bằng 0 nghĩa là không tìm thấy bản ghi nào
	if user.ID == 0 {
		return nil, nil
	}

	return &user, nil
}

// GetByID tìm kiếm User dựa trên ID (dùng cho các tác vụ cần thông tin User sau khi đăng nhập)
func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).First(&user, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &user, nil
}

// Login tìm user bằng định danh (email hoặc username)
func (r *userRepository) GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error) {
	var user domain.User

	// Tìm kiếm linh hoạt: chấp nhận cả Email hoặc Username
	err := r.db.WithContext(ctx).
		Where("email = ? OR username = ?", identifier, identifier).
		First(&user).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserProfileByUsername lấy thông tin profile và check xem currentUserID có đang follow người này không
func (r *userRepository) GetUserProfileByUsername(ctx context.Context, currentUserID int64, username string) (*domain.User, error) {
	var user domain.User

	// Câu SQL lấy thông tin user kèm trạng thái Follow
	query := `
        *,
        EXISTS (
            SELECT 1 FROM follows 
            WHERE follower_id = ? AND following_id = users.id
        ) as is_following
    `

	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Select(query, currentUserID).
		Where("username = ?", username).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Trả về nil nếu không tìm thấy, không đánh lỗi hệ thống
		}
		return nil, err
	}

	return &user, nil
}

// UpdateAvatar cập nhật avatar URL cho user
func (r *userRepository) UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error {
	log.Printf("[UpdateAvatar] Updating avatar for user ID: %d, URL: %s\n", userID, avatarURL)

	// Use Save() with partial update - more reliable for GORM
	user := &domain.User{ID: userID}

	// First check if user exists
	if err := r.db.WithContext(ctx).First(user, userID).Error; err != nil {
		log.Printf("[UpdateAvatar] User not found: %v\n", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("no user found with that ID")
		}
		return err
	}

	// Update the avatar URL
	user.AvatarURL = avatarURL
	result := r.db.WithContext(ctx).Model(user).Update("avatar_url", avatarURL)

	log.Printf("[UpdateAvatar] GORM Result - RowsAffected: %d, Error: %v\n", result.RowsAffected, result.Error)

	if result.Error != nil {
		log.Printf("[UpdateAvatar] ERROR: %v\n", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		log.Printf("[UpdateAvatar] No rows affected - user ID %d\n", userID)
		return errors.New("failed to update user avatar")
	}

	log.Printf("[UpdateAvatar] Successfully updated avatar for user ID: %d\n", userID)
	return nil
}

// UpdateCover cập nhật ảnh bìa (cover URL) cho user
func (r *userRepository) UpdateCover(ctx context.Context, userID int64, coverURL string) error {
	log.Printf("[UpdateCover] Updating cover for user ID: %d, URL: %s\n", userID, coverURL)

	user := &domain.User{ID: userID}

	if err := r.db.WithContext(ctx).First(user, userID).Error; err != nil {
		log.Printf("[UpdateCover] User not found: %v\n", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("no user found with that ID")
		}
		return err
	}

	user.CoverURL = coverURL
	result := r.db.WithContext(ctx).Model(user).Update("cover_url", coverURL)

	if result.Error != nil {
		log.Printf("[UpdateCover] ERROR: %v\n", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("failed to update user cover")
	}

	log.Printf("[UpdateCover] Successfully updated cover for user ID: %d\n", userID)
	return nil
}

// update thong tin profile user
func (r *userRepository) UpdateProfile(ctx context.Context, userID int64, updates map[string]interface{}) error {
	// GORM tự động cập nhật updated_at, nhưng ta cứ thêm cho chắc
	updates["updated_at"] = time.Now()

	result := r.db.WithContext(ctx).
		Table("users").
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("không tìm thấy người dùng để cập nhật")
	}
	return nil
}

// Tìm kiếm người dùng theo username
func (r *userRepository) SearchUsers(ctx context.Context, currentUserID int64, query string, limit, offset int) ([]dto.UserCompact, error) {
	var users []dto.UserCompact // 1. Đổi kiểu dữ liệu sang DTO

	// 2. Câu SQL "ma thuật" ver 2.0:
	// Thêm trường kiểm tra xem họ có đang follow mình không (is_followed_by)
	selectQuery := `
        users.id, 
        users.username, 
        users.avatar_url,
        EXISTS (
            SELECT 1 FROM follows 
            WHERE follower_id = ? AND following_id = users.id
        ) as is_following,
        EXISTS (
            SELECT 1 FROM follows 
            WHERE follower_id = users.id AND following_id = ?
        ) as is_followed_by`
	// 3. Thực thi Query
	err := r.db.WithContext(ctx).
		Table("users").                                    // Chỉ định rõ bảng vì ta đang dùng custom struct
		Select(selectQuery, currentUserID, currentUserID). // Truyền currentUserID 2 lần cho 2 dấu ?
		Where("(username ILIKE ? OR email ILIKE ?) AND id <> ?", "%"+query+"%", "%"+query+"%", currentUserID).
		Limit(limit).
		Offset(offset).
		Scan(&users).Error // 4. Dùng Scan thay vì Find khi mapping vào DTO không phải là Model chuẩn của GORM

	return users, err
}

func (r *userRepository) GetFollowing(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error) {
	var users []dto.UserCompact

	err := r.db.WithContext(ctx).
		Table("users").
		Select(`
			users.id, 
			users.username, 
			users.avatar_url,
			true as is_following,
			EXISTS (SELECT 1 FROM follows f2 WHERE f2.follower_id = users.id AND f2.following_id = ?) as is_followed_by
		`, currentUserID).
		Joins("JOIN follows f1 ON f1.following_id = users.id").
		Where("f1.follower_id = ?", currentUserID).
		Limit(limit).Offset(offset).
		Scan(&users).Error

	return users, err
}

func (r *userRepository) GetFollowers(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error) {
	var users []dto.UserCompact

	err := r.db.WithContext(ctx).
		Table("users").
		Select(`
			users.id, 
			users.username, 
			users.avatar_url,
			EXISTS (SELECT 1 FROM follows f2 WHERE f2.follower_id = ? AND f2.following_id = users.id) as is_following,
			true as is_followed_by
		`, currentUserID).
		Joins("JOIN follows f1 ON f1.follower_id = users.id").
		Where("f1.following_id = ?", currentUserID).
		Limit(limit).Offset(offset).
		Scan(&users).Error

	return users, err
}

func (r *userRepository) GetSuggestedUsers(ctx context.Context, myUserID int64, limit int) ([]domain.SuggestedUser, error) {
	var suggestions []domain.SuggestedUser

	// 🧠 GIẢI THÍCH THUẬT TOÁN FOLLOW:
	// - Lấy u.id KHÁC myUserID (Không tự gợi ý chính mình)
	// - Lấy u.id CHƯA CÓ TRONG danh sách đang theo dõi (following_id) của myUserID
	// - Subquery đếm số người chung (mutual): Đếm số lượng những người mà (myUserID đang theo dõi) VÀ họ (đang theo dõi u.id)
	query := `
		SELECT 
			u.id, 
			u.username, 
			u.avatar_url,
			(
				SELECT COUNT(*) 
				FROM follows f1
				INNER JOIN follows f2 ON f1.following_id = f2.follower_id
				WHERE f1.follower_id = ? 
				  AND f2.following_id = u.id
			) AS mutual_friends_count
		FROM users u
		WHERE u.id != ?
		  AND u.id NOT IN (
			  SELECT following_id FROM follows WHERE follower_id = ?
		  )
		ORDER BY mutual_friends_count DESC, u.created_at DESC
		LIMIT ?;
	`

	// Truyền 4 tham số: myUserID (cho Subquery), myUserID (loại trừ bản thân), myUserID (loại trừ người đã follow), limit
	err := r.db.WithContext(ctx).Raw(query, myUserID, myUserID, myUserID, limit).Scan(&suggestions).Error
	if err != nil {
		return nil, err
	}

	return suggestions, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID int64, newPasswordHash string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("password_hash", newPasswordHash)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("không tìm thấy người dùng để cập nhật mật khẩu")
	}

	return nil
}
