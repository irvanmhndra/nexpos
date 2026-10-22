package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/irvanmhndra/nexpos-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// orderTestData holds all IDs needed for order tests.
type orderTestData struct {
	auth       *testutil.AuthContext
	customerID int64
	categoryID int64
	productID  int64
	variantID  int64
}

// setupOrderTest creates all prerequisite data for order tests via the API.
func setupOrderTest(t *testing.T) *orderTestData {
	t.Helper()
	cleanupDatabase(t)
	auth := registerTestUser(t)

	// Create customer via API
	resp := doPost(t, "/api/v1/customers", map[string]interface{}{"code": "ORD-CUST", "name": "Order Test Customer"}, auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create customer failed: %s", string(resp.Body))
	r := decodeResponse(t, resp)
	var custData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &custData))
	customerID := int64(custData["id"].(float64))

	// Create category + product via API helpers
	categoryID := createTestCategory(t, auth.Token)
	productID, variantID := createTestProduct(t, auth.Token, categoryID, "Order Test Product", "ORD-SKU-001", 100.00, 50.00)

	// Seed initial stock (100 units) so order creation doesn't fail on stock validation
	err := testEnv.Fixtures.CreateStock(context.Background(), variantID, auth.BranchID, 100, 0)
	require.NoError(t, err)

	return &orderTestData{
		auth:       auth,
		customerID: customerID,
		categoryID: categoryID,
		productID:  productID,
		variantID:  variantID,
	}
}

func TestOrderFlow_CreateAndGet(t *testing.T) {
	td := setupOrderTest(t)

	// Create an order
	resp := doPost(t, "/api/v1/orders", map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           2,
				"discount_amount":    0,
			},
		},
	}, td.auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	assert.True(t, r.Success)
	assert.Equal(t, "Order created successfully", r.Message)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	orderID := int64(data["id"].(float64))
	assert.NotZero(t, orderID)
	assert.NotEmpty(t, data["order_no"])
	assert.Equal(t, "draft", data["status"])
	assert.Equal(t, "unpaid", data["payment_status"])
	assert.Equal(t, "counter", data["fulfillment_type"])

	// Verify total calculations (2 items x 100.00 price)
	assert.Equal(t, float64(200), data["total_amount"])
	assert.Equal(t, float64(200), data["grand_total"])

	// Get the order
	resp = doGet(t, fmt.Sprintf("/api/v1/orders/%d", orderID), td.auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)

	var getData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &getData))
	assert.Equal(t, float64(orderID), getData["id"])
}

func TestOrderFlow_CreateConfirmPayComplete(t *testing.T) {
	td := setupOrderTest(t)
	ctx := context.Background()
	token := td.auth.Token

	// Seed initial stock (10 units)
	err := testEnv.Fixtures.CreateStock(ctx, td.variantID, td.auth.BranchID, 10, 0)
	require.NoError(t, err)

	// Step 1: Create an order (delivery to avoid auto-complete on payment)
	resp := doPost(t, "/api/v1/orders", map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "delivery",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	orderID := int64(data["id"].(float64))
	grandTotal := data["grand_total"].(float64)

	assert.Equal(t, "draft", data["status"])

	// Step 2: Confirm the order
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)

	var confirmData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &confirmData))
	assert.Equal(t, "confirmed", confirmData["status"])
	assert.NotNil(t, confirmData["confirmed_at"])

	// Step 3: Add payment
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/payments", orderID), map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "cash",
				"amount": grandTotal,
			},
		},
	}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "payment failed: %s", string(resp.Body))

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)

	var paymentData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &paymentData))
	assert.Equal(t, "paid", paymentData["payment_status"])
	assert.NotNil(t, paymentData["paid_at"])

	// Step 4: Complete the order
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/complete", orderID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "complete failed: %s", string(resp.Body))

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)

	var completeData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &completeData))
	assert.Equal(t, "completed", completeData["status"])
	assert.NotNil(t, completeData["completed_at"])

	// DB assertion: verify stock was deducted
	var stockQty int
	err = testEnv.DB.QueryRowContext(ctx,
		"SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2",
		td.variantID, td.auth.BranchID,
	).Scan(&stockQty)
	require.NoError(t, err)
	assert.Equal(t, 9, stockQty) // started at 10, ordered 1

	// DB assertion: verify stock_movement OUT record was created
	var movementType, refType string
	var movementQty, stockBefore, stockAfter int
	err = testEnv.DB.QueryRowContext(ctx,
		`SELECT type, quantity, stock_before, stock_after, reference_type
		 FROM stock_movements
		 WHERE reference_id = $1 AND reference_type = 'order'`,
		orderID,
	).Scan(&movementType, &movementQty, &stockBefore, &stockAfter, &refType)
	require.NoError(t, err)
	assert.Equal(t, "OUT", movementType)
	assert.Equal(t, 1, movementQty)
	assert.Equal(t, 10, stockBefore)
	assert.Equal(t, 9, stockAfter)
	assert.Equal(t, "order", refType)

	// DB assertion: verify exactly 1 stock movement for this order
	var movementCount int
	err = testEnv.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM stock_movements WHERE reference_id = $1 AND reference_type = 'order'",
		orderID,
	).Scan(&movementCount)
	require.NoError(t, err)
	assert.Equal(t, 1, movementCount)
}

