package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	repoMocks "github.com/irvanmhndra/nexpos-api/internal/repository/mocks"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// =======================
// Test Helpers
// =======================

type orderTestSetup struct {
	svc         *OrderService
	orderRepo   *repoMocks.MockOrderRepository
	itemRepo    *repoMocks.MockOrderItemRepository
	paymentRepo *repoMocks.MockPaymentRepository
	variantRepo *repoMocks.MockProductVariantRepository
	customerRepo *repoMocks.MockCustomerRepository
	userRepo    *repoMocks.MockUserRepository
	settingsRepo *repoMocks.MockCompanySettingsRepository
	stockRepo   *repoMocks.MockStockRepository
	movementRepo *repoMocks.MockStockMovementRepository
	promoRepo   *repoMocks.MockPromotionRepository
}

func setupOrderTest(t *testing.T) *orderTestSetup {
	t.Helper()
	s := &orderTestSetup{
		orderRepo:    repoMocks.NewMockOrderRepository(t),
		itemRepo:     repoMocks.NewMockOrderItemRepository(t),
		paymentRepo:  repoMocks.NewMockPaymentRepository(t),
		variantRepo:  repoMocks.NewMockProductVariantRepository(t),
		customerRepo: repoMocks.NewMockCustomerRepository(t),
		userRepo:     repoMocks.NewMockUserRepository(t),
		settingsRepo: repoMocks.NewMockCompanySettingsRepository(t),
		stockRepo:    repoMocks.NewMockStockRepository(t),
		movementRepo: repoMocks.NewMockStockMovementRepository(t),
		promoRepo:    repoMocks.NewMockPromotionRepository(t),
	}
	s.svc = NewOrderService(
		s.orderRepo, s.itemRepo, s.paymentRepo, s.variantRepo,
		s.customerRepo, s.userRepo, s.settingsRepo,
		s.stockRepo, s.movementRepo, s.promoRepo,
	)
	return s
}

func defaultSettings() *model.CompanySettings {
	return &model.CompanySettings{
		TaxEnabled:                false,
		TaxRate:                   0,
		TaxInclusive:              false,
		RoundingEnabled:           false,
		AutoCompleteCounterOrders: false,
	}
}

func testVariant(id int64, price float64) *model.ProductVariant {
	return &model.ProductVariant{
		ID:           id,
		ProductID:    10,
		SKU:          "SKU-001",
		Name:         "Default",
		ProductName:  "Test Product",
		Price:        price,
		StandardCost: price * 0.6,
	}
}

func testOrder(id, companyID, branchID int64, status string) *model.Order {
	return &model.Order{
		ID:            id,
		CompanyID:     companyID,
		BranchID:      branchID,
		OrderNo:       "ORD-001",
		CashierID:     99,
		Status:        status,
		PaymentStatus: model.PaymentStatusUnpaid,
		FulfillmentType: model.FulfillmentTypeCounter,
		GrandTotal:    100_000,
	}
}

// expectReload sets up the mock calls for GetByID (which loads items, payments,
// cashier name, and optionally customer name).
func (s *orderTestSetup) expectReload(ctx context.Context, companyID, orderID int64, order *model.Order) {
	s.orderRepo.EXPECT().
		GetByID(ctx, companyID, orderID).
		Return(order, nil).
		Once()
	s.itemRepo.EXPECT().
		GetByOrderID(ctx, orderID).
		Return([]*model.OrderItem{}, nil).
		Once()
	s.paymentRepo.EXPECT().
		GetByOrderID(ctx, orderID).
		Return([]*model.Payment{}, nil).
		Once()
	s.userRepo.EXPECT().
		GetByID(ctx, companyID, order.CashierID).
		Return(&model.User{ID: order.CashierID, Name: "Cashier"}, nil).
		Once()
}

// =======================
// Create Tests
// =======================

