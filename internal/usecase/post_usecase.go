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
}

func NewPostUsecase(
	postRepo domain.PostRepository,
	followRepo domain.FollowRepository, // <--- Thêm tham số này
	storage *storage.S3Client,
) domain.PostUsecase {
	return &postUsecase{
		postRepo:   postRepo,
		followRepo: followRepo,
		storage:    storage,
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

func (u *postUsecase) GetPersonalizedFeed(ctx context.Context, userID int64, page int) ([]domain.Post, error) {
	limit := 10
	offset := (page - 1) * limit

	// 1. Lấy danh sách ID những người mình đang follow
	followingIDs, err := u.followRepo.GetFollowingIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Nếu chưa follow ai, có thể gợi ý bài viết của chính mình hoặc bài viết mới nhất toàn sàn
	if len(followingIDs) == 0 {
		followingIDs = []int64{userID} // Tạm thời chỉ xem bài của chính mình
	} else {
		followingIDs = append(followingIDs, userID) // Xem bài của người mình follow + bài của mình
	}

	// 3. Lấy bài viết từ DB
	return u.postRepo.GetNewsfeed(ctx, followingIDs, limit, offset)
}