func TestOrderFlow_Cancel(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token

	// Create an order
	resp := doPost(t, "/api/v1/orders", map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	orderID := int64(data["id"].(float64))

	// Cancel the order
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/cancel", orderID), map[string]interface{}{
		"reason": "Customer changed their mind",
	}, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)

	var cancelData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &cancelData))
	assert.Equal(t, "cancelled", cancelData["status"])
	assert.Equal(t, "Customer changed their mind", cancelData["cancel_reason"])
	assert.NotNil(t, cancelData["cancelled_at"])
}

func TestOrderFlow_VoidConfirmed(t *testing.T) {
	td := setupOrderTest(t)
	ctx := context.Background()
	token := td.auth.Token

	// Seed initial stock
	err := testEnv.Fixtures.CreateStock(ctx, td.variantID, td.auth.BranchID, 10, 0)
	require.NoError(t, err)

	// Create and confirm an order (not completed -- no stock deduction yet)
	resp := doPost(t, "/api/v1/orders", map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	orderID := int64(data["id"].(float64))

	// Confirm the order first
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	// Void the confirmed order (not completed, so no stock restoration)
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/void", orderID), map[string]interface{}{
		"reason": "Fraudulent order",
	}, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)

	var voidData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &voidData))
	assert.Equal(t, "voided", voidData["status"])
	assert.Equal(t, "Fraudulent order", voidData["void_reason"])
	assert.NotNil(t, voidData["voided_at"])

	// DB assertion: stock should remain unchanged (no deduction happened for confirmed-only orders)
	var stockQty int
	err = testEnv.DB.QueryRowContext(ctx,
		"SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2",
		td.variantID, td.auth.BranchID,
	).Scan(&stockQty)
	require.NoError(t, err)
	assert.Equal(t, 10, stockQty) // unchanged

	// DB assertion: no stock movements should exist
	var movementCount int
	err = testEnv.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM stock_movements WHERE reference_id = $1",
		orderID,
	).Scan(&movementCount)
	require.NoError(t, err)
	assert.Equal(t, 0, movementCount)
}

