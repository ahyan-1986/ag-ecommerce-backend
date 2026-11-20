package storage

import (
	"bytes"
	"context"
	"fmt"
	"path"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/config"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type s3Client struct {
	client *minio.Client
	cfg    config.S3Config
}

func NewClient(cfg config.S3Config) (domain.StorageClient, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("S3_ENDPOINT not set")
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &s3Client{client: client, cfg: cfg}, nil
}

func (s *s3Client) UploadFile(bucket, key string, data []byte, contentType string) (string, error) {
	if s.cfg.Prefix != "" {
		key = path.Join(s.cfg.Prefix, key)
	}
	ctx := context.Background()
	_, err := s.client.PutObject(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return s.GetPublicURL(key), nil
}

func (s *s3Client) DeleteFile(bucket, key string) error {
	return s.client.RemoveObject(context.Background(), bucket, key, minio.RemoveObjectOptions{})
}

func (s *s3Client) GetPublicURL(key string) string {
	if s.cfg.PublicURL != "" {
		return fmt.Sprintf("%s/%s/%s", s.cfg.PublicURL, s.cfg.Bucket, key)
	}
	scheme := "http"
	if s.cfg.UseSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.cfg.Endpoint, s.cfg.Bucket, key)
}
