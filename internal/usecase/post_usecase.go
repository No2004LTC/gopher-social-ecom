package usecase

import (
	"context"
	"mime/multipart"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/pkg/storage"
)

type postUsecase struct {
	repo    domain.PostRepository
	storage *storage.S3Client
}

func NewPostUsecase(repo domain.PostRepository, storage *storage.S3Client) domain.PostUsecase {
	return &postUsecase{
		repo:    repo,
		storage: storage,
	}
}

func (u *postUsecase) CreatePost(ctx context.Context, post *domain.Post, file *multipart.FileHeader) error {
	if file != nil {
		url, err := u.storage.UploadFile(file, "posts")
		if err != nil {
			return err
		}
		post.ImageURL = url
	}
	return u.repo.Create(ctx, post)
}

func (u *postUsecase) GetFeed(ctx context.Context, page, limit int, currentUserID int64) ([]domain.Post, error) {
	offset := (page - 1) * limit
	// Chuyền currentUserID xuống đây
	return u.repo.GetList(ctx, offset, limit, currentUserID)
}
