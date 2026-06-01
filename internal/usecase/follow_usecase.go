package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/redis/go-redis/v9"
)

type followUsecase struct {
	repo   domain.FollowRepository
	notiUC domain.NotificationUsecase
	rdb    *redis.Client
}

func NewFollowUsecase(repo domain.FollowRepository, notiUC domain.NotificationUsecase, rdb *redis.Client) domain.FollowUsecase {
	return &followUsecase{
		repo:   repo,
		notiUC: notiUC,
		rdb:    rdb,
	}
}

// FollowUser
func (u *followUsecase) FollowUser(ctx context.Context, followerID, followingID int64) error {
	if followerID == followingID {
		return errors.New("cannot follow yourself")
	}

	err := u.repo.Follow(ctx, followerID, followingID)
	if err != nil {
		return err
	}

	if u.notiUC != nil {
		go func() {
			noti := &domain.Notification{
				UserID:   followingID,
				ActorID:  followerID,
				Type:     "FOLLOW",
				EntityID: followerID,
				Message:  "đã bắt đầu theo dõi bạn.",
			}

			_ = u.notiUC.SendNotification(context.Background(), noti)
		}()
	}

	return nil
}

// UnfollowUser
func (u *followUsecase) UnfollowUser(ctx context.Context, followerID, followingID int64) error {
	err := u.repo.Unfollow(ctx, followerID, followingID)
	if err != nil {
		return err
	}

	if u.rdb != nil {
		go func() {
			// Gửi RELATIONSHIP_CHANGED cho CẢ 2 phía:
			// - followingID: người bị unfollow → cần biết để refresh danh sách
			// - followerID: bản thân người bấm unfollow → cần refresh lại phân loại chat
			for _, targetID := range []int64{followerID, followingID} {
				wsEvent := map[string]interface{}{
					"type": "RELATIONSHIP_CHANGED",
					"data": map[string]interface{}{
						"actor_id": followerID,
					},
				}
				eventBytes, _ := json.Marshal(wsEvent)

				envelope := map[string]interface{}{
					"to_user_id": targetID,
					"payload":    string(eventBytes),
				}
				envelopeBytes, _ := json.Marshal(envelope)

				_ = u.rdb.Publish(context.Background(), "system:ws_messages", envelopeBytes).Err()
			}
		}()
	}

	return nil
}

// GetFollowingList
func (u *followUsecase) GetFollowingList(ctx context.Context, userID int64) ([]int64, error) {
	return u.repo.GetFollowingIDs(ctx, userID)
}
