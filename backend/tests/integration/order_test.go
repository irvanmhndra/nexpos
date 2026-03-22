package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// orderTestData holds all IDs needed for order tests.
type orderTestData struct {
	auth       *authContext
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
	custBody := map[string]interface{}{"code": "ORD-CUST", "name": "Order Test Customer"}
	resp, err := testServer.POST("/api/v1/customers", custBody, auth.Token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create customer failed: %s", string(resp.Body))
	var custResp map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &custResp))
	customerID := int64(custResp["data"].(map[string]interface{})["id"].(float64))

	// Create category + product via API helpers
	categoryID := createTestCategory(t, auth.Token)
	productID, variantID := createTestProduct(t, auth.Token, categoryID, "Order Test Product", "ORD-SKU-001", 100.00, 50.00)

	// Seed initial stock (100 units) so order creation doesn't fail on stock validation
	err = testFixture.CreateStock(context.Background(), variantID, auth.BranchID, 100, 0)
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
	orderBody := map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           2,
				"discount_amount":    0,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, td.auth.Token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	assert.True(t, createResp["success"].(bool))
	assert.Equal(t, "Order created successfully", createResp["message"])

	data := createResp["data"].(map[string]interface{})
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
	resp, err = testServer.GET(fmt.Sprintf("/api/v1/orders/%d", orderID), td.auth.Token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var getResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &getResp)
	require.NoError(t, err)

	assert.True(t, getResp["success"].(bool))
	getData := getResp["data"].(map[string]interface{})
	assert.Equal(t, float64(orderID), getData["id"])
}

func TestOrderFlow_CreateConfirmPayComplete(t *testing.T) {
	td := setupOrderTest(t)
	ctx := context.Background()
	token := td.auth.Token

	// Seed initial stock (10 units)
	err := testFixture.CreateStock(ctx, td.variantID, td.auth.BranchID, 10, 0)
	require.NoError(t, err)

	// Step 1: Create an order (delivery to avoid auto-complete on payment)
	orderBody := map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "delivery",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))
	grandTotal := data["grand_total"].(float64)

	assert.Equal(t, "draft", data["status"])

	// Step 2: Confirm the order
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	var confirmResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &confirmResp)
	require.NoError(t, err)

	assert.True(t, confirmResp["success"].(bool))
	confirmData := confirmResp["data"].(map[string]interface{})
	assert.Equal(t, "confirmed", confirmData["status"])
	assert.NotNil(t, confirmData["confirmed_at"])

	// Step 3: Add payment
	paymentBody := map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "cash",
				"amount": grandTotal,
			},
		},
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/payments", orderID), paymentBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "payment failed: %s", string(resp.Body))

	var paymentResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &paymentResp)
	require.NoError(t, err)

	assert.True(t, paymentResp["success"].(bool))
	paymentData := paymentResp["data"].(map[string]interface{})
	assert.Equal(t, "paid", paymentData["payment_status"])
	assert.NotNil(t, paymentData["paid_at"])

	// Step 4: Complete the order
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/complete", orderID), nil, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "complete failed: %s", string(resp.Body))

	var completeResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &completeResp)
	require.NoError(t, err)

	assert.True(t, completeResp["success"].(bool))
	completeData := completeResp["data"].(map[string]interface{})
	assert.Equal(t, "completed", completeData["status"])
	assert.NotNil(t, completeData["completed_at"])

	// DB assertion: verify stock was deducted
	var stockQty int
	err = testDB.DB.QueryRowContext(ctx,
		"SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2",
		td.variantID, td.auth.BranchID,
	).Scan(&stockQty)
	require.NoError(t, err)
	assert.Equal(t, 9, stockQty) // started at 10, ordered 1

	// DB assertion: verify stock_movement OUT record was created
	var movementType, refType string
	var movementQty, stockBefore, stockAfter int
	err = testDB.DB.QueryRowContext(ctx,
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
	err = testDB.DB.QueryRowContext(ctx,
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
	orderBody := map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))

	// Cancel the order
	cancelBody := map[string]interface{}{
		"reason": "Customer changed their mind",
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/cancel", orderID), cancelBody, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var cancelResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &cancelResp)
	require.NoError(t, err)

	assert.True(t, cancelResp["success"].(bool))
	cancelData := cancelResp["data"].(map[string]interface{})
	assert.Equal(t, "cancelled", cancelData["status"])
	assert.Equal(t, "Customer changed their mind", cancelData["cancel_reason"])
	assert.NotNil(t, cancelData["cancelled_at"])
}

