package usecase

import (
	"context"
	"errors"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
)

type followUsecase struct {
	repo   domain.FollowRepository
	notiUC domain.NotificationUsecase // Đổi từ hub sang cái này
}

func NewFollowUsecase(repo domain.FollowRepository, notiUC domain.NotificationUsecase) domain.FollowUsecase {
	return &followUsecase{
		repo:   repo,
		notiUC: notiUC,
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

	// Bắn thông báo qua NotificationUsecase (Vừa lưu DB vừa Real-time)
	if u.notiUC != nil {
		// Chúng ta dùng goroutine để việc gửi thông báo không làm chậm thao tác Follow
		go func() {
			noti := &domain.Notification{
				UserID:   followingID,
				ActorID:  followerID,
				Type:     "FOLLOW",
				EntityID: followerID,                 // ID của người thực hiện hành động
				Message:  "đã bắt đầu theo dõi bạn.", // Nội dung hiển thị
			}

			// Sử dụng context.Background() vì goroutine có thể chạy lâu hơn request chính
			_ = u.notiUC.SendNotification(context.Background(), noti)
		}()
	}

	return nil
}

func (u *followUsecase) UnfollowUser(ctx context.Context, followerID, followingID int64) error {
	return u.repo.Unfollow(ctx, followerID, followingID)
}

func (u *followUsecase) GetFollowingList(ctx context.Context, userID int64) ([]int64, error) {
	return u.repo.GetFollowingIDs(ctx, userID)
}
