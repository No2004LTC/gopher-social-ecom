package domain

import (
	"context"
	"time"
)

// Bookmark
type Bookmark struct {
	UserID    int64     `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	PostID    int64     `json:"post_id" gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time `json:"created_at"`
}

// BookmarkRepository
type BookmarkRepository interface {
	ToggleSavePost(ctx context.Context, userID int64, postID int64) (bool, error)
	GetSavedPosts(ctx context.Context, userID int64, limit, offset int) ([]Post, error) // Trả về mảng Post
}

// BookmarkUseCase
type BookmarkUseCase interface {
	ToggleSavePost(ctx context.Context, userID int64, postID int64) (bool, error)
	GetSavedPosts(ctx context.Context, userID int64, page int) ([]Post, error)
}
