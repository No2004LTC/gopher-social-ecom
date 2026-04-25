package domain

import (
	"context"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
)

type Message struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	FromUserID int64     `json:"from_user_id"`
	ToUserID   int64     `json:"to_user_id"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`
}

// ChatRepository
type ChatRepository interface {
	SaveMessage(ctx context.Context, msg *Message) error
	GetHistory(ctx context.Context, user1, user2 int64, limit int) ([]Message, error)
	GetUnreadCount(ctx context.Context, userID int64) (int, error)
	GetConversations(ctx context.Context, userID int64) ([]dto.Conversation, error)
	MarkMessagesAsRead(ctx context.Context, myUserID, partnerID int64) error
}

// ChatUsecase
type ChatUsecase interface {
	SaveMessage(ctx context.Context, msg *Message) error
	GetChatHistory(ctx context.Context, user1, user2 int64, limit int) ([]Message, error)
	GetUnreadCount(ctx context.Context, userID int64) (int, error)
	GetCategorizedConversations(ctx context.Context, userID int64) (map[string][]dto.Conversation, error)
	MarkMessagesAsRead(ctx context.Context, myUserID, partnerID int64) error
}