func TestOrderService_Create_Success(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, branchID, cashierID := int64(1), int64(2), int64(99)

	variantID := int64(5)
	variant := testVariant(variantID, 50_000)

	req := dto.CreateOrderRequest{
		Items: []dto.OrderItemInput{
			{ProductVariantID: &variantID, Quantity: 2},
		},
	}

	s.settingsRepo.EXPECT().GetByCompanyID(ctx, companyID).Return(defaultSettings(), nil).Once()
	s.orderRepo.EXPECT().GenerateOrderNo(ctx, companyID, branchID).Return("ORD-001", nil).Once()
	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(variant, nil).Once()
	s.stockRepo.EXPECT().GetByVariantAndBranch(ctx, variantID, branchID).
		Return(&model.Stock{Quantity: 10}, nil).Once()
	s.promoRepo.On("GetActivePromotions", ctx, companyID, mock.AnythingOfType("time.Time")).
		Return(([]*model.Promotion)(nil), nil).Once()
	s.orderRepo.EXPECT().Create(ctx, mock.MatchedBy(func(o *model.Order) bool {
		return o.CompanyID == companyID && o.GrandTotal == 100_000
	})).Return(nil).Once()
	s.itemRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.OrderItem")).Return(nil).Once()

	createdOrder := testOrder(0, companyID, branchID, model.OrderStatusDraft)
	createdOrder.GrandTotal = 100_000
	s.expectReload(ctx, companyID, 0, createdOrder)

	resp, err := s.svc.Create(ctx, companyID, branchID, cashierID, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "ORD-001", resp.OrderNo)
}

func TestOrderService_Create_EmptyItems(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()

	req := dto.CreateOrderRequest{Items: []dto.OrderItemInput{}}

	_, err := s.svc.Create(ctx, 1, 1, 99, req)

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "at least one item")
}

func TestOrderService_Create_InsufficientStock(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, branchID := int64(1), int64(2)
	variantID := int64(5)

	req := dto.CreateOrderRequest{
		Items: []dto.OrderItemInput{
			{ProductVariantID: &variantID, Quantity: 10},
		},
	}

	s.settingsRepo.EXPECT().GetByCompanyID(ctx, companyID).Return(defaultSettings(), nil).Once()
	s.orderRepo.EXPECT().GenerateOrderNo(ctx, companyID, branchID).Return("ORD-001", nil).Once()
	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(testVariant(variantID, 50_000), nil).Once()
	s.stockRepo.EXPECT().GetByVariantAndBranch(ctx, variantID, branchID).
		Return(&model.Stock{Quantity: 3}, nil).Once()

	_, err := s.svc.Create(ctx, companyID, branchID, 99, req)

	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
}

