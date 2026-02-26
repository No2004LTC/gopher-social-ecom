package utils

import (
	"context"
	"io"

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

// PutObject uploads content from reader into the configured bucket
func (s *S3Client) PutObject(ctx context.Context, objectName string, r io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return s.client.PutObject(ctx, s.bucketName, objectName, r, objectSize, opts)
}

// Client returns the underlying minio client (if needed)
func (s *S3Client) Client() *minio.Client {
	return s.client
}

// Endpoint returns the configured MinIO endpoint
func (s *S3Client) Endpoint() string {
	return s.endpoint
}

// Bucket returns the configured bucket name
func (s *S3Client) Bucket() string {
	return s.bucketName
}
