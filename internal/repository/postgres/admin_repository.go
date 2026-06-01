package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"gorm.io/gorm"
)

type adminRepository struct {
	db *gorm.DB
}

// NewAdminRepository khởi tạo thực thi DB cho Admin dùng GORM xịn
func NewAdminRepository(db *gorm.DB) domain.AdminRepository {
	return &adminRepository{db: db}
}

// 1. Đếm tổng số người dùng hệ thống
func (r *adminRepository) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&count).Error
	return count, err
}

// 2. Đếm tổng số bài viết toàn hệ thống
func (r *adminRepository) CountPosts(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Post{}).Count(&count).Error
	return count, err
}

// 3. Đếm số lượng tài khoản đang bị BAN
func (r *adminRepository) CountBannedUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).Where("status = ?", "banned").Count(&count).Error
	return count, err
}

// 4. Lấy danh sách Users kèm tìm kiếm nâng cao theo từ khóa (ILIKE)
func (r *adminRepository) GetAllUsers(ctx context.Context, keyword string) ([]*domain.User, error) {
	var users []*domain.User
	query := r.db.WithContext(ctx).Model(&domain.User{}).Order("created_at DESC")

	if keyword != "" {
		pattern := "%" + keyword + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ?", pattern, pattern)
	}

	err := query.Find(&users).Error
	return users, err
}

// 5. Tìm User theo ID để phục vụ logic kiểm tra trước khi BAN
func (r *adminRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Trả về nil nếu không tìm thấy để giống logic cũ của cậu
		}
		return nil, err
	}
	return &user, nil
}

// 6. Cập nhật trạng thái tài khoản (Hạ lệnh BAN vĩnh viễn)
func (r *adminRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	// GORM sẽ tự động cập nhật cả cột updated_at nếu struct User của cậu có trường đó
	result := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("không tìm thấy người dùng để cập nhật trạng thái")
	}
	return nil
}

// GetDailyGrowth trả về số user mới + post mới mỗi ngày trong N ngày gần nhất
func (r *adminRepository) GetDailyGrowth(ctx context.Context, days int) ([]dto.GrowthPoint, error) {
	result := make([]dto.GrowthPoint, days)
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
		dayEnd := dayStart.AddDate(0, 0, 1)

		var userCount int64
		r.db.WithContext(ctx).Model(&domain.User{}).Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).Count(&userCount)

		var postCount int64
		r.db.WithContext(ctx).Model(&domain.Post{}).Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).Count(&postCount)

		result[days-1-i] = dto.GrowthPoint{
			Date:       fmt.Sprintf("%02d/%02d", day.Day(), day.Month()),
			UsersCount: userCount,
			PostsCount: postCount,
		}
	}

	return result, nil
}

// 7. Lấy dòng thời gian các bài viết mới nhất (Hậu kiểm bằng cơm)
func (r *adminRepository) GetLatestPosts(ctx context.Context, limit int) ([]*domain.Post, error) {
	var posts []*domain.Post
	err := r.db.WithContext(ctx).Model(&domain.Post{}).
		Select(`posts.*,
			(SELECT COUNT(*) FROM likes WHERE likes.post_id = posts.id) AS likes_count,
			(SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) AS comments_count`).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

// 8. Xóa bài viết bất kỳ (Admin quyền tối cao, không cần kiểm tra user_id)
func (r *adminRepository) DeletePost(ctx context.Context, postID int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", postID).Delete(&domain.Post{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("không tìm thấy bài viết để xóa")
	}
	return nil
}
