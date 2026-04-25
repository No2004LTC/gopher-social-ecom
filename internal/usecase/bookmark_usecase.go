package usecase

import (
	"context"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
)

type bookmarkUseCase struct {
	bookmarkRepo domain.BookmarkRepository
}

func NewBookmarkUseCase(br domain.BookmarkRepository) domain.BookmarkUseCase {
	return &bookmarkUseCase{bookmarkRepo: br}
}

// ToggleSavePost
func (u *bookmarkUseCase) ToggleSavePost(ctx context.Context, userID int64, postID int64) (bool, error) {
	return u.bookmarkRepo.ToggleSavePost(ctx, userID, postID)
}

// GetSavedPosts
func (u *bookmarkUseCase) GetSavedPosts(ctx context.Context, userID int64, page int) ([]domain.Post, error) {
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	return u.bookmarkRepo.GetSavedPosts(ctx, userID, limit, offset)
}
