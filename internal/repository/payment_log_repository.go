package repository

import (
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type paymentLogRepository struct{ db *gorm.DB }

func NewPaymentLogRepository(db *gorm.DB) domain.PaymentLogRepository {
	return &paymentLogRepository{db: db}
}

func (r *paymentLogRepository) Create(l *domain.PaymentLog) error {
	return r.db.Create(l).Error
}

func (r *paymentLogRepository) Update(l *domain.PaymentLog) error {
	return r.db.Save(l).Error
}

func (r *paymentLogRepository) FindByID(id uuid.UUID) (*domain.PaymentLog, error) {
	var l domain.PaymentLog
	err := r.db.First(&l, "id = ?", id).Error
	return &l, err
}

func (r *paymentLogRepository) List(gateway, status, eventType string, page, pageSize int) ([]domain.PaymentLog, int64, error) {
	q := r.db.Model(&domain.PaymentLog{})
	if gateway != "" {
		q = q.Where("gateway = ?", gateway)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if eventType != "" {
		q = q.Where("event_type = ?", eventType)
	}
	var total int64
	q.Count(&total)
	var logs []domain.PaymentLog
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}