func TestOrderFlow_VoidCompleted(t *testing.T) {
	td := setupOrderTest(t)
	ctx := context.Background()
	token := td.auth.Token

	// Seed initial stock (10 units)
	err := testEnv.Fixtures.CreateStock(ctx, td.variantID, td.auth.BranchID, 10, 0)
	require.NoError(t, err)

	// Create -> Confirm -> Pay -> Complete an order (delivery to avoid auto-complete)
	resp := doPost(t, "/api/v1/orders", map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "delivery",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           2,
				"discount_amount":    0,
			},
		},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	orderID := int64(data["id"].(float64))
	grandTotal := data["grand_total"].(float64)

	// Confirm
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	// Pay
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/payments", orderID), map[string]interface{}{
		"payments": []map[string]interface{}{
			{"method": "cash", "amount": grandTotal},
		},
	}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "payment failed: %s", string(resp.Body))

	// Complete
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/complete", orderID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "complete failed: %s", string(resp.Body))

	// Verify stock was deducted after completion
	var stockQtyAfterComplete int
	err = testEnv.DB.QueryRowContext(ctx,
		"SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2",
		td.variantID, td.auth.BranchID,
	).Scan(&stockQtyAfterComplete)
	require.NoError(t, err)
	assert.Equal(t, 8, stockQtyAfterComplete) // 10 - 2

	// Now void the completed order
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/void", orderID), map[string]interface{}{
		"reason": "Customer returned items",
	}, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	var voidData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &voidData))
	assert.Equal(t, "voided", voidData["status"])

	// DB assertion: stock should be restored
	var stockQtyAfterVoid int
	err = testEnv.DB.QueryRowContext(ctx,
		"SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2",
		td.variantID, td.auth.BranchID,
	).Scan(&stockQtyAfterVoid)
	require.NoError(t, err)
	assert.Equal(t, 10, stockQtyAfterVoid) // restored to original

	// DB assertion: should have both OUT (order) and IN (order_void) movements
	var outCount, inCount int
	err = testEnv.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM stock_movements WHERE reference_id = $1 AND reference_type = 'order' AND type = 'OUT'",
		orderID,
	).Scan(&outCount)
	require.NoError(t, err)
	assert.Equal(t, 1, outCount)

	err = testEnv.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM stock_movements WHERE reference_id = $1 AND reference_type = 'order_void' AND type = 'IN'",
		orderID,
	).Scan(&inCount)
	require.NoError(t, err)
	assert.Equal(t, 1, inCount)

	// DB assertion: IN movement should show correct stock restoration
	var inMovementQty, inStockBefore, inStockAfter int
	err = testEnv.DB.QueryRowContext(ctx,
		`SELECT quantity, stock_before, stock_after FROM stock_movements
		 WHERE reference_id = $1 AND reference_type = 'order_void'`,
		orderID,
	).Scan(&inMovementQty, &inStockBefore, &inStockAfter)
	require.NoError(t, err)
	assert.Equal(t, 2, inMovementQty)
	assert.Equal(t, 8, inStockBefore)
	assert.Equal(t, 10, inStockAfter)
}

func TestOrderFlow_Refund(t *testing.T) {
	td := setupOrderTest(t)
	ctx := context.Background()
	token := td.auth.Token

	// Create, confirm, and pay an order
	resp := doPost(t, "/api/v1/orders", map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	orderID := int64(data["id"].(float64))
	grandTotal := data["grand_total"].(float64)

	// Confirm
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	// Pay
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/payments", orderID), map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "cash",
				"amount": grandTotal,
			},
		},
	}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "payment failed: %s", string(resp.Body))

	r = decodeResponse(t, resp)
	var paymentData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &paymentData))
	payments := paymentData["payments"].([]interface{})
	payment := payments[0].(map[string]interface{})
	paymentID := int64(payment["id"].(float64))

	// Refund
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/refund", orderID), map[string]interface{}{
		"payment_id":    paymentID,
		"amount":        grandTotal,
		"refund_reason": "Customer returned item",
	}, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)

	var refundData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &refundData))
	assert.Equal(t, grandTotal, refundData["refunded_total"])

	// DB assertion: verify payment status is 'refunded' in database
	var paymentStatus string
	err := testEnv.DB.QueryRowContext(ctx,
		"SELECT status FROM payments WHERE id = $1",
		paymentID,
	).Scan(&paymentStatus)
	require.NoError(t, err)
	assert.Equal(t, "refunded", paymentStatus)
}

