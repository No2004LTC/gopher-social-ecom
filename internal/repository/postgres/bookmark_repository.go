package postgres

import (
	"context"

	"gorm.io/gorm"

	// Thay project_gopher bằng tên module trong go.mod của cậu
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
)

type bookmarkRepository struct {
	db *gorm.DB
}

func NewBookmarkRepository(db *gorm.DB) domain.BookmarkRepository {
	return &bookmarkRepository{db: db}
}

// Logic Toggle
func (r *bookmarkRepository) ToggleSavePost(ctx context.Context, userID int64, postID int64) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&domain.Bookmark{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error

	if err != nil {
		return false, err // Chỉ văng ra khi DB thực sự sập hoặc mất kết nối
	}

	if count == 0 {
		newBookmark := domain.Bookmark{
			UserID: userID,
			PostID: postID,
		}
		if errCreate := r.db.WithContext(ctx).Create(&newBookmark).Error; errCreate != nil {
			return false, errCreate
		}
		return true, nil
	}

	errDelete := r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&domain.Bookmark{}).Error

	if errDelete != nil {
		return false, errDelete
	}

	return false, nil
}

// Lấy danh sách các bài viết đã lưu (Join bảng)
func (r *bookmarkRepository) GetSavedPosts(ctx context.Context, userID int64, limit, offset int) ([]domain.Post, error) {
	var posts []domain.Post

	err := r.db.WithContext(ctx).
		Table("posts").
		Select(`
			posts.*,
			(SELECT COUNT(*) FROM likes WHERE likes.post_id = posts.id) AS likes_count,
			(SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) AS comments_count,
			(EXISTS (SELECT 1 FROM likes WHERE likes.post_id = posts.id AND likes.user_id = ?)) AS is_liked,
			true AS is_saved -- Đã vào trang Saved thì chắc chắn bài này đã được lưu (tránh tốn thêm 1 phép EXISTS)
		`, userID).
		Joins("JOIN bookmarks ON bookmarks.post_id = posts.id").
		Where("bookmarks.user_id = ?", userID).
		Preload("User").
		Order("bookmarks.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error

	return posts, err
}
