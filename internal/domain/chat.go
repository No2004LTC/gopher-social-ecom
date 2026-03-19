package domain

import (
	"context"
	"time"
)

type Message struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	FromUserID int64     `json:"from_user_id"`
	ToUserID   int64     `json:"to_user_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type ChatRepository interface {
	SaveMessage(ctx context.Context, msg *Message) error
	GetHistory(ctx context.Context, user1, user2 int64, limit int) ([]Message, error)
}

type ChatUsecase interface {
	SaveMessage(ctx context.Context, msg *Message) error
	GetChatHistory(ctx context.Context, user1, user2 int64, limit int) ([]Message, error)
}
