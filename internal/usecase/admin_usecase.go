package usecase

import (
	"context"
	"errors"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
)

type adminUsecase struct {
	adminRepo domain.AdminRepository
}

func NewAdminUsecase(ar domain.AdminRepository) domain.AdminUsecase {
	return &adminUsecase{
		adminRepo: ar,
	}
}

func (u *adminUsecase) GetDashboardStats(ctx context.Context) (*dto.AdminDashboardStatsResponse, error) {
	totalUsers, err := u.adminRepo.CountUsers(ctx)
	if err != nil {
		return nil, errors.New("không thể đếm tổng số thành viên")
	}

	totalPosts, err := u.adminRepo.CountPosts(ctx)
	if err != nil {
		return nil, errors.New("không thể đếm tổng số bài viết")
	}

	totalBanned, err := u.adminRepo.CountBannedUsers(ctx)
	if err != nil {
		return nil, errors.New("không thể đếm số tài khoản bị khóa")
	}

	return &dto.AdminDashboardStatsResponse{
		TotalUsers:       totalUsers,
		TotalPosts:       totalPosts,
		TotalBannedUsers: totalBanned,
	}, nil
}

func (u *adminUsecase) GetGrowthStats(ctx context.Context, days int) ([]dto.GrowthPoint, error) {
	if days <= 0 {
		days = 7
	}
	return u.adminRepo.GetDailyGrowth(ctx, days)
}

func (u *adminUsecase) GetAllUsers(ctx context.Context, keyword string) ([]*domain.User, error) {
	return u.adminRepo.GetAllUsers(ctx, keyword)
}

func (u *adminUsecase) BanUser(ctx context.Context, targetUserID int64, adminEmail string) error {
	targetUser, err := u.adminRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return errors.New("không tìm thấy người dùng này trong hệ thống")
	}

	if targetUser.Email == "lethanhcong20052004@gmail.com" || targetUser.Email == adminEmail {
		return errors.New("thao tác không hợp lệ: không thể khóa tài khoản quản trị tối cao")
	}

	err = u.adminRepo.UpdateStatus(ctx, targetUserID, "banned")
	if err != nil {
		return errors.New("thực thi lệnh BAN thất bại, vui lòng thử lại")
	}

	return nil
}

func (u *adminUsecase) UnbanUser(ctx context.Context, targetUserID int64) error {
	_, err := u.adminRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return errors.New("không tìm thấy người dùng này trong hệ thống")
	}

	err = u.adminRepo.UpdateStatus(ctx, targetUserID, "active")
	if err != nil {
		return errors.New("thực thi lệnh GỠ KHÓA thất bại, vui lòng thử lại")
	}

	return nil
}

func (u *adminUsecase) GetModerationFeed(ctx context.Context, limit int) ([]*domain.Post, error) {
	if limit <= 0 {
		limit = 20
	}
	return u.adminRepo.GetLatestPosts(ctx, limit)
}

func (u *adminUsecase) AdminDeletePost(ctx context.Context, postID int64) error {
	return u.adminRepo.DeletePost(ctx, postID)
}
