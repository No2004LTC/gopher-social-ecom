package postgres

import (
	"context"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"gorm.io/gorm"
)

type followRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) domain.FollowRepository {
	return &followRepository{db: db}
}

// Follow
func (r *followRepository) Follow(ctx context.Context, followerID, followingID int64) error {
	return r.db.WithContext(ctx).Create(&domain.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}).Error
}

// Unfollow
func (r *followRepository) Unfollow(ctx context.Context, followerID, followingID int64) error {
	return r.db.WithContext(ctx).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&domain.Follow{}).Error
}

// IsFollowing
func (r *followRepository) IsFollowing(ctx context.Context, followerID, followingID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("follows").
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error
	return count > 0, err
}

// GetFollowingIDs
func (r *followRepository) GetFollowingIDs(ctx context.Context, userID int64) ([]int64, error) {
	var followingIDs []int64
	err := r.db.WithContext(ctx).Table("follows").
		Where("follower_id = ?", userID).
		Pluck("following_id", &followingIDs).Error
	return followingIDs, err
}

// CountFollowers
func (r *followRepository) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("follows").
		Where("following_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountFollowing
func (r *followRepository) CountFollowing(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("follows").
		Where("follower_id = ?", userID).
		Count(&count).Error
	return count, err
}
