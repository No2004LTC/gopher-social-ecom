package postgres

import (
	"context"
	"errors"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository khởi tạo một instance của userRepository
// Nó trả về interface domain.UserRepository để đảm bảo tính trừu tượng
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
