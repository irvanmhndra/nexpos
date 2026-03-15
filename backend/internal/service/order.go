package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type OrderService struct {
	orderRepo        repository.OrderRepository
	orderItemRepo    repository.OrderItemRepository
	paymentRepo      repository.PaymentRepository
	variantRepo      repository.ProductVariantRepository
	customerRepo     repository.CustomerRepository
	userRepo         repository.UserRepository
	settingsRepo     repository.CompanySettingsRepository
	stockRepo        repository.StockRepository
	movementRepo     repository.StockMovementRepository
	promotionRepo    repository.PromotionRepository
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	paymentRepo repository.PaymentRepository,
	variantRepo repository.ProductVariantRepository,
	customerRepo repository.CustomerRepository,
	userRepo repository.UserRepository,
	settingsRepo repository.CompanySettingsRepository,
	stockRepo repository.StockRepository,
	movementRepo repository.StockMovementRepository,
	promotionRepo repository.PromotionRepository,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		paymentRepo:   paymentRepo,
		variantRepo:   variantRepo,
		customerRepo:  customerRepo,
		userRepo:      userRepo,
		settingsRepo:  settingsRepo,
		stockRepo:     stockRepo,
		movementRepo:  movementRepo,
		promotionRepo: promotionRepo,
	}
}

