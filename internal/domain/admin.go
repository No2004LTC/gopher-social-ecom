package domain

import (
	"context"

	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
)

type AdminUsecase interface {
	GetDashboardStats(ctx context.Context) (*dto.AdminDashboardStatsResponse, error)
	GetGrowthStats(ctx context.Context, days int) ([]dto.GrowthPoint, error)
	GetAllUsers(ctx context.Context, keyword string) ([]*User, error)
	BanUser(ctx context.Context, targetUserID int64, adminEmail string) error
	UnbanUser(ctx context.Context, targetUserID int64) error
	GetModerationFeed(ctx context.Context, limit int) ([]*Post, error)
	AdminDeletePost(ctx context.Context, postID int64) error
}

type AdminRepository interface {
	CountUsers(ctx context.Context) (int64, error)
	CountPosts(ctx context.Context) (int64, error)
	GetDailyGrowth(ctx context.Context, days int) ([]dto.GrowthPoint, error)
	CountBannedUsers(ctx context.Context) (int64, error)
	GetAllUsers(ctx context.Context, keyword string) ([]*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	GetLatestPosts(ctx context.Context, limit int) ([]*Post, error)
	DeletePost(ctx context.Context, postID int64) error
}