func TestOrderService_Create_VariantNotFound(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, branchID := int64(1), int64(2)
	variantID := int64(5)

	req := dto.CreateOrderRequest{
		Items: []dto.OrderItemInput{
			{ProductVariantID: &variantID, Quantity: 1},
		},
	}

	s.settingsRepo.EXPECT().GetByCompanyID(ctx, companyID).Return(defaultSettings(), nil).Once()
	s.orderRepo.EXPECT().GenerateOrderNo(ctx, companyID, branchID).Return("ORD-001", nil).Once()
	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(nil, sql.ErrNoRows).Once()

	_, err := s.svc.Create(ctx, companyID, branchID, 99, req)

	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

func TestOrderService_Create_CustomerNotFound(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID := int64(1)
	customerID := int64(42)

	req := dto.CreateOrderRequest{
		CustomerID: &customerID,
		Items:      []dto.OrderItemInput{{ProductVariantID: func() *int64 { v := int64(5); return &v }(), Quantity: 1}},
	}

	s.settingsRepo.EXPECT().GetByCompanyID(ctx, companyID).Return(defaultSettings(), nil).Once()
	s.customerRepo.EXPECT().GetByID(ctx, companyID, customerID).Return(nil, sql.ErrNoRows).Once()

	_, err := s.svc.Create(ctx, companyID, 2, 99, req)

	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

func TestOrderService_Create_WithTax_Exclusive(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, branchID := int64(1), int64(2)
	variantID := int64(5)

	settings := &model.CompanySettings{
		TaxEnabled:   true,
		TaxInclusive: false,
		TaxRate:      10, // 10%
	}

	req := dto.CreateOrderRequest{
		Items: []dto.OrderItemInput{{ProductVariantID: &variantID, Quantity: 1}},
	}

	variant := testVariant(variantID, 100_000)
	s.settingsRepo.EXPECT().GetByCompanyID(ctx, companyID).Return(settings, nil).Once()
	s.orderRepo.EXPECT().GenerateOrderNo(ctx, companyID, branchID).Return("ORD-001", nil).Once()
	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(variant, nil).Once()
	s.stockRepo.EXPECT().GetByVariantAndBranch(ctx, variantID, branchID).
		Return(&model.Stock{Quantity: 5}, nil).Once()
	s.promoRepo.On("GetActivePromotions", ctx, companyID, mock.AnythingOfType("time.Time")).
		Return(([]*model.Promotion)(nil), nil).Once()
	s.orderRepo.EXPECT().Create(ctx, mock.MatchedBy(func(o *model.Order) bool {
		// tax = 100000 * 10/100 = 10000; grand = 110000
		return o.TotalTax == 10_000 && o.GrandTotal == 110_000
	})).Return(nil).Once()
	s.itemRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.OrderItem")).Return(nil).Once()

	order := testOrder(0, companyID, branchID, model.OrderStatusDraft)
	order.GrandTotal = 110_000
	s.expectReload(ctx, companyID, 0, order)

	resp, err := s.svc.Create(ctx, companyID, branchID, 99, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestOrderService_Create_DeliveryRequiresCustomer(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID := int64(1)

	settings := &model.CompanySettings{RequireCustomerForDelivery: true}
	req := dto.CreateOrderRequest{
		FulfillmentType: model.FulfillmentTypeDelivery,
		Items:           []dto.OrderItemInput{{ProductVariantID: func() *int64 { v := int64(5); return &v }(), Quantity: 1}},
	}

	s.settingsRepo.EXPECT().GetByCompanyID(ctx, companyID).Return(settings, nil).Once()

	_, err := s.svc.Create(ctx, companyID, 2, 99, req)

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "Customer is required")
}

func TestOrderService_Create_WithPromoCode(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, branchID := int64(1), int64(2)
	variantID := int64(5)
	code := "DISC10"
	discountVal := 10.0
	promoNow := time.Now().Add(-time.Hour)

	promo := &model.Promotion{
		ID:            1,
		Code:          code,
		Name:          "10% Off",
		IsActive:      true,
		DiscountType:  strPtr("percentage"),
		DiscountValue: &discountVal,
		StartAt:       promoNow,
	}

	req := dto.CreateOrderRequest{
		Items:     []dto.OrderItemInput{{ProductVariantID: &variantID, Quantity: 1}},
		PromoCode: &code,
	}

	variant := testVariant(variantID, 100_000)
	s.settingsRepo.EXPECT().GetByCompanyID(ctx, companyID).Return(defaultSettings(), nil).Once()
	s.orderRepo.EXPECT().GenerateOrderNo(ctx, companyID, branchID).Return("ORD-001", nil).Once()
	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(variant, nil).Once()
	s.stockRepo.EXPECT().GetByVariantAndBranch(ctx, variantID, branchID).
		Return(&model.Stock{Quantity: 5}, nil).Once()
	s.promoRepo.EXPECT().GetByCode(ctx, companyID, code).Return(promo, nil).Once()
	s.orderRepo.EXPECT().Create(ctx, mock.MatchedBy(func(o *model.Order) bool {
		// discount = 10% of 100000 = 10000; grand = 90000
		return o.TotalDiscount == 10_000 && o.GrandTotal == 90_000
	})).Return(nil).Once()
	s.itemRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.OrderItem")).Return(nil).Once()

	order := testOrder(0, companyID, branchID, model.OrderStatusDraft)
	order.GrandTotal = 90_000
	s.expectReload(ctx, companyID, 0, order)

	resp, err := s.svc.Create(ctx, companyID, branchID, 99, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// =======================
// Preview Tests
// =======================

func TestOrderService_Preview_Success(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, branchID := int64(1), int64(2)
	variantID := int64(5)

	req := dto.PreviewOrderRequest{
		Items: []dto.OrderItemInput{{ProductVariantID: &variantID, Quantity: 2}},
	}

	s.settingsRepo.EXPECT().GetByCompanyID(ctx, companyID).Return(defaultSettings(), nil).Once()
	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(testVariant(variantID, 50_000), nil).Once()
	s.promoRepo.On("GetActivePromotions", ctx, companyID, mock.AnythingOfType("time.Time")).
		Return(([]*model.Promotion)(nil), nil).Once()

	resp, err := s.svc.Preview(ctx, companyID, branchID, req)

	require.NoError(t, err)
	assert.Equal(t, float64(100_000), resp.Subtotal)
	assert.Equal(t, float64(100_000), resp.GrandTotal)
}

func TestOrderService_Preview_EmptyItems(t *testing.T) {
	s := setupOrderTest(t)

	_, err := s.svc.Preview(context.Background(), 1, 2, dto.PreviewOrderRequest{})

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
}

// =======================
// ConfirmOrder Tests
// =======================

func TestOrderService_ConfirmOrder_Success(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusDraft)
	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()
	s.orderRepo.EXPECT().Update(ctx, mock.MatchedBy(func(o *model.Order) bool {
		return o.Status == model.OrderStatusConfirmed && o.ConfirmedAt != nil
	})).Return(nil).Once()
	s.expectReload(ctx, companyID, orderID, testOrder(orderID, companyID, 2, model.OrderStatusConfirmed))

	resp, err := s.svc.ConfirmOrder(ctx, companyID, orderID)

	require.NoError(t, err)
	assert.Equal(t, model.OrderStatusConfirmed, resp.Status)
}

func TestOrderService_ConfirmOrder_NotFound(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()

	s.orderRepo.EXPECT().GetByID(ctx, int64(1), int64(99)).Return(nil, sql.ErrNoRows).Once()

	_, err := s.svc.ConfirmOrder(ctx, 1, 99)

	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

func TestOrderService_ConfirmOrder_WrongStatus(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusConfirmed) // already confirmed
	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()

	_, err := s.svc.ConfirmOrder(ctx, companyID, orderID)

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
}

// =======================
// AddPayment Tests
// =======================

func TestOrderService_AddPayment_Success(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusDraft)
	order.GrandTotal = 100_000

	req := dto.AddPaymentRequest{
		Payments: []dto.PaymentInput{{Method: "cash", Amount: 100_000}},
	}

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()
	s.settingsRepo.EXPECT().GetByCompanyID(ctx, companyID).Return(defaultSettings(), nil).Once()
	s.paymentRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.Payment")).Return(nil).Once()
	s.paymentRepo.EXPECT().GetTotalPaidByOrderID(ctx, orderID).Return(100_000.0, nil).Once()
	s.orderRepo.EXPECT().Update(ctx, mock.MatchedBy(func(o *model.Order) bool {
		return o.PaymentStatus == model.PaymentStatusPaid
	})).Return(nil).Once()
	s.expectReload(ctx, companyID, orderID, testOrder(orderID, companyID, 2, model.OrderStatusDraft))

	resp, err := s.svc.AddPayment(ctx, companyID, orderID, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestOrderService_AddPayment_PartialPayment(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusDraft)
	order.GrandTotal = 100_000

	req := dto.AddPaymentRequest{
		Payments: []dto.PaymentInput{{Method: "cash", Amount: 50_000}},
	}

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()
	s.settingsRepo.EXPECT().GetByCompanyID(ctx, companyID).Return(defaultSettings(), nil).Once()
	s.paymentRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.Payment")).Return(nil).Once()
	s.paymentRepo.EXPECT().GetTotalPaidByOrderID(ctx, orderID).Return(50_000.0, nil).Once()
	s.orderRepo.EXPECT().Update(ctx, mock.MatchedBy(func(o *model.Order) bool {
		return o.PaymentStatus == model.PaymentStatusPartial
	})).Return(nil).Once()
	s.expectReload(ctx, companyID, orderID, testOrder(orderID, companyID, 2, model.OrderStatusDraft))

	resp, err := s.svc.AddPayment(ctx, companyID, orderID, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestOrderService_AddPayment_WrongStatus(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusCompleted)
	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()

	_, err := s.svc.AddPayment(ctx, companyID, orderID, dto.AddPaymentRequest{
		Payments: []dto.PaymentInput{{Method: "cash", Amount: 100_000}},
	})

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
}

// =======================
// CompleteOrder Tests
// =======================

func TestOrderService_CompleteOrder_Success(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusConfirmed)
	order.PaymentStatus = model.PaymentStatusPaid
	order.Items = []*model.OrderItem{}

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()
	s.orderRepo.EXPECT().Update(ctx, mock.MatchedBy(func(o *model.Order) bool {
		return o.Status == model.OrderStatusCompleted && o.CompletedAt != nil
	})).Return(nil).Once()
	// deductStockForOrder calls itemRepo.GetByOrderID internally (empty → no stock ops)
	s.itemRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]*model.OrderItem{}, nil).Once()
	s.expectReload(ctx, companyID, orderID, testOrder(orderID, companyID, 2, model.OrderStatusCompleted))

	resp, err := s.svc.CompleteOrder(ctx, companyID, orderID, dto.CompleteOrderRequest{})

	require.NoError(t, err)
	assert.Equal(t, model.OrderStatusCompleted, resp.Status)
}

func TestOrderService_CompleteOrder_NotPaid(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusConfirmed)
	order.PaymentStatus = model.PaymentStatusUnpaid

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()

	_, err := s.svc.CompleteOrder(ctx, companyID, orderID, dto.CompleteOrderRequest{})

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "fully paid")
}