func TestOrderFlow_VoidConfirmed(t *testing.T) {
	td := setupOrderTest(t)
	ctx := context.Background()
	token := td.auth.Token

	// Seed initial stock
	err := testFixture.CreateStock(ctx, td.variantID, td.auth.BranchID, 10, 0)
	require.NoError(t, err)

	// Create and confirm an order (not completed — no stock deduction yet)
	orderBody := map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))

	// Confirm the order first
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	// Void the confirmed order (not completed, so no stock restoration)
	voidBody := map[string]interface{}{
		"reason": "Fraudulent order",
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/void", orderID), voidBody, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var voidResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &voidResp)
	require.NoError(t, err)

	assert.True(t, voidResp["success"].(bool))
	voidData := voidResp["data"].(map[string]interface{})
	assert.Equal(t, "voided", voidData["status"])
	assert.Equal(t, "Fraudulent order", voidData["void_reason"])
	assert.NotNil(t, voidData["voided_at"])

	// DB assertion: stock should remain unchanged (no deduction happened for confirmed-only orders)
	var stockQty int
	err = testDB.DB.QueryRowContext(ctx,
		"SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2",
		td.variantID, td.auth.BranchID,
	).Scan(&stockQty)
	require.NoError(t, err)
	assert.Equal(t, 10, stockQty) // unchanged

	// DB assertion: no stock movements should exist
	var movementCount int
	err = testDB.DB.QueryRowContext(ctx,
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
	err := testFixture.CreateStock(ctx, td.variantID, td.auth.BranchID, 10, 0)
	require.NoError(t, err)

	// Create → Confirm → Pay → Complete an order (delivery to avoid auto-complete)
	orderBody := map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "delivery",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           2,
				"discount_amount":    0,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))
	grandTotal := data["grand_total"].(float64)

	// Confirm
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	// Pay
	paymentBody := map[string]interface{}{
		"payments": []map[string]interface{}{
			{"method": "cash", "amount": grandTotal},
		},
	}
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/payments", orderID), paymentBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "payment failed: %s", string(resp.Body))

	// Complete
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/complete", orderID), nil, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "complete failed: %s", string(resp.Body))

	// Verify stock was deducted after completion
	var stockQtyAfterComplete int
	err = testDB.DB.QueryRowContext(ctx,
		"SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2",
		td.variantID, td.auth.BranchID,
	).Scan(&stockQtyAfterComplete)
	require.NoError(t, err)
	assert.Equal(t, 8, stockQtyAfterComplete) // 10 - 2

	// Now void the completed order
	voidBody := map[string]interface{}{
		"reason": "Customer returned items",
	}
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/void", orderID), voidBody, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var voidResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &voidResp)
	require.NoError(t, err)
	assert.Equal(t, "voided", voidResp["data"].(map[string]interface{})["status"])

	// DB assertion: stock should be restored
	var stockQtyAfterVoid int
	err = testDB.DB.QueryRowContext(ctx,
		"SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2",
		td.variantID, td.auth.BranchID,
	).Scan(&stockQtyAfterVoid)
	require.NoError(t, err)
	assert.Equal(t, 10, stockQtyAfterVoid) // restored to original

	// DB assertion: should have both OUT (order) and IN (order_void) movements
	var outCount, inCount int
	err = testDB.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM stock_movements WHERE reference_id = $1 AND reference_type = 'order' AND type = 'OUT'",
		orderID,
	).Scan(&outCount)
	require.NoError(t, err)
	assert.Equal(t, 1, outCount)

	err = testDB.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM stock_movements WHERE reference_id = $1 AND reference_type = 'order_void' AND type = 'IN'",
		orderID,
	).Scan(&inCount)
	require.NoError(t, err)
	assert.Equal(t, 1, inCount)

	// DB assertion: IN movement should show correct stock restoration
	var inMovementQty, inStockBefore, inStockAfter int
	err = testDB.DB.QueryRowContext(ctx,
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
	orderBody := map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))
	grandTotal := data["grand_total"].(float64)

	// Confirm
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	// Pay
	paymentBody := map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "cash",
				"amount": grandTotal,
			},
		},
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/payments", orderID), paymentBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "payment failed: %s", string(resp.Body))

	var paymentResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &paymentResp)
	require.NoError(t, err)

	paymentData := paymentResp["data"].(map[string]interface{})
	payments := paymentData["payments"].([]interface{})
	payment := payments[0].(map[string]interface{})
	paymentID := int64(payment["id"].(float64))

	// Refund
	refundBody := map[string]interface{}{
		"payment_id":    paymentID,
		"amount":        grandTotal,
		"refund_reason": "Customer returned item",
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/refund", orderID), refundBody, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var refundResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &refundResp)
	require.NoError(t, err)

	assert.True(t, refundResp["success"].(bool))
	refundData := refundResp["data"].(map[string]interface{})
	assert.Equal(t, grandTotal, refundData["refunded_total"])

	// DB assertion: verify payment status is 'refunded' in database
	var paymentStatus string
	err = testDB.DB.QueryRowContext(ctx,
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
		orderBody := map[string]interface{}{
			"customer_id":      td.customerID,
			"fulfillment_type": "counter",
			"items": []map[string]interface{}{
				{
					"product_variant_id": td.variantID,
					"quantity":           1,
					"discount_amount":    0,
				},
			},
		}

		resp, err := testServer.POST("/api/v1/orders", orderBody, token)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))
	}

	// List orders
	resp, err := testServer.GET("/api/v1/orders", token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &listResp)
	require.NoError(t, err)

	assert.True(t, listResp["success"].(bool))
	data := listResp["data"].([]interface{})
	assert.Len(t, data, 3)

	// Verify pagination in meta
	meta := listResp["meta"].(map[string]interface{})
	pagination := meta["pagination"].(map[string]interface{})
	assert.Equal(t, float64(3), pagination["total_records"])
}

