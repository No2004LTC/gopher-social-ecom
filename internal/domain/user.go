package domain

import (
	"context"
	"time"
)

// User thực thể người dùng
type User struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"column:username" json:"username"`
	Email        string    `gorm:"column:email;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	AvatarURL    string    `gorm:"column:avatar_url" json:"avatar_url"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// UserRepository - Hợp đồng cho tầng lưu trữ dữ liệu
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error
	Update(ctx context.Context, user *User) error
}

// UserUsecase - Hợp đồng cho tầng xử lý nghiệp vụ
type UserUsecase interface {
	Register(ctx context.Context, username, email, password string) error
	Login(ctx context.Context, email, password string) (string, error) // Trả về JWT Token
	UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error
	GetProfile(ctx context.Context, userID int64) (*User, error)
	UpdateProfile(ctx context.Context, userID int64, username string) error
}
