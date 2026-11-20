package service

import (
	"errors"
	"fmt"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
)

type inventoryService struct {
	inventoryRepo domain.InventoryRepository
	productRepo   domain.ProductRepository
}

func NewInventoryService(inventoryRepo domain.InventoryRepository, productRepo domain.ProductRepository) domain.InventoryService {
	return &inventoryService{inventoryRepo: inventoryRepo, productRepo: productRepo}
}

func (s *inventoryService) AdjustStock(adminID uuid.UUID, req domain.StockAdjustRequest) (*domain.InventoryMovement, error) {
	variant, err := s.productRepo.FindVariantByID(req.VariantID)
	if err != nil {
		return nil, errors.New("variant not found")
	}

	balanceBefore := variant.StockOnHand
	balanceAfter := balanceBefore + req.Quantity

	if balanceAfter < 0 {
		return nil, errors.New("adjustment would result in negative stock")
	}

	m := &domain.InventoryMovement{
		VariantID:     req.VariantID,
		Type:          req.Type,
		Quantity:      req.Quantity,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Reference:     req.Reference,
		Note:          req.Note,
		CreatedBy:     adminID,
	}

	if err := s.inventoryRepo.CreateMovement(m); err != nil {
		return nil, err
	}

	// Update variant stock
	variant.StockOnHand = balanceAfter
	if err := s.productRepo.UpdateVariant(variant); err != nil {
		return nil, err
	}

	return m, nil
}

func (s *inventoryService) ListMovements(variantID *uuid.UUID, movType domain.MovementType, page, pageSize int) ([]domain.InventoryMovement, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.inventoryRepo.ListMovements(variantID, movType, page, pageSize)
}

func (s *inventoryService) GetLowStockItems() ([]domain.LowStockItem, error) {
	return s.inventoryRepo.GetLowStockItems()
}

func (s *inventoryService) GetSummary() (*domain.InventorySummary, error) {
	return s.inventoryRepo.GetSummary()
}

// BulkReceive processes a purchase-order style receipt — multiple variants in one go.
func (s *inventoryService) BulkReceive(adminID uuid.UUID, req domain.BulkReceiveRequest) ([]domain.InventoryMovement, error) {
	var movements []domain.InventoryMovement
	for _, item := range req.Items {
		variant, err := s.productRepo.FindVariantByID(item.VariantID)
		if err != nil {
			return nil, fmt.Errorf("variant %s not found", item.VariantID)
		}
		balanceBefore := variant.StockOnHand
		balanceAfter := balanceBefore + item.Quantity
		note := item.Note
		if note == "" {
			note = req.Note
		}
		m := &domain.InventoryMovement{
			VariantID:     item.VariantID,
			Type:          domain.MovementReceive,
			Quantity:      item.Quantity,
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceAfter,
			Reference:     req.Reference,
			Note:          note,
			CreatedBy:     adminID,
		}
		if err := s.inventoryRepo.CreateMovement(m); err != nil {
			return nil, err
		}
		variant.StockOnHand = balanceAfter
		if err := s.productRepo.UpdateVariant(variant); err != nil {
			return nil, err
		}
		movements = append(movements, *m)
	}
	return movements, nil
}