// Create creates a new order as draft
func (s *OrderService) Create(ctx context.Context, companyID, branchID, cashierID int64, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// Validate items
	if len(req.Items) == 0 {
		return nil, apperror.BadRequest("Order must have at least one item")
	}

	// Get company settings for tax calculation
	settings, err := s.settingsRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	// Validate customer if provided
	if req.CustomerID != nil {
		_, err := s.customerRepo.GetByID(ctx, companyID, *req.CustomerID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, apperror.NotFound("Customer not found")
			}
			return nil, apperror.InternalError(err)
		}
	}

	// Validate fulfillment type
	fulfillmentType := model.FulfillmentTypeCounter
	if req.FulfillmentType != "" {
		fulfillmentType = req.FulfillmentType
	}

	// Validate delivery order requirements
	if fulfillmentType == model.FulfillmentTypeDelivery {
		if settings.RequireCustomerForDelivery && req.CustomerID == nil {
			return nil, apperror.BadRequest("Customer is required for delivery orders")
		}
	}

	// Generate order number
	orderNo, err := s.orderRepo.GenerateOrderNo(ctx, companyID, branchID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	// Calculate totals
	var totalAmount, totalDiscount, totalTax float64
	orderItems := make([]*model.OrderItem, 0, len(req.Items))

	for _, itemInput := range req.Items {
		// Get variant details
		if itemInput.ProductVariantID == nil {
			return nil, apperror.BadRequest("Product variant is required")
		}

		variant, err := s.variantRepo.GetByID(ctx, *itemInput.ProductVariantID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, apperror.NotFound("Product variant not found")
			}
			return nil, apperror.InternalError(err)
		}

		// Check stock availability
		stock, err := s.stockRepo.GetByVariantAndBranch(ctx, *itemInput.ProductVariantID, branchID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		availableStock := 0
		if stock != nil {
			availableStock = stock.Quantity
		}
		if availableStock < itemInput.Quantity {
			return nil, apperror.ValidationError(
				fmt.Sprintf("Stok %s tidak cukup (tersedia: %d, diminta: %d)", variant.SKU, availableStock, itemInput.Quantity),
				nil,
			)
		}

		// Calculate item amounts
		unitPrice := variant.Price
		quantity := float64(itemInput.Quantity)
		discountAmount := itemInput.DiscountAmount
		subtotal := (unitPrice * quantity) - discountAmount

		// Tax calculation
		var taxAmount float64
		if settings.TaxEnabled {
			if settings.TaxInclusive {
				// Price includes tax, extract it
				taxAmount = subtotal * settings.TaxRate / (100 + settings.TaxRate)
			} else {
				// Price excludes tax, add it
				taxAmount = subtotal * settings.TaxRate / 100
			}
		}

		// COGS
		cogsAmount := variant.StandardCost * quantity

		totalAmount += unitPrice * quantity
		totalDiscount += discountAmount
		totalTax += taxAmount

		orderItems = append(orderItems, &model.OrderItem{
			ProductID:         &variant.ProductID,
			ProductVariantID:  itemInput.ProductVariantID,
			SKU:               variant.SKU,
			ProductName:       variant.ProductName,
			VariantName:       &variant.Name,
			VariantAttributes: variant.Attributes,
			UnitPrice:         unitPrice,
			UnitCost:          variant.StandardCost,
			Quantity:          itemInput.Quantity,
			DiscountAmount:    discountAmount,
			TaxAmount:         taxAmount,
			Subtotal:          subtotal,
			CogsAmount:        cogsAmount,
		})
	}

	// Evaluate and apply promotion
	subtotal := totalAmount - totalDiscount
	promo, promoDiscount := s.evaluatePromotion(ctx, companyID, req.PromoCode, subtotal)
	if promo != nil {
		totalDiscount += promoDiscount
	}

	grandTotal := totalAmount - totalDiscount + totalTax

	// Apply rounding if enabled
	if settings.RoundingEnabled && settings.RoundingAmount > 0 {
		grandTotal = roundToNearest(grandTotal, settings.RoundingAmount)
	}

	// Initialize fulfillment status for delivery orders
	var fulfillmentStatus *string
	if fulfillmentType == model.FulfillmentTypeDelivery {
		status := model.FulfillmentStatusPending
		fulfillmentStatus = &status
	}

	// Create order as draft
	order := &model.Order{
		CompanyID:         companyID,
		BranchID:          branchID,
		OrderNo:           orderNo,
		CustomerID:        req.CustomerID,
		CashierID:         cashierID,
		Status:            model.OrderStatusDraft,
		PaymentStatus:     model.PaymentStatusUnpaid,
		FulfillmentType:   fulfillmentType,
		FulfillmentStatus: fulfillmentStatus,
		ShippingAddress:   req.ShippingAddress,
		TotalAmount:       totalAmount,
		TotalDiscount:     totalDiscount,
		TotalTax:          totalTax,
		GrandTotal:        grandTotal,
		Notes:             req.Notes,
		OfflineID:         req.OfflineID,
	}
	if promo != nil {
		order.AppliedPromotions = model.JSONMap{
			"id":              promo.ID,
			"code":            promo.Code,
			"name":            promo.Name,
			"discount_amount": promoDiscount,
		}
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Create order items
	for _, item := range orderItems {
		item.OrderID = order.ID
		if err := s.orderItemRepo.Create(ctx, item); err != nil {
			return nil, apperror.InternalError(err)
		}
	}

	order.Items = orderItems

	// Process payments if provided
	if len(req.Payments) > 0 {
		_, err := s.processPayments(ctx, order, req.Payments, settings)
		if err != nil {
			return nil, err
		}
	}

	// Reload order
	return s.GetByID(ctx, companyID, order.ID)
}

// Preview calculates order totals without persisting anything.
// Used by the POS to show accurate pricing before checkout.
func (s *OrderService) Preview(ctx context.Context, companyID, branchID int64, req dto.PreviewOrderRequest) (*dto.PreviewOrderResponse, error) {
	if len(req.Items) == 0 {
		return nil, apperror.BadRequest("Order must have at least one item")
	}

	settings, err := s.settingsRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	var totalAmount, itemDiscount, totalTax float64

	for _, itemInput := range req.Items {
		if itemInput.ProductVariantID == nil {
			return nil, apperror.BadRequest("Product variant is required")
		}

		variant, err := s.variantRepo.GetByID(ctx, *itemInput.ProductVariantID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, apperror.NotFound("Product variant not found")
			}
			return nil, apperror.InternalError(err)
		}

		unitPrice := variant.Price
		quantity := float64(itemInput.Quantity)
		discount := itemInput.DiscountAmount
		subtotalItem := (unitPrice * quantity) - discount

		var taxAmount float64
		if settings.TaxEnabled {
			if settings.TaxInclusive {
				taxAmount = subtotalItem * settings.TaxRate / (100 + settings.TaxRate)
			} else {
				taxAmount = subtotalItem * settings.TaxRate / 100
			}
		}

		totalAmount += unitPrice * quantity
		itemDiscount += discount
		totalTax += taxAmount
	}

	subtotalAfterItemDiscounts := totalAmount - itemDiscount
	promo, promoDiscount := s.evaluatePromotion(ctx, companyID, req.PromoCode, subtotalAfterItemDiscounts)

	grandTotal := subtotalAfterItemDiscounts - promoDiscount + totalTax
	if settings.RoundingEnabled && settings.RoundingAmount > 0 {
		grandTotal = roundToNearest(grandTotal, settings.RoundingAmount)
	}

	resp := &dto.PreviewOrderResponse{
		Subtotal:      subtotalAfterItemDiscounts,
		ItemDiscount:  itemDiscount,
		PromoDiscount: promoDiscount,
		Tax:           totalTax,
		GrandTotal:    grandTotal,
	}
	if promo != nil {
		resp.AppliedPromotion = &dto.AppliedPromotionDTO{
			ID:             promo.ID,
			Code:           promo.Code,
			Name:           promo.Name,
			DiscountAmount: promoDiscount,
		}
	}

	return resp, nil
}

