package domain

// EmailService is a thin interface for sending transactional emails.
type EmailService interface {
	SendWelcome(to, firstName string) error
	SendEmailVerification(to, firstName, verifyURL string) error
	SendPasswordReset(to, firstName, resetURL string) error
	SendOrderConfirmation(to, firstName, orderNo string, total float64) error
	SendOrderShipped(to, firstName, orderNo, courier, trackingNo string) error
	SendReceiptWithPDF(to, firstName, orderNo string, pdfData []byte) error
}

// StorageClient abstracts the object storage (MinIO / S3-compatible).
type StorageClient interface {
	UploadFile(bucket, key string, data []byte, contentType string) (string, error)
	DeleteFile(bucket, key string) error
	GetPublicURL(key string) string
}

// AIService abstracts the AI analytics and customer service.
type AIService interface {
	CustomerChat(userID, sessionID, message string) (string, error)
	AdminAnalytics(adminID, query string, dbContext string) (string, error)
}
