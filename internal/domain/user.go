package domain

import (
	"context"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
)

// User (Model chính)
type User struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"column:username" json:"username"`
	Email        string    `gorm:"column:email;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	AvatarURL    string    `gorm:"column:avatar_url" json:"avatar_url"`
	CoverURL     string    `gorm:"column:cover_url" json:"cover_url"`
	Bio          string    `json:"bio"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Các trường ảo dùng để map dữ liệu Count từ Repository/Usecase
	FollowersCount int  `json:"followers_count" gorm:"-"`
	FollowingCount int  `json:"following_count" gorm:"-"`
	PostsCount     int  `json:"posts_count" gorm:"-"`
	IsFollowing    bool `json:"is_following" gorm:"->"`
}

type SuggestedUser struct {
	ID                 int64  `json:"id"`
	Username           string `json:"username"`
	AvatarURL          string `json:"avatar_url"`
	MutualFriendsCount int    `json:"mutual_friends_count"`
}

// UserRepository: Chỉ chứa các hàm về thông tin người dùng
type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetUserProfileByUsername(ctx context.Context, currentUserID int64, username string) (*User, error)
	UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error
	UpdateCover(ctx context.Context, userID int64, coverURL string) error
	UpdateProfile(ctx context.Context, userID int64, updates map[string]interface{}) error
	SearchUsers(ctx context.Context, currentUserID int64, query string, limit, offset int) ([]dto.UserCompact, error)
	GetFollowing(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error)
	GetFollowers(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error)
	GetSuggestedUsers(ctx context.Context, userID int64, limit int) ([]SuggestedUser, error)
}

// UserUsecase: Business logic về Profile/Mạng xã hội
type UserUsecase interface {
	GetProfile(ctx context.Context, userID int64) (*User, error)
	GetUserProfileByUsername(ctx context.Context, currentUserID int64, username string) (*dto.UserProfileResponse, error)
	UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error
	UpdateCover(ctx context.Context, userID int64, coverURL string) error
	UpdateProfile(ctx context.Context, userID int64, input dto.UpdateProfileInput) error
	SearchUsers(ctx context.Context, currentUserID int64, query string, limit, offset int) ([]dto.UserCompact, error)
	GetFollowing(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error)
	GetFollowers(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error)
	GetFriendSuggestions(ctx context.Context, userID int64) ([]SuggestedUser, error)
	GetOnlineContacts(ctx context.Context, userID int64) ([]dto.UserCompact, error)
}
