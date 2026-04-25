package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
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

// UploadFile
func (s *S3Client) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("không thể mở file: %v", err)
	}
	defer src.Close()

	objectName := fmt.Sprintf("%s/%d_%s", folder, time.Now().Unix(), file.Filename)

	_, err = s.client.PutObject(context.Background(), s.bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", fmt.Errorf("lỗi khi đẩy file lên MinIO: %v", err)
	}

	return fmt.Sprintf("http://%s/%s/%s", s.endpoint, s.bucketName, objectName), nil
}

// PutObject
func (s *S3Client) PutObject(ctx context.Context, objectName string, r io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return s.client.PutObject(ctx, s.bucketName, objectName, r, objectSize, opts)
}

// Endpoint
func (s *S3Client) Endpoint() string { return s.endpoint }

// Bucket
func (s *S3Client) Bucket() string { return s.bucketName }
