package usecase

import (
	"context"
	"mime/multipart"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/pkg/storage"
)

type postUsecase struct {
	postRepo domain.PostRepository
	storage  *storage.S3Client
	notiUC   domain.NotificationUsecase
}

func NewPostUsecase(
	postRepo domain.PostRepository,
	storage *storage.S3Client,
	nuc domain.NotificationUsecase,
) domain.PostUsecase {
	return &postUsecase{
		postRepo: postRepo,
		storage:  storage,
		notiUC:   nuc,
	}
}

// CREATE - Xử lý nghiệp vụ upload ảnh trước khi lưu DB
func (u *postUsecase) CreatePost(ctx context.Context, post *domain.Post, file *multipart.FileHeader) error {
	if file != nil {
		url, err := u.storage.UploadFile(file, "posts")
		if err != nil {
			return err
		}
		post.ImageURL = url
	}
	return u.postRepo.Create(ctx, post)
}

// DELETE - Xóa bài viết
func (u *postUsecase) DeletePost(ctx context.Context, postID int64, currentUserID int64) error {
	return u.postRepo.DeletePost(ctx, postID, currentUserID)
}

// UPDATE - Sửa bài viết
func (u *postUsecase) UpdatePost(ctx context.Context, postID int64, currentUserID int64, newContent string) error {
	return u.postRepo.UpdatePost(ctx, postID, currentUserID, newContent)
}

/*
🎯 HÀM CHIẾN LƯỢC: GetPosts

	Hàm này nhận targetUserID từ Handler để quyết định lấy bài cho Trang chủ hay Profile.
	Logic phân trang (Pagination) được tính toán tập trung tại đây.
*/
func (u *postUsecase) GetPosts(ctx context.Context, currentUserID int64, targetUserID int64, page, limit int) ([]domain.Post, error) {
	// Tránh trường hợp limit quá ảo
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	return u.postRepo.GetPosts(ctx, currentUserID, targetUserID, limit, offset)
}
