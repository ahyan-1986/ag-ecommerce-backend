package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentLogStatus string

const (
	PaymentLogReceived  PaymentLogStatus = "received"
	PaymentLogProcessed PaymentLogStatus = "processed"
	PaymentLogFailed    PaymentLogStatus = "failed"
	PaymentLogIgnored   PaymentLogStatus = "ignored"
)

// PaymentLog records every inbound payment gateway webhook event.
type PaymentLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Gateway      string     `gorm:"type:varchar(50);not null;default:'stripe';index" json:"gateway"`
	EventType    string     `gorm:"type:varchar(100);not null;index" json:"eventType"`
	GatewayTxnID string     `gorm:"type:varchar(255);not null;default:''" json:"gatewayTxnId"`
	OrderID      *uuid.UUID `gorm:"type:uuid;index" json:"orderId,omitempty"`
	OrderNo      string     `gorm:"type:varchar(50);not null;default:''" json:"orderNo"`
	// AmountCents is the transaction amount in the smallest currency unit (e.g. cents for MYR).
	AmountCents  int64            `gorm:"not null;default:0" json:"amountCents"`
	Currency     string           `gorm:"type:varchar(10);not null;default:''" json:"currency"`
	Status       PaymentLogStatus `gorm:"type:varchar(30);not null;default:'received';index" json:"status"`
	ErrorMessage string           `gorm:"type:text;not null;default:''" json:"errorMessage"`
	IPAddress    string           `gorm:"type:varchar(50);not null;default:''" json:"ipAddress"`
	// RawPayload stores the full webhook JSON for audit. Not exposed in list endpoint.
	RawPayload string    `gorm:"type:text;not null;default:''" json:"-"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index" json:"createdAt"`
}

func (PaymentLog) TableName() string { return "payment_logs" }

// PaymentLogRepository defines persistence for payment event logs.
type PaymentLogRepository interface {
	Create(log *PaymentLog) error
	Update(log *PaymentLog) error
	FindByID(id uuid.UUID) (*PaymentLog, error)
	List(gateway, status, eventType string, page, pageSize int) ([]PaymentLog, int64, error)
}
