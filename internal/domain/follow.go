package domain

import (
	"context"
	"time"
)

type Follow struct {
	FollowerID  int64     `json:"follower_id" gorm:"primaryKey"`
	FollowingID int64     `json:"following_id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
}

// FollowRepository
type FollowRepository interface {
	Follow(ctx context.Context, followerID, followingID int64) error
	Unfollow(ctx context.Context, followerID, followingID int64) error
	IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error)
	GetFollowingIDs(ctx context.Context, userID int64) ([]int64, error)
	CountFollowers(ctx context.Context, userID int64) (int64, error)
	CountFollowing(ctx context.Context, userID int64) (int64, error)
}

// FollowUsecase
type FollowUsecase interface {
	FollowUser(ctx context.Context, followerID, followingID int64) error
	UnfollowUser(ctx context.Context, followerID, followingID int64) error
	GetFollowingList(ctx context.Context, userID int64) ([]int64, error)
}
