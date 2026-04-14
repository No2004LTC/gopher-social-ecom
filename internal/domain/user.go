package domain

import (
	"context"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
)

// User (model)
type User struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"column:username" json:"username"`
	Email        string    `gorm:"column:email;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	AvatarURL    string    `gorm:"column:avatar_url" json:"avatar_url"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	// -> để GORM cho phép "đọc" từ DB vào nhưng không "ghi" xuống DB
	IsFollowing bool `json:"is_following" gorm:"->"`
}

// UserRepository (Các hàm tôi tạo để truy vấn)
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetUserByIdentifier(ctx context.Context, identifier string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error
	Update(ctx context.Context, user *User) error
	SearchUsers(ctx context.Context, currentUserID int64, query string, limit, offset int) ([]dto.UserCompact, error)
	GetFollowing(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error)
	GetFollowers(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error)
}

// UserUsecase (Chứa business logic)
type UserUsecase interface {
	Register(ctx context.Context, username, email, password string) error
	Login(ctx context.Context, identifier, password string) (string, *User, error) // Trả về JWT Token
	UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error
	GetProfile(ctx context.Context, userID int64) (*User, error)
	UpdateProfile(ctx context.Context, userID int64, username string) error
	SearchUsers(ctx context.Context, currentUserID int64, query string, limit, offset int) ([]dto.UserCompact, error)
	GetFollowing(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error)
	GetFollowers(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error)
}
