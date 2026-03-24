package usecase

import (
	"context"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
)

// Interface này giúp usecase gọi được Hub mà không cần import package ws
type NotificationHub interface {
	BroadcastNotification(noti domain.Notification)
}

type interactionUsecase struct {
	repo domain.InteractionRepository
	hub  NotificationHub
}

func NewInteractionUsecase(repo domain.InteractionRepository, hub NotificationHub) domain.InteractionUsecase {
	return &interactionUsecase{
		repo: repo,
		hub:  hub,
	}
}

func (u *interactionUsecase) ToggleLike(ctx context.Context, userID, postID int64) (bool, error) {
	// 1. Kiểm tra xem đã like chưa
	isLiked, err := u.repo.IsLiked(ctx, userID, postID)
	if err != nil {
		return false, err
	}

	if isLiked {
		// 2. Nếu đã like rồi thì Unlike
		err = u.repo.UnlikePost(ctx, userID, postID)
		return false, err
	}

	// 3. Nếu chưa like thì Like
	err = u.repo.LikePost(ctx, userID, postID)
	if err != nil {
		return false, err
	}

	// 4. Gửi thông báo real-time
	ownerID := u.repo.GetPostOwner(ctx, postID)
	//log.Printf("DEBUG: UserID=%d, OwnerID=%d", userID, ownerID)
	if userID != ownerID && u.hub != nil {
		// Chạy ngầm việc gửi thông báo để API Like trả về ngay lập tức
		go func(n domain.Notification) {
			u.hub.BroadcastNotification(n)
		}(domain.Notification{
			UserID:   ownerID,
			ActorID:  userID,
			Type:     "LIKE",
			EntityID: postID,
		})
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

	// Gửi thông báo cho chủ bài viết khi có comment mới
	ownerID := u.repo.GetPostOwner(ctx, postID)
	if userID != ownerID && u.hub != nil {
		u.hub.BroadcastNotification(domain.Notification{
			UserID:   ownerID,
			ActorID:  userID,
			Type:     "COMMENT",
			EntityID: postID,
		})
	}

	return comment, nil
}

func (u *interactionUsecase) GetPostComments(ctx context.Context, postID int64) ([]domain.Comment, error) {
	return u.repo.GetCommentsByPostID(ctx, postID)
}
