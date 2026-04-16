package postgres

import (
	"context"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"gorm.io/gorm"
)

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) domain.ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *chatRepository) GetHistory(ctx context.Context, user1, user2 int64, limit int) ([]domain.Message, error) {
	var msgs []domain.Message
	err := r.db.WithContext(ctx).
		Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)", user1, user2, user2, user1).
		Order("created_at desc").
		Limit(limit).
		Find(&msgs).Error
	return msgs, err
}

func (r *chatRepository) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Message{}).
		Where("to_user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return int(count), err
}

func (r *chatRepository) GetConversations(ctx context.Context, userID int64) ([]dto.Conversation, error) {
	var conversations []dto.Conversation

	query := `
       WITH UserMessages AS (
          SELECT 
             CASE WHEN from_user_id = ? THEN to_user_id ELSE from_user_id END AS partner_id,
             content,
             created_at
          FROM messages
          WHERE from_user_id = ? OR to_user_id = ?
       ),
       LatestMessages AS (
          SELECT DISTINCT ON (partner_id)
             partner_id,
             content AS last_message,
             created_at AS last_message_at
          FROM UserMessages
          ORDER BY partner_id, created_at DESC
       ),
       UnreadCounts AS (
          SELECT from_user_id AS partner_id, COUNT(id) AS unread_count
          FROM messages
          WHERE to_user_id = ? AND is_read = false
          GROUP BY from_user_id
       )
       SELECT 
          lm.partner_id, 
          COALESCE(u.username, 'Người dùng hệ thống') as partner_username, 
          COALESCE(u.avatar_url, '') as partner_avatar_url,
          lm.last_message, 
          lm.last_message_at,
          COALESCE(uc.unread_count, 0) AS unread_count
       FROM LatestMessages lm
       -- 👉 ĐỔI THÀNH LEFT JOIN: Để nếu User có bị lỗi data thì tin nhắn vẫn hiện ra
       LEFT JOIN users u ON u.id = lm.partner_id
       LEFT JOIN UnreadCounts uc ON uc.partner_id = lm.partner_id
       ORDER BY lm.last_message_at DESC;
    `

	err := r.db.WithContext(ctx).Raw(query, userID, userID, userID, userID).Scan(&conversations).Error
	return conversations, err
}

func (r *chatRepository) MarkMessagesAsRead(ctx context.Context, myUserID, partnerID int64) error {
	// Chỉ update những tin nhắn GỬI ĐẾN mình (to_user_id) TỪ người đó (from_user_id)
	result := r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Where("to_user_id = ? AND from_user_id = ? AND is_read = false", myUserID, partnerID).
		Update("is_read", true)

	return result.Error
}
