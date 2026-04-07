package domain

import (
	"context"
	"time"
)

type Notification struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	UserID    int64     `json:"user_id"`   // Người nhận
	ActorID   int64     `json:"actor_id"`  // Người thực hiện (ví dụ: người Like)
	Type      string    `json:"type"`      // FOLLOW, LIKE, NEW_POST...
	EntityID  int64     `json:"entity_id"` // ID của Post hoặc Comment liên quan
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	Actor     *User     `json:"actor,omitempty" gorm:"foreignKey:ActorID"`
}

type NotificationRepository interface {
	Create(ctx context.Context, noti *Notification) error
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]Notification, error)
	MarkAsRead(ctx context.Context, notiID int64) error
}

type NotificationUsecase interface {
	SendNotification(ctx context.Context, noti *Notification) error
	GetNotifications(ctx context.Context, userID int64, page int) ([]Notification, error)
	MarkAsRead(ctx context.Context, notiID int64) error
}