func TestOrderFlow_UpdateOrder(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token

	// Create an order
	orderBody := map[string]interface{}{
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
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))

	// Update the order
	updateBody := map[string]interface{}{
		"fulfillment_type": "delivery",
		"shipping_address": "123 Main St, City",
		"notes":            "Updated notes",
	}

	resp, err = testServer.PUT(fmt.Sprintf("/api/v1/orders/%d", orderID), updateBody, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updateResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &updateResp)
	require.NoError(t, err)

	assert.True(t, updateResp["success"].(bool))
	updateData := updateResp["data"].(map[string]interface{})
	assert.Equal(t, "delivery", updateData["fulfillment_type"])
	assert.Equal(t, "123 Main St, City", updateData["shipping_address"])
	assert.Equal(t, "Updated notes", updateData["notes"])
}

func TestOrder_CreateWithMultipleItems(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token

	// Create additional product variant via API
	productBody := map[string]interface{}{
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
	}
	resp, err := testServer.POST("/api/v1/products", productBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create product failed: %s", string(resp.Body))

	var prodResp map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &prodResp))
	prodData := prodResp["data"].(map[string]interface{})
	variants := prodData["variants"].([]interface{})
	variant2ID := int64(variants[0].(map[string]interface{})["id"].(float64))

	// Seed stock for second variant
	err = testFixture.CreateStock(context.Background(), variant2ID, td.auth.BranchID, 100, 0)
	require.NoError(t, err)

	// Create order with multiple items
	orderBody := map[string]interface{}{
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
	}

	resp, err = testServer.POST("/api/v1/orders", orderBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})

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
			resp, err := testServer.POST("/api/v1/orders", tt.body, auth.Token)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestOrder_PaymentValidation(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token

	// Create and confirm an order
	orderBody := map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))

	// Confirm
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	// Test invalid payment method
	invalidPayment := map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "invalid_method",
				"amount": 100,
			},
		},
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/payments", orderID), invalidPayment, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestOrder_PartialPayment(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token

	// Create order (grand_total = 100)
	orderBody := map[string]interface{}{
		"customer_id":      td.customerID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": td.variantID,
				"quantity":           1,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))

	// Confirm
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	// Pay partial amount (50 out of 100)
	partialPayment := map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "cash",
				"amount": 50,
			},
		},
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/payments", orderID), partialPayment, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var paymentResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &paymentResp)
	require.NoError(t, err)

	paymentData := paymentResp["data"].(map[string]interface{})
	assert.Equal(t, "partial", paymentData["payment_status"])
	assert.Equal(t, float64(50), paymentData["paid_amount"])
	assert.Equal(t, float64(50), paymentData["balance_due"])

	// Pay remaining
	remainingPayment := map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "credit_card",
				"amount": 50,
			},
		},
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/payments", orderID), remainingPayment, token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.Unmarshal(resp.Body, &paymentResp)
	require.NoError(t, err)

	paymentData = paymentResp["data"].(map[string]interface{})
	assert.Equal(t, "paid", paymentData["payment_status"])
	assert.Equal(t, float64(100), paymentData["paid_amount"])
	assert.Equal(t, float64(0), paymentData["balance_due"])
}
