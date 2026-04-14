package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
)

// XÓA cái interface NotificationHub ở đây đi vì ta dùng domain.NotificationUsecase rồi

type interactionUsecase struct {
	repo   domain.InteractionRepository
	notiUC domain.NotificationUsecase // Dùng cái này đồng bộ
}

func NewInteractionUsecase(repo domain.InteractionRepository, notiUC domain.NotificationUsecase) domain.InteractionUsecase {
	return &interactionUsecase{
		repo:   repo,
		notiUC: notiUC,
	}
}

func (u *interactionUsecase) ToggleLike(ctx context.Context, userID, postID int64) (bool, error) {
	isLiked, err := u.repo.IsLiked(ctx, userID, postID)
	if err != nil {
		return false, err
	}

	if isLiked {
		err = u.repo.UnlikePost(ctx, userID, postID)
		return false, err
	}

	err = u.repo.LikePost(ctx, userID, postID)
	if err != nil {
		return false, err
	}

	// Gửi thông báo real-time khi LIKE
	ownerID := u.repo.GetPostOwner(ctx, postID)
	if userID != ownerID && u.notiUC != nil {
		go func() {
			noti := &domain.Notification{
				UserID:   ownerID,
				ActorID:  userID,
				Type:     "LIKE",
				EntityID: postID,
				Message:  "đã thích bài viết của bạn.",
			}
			_ = u.notiUC.SendNotification(context.Background(), noti)
		}()
	}

	return true, nil
}

func (u *interactionUsecase) CommentPost(ctx context.Context, userID, postID int64, content string) (*domain.Comment, error) {
	comment := &domain.Comment{
		UserID:  userID,
		PostID:  postID,
		Content: content,
	}
	err := u.repo.CreateComment(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Gửi thông báo cho chủ bài viết khi có COMMENT mới
	ownerID := u.repo.GetPostOwner(ctx, postID)
	if userID != ownerID && u.notiUC != nil {
		go func() {
			noti := &domain.Notification{
				UserID:   ownerID,
				ActorID:  userID,
				Type:     "COMMENT",
				EntityID: postID,
				Message:  "đã bình luận về bài viết của bạn.",
			}
			_ = u.notiUC.SendNotification(context.Background(), noti)
		}()
	}

	return comment, nil
}

func (u *interactionUsecase) DeleteComment(ctx context.Context, commentID int64, currentUserID int64) error {
	return u.repo.DeleteComment(ctx, commentID, currentUserID)
}

func (u *interactionUsecase) UpdateComment(ctx context.Context, commentID int64, currentUserID int64, newContent string) error {
	if strings.TrimSpace(newContent) == "" {
		return errors.New("nội dung bình luận không được để trống")
	}
	return u.repo.UpdateComment(ctx, commentID, currentUserID, newContent)
}

func (u *interactionUsecase) GetPostComments(ctx context.Context, postID int64) ([]domain.Comment, error) {
	return u.repo.GetCommentsByPostID(ctx, postID)
}
