package service

import (
	"context"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type ReceiptService struct {
	receiptRepo repository.ReceiptRepository
}

func NewReceiptService(receiptRepo repository.ReceiptRepository) *ReceiptService {
	return &ReceiptService{receiptRepo: receiptRepo}
}

func (s *ReceiptService) CreateFromOrder(ctx context.Context, companyID int64, order *dto.OrderResponse) error {
	items := make([]model.ReceiptItem, len(order.Items))
	for i, it := range order.Items {
		items[i] = model.ReceiptItem{
			ProductName: it.ProductName,
			VariantName: it.VariantName,
			SKU:         it.SKU,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
			Discount:    it.DiscountAmount,
			Tax:         it.TaxAmount,
			Subtotal:    it.Subtotal,
		}
	}

	payments := make([]model.ReceiptPayment, len(order.Payments))
	for i, p := range order.Payments {
		payments[i] = model.ReceiptPayment{
			Method:      p.Method,
			Amount:      p.Amount,
			ReferenceNo: p.ReferenceNo,
		}
	}

	completedAt := time.Now()
	if order.CompletedAt != nil {
		completedAt = *order.CompletedAt
	}

	rec := &model.Receipt{
		OrderID:       order.ID,
		CompanyID:     companyID,
		OrderNo:       order.OrderNo,
		CashierID:     order.CashierID,
		CashierName:   order.CashierName,
		CustomerID:    order.CustomerID,
		CustomerName:  order.CustomerName,
		Items:         items,
		Payments:      payments,
		TotalAmount:   order.TotalAmount,
		TotalDiscount: order.TotalDiscount,
		TotalTax:      order.TotalTax,
		GrandTotal:    order.GrandTotal,
		Notes:         order.Notes,
		CompletedAt:   completedAt,
		CreatedAt:     time.Now(),
	}

	return s.receiptRepo.Save(ctx, rec)
}

func (s *ReceiptService) GetByOrderID(ctx context.Context, companyID, orderID int64) (*model.Receipt, error) {
	rec, err := s.receiptRepo.GetByOrderID(ctx, companyID, orderID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if rec == nil {
		return nil, apperror.NotFound("receipt")
	}
	return rec, nil
}
