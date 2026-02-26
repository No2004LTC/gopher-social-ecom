package domain

import (
	"context"
	"time"
)

// User thực thể người dùng
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Không bao giờ trả về mật khẩu trong JSON(marshal or unmarshal)
	AvatarURL    string    `json:"avatar_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRepository - Hợp đồng cho tầng lưu trữ dữ liệu
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
}

// UserUsecase - Hợp đồng cho tầng xử lý nghiệp vụ
type UserUsecase interface {
	Register(ctx context.Context, username, email, password string) error
	Login(ctx context.Context, email, password string) (string, error) // Trả về JWT Token
}
