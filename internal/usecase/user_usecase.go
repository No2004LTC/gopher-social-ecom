package usecase

import (
	"context"
	"fmt"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"github.com/redis/go-redis/v9"
	"strconv"
)

type userUsecase struct {
	userRepo    domain.UserRepository
	followRepo  domain.FollowRepository
	postRepo    domain.PostRepository
	redisClient *redis.Client
}

func NewUserUsecase(ur domain.UserRepository, fr domain.FollowRepository, pr domain.PostRepository, rdb *redis.Client) domain.UserUsecase {
	return &userUsecase{
		userRepo:    ur,
		followRepo:  fr,
		postRepo:    pr,
		redisClient: rdb,
	}
}

// 1. GetProfile
func (uc *userUsecase) GetProfile(ctx context.Context, userID int64) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

// 2. GetUserProfileByUsername (Hàm quan trọng nhất để hiện số Follower)
func (uc *userUsecase) GetUserProfileByUsername(ctx context.Context, currentUserID int64, username string) (*dto.UserProfileResponse, error) {
	user, err := uc.userRepo.GetUserProfileByUsername(ctx, currentUserID, username)
	if err != nil || user == nil {
		return nil, fmt.Errorf("không tìm thấy người dùng")
	}

	followers, _ := uc.followRepo.CountFollowers(ctx, user.ID)
	following, _ := uc.followRepo.CountFollowing(ctx, user.ID)
	posts, _ := uc.postRepo.CountPosts(ctx, user.ID)

	return &dto.UserProfileResponse{
		ID:             user.ID,
		Username:       user.Username,
		Bio:            user.Bio,
		AvatarURL:      user.AvatarURL,
		CoverURL:       user.CoverURL,
		FollowersCount: int(followers),
		FollowingCount: int(following),
		PostsCount:     int(posts),
		IsFollowing:    user.IsFollowing,
	}, nil
}

// 3. UpdateAvatar
func (uc *userUsecase) UpdateAvatar(ctx context.Context, userID int64, url string) error {
	return uc.userRepo.UpdateAvatar(ctx, userID, url)
}

// 4. UpdateCover
func (uc *userUsecase) UpdateCover(ctx context.Context, userID int64, url string) error {
	return uc.userRepo.UpdateCover(ctx, userID, url)
}

// 5. UpdateProfile
func (uc *userUsecase) UpdateProfile(ctx context.Context, userID int64, input dto.UpdateProfileInput) error {
	updates := make(map[string]interface{})
	if input.Username != nil {
		updates["username"] = *input.Username
	}
	if input.Bio != nil {
		updates["bio"] = *input.Bio
	}
	return uc.userRepo.UpdateProfile(ctx, userID, updates)
}

// 6. SearchUsers
func (uc *userUsecase) SearchUsers(ctx context.Context, currentUserID int64, query string, limit, offset int) ([]dto.UserCompact, error) {
	return uc.userRepo.SearchUsers(ctx, currentUserID, query, limit, offset)
}

// 7. GetFollowing
func (uc *userUsecase) GetFollowing(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error) {
	return uc.userRepo.GetFollowing(ctx, currentUserID, limit, offset)
}

// 8. GetFollowers
func (uc *userUsecase) GetFollowers(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error) {
	return uc.userRepo.GetFollowers(ctx, currentUserID, limit, offset)
}

// 9. GetFriendSuggestions
func (uc *userUsecase) GetFriendSuggestions(ctx context.Context, userID int64) ([]domain.SuggestedUser, error) {
	return uc.userRepo.GetSuggestedUsers(ctx, userID, 10)
}

// 10. GetOnlineContacts
func (uc *userUsecase) GetOnlineContacts(ctx context.Context, userID int64) ([]dto.UserCompact, error) {
	contacts, _ := uc.userRepo.GetFollowers(ctx, userID, 50, 0)
	onlineMap, _ := uc.redisClient.HGetAll(ctx, "system:online_users").Result()

	for i := range contacts {
		userIDStr := strconv.FormatInt(contacts[i].ID, 10)
		if _, exists := onlineMap[userIDStr]; exists {
			contacts[i].IsOnline = true
		}
	}
	return contacts, nil
}