func TestOrderService_CompleteOrder_WrongStatus(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusDraft)
	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()

	_, err := s.svc.CompleteOrder(ctx, companyID, orderID, dto.CompleteOrderRequest{})

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
}

// =======================
// CancelOrder Tests
// =======================

func TestOrderService_CancelOrder_Success(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusDraft)

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()
	s.orderRepo.EXPECT().Update(ctx, mock.MatchedBy(func(o *model.Order) bool {
		return o.Status == model.OrderStatusCancelled && o.CancelReason != nil
	})).Return(nil).Once()
	s.expectReload(ctx, companyID, orderID, testOrder(orderID, companyID, 2, model.OrderStatusCancelled))

	resp, err := s.svc.CancelOrder(ctx, companyID, orderID, dto.CancelOrderRequest{Reason: "customer request"})

	require.NoError(t, err)
	assert.Equal(t, model.OrderStatusCancelled, resp.Status)
}

func TestOrderService_CancelOrder_FullyPaid(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusConfirmed)
	order.PaymentStatus = model.PaymentStatusPaid

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()

	_, err := s.svc.CancelOrder(ctx, companyID, orderID, dto.CancelOrderRequest{Reason: "mistake"})

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "void")
}

func TestOrderService_CancelOrder_CompletedOrder(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusCompleted)
	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()

	_, err := s.svc.CancelOrder(ctx, companyID, orderID, dto.CancelOrderRequest{Reason: "mistake"})

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
}

