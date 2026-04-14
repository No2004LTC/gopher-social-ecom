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

func (u *notificationUsecase) SendNotification(ctx context.Context, noti *domain.Notification) error {
	// 1. Lưu vào Database (Persistence)
	if err := u.repo.Create(ctx, noti); err != nil {
		return err
	}

	// 2. Bắn qua WebSocket (Real-time)
	// Đẩy vào channel Notifications mà cậu đã viết ở Hub
	u.hub.Notifications <- *noti

	return nil
}

func (u *notificationUsecase) GetNotifications(ctx context.Context, userID int64, page int) ([]domain.Notification, error) {
	limit := 20
	offset := (page - 1) * limit
	return u.repo.GetByUserID(ctx, userID, limit, offset)
}

func (u *notificationUsecase) MarkAsRead(ctx context.Context, notiID int64) error {
	// Usecase không trực tiếp sửa DB, nó gọi Repo làm
	return u.repo.MarkAsRead(ctx, notiID)
}

func (u *notificationUsecase) GetUserNotifications(ctx context.Context, userID int64, limit, offset int) ([]dto.NotificationResponse, error) {
	return u.repo.GetUserNotifications(ctx, userID, limit, offset)
}
