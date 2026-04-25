package domain

import "context"

// AuthUsecase
type AuthUsecase interface {
	Register(ctx context.Context, username, email, password string) error
	Login(ctx context.Context, email, password string) (string, *User, error)
	SendPasswordOTP(ctx context.Context, email string) error
	ChangePasswordWithOTP(ctx context.Context, email, otp, newPassword string) error
}

// AuthRepository
type AuthRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetUserByIdentifier(ctx context.Context, identifier string) (*User, error)
	UpdatePassword(ctx context.Context, userID int64, newPasswordHash string) error
}