// =======================
// VoidOrder Tests
// =======================

func TestOrderService_VoidOrder_FromConfirmed(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusConfirmed)
	order.PaymentStatus = model.PaymentStatusUnpaid

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()
	s.orderRepo.EXPECT().Update(ctx, mock.MatchedBy(func(o *model.Order) bool {
		return o.Status == model.OrderStatusVoided && o.VoidReason != nil
	})).Return(nil).Once()
	// wasCompleted = false → no stock restore
	s.expectReload(ctx, companyID, orderID, testOrder(orderID, companyID, 2, model.OrderStatusVoided))

	resp, err := s.svc.VoidOrder(ctx, companyID, orderID, dto.VoidOrderRequest{Reason: "test void"})

	require.NoError(t, err)
	assert.Equal(t, model.OrderStatusVoided, resp.Status)
}

func TestOrderService_VoidOrder_FromCompleted_RestoresStock(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	variantID := int64(5)
	order := testOrder(orderID, companyID, 2, model.OrderStatusCompleted)
	order.PaymentStatus = model.PaymentStatusPaid
	order.BranchID = 2
	order.Items = []*model.OrderItem{
		{ProductVariantID: &variantID, Quantity: 2, SKU: "SKU-001"},
	}

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()
	s.orderRepo.EXPECT().Update(ctx, mock.MatchedBy(func(o *model.Order) bool {
		return o.Status == model.OrderStatusVoided && o.PaymentStatus == model.PaymentStatusRefunded
	})).Return(nil).Once()

	// restoreStockForOrder calls itemRepo.GetByOrderID internally, then per-item stock ops
	s.itemRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]*model.OrderItem{
		{ProductVariantID: &variantID, Quantity: 2, SKU: "SKU-001"},
	}, nil).Once()
	s.stockRepo.EXPECT().
		GetByVariantAndBranch(ctx, variantID, int64(2)).
		Return(&model.Stock{Quantity: 0}, nil).Once()
	s.movementRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.StockMovement")).Return(nil).Once()
	s.stockRepo.EXPECT().Upsert(ctx, mock.AnythingOfType("*model.Stock")).Return(nil).Once()

	s.expectReload(ctx, companyID, orderID, testOrder(orderID, companyID, 2, model.OrderStatusVoided))

	resp, err := s.svc.VoidOrder(ctx, companyID, orderID, dto.VoidOrderRequest{Reason: "error"})

	require.NoError(t, err)
	assert.Equal(t, model.OrderStatusVoided, resp.Status)
}

