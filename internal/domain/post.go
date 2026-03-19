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

	// Quan hệ: Giúp GORM lấy thông tin User khi hiển thị Newsfeed
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// Các trường bổ sung (Computed fields)
	LikesCount    int64 `json:"likes_count" gorm:"-"`
	CommentsCount int64 `json:"comments_count" gorm:"-"`
	IsLiked       bool  `json:"is_liked" gorm:"-"`
}

type PostRepository interface {
	Create(ctx context.Context, post *Post) error
	GetList(ctx context.Context, offset, limit int, currentUserID int64) ([]Post, error)
}

type PostUsecase interface {
	CreatePost(ctx context.Context, post *Post, file *multipart.FileHeader) error
	GetFeed(ctx context.Context, page, limit int, currentUserID int64) ([]Post, error)
}
