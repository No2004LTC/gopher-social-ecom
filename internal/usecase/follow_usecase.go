package usecase

import (
	"context"
	"errors"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
)

type followUsecase struct {
	repo domain.FollowRepository
	hub  NotificationHub // Sử dụng lại interface NotificationHub đã có
}

func NewFollowUsecase(repo domain.FollowRepository, hub NotificationHub) domain.FollowUsecase {
	return &followUsecase{
		repo: repo,
		hub:  hub,
	}
}

func (u *followUsecase) FollowUser(ctx context.Context, followerID, followingID int64) error {
	if followerID == followingID {
		return errors.New("cannot follow yourself")
	}

	err := u.repo.Follow(ctx, followerID, followingID)
	if err != nil {
		return err
	}

	// Bắn thông báo Real-time cho người được follow
	if u.hub != nil {
		go u.hub.BroadcastNotification(domain.Notification{
			UserID:   followingID,
			ActorID:  followerID,
			Type:     "FOLLOW",
			EntityID: followerID,
		})
	}

	return nil
}

func (u *followUsecase) UnfollowUser(ctx context.Context, followerID, followingID int64) error {
	return u.repo.Unfollow(ctx, followerID, followingID)
}

func (u *followUsecase) GetFollowingList(ctx context.Context, userID int64) ([]int64, error) {
	return u.repo.GetFollowingIDs(ctx, userID)
}