func TestOrderService_VoidOrder_WrongStatus(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusDraft)
	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()

	_, err := s.svc.VoidOrder(ctx, companyID, orderID, dto.VoidOrderRequest{Reason: "test"})

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
}

// =======================
// RefundPayment Tests
// =======================

func TestOrderService_RefundPayment_Success(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID, paymentID := int64(1), int64(10), int64(20)

	order := testOrder(orderID, companyID, 2, model.OrderStatusCompleted)
	order.PaymentStatus = model.PaymentStatusPaid

	payment := &model.Payment{
		ID:             paymentID,
		OrderID:        orderID,
		Amount:         100_000,
		RefundedAmount: 0,
		Status:         model.PaymentStatusCompleted,
		PaidAt:         time.Now(),
	}

	req := dto.RefundPaymentRequest{
		PaymentID:    paymentID,
		Amount:       50_000,
		RefundReason: "partial refund",
	}

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()
	s.paymentRepo.EXPECT().GetByID(ctx, paymentID).Return(payment, nil).Once()
	s.paymentRepo.EXPECT().Update(ctx, mock.MatchedBy(func(p *model.Payment) bool {
		return p.RefundedAmount == 50_000 && p.Status == model.PaymentStatusPartiallyRefunded
	})).Return(nil).Once()
	s.paymentRepo.EXPECT().GetTotalPaidByOrderID(ctx, orderID).Return(50_000.0, nil).Once()
	// updateOrderPaymentStatus is in-memory only — no orderRepo.Update call in RefundPayment
	s.expectReload(ctx, companyID, orderID, testOrder(orderID, companyID, 2, model.OrderStatusCompleted))

	resp, err := s.svc.RefundPayment(ctx, companyID, orderID, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestOrderService_RefundPayment_FullRefund(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID, paymentID := int64(1), int64(10), int64(20)

	order := testOrder(orderID, companyID, 2, model.OrderStatusCompleted)
	payment := &model.Payment{
		ID: paymentID, OrderID: orderID, Amount: 100_000,
		Status: model.PaymentStatusCompleted, PaidAt: time.Now(),
	}

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()
	s.paymentRepo.EXPECT().GetByID(ctx, paymentID).Return(payment, nil).Once()
	s.paymentRepo.EXPECT().Update(ctx, mock.MatchedBy(func(p *model.Payment) bool {
		return p.RefundedAmount == 100_000 && p.Status == model.PaymentStatusRefunded
	})).Return(nil).Once()
	s.paymentRepo.EXPECT().GetTotalPaidByOrderID(ctx, orderID).Return(0.0, nil).Once()
	// updateOrderPaymentStatus is in-memory only — no orderRepo.Update call in RefundPayment
	s.expectReload(ctx, companyID, orderID, testOrder(orderID, companyID, 2, model.OrderStatusCompleted))

	resp, err := s.svc.RefundPayment(ctx, companyID, orderID, dto.RefundPaymentRequest{
		PaymentID: paymentID, Amount: 100_000, RefundReason: "customer complaint",
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestOrderService_RefundPayment_ExceedsAvailable(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID, paymentID := int64(1), int64(10), int64(20)

	order := testOrder(orderID, companyID, 2, model.OrderStatusCompleted)
	payment := &model.Payment{
		ID: paymentID, OrderID: orderID, Amount: 100_000, RefundedAmount: 80_000,
		PaidAt: time.Now(),
	}

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()
	s.paymentRepo.EXPECT().GetByID(ctx, paymentID).Return(payment, nil).Once()

	_, err := s.svc.RefundPayment(ctx, companyID, orderID, dto.RefundPaymentRequest{
		PaymentID: paymentID, Amount: 50_000, RefundReason: "too much",
	})

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "exceeds")
}

func TestOrderService_RefundPayment_WrongOrder(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID, paymentID := int64(1), int64(10), int64(20)

	order := testOrder(orderID, companyID, 2, model.OrderStatusCompleted)
	payment := &model.Payment{
		ID: paymentID, OrderID: 999, // belongs to different order
		Amount: 100_000, PaidAt: time.Now(),
	}

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()
	s.paymentRepo.EXPECT().GetByID(ctx, paymentID).Return(payment, nil).Once()

	_, err := s.svc.RefundPayment(ctx, companyID, orderID, dto.RefundPaymentRequest{
		PaymentID: paymentID, Amount: 50_000, RefundReason: "test",
	})

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
}

// =======================
// AutoComplete Tests
// =======================

func TestOrderService_AddPayment_AutoCompleteCounterOrder(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID, orderID := int64(1), int64(10)

	order := testOrder(orderID, companyID, 2, model.OrderStatusDraft)
	order.GrandTotal = 100_000
	order.Items = []*model.OrderItem{} // empty → deductStock is no-op

	settings := &model.CompanySettings{AutoCompleteCounterOrders: true}

	req := dto.AddPaymentRequest{
		Payments: []dto.PaymentInput{{Method: "cash", Amount: 100_000}},
	}

	s.orderRepo.EXPECT().GetByID(ctx, companyID, orderID).Return(order, nil).Once()
	s.settingsRepo.EXPECT().GetByCompanyID(ctx, companyID).Return(settings, nil).Once()
	s.paymentRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.Payment")).Return(nil).Once()
	s.paymentRepo.EXPECT().GetTotalPaidByOrderID(ctx, orderID).Return(100_000.0, nil).Once()
	s.orderRepo.EXPECT().Update(ctx, mock.MatchedBy(func(o *model.Order) bool {
		return o.Status == model.OrderStatusCompleted
	})).Return(nil).Once()
	// deductStockForOrder calls itemRepo.GetByOrderID internally (empty → no stock ops)
	s.itemRepo.EXPECT().GetByOrderID(ctx, orderID).Return([]*model.OrderItem{}, nil).Once()
	s.expectReload(ctx, companyID, orderID, testOrder(orderID, companyID, 2, model.OrderStatusCompleted))

	resp, err := s.svc.AddPayment(ctx, companyID, orderID, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// =======================
// List Tests
// =======================

func TestOrderService_List_DefaultPagination(t *testing.T) {
	s := setupOrderTest(t)
	ctx := context.Background()
	companyID := int64(1)

	s.orderRepo.EXPECT().List(ctx, companyID, mock.MatchedBy(func(p interface{}) bool {
		return true // pagination handled internally
	})).Return([]*model.Order{}, 0, nil).Once()

	resp, err := s.svc.List(ctx, companyID, dto.ListOrderRequest{Page: 0, PerPage: 0})

	require.NoError(t, err)
	assert.Empty(t, resp.Orders)
	assert.Equal(t, 1, resp.Pagination.CurrentPage)
	assert.Equal(t, 20, resp.Pagination.PerPage)
}
