package postgres

import (
	"context"
	"errors"
	"log"

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
	like := domain.Like{UserID: userID, PostID: postID}
	return r.db.WithContext(ctx).Create(&like).Error
}

// UnlikePost: Xóa bản ghi khỏi bảng likes
func (r *interactionRepository) UnlikePost(ctx context.Context, userID, postID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&domain.Like{}).Error
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
	return r.db.WithContext(ctx).Create(comment).Error
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
	var post struct {
		UserID int64 `gorm:"column:user_id"` // Chỉ định rõ cột
	}
	// Dùng bảng posts
	err := r.db.WithContext(ctx).Table("posts").Select("user_id").Where("id = ?", postID).First(&post).Error
	if err != nil {
		log.Printf("Lỗi GetPostOwner: %v", err)
		return 0
	}
	return post.UserID
}
