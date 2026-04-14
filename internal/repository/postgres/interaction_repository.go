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
	var count int64
	// Dùng Count() sẽ trả về 0 nếu không có, hoàn toàn không văng lỗi "record not found"
	err := r.db.Model(&domain.Like{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
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

func (r *interactionRepository) UpdateComment(ctx context.Context, commentID int64, currentUserID int64, newContent string) error {
	// 1. Đổi model sang bảng Comment
	result := r.db.WithContext(ctx).Model(&domain.Comment{}).
		Where("id = ? AND user_id = ?", commentID, currentUserID). // 2. Đổi tham số thành commentID
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
		Delete(&domain.Comment{}) // 3. Xóa model ở bảng Comment

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("không tìm thấy comment hoặc bạn không có quyền xóa")
	}
	return nil
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