// ConfirmOrder moves order from draft to confirmed
func (s *OrderService) ConfirmOrder(ctx context.Context, companyID, id int64) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("Order not found")
		}
		return nil, apperror.InternalError(err)
	}

	// Validate state transition
	if order.Status != model.OrderStatusDraft {
		return nil, apperror.BadRequest("Only draft orders can be confirmed")
	}

	// Update order
	now := time.Now()
	order.Status = model.OrderStatusConfirmed
	order.ConfirmedAt = &now

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.GetByID(ctx, companyID, id)
}

// AddPayment adds a payment to an order (supports split payments)
func (s *OrderService) AddPayment(ctx context.Context, companyID, id int64, req dto.AddPaymentRequest) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("Order not found")
		}
		return nil, apperror.InternalError(err)
	}

	// Only allow payments on draft or confirmed orders
	if order.Status != model.OrderStatusDraft && order.Status != model.OrderStatusConfirmed {
		return nil, apperror.BadRequest("Cannot add payment to this order")
	}

	settings, err := s.settingsRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	_, err = s.processPayments(ctx, order, req.Payments, settings)
	if err != nil {
		return nil, err
	}

	return s.GetByID(ctx, companyID, id)
}

// CompleteOrder completes an order
func (s *OrderService) CompleteOrder(ctx context.Context, companyID, id int64, req dto.CompleteOrderRequest) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("Order not found")
		}
		return nil, apperror.InternalError(err)
	}

	// Validate state transition
	if order.Status != model.OrderStatusConfirmed {
		return nil, apperror.BadRequest("Only confirmed orders can be completed")
	}

	// Check payment status
	if order.PaymentStatus != model.PaymentStatusPaid {
		return nil, apperror.BadRequest("Order must be fully paid before completion")
	}

	// For delivery orders, validate fulfillment
	if order.FulfillmentType == model.FulfillmentTypeDelivery {
		if req.FulfillmentStatus != nil {
			order.FulfillmentStatus = req.FulfillmentStatus
		}
	}

	// Update order
	now := time.Now()
	order.Status = model.OrderStatusCompleted
	order.CompletedAt = &now

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Deduct stock (best-effort — order is already completed)
	if err := s.deductStockForOrder(ctx, order); err != nil {
		slog.Error("failed to deduct stock on order completion", "order_id", order.ID, "error", err)
	}

	return s.GetByID(ctx, companyID, id)
}

// CancelOrder cancels an order (for unpaid or partially paid orders)
func (s *OrderService) CancelOrder(ctx context.Context, companyID, id int64, req dto.CancelOrderRequest) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("Order not found")
		}
		return nil, apperror.InternalError(err)
	}

	// Can only cancel draft or confirmed orders
	if order.Status != model.OrderStatusDraft && order.Status != model.OrderStatusConfirmed {
		return nil, apperror.BadRequest("Cannot cancel this order")
	}

	// If order has payments, use void instead
	if order.PaymentStatus == model.PaymentStatusPaid {
		return nil, apperror.BadRequest("Cannot cancel a fully paid order. Use void with refund instead.")
	}

	now := time.Now()
	order.Status = model.OrderStatusCancelled
	order.CancelledAt = &now
	order.CancelReason = &req.Reason

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.GetByID(ctx, companyID, id)
}

