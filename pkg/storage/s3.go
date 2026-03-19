package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart" // Thêm cái này để nhận file từ Handler
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	client     *minio.Client
	bucketName string
	endpoint   string
}

func NewS3Client(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*S3Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &S3Client{client: minioClient, bucketName: bucketName, endpoint: endpoint}, nil
}

// --- ĐÂY LÀ HÀM CẬU CẦN THÊM VÀO ---

// UploadFile giúp Usecase upload ảnh nhanh mà không cần quan tâm logic bên dưới
func (s *S3Client) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	// 1. Mở file từ header
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("không thể mở file: %v", err)
	}
	defer src.Close()

	// 2. Tạo tên file duy nhất (folder/timestamp_tên-file)
	// Tránh việc 2 người cùng upload file "anh.jpg" làm ghi đè nhau
	objectName := fmt.Sprintf("%s/%d_%s", folder, time.Now().Unix(), file.Filename)

	// 3. Upload lên MinIO
	_, err = s.client.PutObject(context.Background(), s.bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", fmt.Errorf("lỗi khi đẩy file lên MinIO: %v", err)
	}

	// 4. Trả về URL công khai để lưu vào Postgres
	// Format: http://localhost:9000/gopher-bucket/posts/12345_anh.jpg
	return fmt.Sprintf("http://%s/%s/%s", s.endpoint, s.bucketName, objectName), nil
}

// --- CÁC HÀM CŨ GIỮ NGUYÊN ---

func (s *S3Client) PutObject(ctx context.Context, objectName string, r io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return s.client.PutObject(ctx, s.bucketName, objectName, r, objectSize, opts)
}

func (s *S3Client) Endpoint() string { return s.endpoint }
func (s *S3Client) Bucket() string   { return s.bucketName }
