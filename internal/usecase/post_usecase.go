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

// CREATE
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

// DELETE
func (u *postUsecase) DeletePost(ctx context.Context, postID int64, currentUserID int64) error {
	return u.postRepo.DeletePost(ctx, postID, currentUserID)
}

// UPDATE
func (u *postUsecase) UpdatePost(ctx context.Context, postID int64, currentUserID int64, newContent string) error {
	return u.postRepo.UpdatePost(ctx, postID, currentUserID, newContent)
}

// GET POSTS
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