// VoidOrder voids a completed order (requires refund processing)
func (s *OrderService) VoidOrder(ctx context.Context, companyID, id int64, req dto.VoidOrderRequest) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("Order not found")
		}
		return nil, apperror.InternalError(err)
	}

	// Can void confirmed or completed orders
	if order.Status != model.OrderStatusConfirmed && order.Status != model.OrderStatusCompleted {
		return nil, apperror.BadRequest("Cannot void this order")
	}

	wasCompleted := order.Status == model.OrderStatusCompleted

	now := time.Now()
	order.Status = model.OrderStatusVoided
	order.VoidedAt = &now
	order.VoidReason = &req.Reason

	// Mark order payment status as refunded if there were payments
	if order.PaymentStatus == model.PaymentStatusPaid || order.PaymentStatus == model.PaymentStatusPartial {
		order.PaymentStatus = model.PaymentStatusRefunded
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Restore stock only if order was completed (stock was deducted)
	if wasCompleted {
		if err := s.restoreStockForOrder(ctx, order); err != nil {
			slog.Error("failed to restore stock on order void", "order_id", order.ID, "error", err)
		}
	}

	return s.GetByID(ctx, companyID, id)
}

// RefundPayment processes a refund for a specific payment
func (s *OrderService) RefundPayment(ctx context.Context, companyID, orderID int64, req dto.RefundPaymentRequest) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, companyID, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("Order not found")
		}
		return nil, apperror.InternalError(err)
	}

	// Get payment
	payment, err := s.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("Payment not found")
		}
		return nil, apperror.InternalError(err)
	}

	// Verify payment belongs to order
	if payment.OrderID != orderID {
		return nil, apperror.BadRequest("Payment does not belong to this order")
	}

	// Validate refund amount
	availableForRefund := payment.Amount - payment.RefundedAmount
	if req.Amount > availableForRefund {
		return nil, apperror.BadRequest("Refund amount exceeds available amount")
	}

	// Update payment
	now := time.Now()
	payment.RefundedAmount += req.Amount
	payment.RefundedAt = &now
	payment.RefundReason = &req.RefundReason

	if payment.RefundedAmount >= payment.Amount {
		payment.Status = model.PaymentStatusRefunded
	} else {
		payment.Status = model.PaymentStatusPartiallyRefunded
	}

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Update order payment status
	totalPaid, _ := s.paymentRepo.GetTotalPaidByOrderID(ctx, orderID)
	s.updateOrderPaymentStatus(ctx, order, totalPaid)

	return s.GetByID(ctx, companyID, orderID)
}

