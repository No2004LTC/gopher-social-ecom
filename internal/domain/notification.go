package domain

import (
	"context"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
)

type Notification struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	UserID    int64     `json:"user_id"`
	ActorID   int64     `json:"actor_id"`
	Type      string    `json:"type"`
	EntityID  int64     `json:"entity_id"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	Actor     *User     `json:"actor,omitempty" gorm:"foreignKey:ActorID"`
}

// NotificationRepository
type NotificationRepository interface {
	Create(ctx context.Context, noti *Notification) error
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]Notification, error)
	MarkAsRead(ctx context.Context, notiID int64) error
	GetUserNotifications(ctx context.Context, userID int64, limit, offset int) ([]dto.NotificationResponse, error)
	GetUnreadCount(ctx context.Context, userID int64) (int, error)
	MarkAllAsRead(ctx context.Context, userID int64) error
}

// NotificationUsecase
type NotificationUsecase interface {
	SendNotification(ctx context.Context, noti *Notification) error
	GetNotifications(ctx context.Context, userID int64, page int) ([]Notification, error)
	MarkAsRead(ctx context.Context, notiID int64) error
	GetUserNotifications(ctx context.Context, userID int64, limit, offset int) ([]dto.NotificationResponse, error)
	GetUnreadCount(ctx context.Context, userID int64) (int, error)
	MarkAllAsRead(ctx context.Context, userID int64) error
}
