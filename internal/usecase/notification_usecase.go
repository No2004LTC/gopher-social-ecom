package usecase

import (
	"context"

	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/ws"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
)

type notificationUsecase struct {
	repo domain.NotificationRepository
	hub  *ws.Hub
}

func NewNotificationUsecase(repo domain.NotificationRepository, hub *ws.Hub) domain.NotificationUsecase {
	return &notificationUsecase{
		repo: repo,
		hub:  hub,
	}
}

// SendNotification
func (u *notificationUsecase) SendNotification(ctx context.Context, noti *domain.Notification) error {
	err := u.repo.Create(ctx, noti)
	if err != nil {
		return err
	}
	u.hub.BroadcastNotification(*noti)

	return nil
}

// GetNotifications
func (u *notificationUsecase) GetNotifications(ctx context.Context, userID int64, page int) ([]domain.Notification, error) {
	limit := 20
	offset := (page - 1) * limit
	return u.repo.GetByUserID(ctx, userID, limit, offset)
}

// MarkAsRead
func (u *notificationUsecase) MarkAsRead(ctx context.Context, notiID int64) error {

	return u.repo.MarkAsRead(ctx, notiID)
}

// GetUserNotifications
func (u *notificationUsecase) GetUserNotifications(ctx context.Context, userID int64, limit, offset int) ([]dto.NotificationResponse, error) {
	return u.repo.GetUserNotifications(ctx, userID, limit, offset)
}

// GetUnreadCount
func (uc *notificationUsecase) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	return uc.repo.GetUnreadCount(ctx, userID)
}

// MarkAllAsRead
func (uc *notificationUsecase) MarkAllAsRead(ctx context.Context, userID int64) error {
	return uc.repo.MarkAllAsRead(ctx, userID)
}