// UpdateOrder updates a draft order
func (s *OrderService) UpdateOrder(ctx context.Context, companyID, id int64, req dto.UpdateOrderRequest) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("Order not found")
		}
		return nil, apperror.InternalError(err)
	}

	// Only draft orders can be updated
	if order.Status != model.OrderStatusDraft {
		return nil, apperror.BadRequest("Only draft orders can be updated")
	}

	// Update fields
	if req.CustomerID != nil {
		order.CustomerID = req.CustomerID
	}
	if req.FulfillmentType != nil {
		order.FulfillmentType = *req.FulfillmentType
	}
	if req.ShippingAddress != nil {
		order.ShippingAddress = req.ShippingAddress
	}
	if req.Notes != nil {
		order.Notes = req.Notes
	}

	// Handle items update if provided
	if req.Items != nil && len(req.Items) > 0 {
		// Delete existing items
		if err := s.orderItemRepo.DeleteByOrderID(ctx, order.ID); err != nil {
			return nil, apperror.InternalError(err)
		}

		settings, _ := s.settingsRepo.GetByCompanyID(ctx, companyID)

		// Recalculate and create new items
		var totalAmount, totalDiscount, totalTax float64
		for _, itemInput := range req.Items {
			if itemInput.ProductVariantID == nil {
				return nil, apperror.BadRequest("Product variant is required")
			}

			variant, err := s.variantRepo.GetByID(ctx, *itemInput.ProductVariantID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil, apperror.NotFound("Product variant not found")
				}
				return nil, apperror.InternalError(err)
			}

			unitPrice := variant.Price
			quantity := float64(itemInput.Quantity)
			discountAmount := itemInput.DiscountAmount
			subtotal := (unitPrice * quantity) - discountAmount

			var taxAmount float64
			if settings.TaxEnabled {
				if settings.TaxInclusive {
					taxAmount = subtotal * settings.TaxRate / (100 + settings.TaxRate)
				} else {
					taxAmount = subtotal * settings.TaxRate / 100
				}
			}

			cogsAmount := variant.StandardCost * quantity

			totalAmount += unitPrice * quantity
			totalDiscount += discountAmount
			totalTax += taxAmount

			item := &model.OrderItem{
				OrderID:           order.ID,
				ProductID:         &variant.ProductID,
				ProductVariantID:  itemInput.ProductVariantID,
				SKU:               variant.SKU,
				ProductName:       variant.ProductName,
				VariantName:       &variant.Name,
				VariantAttributes: variant.Attributes,
				UnitPrice:         unitPrice,
				UnitCost:          variant.StandardCost,
				Quantity:          itemInput.Quantity,
				DiscountAmount:    discountAmount,
				TaxAmount:         taxAmount,
				Subtotal:          subtotal,
				CogsAmount:        cogsAmount,
			}

			if err := s.orderItemRepo.Create(ctx, item); err != nil {
				return nil, apperror.InternalError(err)
			}
		}

		order.TotalAmount = totalAmount
		order.TotalDiscount = totalDiscount
		order.TotalTax = totalTax
		order.GrandTotal = totalAmount - totalDiscount + totalTax

		if settings.RoundingEnabled && settings.RoundingAmount > 0 {
			order.GrandTotal = roundToNearest(order.GrandTotal, settings.RoundingAmount)
		}
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.GetByID(ctx, companyID, id)
}

func (s *OrderService) GetByID(ctx context.Context, companyID, id int64) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, companyID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("Order not found")
		}
		return nil, apperror.InternalError(err)
	}

	// Load items
	items, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	order.Items = items

	// Load payments
	payments, err := s.paymentRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	order.Payments = payments

	return s.toResponse(ctx, companyID, order), nil
}

func (s *OrderService) List(ctx context.Context, companyID int64, req dto.ListOrderRequest) (*dto.OrderListResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 {
		req.PerPage = 20
	}
	if req.PerPage > 100 {
		req.PerPage = 100
	}

	offset := (req.Page - 1) * req.PerPage

	params := &repository.OrderListParams{
		Search:            req.Search,
		Status:            req.Status,
		PaymentStatus:     req.PaymentStatus,
		FulfillmentType:   req.FulfillmentType,
		FulfillmentStatus: req.FulfillmentStatus,
		CustomerID:        req.CustomerID,
		DateFrom:          req.DateFrom,
		DateTo:            req.DateTo,
		Limit:             req.PerPage,
		Offset:            offset,
	}

	orders, total, err := s.orderRepo.List(ctx, companyID, params)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	responses := make([]*dto.OrderResponse, 0, len(orders))
	for _, order := range orders {
		items, _ := s.orderItemRepo.GetByOrderID(ctx, order.ID)
		order.Items = items

		payments, _ := s.paymentRepo.GetByOrderID(ctx, order.ID)
		order.Payments = payments

		responses = append(responses, s.toResponse(ctx, companyID, order))
	}

	// Calculate pagination
	totalPages := (total + req.PerPage - 1) / req.PerPage
	var nextPage, prevPage *int
	if req.Page < totalPages {
		next := req.Page + 1
		nextPage = &next
	}
	if req.Page > 1 {
		prev := req.Page - 1
		prevPage = &prev
	}

	return &dto.OrderListResponse{
		Orders: responses,
		Pagination: &dto.PaginationMeta{
			TotalRecords: total,
			TotalPages:   totalPages,
			CurrentPage:  req.Page,
			PerPage:      req.PerPage,
			NextPage:     nextPage,
			PrevPage:     prevPage,
		},
	}, nil
}

