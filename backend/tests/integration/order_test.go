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

func TestOrderFlow_CreateAndGet(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Setup test data
	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create an order
	orderBody := map[string]interface{}{
		"customer_id":      testData.Customer.ID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": testData.ProductVariant.ID,
				"quantity":           2,
				"discount_amount":    0,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

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
	resp, err = testServer.GET(fmt.Sprintf("/api/v1/orders/%d", orderID), "")
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
	cleanupDatabase(t)
	ctx := context.Background()

	// Setup test data
	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Step 1: Create an order
	orderBody := map[string]interface{}{
		"customer_id":      testData.Customer.ID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": testData.ProductVariant.ID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))
	grandTotal := data["grand_total"].(float64)

	assert.Equal(t, "draft", data["status"])

	// Step 2: Confirm the order
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

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

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/payments", orderID), paymentBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var paymentResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &paymentResp)
	require.NoError(t, err)

	assert.True(t, paymentResp["success"].(bool))
	paymentData := paymentResp["data"].(map[string]interface{})
	assert.Equal(t, "paid", paymentData["payment_status"])
	assert.NotNil(t, paymentData["paid_at"])

	// Step 4: Complete the order
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/complete", orderID), nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var completeResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &completeResp)
	require.NoError(t, err)

	assert.True(t, completeResp["success"].(bool))
	completeData := completeResp["data"].(map[string]interface{})
	assert.Equal(t, "completed", completeData["status"])
	assert.NotNil(t, completeData["completed_at"])
}

func TestOrderFlow_Cancel(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Setup test data
	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create an order
	orderBody := map[string]interface{}{
		"customer_id":      testData.Customer.ID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": testData.ProductVariant.ID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))

	// Cancel the order
	cancelBody := map[string]interface{}{
		"reason": "Customer changed their mind",
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/cancel", orderID), cancelBody, "")
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

func TestOrderFlow_Void(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Setup test data
	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create and confirm an order
	orderBody := map[string]interface{}{
		"customer_id":      testData.Customer.ID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": testData.ProductVariant.ID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, "")
	require.NoError(t, err)

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))

	// Confirm the order first
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Void the order
	voidBody := map[string]interface{}{
		"reason": "Fraudulent order",
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/void", orderID), voidBody, "")
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
}

func TestOrderFlow_Refund(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Setup test data
	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create, confirm, and pay an order
	orderBody := map[string]interface{}{
		"customer_id":      testData.Customer.ID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": testData.ProductVariant.ID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, "")
	require.NoError(t, err)

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))
	grandTotal := data["grand_total"].(float64)

	// Confirm
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Pay
	paymentBody := map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "cash",
				"amount": grandTotal,
			},
		},
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/payments", orderID), paymentBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

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

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/refund", orderID), refundBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var refundResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &refundResp)
	require.NoError(t, err)

	assert.True(t, refundResp["success"].(bool))
	refundData := refundResp["data"].(map[string]interface{})
	assert.Equal(t, grandTotal, refundData["refunded_total"])
}

func TestOrderFlow_ListOrders(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Setup test data
	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create multiple orders
	for i := 0; i < 3; i++ {
		orderBody := map[string]interface{}{
			"customer_id":      testData.Customer.ID,
			"fulfillment_type": "counter",
			"items": []map[string]interface{}{
				{
					"product_variant_id": testData.ProductVariant.ID,
					"quantity":           1,
					"discount_amount":    0,
				},
			},
		}

		resp, err := testServer.POST("/api/v1/orders", orderBody, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List orders
	resp, err := testServer.GET("/api/v1/orders", "")
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
	cleanupDatabase(t)
	ctx := context.Background()

	// Setup test data
	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create an order
	orderBody := map[string]interface{}{
		"customer_id":      testData.Customer.ID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": testData.ProductVariant.ID,
				"quantity":           1,
				"discount_amount":    0,
			},
		},
		"notes": "Initial notes",
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

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

	resp, err = testServer.PUT(fmt.Sprintf("/api/v1/orders/%d", orderID), updateBody, "")
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
	cleanupDatabase(t)
	ctx := context.Background()

	// Setup test data
	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create additional product variants
	variant2, err := testFixture.CreateProductVariant(ctx, testData.Product.ID, "SKU-002", "Large", 150.00, 75.00)
	require.NoError(t, err)

	// Create order with multiple items
	orderBody := map[string]interface{}{
		"customer_id":      testData.Customer.ID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": testData.ProductVariant.ID, // 100.00 x 2 = 200.00
				"quantity":           2,
				"discount_amount":    10,
			},
			{
				"product_variant_id": variant2.ID, // 150.00 x 1 = 150.00
				"quantity":           1,
				"discount_amount":    0,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

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
			resp, err := testServer.POST("/api/v1/orders", tt.body, "")
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestOrder_PaymentValidation(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Setup test data
	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create and confirm an order
	orderBody := map[string]interface{}{
		"customer_id":      testData.Customer.ID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": testData.ProductVariant.ID,
				"quantity":           1,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, "")
	require.NoError(t, err)

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))

	// Confirm
	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test invalid payment method
	invalidPayment := map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "invalid_method",
				"amount": 100,
			},
		},
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/payments", orderID), invalidPayment, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestOrder_PartialPayment(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Setup test data
	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create order (grand_total = 100)
	orderBody := map[string]interface{}{
		"customer_id":      testData.Customer.ID,
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{
				"product_variant_id": testData.ProductVariant.ID,
				"quantity":           1,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/orders", orderBody, "")
	require.NoError(t, err)

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	orderID := int64(data["id"].(float64))

	// Confirm
	_, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, "")
	require.NoError(t, err)

	// Pay partial amount (50 out of 100)
	partialPayment := map[string]interface{}{
		"payments": []map[string]interface{}{
			{
				"method": "cash",
				"amount": 50,
			},
		},
	}

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/payments", orderID), partialPayment, "")
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

	resp, err = testServer.POST(fmt.Sprintf("/api/v1/orders/%d/payments", orderID), remainingPayment, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.Unmarshal(resp.Body, &paymentResp)
	require.NoError(t, err)

	paymentData = paymentResp["data"].(map[string]interface{})
	assert.Equal(t, "paid", paymentData["payment_status"])
	assert.Equal(t, float64(100), paymentData["paid_amount"])
	assert.Equal(t, float64(0), paymentData["balance_due"])
}
