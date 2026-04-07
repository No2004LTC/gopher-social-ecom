package postgres

import (
	"context"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"gorm.io/gorm"
)

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) domain.PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *domain.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *postRepository) GetList(ctx context.Context, offset, limit int, currentUserID int64) ([]domain.Post, error) {
	var posts []domain.Post

	err := r.db.WithContext(ctx).
		Preload("User").
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	// Với mỗi bài post, ta đếm Like và Comment
	for i := range posts {
		// 1. Đếm Like
		r.db.Model(&domain.Like{}).Where("post_id = ?", posts[i].ID).Count(&posts[i].LikesCount)

		// 2. Đếm Comment
		r.db.Model(&domain.Comment{}).Where("post_id = ?", posts[i].ID).Count(&posts[i].CommentsCount)

		// 3. Kiểm tra User hiện tại đã Like chưa
		var count int64
		r.db.Model(&domain.Like{}).Where("post_id = ? AND user_id = ?", posts[i].ID, currentUserID).Count(&count)
		posts[i].IsLiked = count > 0
	}

	return posts, nil
}

func (r *postRepository) GetNewsfeed(ctx context.Context, followingIDs []int64, limit, offset int) ([]domain.Post, error) {
	var posts []domain.Post
	err := r.db.WithContext(ctx).
		Preload("User"). // Để hiển thị tên tác giả bài viết
		Where("user_id IN ?", followingIDs).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	return posts, err
}

func (r *postRepository) GetTrendingPosts(ctx context.Context, limit, offset int) ([]domain.Post, error) {
	var posts []domain.Post

	// Query lấy các bài viết có nhiều Like/Comment nhất
	// Sắp xếp theo tổng lượt tương tác giảm dần
	err := r.db.WithContext(ctx).
		Preload("User").
		Order("(likes_count * 2 + comments_count * 5) DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, err
}

func (r *postRepository) GetMixedFeed(ctx context.Context, userID int64, followingIDs []int64, limit, offset int) ([]domain.Post, error) {
	var posts []domain.Post

	// Gộp cả chính mình vào danh sách ưu tiên
	targetIDs := append(followingIDs, userID)

	err := r.db.WithContext(ctx).
		Preload("User").
		// Ưu tiên bài của người quen trước (CASE WHEN), sau đó đến điểm Hot (Like*2 + Comment*5)
		Order(r.db.Raw("CASE WHEN user_id IN (?) THEN 0 ELSE 1 END", targetIDs)).
		Order("(likes_count * 2 + comments_count * 5) DESC"). // Cột đã tồn tại nên chạy rất mượt
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error

	return posts, err
}
