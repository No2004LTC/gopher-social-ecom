package postgres

import (
	"context"
	"errors"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"gorm.io/gorm"
)

type interactionRepository struct {
	db *gorm.DB
}

func NewInteractionRepository(db *gorm.DB) domain.InteractionRepository {
	return &interactionRepository{db: db}
}

// LikePost: Thêm một bản ghi vào bảng likes
func (r *interactionRepository) LikePost(ctx context.Context, userID, postID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Tạo bản ghi like
		if err := tx.Create(&domain.Like{UserID: userID, PostID: postID}).Error; err != nil {
			return err
		}
		// 2. Tăng likes_count ở bảng posts
		return tx.Model(&domain.Post{}).Where("id = ?", postID).
			UpdateColumn("likes_count", gorm.Expr("likes_count + ?", 1)).Error
	})
}

// UnlikePost: Xóa bản ghi khỏi bảng likes
func (r *interactionRepository) UnlikePost(ctx context.Context, userID, postID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Xóa bản ghi like
		if err := tx.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&domain.Like{}).Error; err != nil {
			return err
		}
		// 2. Giảm likes_count ở bảng posts
		return tx.Model(&domain.Post{}).Where("id = ?", postID).
			UpdateColumn("likes_count", gorm.Expr("likes_count - ?", 1)).Error
	})
}

// IsLiked: Kiểm tra xem user đã like chưa
func (r *interactionRepository) IsLiked(ctx context.Context, userID, postID int64) (bool, error) {
	var like domain.Like
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		First(&like).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

// CreateComment: Lưu comment mới
func (r *interactionRepository) CreateComment(ctx context.Context, comment *domain.Comment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		// Tăng comments_count
		return tx.Model(&domain.Post{}).Where("id = ?", comment.PostID).
			UpdateColumn("comments_count", gorm.Expr("comments_count + ?", 1)).Error
	})
}

// GetCommentsByPostID: Lấy danh sách comment kèm thông tin User
func (r *interactionRepository) GetCommentsByPostID(ctx context.Context, postID int64) ([]domain.Comment, error) {
	var comments []domain.Comment
	err := r.db.WithContext(ctx).
		Preload("User"). // Để biết ai là người comment
		Where("post_id = ?", postID).
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}

func (r *interactionRepository) GetPostOwner(ctx context.Context, postID int64) int64 {
	var userID int64
	// Lấy user_id từ bảng posts dựa trên postID
	r.db.WithContext(ctx).Table("posts").Select("user_id").Where("id = ?", postID).Scan(&userID)
	return userID
}
