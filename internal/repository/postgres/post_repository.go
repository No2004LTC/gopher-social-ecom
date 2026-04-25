package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"gorm.io/gorm"
)

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) domain.PostRepository {
	return &postRepository{db: db}
}

// CREATE
func (r *postRepository) Create(ctx context.Context, post *domain.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

// DELETE
func (r *postRepository) DeletePost(ctx context.Context, postID int64, currentUserID int64) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", postID, currentUserID).
		Delete(&domain.Post{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("không tìm thấy bài viết hoặc bạn không có quyền xóa")
	}
	return nil
}

// UPDATE
func (r *postRepository) UpdatePost(ctx context.Context, postID int64, currentUserID int64, newContent string) error {
	result := r.db.WithContext(ctx).Model(&domain.Post{}).
		Where("id = ? AND user_id = ?", postID, currentUserID).
		Updates(map[string]interface{}{
			"content":    newContent,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("không tìm thấy bài viết hoặc bạn không có quyền sửa")
	}
	return nil
}

// GET POSTS
func (r *postRepository) GetPosts(ctx context.Context, currentUserID int64, targetUserID int64, limit, offset int) ([]domain.Post, error) {
	posts := make([]domain.Post, 0)

	query := r.db.WithContext(ctx).Model(&domain.Post{})

	if targetUserID > 0 {
		query = query.Where("user_id = ?", targetUserID)
	}

	err := query.
		Select(`
          posts.*,
          (SELECT COUNT(*) FROM likes WHERE likes.post_id = posts.id) AS likes_count,
          (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) AS comments_count,
          -- Kiểm tra trạng thái của User đang lướt (currentUserID) với bài viết
          (EXISTS (SELECT 1 FROM likes WHERE likes.post_id = posts.id AND likes.user_id = ?)) AS is_liked,
          (EXISTS (SELECT 1 FROM bookmarks WHERE bookmarks.post_id = posts.id AND bookmarks.user_id = ?)) AS is_saved
       `, currentUserID, currentUserID).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, err
}

// CountPosts
func (r *postRepository) CountPosts(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Post{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
