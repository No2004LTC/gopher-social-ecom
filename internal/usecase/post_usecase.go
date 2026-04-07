package usecase

import (
	"context"
	"mime/multipart"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/pkg/storage"
)

type postUsecase struct {
	postRepo   domain.PostRepository
	followRepo domain.FollowRepository
	storage    *storage.S3Client
	notiUC     domain.NotificationUsecase
}

func NewPostUsecase(
	postRepo domain.PostRepository,
	followRepo domain.FollowRepository, // <--- Thêm tham số này
	storage *storage.S3Client,
	nuc domain.NotificationUsecase,
) domain.PostUsecase {
	return &postUsecase{
		postRepo:   postRepo,
		followRepo: followRepo,
		storage:    storage,
		notiUC:     nuc,
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
	return u.postRepo.Create(ctx, post)
}

func (u *postUsecase) GetFeed(ctx context.Context, page, limit int, currentUserID int64) ([]domain.Post, error) {
	offset := (page - 1) * limit
	// Chuyền currentUserID xuống đây
	return u.postRepo.GetList(ctx, offset, limit, currentUserID)
}

func (u *postUsecase) GetDiscoveryFeed(ctx context.Context, userID int64, page int) ([]domain.Post, error) {
	limit := 10
	offset := (page - 1) * limit

	followingIDs, _ := u.followRepo.GetFollowingIDs(ctx, userID)

	// Gọi xuống Repo
	return u.postRepo.GetMixedFeed(ctx, userID, followingIDs, limit, offset)
}
