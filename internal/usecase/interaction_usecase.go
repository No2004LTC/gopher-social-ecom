package usecase

import (
	"context"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
)

type interactionUsecase struct {
	repo domain.InteractionRepository
}

func NewInteractionUsecase(repo domain.InteractionRepository) domain.InteractionUsecase {
	return &interactionUsecase{repo: repo}
}

func (u *interactionUsecase) ToggleLike(ctx context.Context, userID, postID int64) (bool, error) {
	liked, err := u.repo.IsLiked(ctx, userID, postID)
	if err != nil {
		return false, err
	}

	if liked {
		// Đã like rồi thì Unlike
		err = u.repo.UnlikePost(ctx, userID, postID)
		return false, err
	}

	// Chưa like thì Like
	err = u.repo.LikePost(ctx, userID, postID)
	return true, err
}

func (u *interactionUsecase) CommentPost(ctx context.Context, userID, postID int64, content string) (*domain.Comment, error) {
	comment := &domain.Comment{
		UserID:  userID,
		PostID:  postID,
		Content: content,
	}
	err := u.repo.CreateComment(ctx, comment)
	return comment, err
}

func (u *interactionUsecase) GetPostComments(ctx context.Context, postID int64) ([]domain.Comment, error) {
	return u.repo.GetCommentsByPostID(ctx, postID)
}
