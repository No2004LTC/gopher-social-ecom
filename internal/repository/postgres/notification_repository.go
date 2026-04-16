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

func (r *notificationRepository) GetUserNotifications(ctx context.Context, userID int64, limit, offset int) ([]dto.NotificationResponse, error) {
	// 1. Lấy dữ liệu nguyên bản từ bảng Notifications
	var notis []domain.Notification
	err := r.db.WithContext(ctx).
		Preload("Actor"). // "Phép thuật" của GORM: Tự động móc nối lấy data user nhét vào trường Actor
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notis).Error

	if err != nil {
		return nil, err
	}

	// 2. Chuyển đổi (Mapping) sang DTO
	// Cách này đảm bảo 100% dữ liệu không bao giờ bị thất lạc
	var responses []dto.NotificationResponse
	for _, n := range notis {
		var actorCompact dto.ActorCompact

		// Tránh lỗi Nil Pointer (Panic) lỡ có noti nào không có người gửi
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
			Actor:     actorCompact, // Nhét cục Actor đã chuẩn bị vào đây
		})
	}

	return responses, nil
}

// GetUnreadCount lấy tổng số thông báo chưa đọc của 1 user
func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	var count int64
	// Đếm số dòng có user_id = mình và is_read = false
	err := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error

	return int(count), err
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID int64) error {
	// Tìm tất cả thông báo CỦA MÌNH đang chưa đọc và update thành true
	return r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}
