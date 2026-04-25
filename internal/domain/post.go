package domain

import (
	"context"
	"mime/multipart"
	"time"
)

type Post struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	LikesCount    int64 `json:"likes_count" gorm:"->"`
	CommentsCount int64 `json:"comments_count" gorm:"->"`
	IsLiked       bool  `json:"is_liked" gorm:"->"`
	IsSaved       bool  `json:"is_saved" gorm:"->"`
}

// PostRepository
type PostRepository interface {
	Create(ctx context.Context, post *Post) error
	DeletePost(ctx context.Context, postID int64, currentUserID int64) error
	UpdatePost(ctx context.Context, postID int64, currentUserID int64, newContent string) error
	GetPosts(ctx context.Context, currentUserID int64, targetUserID int64, limit, offset int) ([]Post, error)
	CountPosts(ctx context.Context, userID int64) (int64, error)
}

// PostUsecase
type PostUsecase interface {
	CreatePost(ctx context.Context, post *Post, file *multipart.FileHeader) error
	DeletePost(ctx context.Context, postID int64, currentUserID int64) error
	UpdatePost(ctx context.Context, postID int64, currentUserID int64, newContent string) error
	GetPosts(ctx context.Context, currentUserID int64, targetUserID int64, page, limit int) ([]Post, error)
}
