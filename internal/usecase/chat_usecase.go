package usecase

import (
	"context"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
)

type chatUsecase struct {
	repo domain.ChatRepository
}

func NewChatUsecase(repo domain.ChatRepository) domain.ChatUsecase {
	return &chatUsecase{repo: repo}
}

func (u *chatUsecase) SaveMessage(ctx context.Context, msg *domain.Message) error {
	// Ở đây cậu có thể thêm logic như: check xem 2 user có phải là bạn bè không,
	// hoặc filter những từ ngữ nhạy cảm trước khi lưu.
	return u.repo.SaveMessage(ctx, msg)
}

func (u *chatUsecase) GetChatHistory(ctx context.Context, user1, user2 int64, limit int) ([]domain.Message, error) {
	return u.repo.GetHistory(ctx, user1, user2, limit)
}
