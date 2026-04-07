package postgres

import (
	"context"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) domain.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, noti *domain.Notification) error {
	return r.db.WithContext(ctx).Create(noti).Error
}

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

func (r *notificationRepository) MarkAsRead(ctx context.Context, notiID int64) error {
	return r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("id = ?", notiID).Update("is_read", true).Error
}
