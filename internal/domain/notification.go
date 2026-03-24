package domain

import (
	"context"
	"time"
)

type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`   // Người nhận thông báo
	ActorID   int64     `json:"actor_id"`  // Người gây ra hành động (người like/comment)
	Type      string    `json:"type"`      // "LIKE", "COMMENT", "NEW_FOLLOWER"
	EntityID  int64     `json:"entity_id"` // ID của bài viết hoặc comment liên quan
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
	// Bạn có thể thêm ActorName hoặc ActorAvatar để hiển thị luôn trên App
}

type NotificationUsecase interface {
	Create(ctx context.Context, noti *Notification) error
	GetList(ctx context.Context, userID int64) ([]Notification, error)
	MarkAsRead(ctx context.Context, notiID int64) error
}
