package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) GetUserProfileByUsername(ctx context.Context, currentUserID int64, username string) (*domain.User, error) {
	var user domain.User
	query := `*, EXISTS (SELECT 1 FROM follows WHERE follower_id = ? AND following_id = users.id) as is_following`

	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Select(query, currentUserID).
		Where("username = ?", username).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateAvatar(ctx context.Context, userID int64, avatarURL string) error {
	result := r.db.WithContext(ctx).Model(&domain.User{ID: userID}).Update("avatar_url", avatarURL)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("failed to update user avatar")
	}
	return nil
}

func (r *userRepository) UpdateCover(ctx context.Context, userID int64, coverURL string) error {
	result := r.db.WithContext(ctx).Model(&domain.User{ID: userID}).Update("cover_url", coverURL)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("failed to update user cover")
	}
	return nil
}

func (r *userRepository) UpdateProfile(ctx context.Context, userID int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	result := r.db.WithContext(ctx).Table("users").Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("không tìm thấy người dùng")
	}
	return nil
}

func (r *userRepository) SearchUsers(ctx context.Context, currentUserID int64, query string, limit, offset int) ([]dto.UserCompact, error) {
	var users []dto.UserCompact
	selectQuery := `
        users.id, users.username, users.avatar_url,
        EXISTS (SELECT 1 FROM follows WHERE follower_id = ? AND following_id = users.id) as is_following,
        EXISTS (SELECT 1 FROM follows WHERE follower_id = users.id AND following_id = ?) as is_followed_by`

	err := r.db.WithContext(ctx).Table("users").
		Select(selectQuery, currentUserID, currentUserID).
		Where("(username ILIKE ? OR email ILIKE ?) AND id <> ?", "%"+query+"%", "%"+query+"%", currentUserID).
		Limit(limit).Offset(offset).Scan(&users).Error
	return users, err
}

func (r *userRepository) GetFollowing(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error) {
	var users []dto.UserCompact
	err := r.db.WithContext(ctx).Table("users").
		Select(`users.id, users.username, users.avatar_url, true as is_following, EXISTS (SELECT 1 FROM follows f2 WHERE f2.follower_id = users.id AND f2.following_id = ?) as is_followed_by`, currentUserID).
		Joins("JOIN follows f1 ON f1.following_id = users.id").
		Where("f1.follower_id = ?", currentUserID).
		Limit(limit).Offset(offset).Scan(&users).Error
	return users, err
}

func (r *userRepository) GetFollowers(ctx context.Context, currentUserID int64, limit, offset int) ([]dto.UserCompact, error) {
	var users []dto.UserCompact
	err := r.db.WithContext(ctx).Table("users").
		Select(`users.id, users.username, users.avatar_url, EXISTS (SELECT 1 FROM follows f2 WHERE f2.follower_id = ? AND f2.following_id = users.id) as is_following, true as is_followed_by`, currentUserID).
		Joins("JOIN follows f1 ON f1.follower_id = users.id").
		Where("f1.following_id = ?", currentUserID).
		Limit(limit).Offset(offset).Scan(&users).Error
	return users, err
}

func (r *userRepository) GetSuggestedUsers(ctx context.Context, myUserID int64, limit int) ([]domain.SuggestedUser, error) {
	var suggestions []domain.SuggestedUser
	query := `
       SELECT u.id, u.username, u.avatar_url,
          (SELECT COUNT(*) FROM follows f1 INNER JOIN follows f2 ON f1.following_id = f2.follower_id WHERE f1.follower_id = ? AND f2.following_id = u.id) AS mutual_friends_count
       FROM users u
       WHERE u.id != ? AND u.id NOT IN (SELECT following_id FROM follows WHERE follower_id = ?)
       ORDER BY mutual_friends_count DESC, u.created_at DESC LIMIT ?;
    `
	err := r.db.WithContext(ctx).Raw(query, myUserID, myUserID, myUserID, limit).Scan(&suggestions).Error
	return suggestions, err
}
