package domain

import (
	"context"
	"time"
)

type Like struct {
	UserID    int64     `json:"user_id" gorm:"primaryKey"`
	PostID    int64     `json:"post_id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	UserID    int64     `json:"user_id"`
	PostID    int64     `json:"post_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	User      *User     `json:"user,omitempty"` // Để hiển thị ai đã comment
}

type InteractionRepository interface {
	// Like logic
	LikePost(ctx context.Context, userID, postID int64) error
	UnlikePost(ctx context.Context, userID, postID int64) error
	IsLiked(ctx context.Context, userID, postID int64) (bool, error)

	// Comment logic
	CreateComment(ctx context.Context, comment *Comment) error
	GetCommentsByPostID(ctx context.Context, postID int64) ([]Comment, error)
}

type InteractionUsecase interface {
	ToggleLike(ctx context.Context, userID, postID int64) (bool, error) // Trả về true nếu là Like, false nếu là Unlike
	CommentPost(ctx context.Context, userID, postID int64, content string) (*Comment, error)
	GetPostComments(ctx context.Context, postID int64) ([]Comment, error)
}