// evaluatePromotion finds the best applicable promotion for an order.
// If promoCode is provided, validates that specific promo.
// Otherwise auto-applies the highest-discount active promotion.
func (s *OrderService) evaluatePromotion(ctx context.Context, companyID int64, promoCode *string, subtotalAfterItemDiscounts float64) (*model.Promotion, float64) {
	now := time.Now()
	var candidates []*model.Promotion

	if promoCode != nil && *promoCode != "" {
		p, err := s.promotionRepo.GetByCode(ctx, companyID, *promoCode)
		if err != nil || p == nil || !p.IsActive {
			return nil, 0
		}
		if p.StartAt.After(now) || (p.EndAt != nil && p.EndAt.Before(now)) {
			return nil, 0
		}
		candidates = []*model.Promotion{p}
	} else {
		var err error
		candidates, err = s.promotionRepo.GetActivePromotions(ctx, companyID, now)
		if err != nil {
			return nil, 0
		}
	}

	var bestPromo *model.Promotion
	var bestDiscount float64

	for _, p := range candidates {
		if p.DiscountType == nil || p.DiscountValue == nil {
			continue
		}
		if p.MinPurchase != nil && subtotalAfterItemDiscounts < *p.MinPurchase {
			continue
		}

		var discount float64
		switch *p.DiscountType {
		case "percentage":
			discount = subtotalAfterItemDiscounts * (*p.DiscountValue) / 100
			if p.MaxDiscount != nil && discount > *p.MaxDiscount {
				discount = *p.MaxDiscount
			}
		case "fixed":
			discount = *p.DiscountValue
			if discount > subtotalAfterItemDiscounts {
				discount = subtotalAfterItemDiscounts
			}
		}

		if discount > bestDiscount {
			bestDiscount = discount
			bestPromo = p
		}
	}

	return bestPromo, bestDiscount
}

