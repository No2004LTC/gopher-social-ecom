package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"gorm.io/gorm"
)

type interactionRepository struct {
	db *gorm.DB
}

func NewInteractionRepository(db *gorm.DB) domain.InteractionRepository {
	return &interactionRepository{db: db}
}

// LikePost
func (r *interactionRepository) LikePost(ctx context.Context, userID, postID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&domain.Like{UserID: userID, PostID: postID}).Error; err != nil {
			return err
		}
		return tx.Model(&domain.Post{}).Where("id = ?", postID).
			UpdateColumn("likes_count", gorm.Expr("likes_count + ?", 1)).Error
	})
}

// UnlikePost
func (r *interactionRepository) UnlikePost(ctx context.Context, userID, postID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&domain.Like{}).Error; err != nil {
			return err
		}

		return tx.Model(&domain.Post{}).Where("id = ?", postID).
			UpdateColumn("likes_count", gorm.Expr("likes_count - ?", 1)).Error
	})
}

// IsLiked
func (r *interactionRepository) IsLiked(ctx context.Context, userID, postID int64) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Like{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CreateComment
func (r *interactionRepository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		return tx.Model(&domain.Post{}).Where("id = ?", comment.PostID).
			UpdateColumn("comments_count", gorm.Expr("comments_count + ?", 1)).Error
	})
}

func (r *interactionRepository) UpdateComment(ctx context.Context, commentID int64, currentUserID int64, newContent string) error {
	result := r.db.WithContext(ctx).Model(&domain.Comment{}).
		Where("id = ? AND user_id = ?", commentID, currentUserID).
		Updates(map[string]interface{}{
			"content":    newContent,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("không tìm thấy comment hoặc bạn không có quyền sửa")
	}
	return nil
}

func (r *interactionRepository) DeleteComment(ctx context.Context, commentID int64, currentUserID int64) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", commentID, currentUserID).
		Delete(&domain.Comment{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("không tìm thấy comment hoặc bạn không có quyền xóa")
	}
	return nil
}

// GetCommentsByPostID
func (r *interactionRepository) GetCommentsByPostID(ctx context.Context, postID int64) ([]domain.Comment, error) {
	var comments []domain.Comment
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("post_id = ?", postID).
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}

func (r *interactionRepository) GetPostOwner(ctx context.Context, postID int64) int64 {
	var userID int64
	r.db.WithContext(ctx).Table("posts").Select("user_id").Where("id = ?", postID).Scan(&userID)
	return userID
}
