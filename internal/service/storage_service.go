package service

import "github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"

// noopStorageClient is a no-op implementation of domain.StorageClient used when
// S3 credentials are not configured.
type noopStorageClient struct{}

func NewNoopStorageClient() domain.StorageClient { return &noopStorageClient{} }

func (n *noopStorageClient) UploadFile(bucket, key string, data []byte, contentType string) (string, error) {
	return "", nil
}

func (n *noopStorageClient) DeleteFile(bucket, key string) error {
	return nil
}

func (n *noopStorageClient) GetPublicURL(key string) string {
	return ""
}