// processPayments handles payment creation and order status updates
func (s *OrderService) processPayments(ctx context.Context, order *model.Order, paymentInputs []dto.PaymentInput, settings *model.CompanySettings) ([]*model.Payment, error) {
	payments := make([]*model.Payment, 0, len(paymentInputs))
	now := time.Now()

	for _, input := range paymentInputs {
		payment := &model.Payment{
			OrderID:     order.ID,
			Method:      input.Method,
			Amount:      input.Amount,
			ReferenceNo: input.ReferenceNo,
			Status:      model.PaymentStatusCompleted,
			PaidAt:      now,
		}

		if err := s.paymentRepo.Create(ctx, payment); err != nil {
			return nil, apperror.InternalError(err)
		}
		payments = append(payments, payment)
	}

	// Calculate total paid
	totalPaid, err := s.paymentRepo.GetTotalPaidByOrderID(ctx, order.ID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	// Update order payment status
	s.updateOrderPaymentStatus(ctx, order, totalPaid)

	// Auto-complete counter orders when fully paid (if setting enabled)
	autoCompleted := false
	if settings.AutoCompleteCounterOrders &&
		order.FulfillmentType == model.FulfillmentTypeCounter &&
		order.PaymentStatus == model.PaymentStatusPaid &&
		(order.Status == model.OrderStatusDraft || order.Status == model.OrderStatusConfirmed) {

		// Auto-confirm if still draft
		if order.Status == model.OrderStatusDraft {
			order.Status = model.OrderStatusConfirmed
			order.ConfirmedAt = &now
		}

		// Auto-complete
		order.Status = model.OrderStatusCompleted
		order.CompletedAt = &now
		order.PaidAt = &now
		autoCompleted = true
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Deduct stock on auto-complete (best-effort)
	if autoCompleted {
		if err := s.deductStockForOrder(ctx, order); err != nil {
			slog.Error("failed to deduct stock on auto-complete", "order_id", order.ID, "error", err)
		}
	}

	return payments, nil
}

func (s *OrderService) updateOrderPaymentStatus(ctx context.Context, order *model.Order, totalPaid float64) {
	if totalPaid <= 0 {
		order.PaymentStatus = model.PaymentStatusUnpaid
		order.PaidAt = nil
	} else if totalPaid >= order.GrandTotal {
		order.PaymentStatus = model.PaymentStatusPaid
		if order.PaidAt == nil {
			now := time.Now()
			order.PaidAt = &now
		}
	} else {
		order.PaymentStatus = model.PaymentStatusPartial
		order.PaidAt = nil
	}
}

func (s *OrderService) toResponse(ctx context.Context, companyID int64, order *model.Order) *dto.OrderResponse {
	resp := &dto.OrderResponse{
		ID:                order.ID,
		OrderNo:           order.OrderNo,
		CustomerID:        order.CustomerID,
		CashierID:         order.CashierID,
		Status:            order.Status,
		TotalAmount:       order.TotalAmount,
		TotalDiscount:     order.TotalDiscount,
		TotalTax:          order.TotalTax,
		GrandTotal:        order.GrandTotal,
		Notes:             order.Notes,
		PaymentStatus:     order.PaymentStatus,
		FulfillmentType:   order.FulfillmentType,
		FulfillmentStatus: order.FulfillmentStatus,
		ShippingAddress:   order.ShippingAddress,
		ConfirmedAt:       order.ConfirmedAt,
		PaidAt:            order.PaidAt,
		CompletedAt:       order.CompletedAt,
		CancelledAt:       order.CancelledAt,
		VoidedAt:          order.VoidedAt,
		CancelReason:      order.CancelReason,
		VoidReason:        order.VoidReason,
		OfflineID:         order.OfflineID,
		SyncedAt:          order.SyncedAt,
		CreatedAt:         order.CreatedAt,
		UpdatedAt:         order.UpdatedAt,
	}

	// Calculate paid amounts
	var paidAmount, refundedTotal float64
	if order.Payments != nil {
		for _, p := range order.Payments {
			if p.Status != model.PaymentStatusFailed {
				paidAmount += p.Amount - p.RefundedAmount
				refundedTotal += p.RefundedAmount
			}
		}
	}
	resp.PaidAmount = paidAmount
	resp.BalanceDue = order.GrandTotal - paidAmount
	resp.RefundedTotal = refundedTotal

	// Get customer name
	if order.CustomerID != nil {
		customer, err := s.customerRepo.GetByID(ctx, companyID, *order.CustomerID)
		if err == nil && customer != nil {
			resp.CustomerName = &customer.Name
		}
	}

	// Get cashier name
	cashier, err := s.userRepo.GetByID(ctx, companyID, order.CashierID)
	if err == nil && cashier != nil {
		resp.CashierName = cashier.Name
	}

	// Map items
	if order.Items != nil {
		resp.Items = make([]*dto.OrderItemResponse, 0, len(order.Items))
		for _, item := range order.Items {
			resp.Items = append(resp.Items, &dto.OrderItemResponse{
				ID:                item.ID,
				ProductID:         item.ProductID,
				ProductVariantID:  item.ProductVariantID,
				SKU:               item.SKU,
				ProductName:       item.ProductName,
				VariantName:       item.VariantName,
				VariantAttributes: item.VariantAttributes,
				UnitPrice:         item.UnitPrice,
				UnitCost:          item.UnitCost,
				Quantity:          item.Quantity,
				DiscountAmount:    item.DiscountAmount,
				TaxAmount:         item.TaxAmount,
				Subtotal:          item.Subtotal,
				CogsAmount:        item.CogsAmount,
			})
		}
	}

	// Map payments
	if order.Payments != nil {
		resp.Payments = make([]*dto.PaymentResponse, 0, len(order.Payments))
		for _, payment := range order.Payments {
			resp.Payments = append(resp.Payments, &dto.PaymentResponse{
				ID:             payment.ID,
				Method:         payment.Method,
				Amount:         payment.Amount,
				ReferenceNo:    payment.ReferenceNo,
				Status:         payment.Status,
				RefundedAmount: payment.RefundedAmount,
				RefundedAt:     payment.RefundedAt,
				RefundReason:   payment.RefundReason,
				PaidAt:         payment.PaidAt,
			})
		}
	}

	// Populate applied promotion if present
	if len(order.AppliedPromotions) > 0 {
		ap := &dto.AppliedPromotionDTO{}
		if id, ok := order.AppliedPromotions["id"]; ok {
			switch v := id.(type) {
			case float64:
				ap.ID = int64(v)
			case int64:
				ap.ID = v
			}
		}
		if code, ok := order.AppliedPromotions["code"].(string); ok {
			ap.Code = code
		}
		if name, ok := order.AppliedPromotions["name"].(string); ok {
			ap.Name = name
		}
		if disc, ok := order.AppliedPromotions["discount_amount"]; ok {
			switch v := disc.(type) {
			case float64:
				ap.DiscountAmount = v
			}
		}
		if ap.Code != "" {
			resp.AppliedPromotion = ap
		}
	}

	return resp
}

func roundToNearest(value, nearest float64) float64 {
	if nearest <= 0 {
		return value
	}
	return float64(int64((value+nearest/2)/nearest)) * nearest
}

// deductStockForOrder creates OUT movements for each order item (best-effort).
func (s *OrderService) deductStockForOrder(ctx context.Context, order *model.Order) error {
	items, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		return err
	}

	refType := "order"
	for _, item := range items {
		if item.ProductVariantID == nil {
			continue // skip custom/untracked items
		}
		variantID := *item.ProductVariantID

		current, err := s.stockRepo.GetByVariantAndBranch(ctx, variantID, order.BranchID)
		if err != nil {
			slog.Error("deductStock: get stock failed", "variant_id", variantID, "error", err)
			continue
		}

		currentQty := 0
		minQty := 0
		if current != nil {
			currentQty = current.Quantity
			minQty = current.MinQuantity
		}

		newQty := currentQty - item.Quantity
		if newQty < 0 {
			slog.Warn("deductStock: stock went negative (race condition), clamping to 0", "variant_id", variantID, "order_id", order.ID)
			newQty = 0
		}

		refID := order.ID
		movement := &model.StockMovement{
			ProductVariantID: variantID,
			BranchID:         order.BranchID,
			Type:             model.StockMovementOut,
			Quantity:         item.Quantity,
			StockBefore:      currentQty,
			StockAfter:       newQty,
			ReferenceType:    &refType,
			ReferenceID:      &refID,
		}
		if err := s.movementRepo.Create(ctx, movement); err != nil {
			slog.Error("deductStock: create movement failed", "variant_id", variantID, "error", err)
			continue
		}

		stock := &model.Stock{
			ProductVariantID: variantID,
			BranchID:         order.BranchID,
			Quantity:         newQty,
			MinQuantity:      minQty,
		}
		if err := s.stockRepo.Upsert(ctx, stock); err != nil {
			slog.Error("deductStock: upsert stock failed", "variant_id", variantID, "error", err)
		}
	}
	return nil
}

// restoreStockForOrder creates IN movements to reverse a completed order's stock deductions.
func (s *OrderService) restoreStockForOrder(ctx context.Context, order *model.Order) error {
	items, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		return err
	}

	refType := "order_void"
	for _, item := range items {
		if item.ProductVariantID == nil {
			continue
		}
		variantID := *item.ProductVariantID

		current, err := s.stockRepo.GetByVariantAndBranch(ctx, variantID, order.BranchID)
		if err != nil {
			slog.Error("restoreStock: get stock failed", "variant_id", variantID, "error", err)
			continue
		}

		currentQty := 0
		minQty := 0
		if current != nil {
			currentQty = current.Quantity
			minQty = current.MinQuantity
		}

		newQty := currentQty + item.Quantity
		refID := order.ID
		movement := &model.StockMovement{
			ProductVariantID: variantID,
			BranchID:         order.BranchID,
			Type:             model.StockMovementIn,
			Quantity:         item.Quantity,
			StockBefore:      currentQty,
			StockAfter:       newQty,
			ReferenceType:    &refType,
			ReferenceID:      &refID,
		}
		if err := s.movementRepo.Create(ctx, movement); err != nil {
			slog.Error("restoreStock: create movement failed", "variant_id", variantID, "error", err)
			continue
		}

		stock := &model.Stock{
			ProductVariantID: variantID,
			BranchID:         order.BranchID,
			Quantity:         newQty,
			MinQuantity:      minQty,
		}
		if err := s.stockRepo.Upsert(ctx, stock); err != nil {
			slog.Error("restoreStock: upsert stock failed", "variant_id", variantID, "error", err)
		}
	}
	return nil
}
