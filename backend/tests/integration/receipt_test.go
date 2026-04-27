package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReceipt_GetAfterComplete(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	// Create customer (required for delivery orders)
	custResp := doPost(t, "/api/v1/customers", map[string]interface{}{
		"code": "RCPT-CUST", "name": "Receipt Customer",
	}, auth.Token)
	require.Equal(t, http.StatusCreated, custResp.StatusCode)
	custR := decodeResponse(t, custResp)
	var custData map[string]interface{}
	require.NoError(t, json.Unmarshal(custR.Data, &custData))
	customerID := int64(custData["id"].(float64))

	// Create category + product
	categoryID := createTestCategory(t, auth.Token)
	productID, variantID := createTestProduct(t, auth.Token, categoryID, "Receipt Product", "RCPT-SKU-001", 250.00, 100.00)
	_ = productID

	// Seed stock
	err := testEnv.Fixtures.CreateStock(context.Background(), variantID, auth.BranchID, 50, 0)
	require.NoError(t, err)

	// Create order (delivery to avoid auto-complete on full payment)
	resp := doPost(t, "/api/v1/orders", map[string]interface{}{
		"customer_id":      customerID,
		"fulfillment_type": "delivery",
		"items": []map[string]interface{}{
			{
				"product_variant_id": variantID,
				"quantity":           2,
				"discount_amount":    0,
			},
		},
	}, auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var orderData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &orderData))
	orderID := int64(orderData["id"].(float64))
	grandTotal := orderData["grand_total"].(float64)

	// Receipt should not exist yet (404)
	resp = doGet(t, fmt.Sprintf("/api/v1/orders/%d/receipt", orderID), auth.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Confirm
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm: %s", string(resp.Body))

	// Pay
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/payments", orderID), map[string]interface{}{
		"payments": []map[string]interface{}{
			{"method": "cash", "amount": grandTotal},
		},
	}, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "pay: %s", string(resp.Body))

	// Complete → triggers receipt snapshot creation
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/complete", orderID), nil, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "complete: %s", string(resp.Body))

	// Receipt should now exist
	resp = doGet(t, fmt.Sprintf("/api/v1/orders/%d/receipt", orderID), auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "get receipt: %s", string(resp.Body))

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)
	assert.Equal(t, "Receipt retrieved successfully", r.Message)

	var rec map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &rec))

	assert.Equal(t, float64(orderID), rec["order_id"])
	assert.NotEmpty(t, rec["order_no"])
	assert.Equal(t, grandTotal, rec["grand_total"])
	assert.Equal(t, float64(500), rec["total_amount"]) // 250 x 2
	assert.Equal(t, float64(0), rec["total_discount"])

	items := rec["items"].([]interface{})
	require.Len(t, items, 1)
	item := items[0].(map[string]interface{})
	assert.Equal(t, "Receipt Product", item["product_name"])
	assert.Equal(t, float64(2), item["quantity"])
	assert.Equal(t, float64(250), item["unit_price"])
	assert.Equal(t, float64(500), item["subtotal"])

	payments := rec["payments"].([]interface{})
	require.Len(t, payments, 1)
	pay := payments[0].(map[string]interface{})
	assert.Equal(t, "cash", pay["method"])
	assert.Equal(t, grandTotal, pay["amount"])

	assert.NotEmpty(t, rec["completed_at"])
	assert.NotEmpty(t, rec["created_at"])

	// Fetching again returns the same immutable document
	resp2 := doGet(t, fmt.Sprintf("/api/v1/orders/%d/receipt", orderID), auth.Token)
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	r2 := decodeResponse(t, resp2)
	var rec2 map[string]interface{}
	require.NoError(t, json.Unmarshal(r2.Data, &rec2))
	assert.Equal(t, rec["order_no"], rec2["order_no"])
	assert.Equal(t, rec["grand_total"], rec2["grand_total"])
}

func TestReceipt_NotFoundForNonExistentOrder(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	resp := doGet(t, "/api/v1/orders/99999/receipt", auth.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestReceipt_CancelledOrderHasNoReceipt(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	categoryID := createTestCategory(t, auth.Token)
	_, variantID := createTestProduct(t, auth.Token, categoryID, "Cancel Product", "RCPT-CANCEL-001", 100.00, 50.00)

	err := testEnv.Fixtures.CreateStock(context.Background(), variantID, auth.BranchID, 10, 0)
	require.NoError(t, err)

	// Create order
	resp := doPost(t, "/api/v1/orders", map[string]interface{}{
		"fulfillment_type": "counter",
		"items": []map[string]interface{}{
			{"product_variant_id": variantID, "quantity": 1},
		},
	}, auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	orderID := int64(data["id"].(float64))

	// Cancel
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/cancel", orderID), map[string]interface{}{
		"reason": "test cancel",
	}, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// No receipt for cancelled order
	resp = doGet(t, fmt.Sprintf("/api/v1/orders/%d/receipt", orderID), auth.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
