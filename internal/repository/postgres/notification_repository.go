package postgres

import (
	"context"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) domain.NotificationRepository {
	return &notificationRepository{db: db}
}

// Create
func (r *notificationRepository) Create(ctx context.Context, noti *domain.Notification) error {
	return r.db.WithContext(ctx).Create(noti).Error
}

// GetByUserID
func (r *notificationRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.WithContext(ctx).
		Preload("Actor").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// MarkAsRead
func (r *notificationRepository) MarkAsRead(ctx context.Context, notiID int64) error {
	return r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("id = ?", notiID).Update("is_read", true).Error
}

// GetUserNotifications
func (r *notificationRepository) GetUserNotifications(ctx context.Context, userID int64, limit, offset int) ([]dto.NotificationResponse, error) {
	var notis []domain.Notification
	err := r.db.WithContext(ctx).
		Preload("Actor").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notis).Error

	if err != nil {
		return nil, err
	}

	var responses []dto.NotificationResponse
	for _, n := range notis {
		var actorCompact dto.ActorCompact

		if n.Actor != nil {
			actorCompact = dto.ActorCompact{
				ID:        n.Actor.ID,
				Username:  n.Actor.Username,
				AvatarURL: n.Actor.AvatarURL,
			}
		}

		responses = append(responses, dto.NotificationResponse{
			ID:        n.ID,
			Type:      n.Type,
			Message:   n.Message,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
			Actor:     actorCompact,
		})
	}

	return responses, nil
}

// GetUnreadCount
func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error

	return int(count), err
}

// MarkAllAsRead
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}