func TestOrderFlow_ListOrders(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token

	// Create multiple orders
	for i := 0; i < 3; i++ {
		resp := doPost(t, "/api/v1/orders", map[string]interface{}{
			"customer_id":      td.customerID,
			"fulfillment_type": "counter",
			"items": []map[string]interface{}{
				{
					"product_variant_id": td.variantID,
					"quantity":           1,
					"discount_amount":    0,
				},
			},
		}, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))
	}

	// List orders
	resp := doGet(t, "/api/v1/orders", token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r := decodeResponse(t, resp)
	assert.True(t, r.Success)

	var items []map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &items))
	assert.Len(t, items, 3)

	// Verify pagination in meta
	var meta struct {
		Pagination struct {
			TotalRecords float64 `json:"total_records"`
		} `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(r.Meta, &meta))
	assert.Equal(t, float64(3), meta.Pagination.TotalRecords)
}

func TestOrderFlow_UpdateOrder(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token

	// Create an order
	resp := doPost(t, "/api/v1/orders", map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
		"notes": "Initial notes",
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	orderID := int64(data["id"].(float64))

	// Update the order
	resp = doPut(t, fmt.Sprintf("/api/v1/orders/%d", orderID), map[string]interface{}{
		"fulfillment_type": "delivery",
		"shipping_address": "123 Main St, City",
		"notes":            "Updated notes",
	}, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)

	var updateData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &updateData))
	assert.Equal(t, "delivery", updateData["fulfillment_type"])
	assert.Equal(t, "123 Main St, City", updateData["shipping_address"])
	assert.Equal(t, "Updated notes", updateData["notes"])
}

func TestOrder_CreateWithMultipleItems(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token

	// Create additional product variant via API
	resp := doPost(t, "/api/v1/products", map[string]interface{}{
		"name":                "Second Product",
		"product_category_id": td.categoryID,
		"variants": []map[string]interface{}{
			{
				"sku":           "ORD-SKU-002",
				"name":          "Large",
				"price":         150.00,
				"standard_cost": 75.00,
				"is_default":    true,
			},
		},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create product failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var prodData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &prodData))
	variants := prodData["variants"].([]interface{})
	variant2ID := int64(variants[0].(map[string]interface{})["id"].(float64))

	// Seed stock for second variant
	err := testEnv.Fixtures.CreateStock(context.Background(), variant2ID, td.auth.BranchID, 100, 0)
	require.NoError(t, err)

	// Create order with multiple items
	resp = doPost(t, "/api/v1/orders", map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID, // 100.00 x 2 = 200.00
				"quantity":           2,
				"discount_amount":    10,
			},
			{
				"product_variant_id": variant2ID, // 150.00 x 1 = 150.00
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	r = decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))

	// Total: 200 + 150 = 350
	// Discount: 10
	// Grand total: 340
	assert.Equal(t, float64(350), data["total_amount"])
	assert.Equal(t, float64(10), data["total_discount"])
	assert.Equal(t, float64(340), data["grand_total"])

	items := data["items"].([]interface{})
	assert.Len(t, items, 2)
}

func TestOrder_ValidationErrors(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
	}{
		{
			name: "missing items",
			body: map[string]interface{}{
				"fulfillment_type": "counter",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "empty items array",
			body: map[string]interface{}{
				"fulfillment_type": "counter",
				"items":            []map[string]interface{}{},
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid fulfillment type",
			body: map[string]interface{}{
				"fulfillment_type": "invalid",
				"items": []map[string]interface{}{
					{
						"product_variant_id": 1,
						"quantity":           1,
					},
				},
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "zero quantity",
			body: map[string]interface{}{
				"fulfillment_type": "counter",
				"items": []map[string]interface{}{
					{
						"product_variant_id": 1,
						"quantity":           0,
					},
				},
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := doPost(t, "/api/v1/orders", tt.body, auth.Token)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestOrder_PaymentValidation(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token

	// Create and confirm an order
	resp := doPost(t, "/api/v1/orders", map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
			},
		},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	orderID := int64(data["id"].(float64))

	// Confirm
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	// Test invalid payment method
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/payments", orderID), map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "invalid_method",
				"amount": 100,
			},
		},
	}, token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestOrder_PartialPayment(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token

	// Create order (grand_total = 100)
	resp := doPost(t, "/api/v1/orders", map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
			},
		},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	orderID := int64(data["id"].(float64))

	// Confirm
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	// Pay partial amount (50 out of 100)
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/payments", orderID), map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "cash",
				"amount": 50,
			},
		},
	}, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	var paymentData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &paymentData))
	assert.Equal(t, "partial", paymentData["payment_status"])
	assert.Equal(t, float64(50), paymentData["paid_amount"])
	assert.Equal(t, float64(50), paymentData["balance_due"])

	// Pay remaining
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/payments", orderID), map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "credit_card",
				"amount": 50,
			},
		},
	}, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	require.NoError(t, json.Unmarshal(r.Data, &paymentData))
	assert.Equal(t, "paid", paymentData["payment_status"])
	assert.Equal(t, float64(100), paymentData["paid_amount"])
	assert.Equal(t, float64(0), paymentData["balance_due"])
}
