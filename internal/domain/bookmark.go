package domain

import (
	"context"
	"time"
)

// Thực thể Bookmark ánh xạ với Database
type Bookmark struct {
	UserID    int64     `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	PostID    int64     `json:"post_id" gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time `json:"created_at"`
}

// Hợp đồng cho tầng Repository (Giao tiếp với DB)
type BookmarkRepository interface {
	ToggleSavePost(ctx context.Context, userID int64, postID int64) (bool, error)
	GetSavedPosts(ctx context.Context, userID int64, limit, offset int) ([]Post, error) // Trả về mảng Post
}

// Hợp đồng cho tầng Usecase (Xử lý nghiệp vụ)
type BookmarkUseCase interface {
	ToggleSavePost(ctx context.Context, userID int64, postID int64) (bool, error)
	GetSavedPosts(ctx context.Context, userID int64, page int) ([]Post, error)
}
