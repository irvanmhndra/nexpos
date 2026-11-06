package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const reportRange = "date_from=2020-01-01&date_to=2099-12-31"

// reportProduct creates a stocked product, optionally in a category, and returns its variant ID.
func reportProduct(t *testing.T, token string, branchID int64, categoryID *int64, sku, name string, price, cost float64) int64 {
	t.Helper()
	resp := doPost(t, "/api/v1/products", map[string]any{
		"name": name, "category_id": categoryID,
		"variants": []map[string]any{{"sku": sku, "name": "Default", "price": price, "standard_cost": cost, "is_default": true}},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "%s", resp.Body)
	var product struct {
		Variants []struct {
			ID int64 `json:"id"`
		} `json:"variants"`
	}
	require.NoError(t, json.Unmarshal(decodeResponse(t, resp).Data, &product))
	require.Len(t, product.Variants, 1)
	require.NoError(t, testEnv.Fixtures.CreateStock(context.Background(), product.Variants[0].ID, branchID, 1000, 0))
	return product.Variants[0].ID
}

// reportOrder creates an order and returns its ID and grand total.
func reportOrder(t *testing.T, token string, customerID *int64, variantID int64, qty int) (int64, float64) {
	t.Helper()
	resp := doPost(t, "/api/v1/orders", map[string]any{
		"customer_id": customerID, "fulfillment_type": "delivery",
		"items": []map[string]any{{"product_variant_id": variantID, "quantity": qty, "discount_amount": 0}},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "%s", resp.Body)
	order := decodeObject(t, decodeResponse(t, resp).Data)
	return int64(order["id"].(float64)), order["grand_total"].(float64)
}

// reportCompleteOrder takes an order through confirm, payment, and complete.
func reportCompleteOrder(t *testing.T, token string, customerID *int64, variantID int64, qty int, method string) {
	t.Helper()
	id, total := reportOrder(t, token, customerID, variantID, qty)
	resp := doPost(t, fmt.Sprintf("/api/v1/orders/%d/confirm", id), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/payments", id), map[string]any{
		"payments": []map[string]any{{"method": method, "amount": total}},
	}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/complete", id), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
}

func reportBody(t *testing.T, token, path string) string {
	t.Helper()
	resp := doGet(t, "/api/v1/reports/"+path, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s: %s", path, resp.Body)
	// Echo's JSON encoder ends every body with a newline.
	require.True(t, strings.HasSuffix(string(resp.Body), "}\n"), "%q", resp.Body)
	require.Equal(t, "application/json", resp.Headers.Get("Content-Type"))
	return strings.TrimSuffix(string(resp.Body), "\n")
}

func TestReport_ExactResponses(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	token := auth.Token

	category := createCategory(t, token, map[string]any{"code": "DRINK", "name": "Minuman"})
	categoryID := int64(category["id"].(float64))
	resp := doPost(t, "/api/v1/customers", map[string]any{"code": "R-1", "name": "Rina"}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	customerID := int64(decodeObject(t, decodeResponse(t, resp).Data)["id"].(float64))

	kopi := reportProduct(t, token, auth.BranchID, &categoryID, "KOPI", "Kopi", 10000, 4000)
	teh := reportProduct(t, token, auth.BranchID, nil, "TEH", "Teh", 3333, 1000)

	reportCompleteOrder(t, token, &customerID, kopi, 3, "cash")
	reportCompleteOrder(t, token, &customerID, teh, 1, "qris")
	reportOrder(t, token, &customerID, kopi, 9) // still a draft: not a sale
	cancelled, _ := reportOrder(t, token, &customerID, kopi, 9)
	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/cancel", cancelled), map[string]any{"reason": "test"}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)

	// Date and hour as the database's own TimeZone sees them.
	var day string
	var hour int
	require.NoError(t, testEnv.DB.QueryRow(
		`SELECT to_char(created_at, 'YYYY-MM-DD'), EXTRACT(HOUR FROM created_at)::int FROM orders ORDER BY id LIMIT 1`,
	).Scan(&day, &hour))

	// Compared as strings: key order and number formatting are part of the contract.
	assert.Equal(t, `{"success":true,"message":"Report summary retrieved successfully","data":{"total_revenue":33333,"total_orders":2,"total_items_sold":4,"avg_order_value":16666.5,"total_customers":1,"new_customers":1,"completed_orders":2,"cancelled_orders":1,"pending_orders":0,"total_discount":0,"total_tax":0,"gross_profit":20333,"gross_profit_margin":60.99960999609996}}`,
		reportBody(t, token, "summary?"+reportRange))
	assert.Equal(t, fmt.Sprintf(`{"success":true,"message":"Sales trend retrieved successfully","data":{"data":[{"date":"%sT00:00:00Z","sales":33333,"orders":2,"items":4}]}}`, day),
		reportBody(t, token, "sales-trend?"+reportRange))
	assert.Equal(t, `{"success":true,"message":"Top products retrieved successfully","data":{"data":[{"product_id":1,"product_name":"Kopi","total_sold":3,"total_amount":30000},{"product_id":2,"product_name":"Teh","total_sold":1,"total_amount":3333}]}}`,
		reportBody(t, token, "top-products?"+reportRange))
	assert.Equal(t, `{"success":true,"message":"Top products retrieved successfully","data":{"data":[{"product_id":1,"product_name":"Kopi","total_sold":3,"total_amount":30000}]}}`,
		reportBody(t, token, "top-products?limit=1&"+reportRange))
	assert.Equal(t, reportBody(t, token, "top-products?"+reportRange), reportBody(t, token, "top-products?limit=abc&"+reportRange))
	assert.Equal(t, reportBody(t, token, "top-products?"+reportRange), reportBody(t, token, "top-products?limit=-1&"+reportRange))
	assert.Equal(t, fmt.Sprintf(`{"success":true,"message":"Category revenue retrieved successfully","data":{"data":[{"category_id":%d,"category_name":"Minuman","total_amount":30000,"order_count":1},{"category_id":0,"category_name":"Tanpa Kategori","total_amount":3333,"order_count":1}]}}`, categoryID),
		reportBody(t, token, "category-revenue?"+reportRange))
	assert.Equal(t, `{"success":true,"message":"Payment methods retrieved successfully","data":{"data":[{"method":"cash","total_amount":30000,"count":1},{"method":"qris","total_amount":3333,"count":1}]}}`,
		reportBody(t, token, "payment-methods?"+reportRange))
	assert.Equal(t, fmt.Sprintf(`{"success":true,"message":"Hourly sales retrieved successfully","data":{"data":[{"hour":%d,"total_amount":33333,"order_count":2}]}}`, hour),
		reportBody(t, token, "hourly-sales?"+reportRange))

	// A range with no orders, and a company with no orders: zeros and empty lists.
	emptySummary := `{"success":true,"message":"Report summary retrieved successfully","data":{"total_revenue":0,"total_orders":0,"total_items_sold":0,"avg_order_value":0,"total_customers":0,"new_customers":0,"completed_orders":0,"cancelled_orders":0,"pending_orders":0,"total_discount":0,"total_tax":0,"gross_profit":0,"gross_profit_margin":0}}`
	const past = "date_from=2001-01-01&date_to=2001-01-31"
	assert.Equal(t, emptySummary, reportBody(t, token, "summary?"+past))
	assert.Equal(t, `{"success":true,"message":"Sales trend retrieved successfully","data":{"data":[]}}`, reportBody(t, token, "sales-trend?"+past))
	other := registerTestUser(t)
	assert.Equal(t, emptySummary, reportBody(t, other.Token, "summary?"+reportRange))
	for path, message := range map[string]string{
		"top-products": "Top products", "category-revenue": "Category revenue",
		"payment-methods": "Payment methods", "hourly-sales": "Hourly sales", "sales-trend": "Sales trend",
	} {
		assert.Equal(t, fmt.Sprintf(`{"success":true,"message":"%s retrieved successfully","data":{"data":[]}}`, message),
			reportBody(t, other.Token, path+"?"+reportRange), path)
	}
}

// A margin that is a whole number renders without a fraction.
func TestReport_WholeMarginAndRoundedAverage(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	token := auth.Token
	resp := doPost(t, "/api/v1/customers", map[string]any{"code": "R-1", "name": "Rina"}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	customerID := int64(decodeObject(t, decodeResponse(t, resp).Data)["id"].(float64))

	a := reportProduct(t, token, auth.BranchID, nil, "A", "A", 100, 50)
	reportCompleteOrder(t, token, &customerID, a, 1, "cash")
	assert.Contains(t, reportBody(t, token, "summary?"+reportRange), `"gross_profit_margin":50}`)

	// 100 + 100 + 33.33 over 3 orders rounds to 2 places.
	b := reportProduct(t, token, auth.BranchID, nil, "B", "B", 33.33, 0)
	reportCompleteOrder(t, token, &customerID, a, 1, "cash")
	reportCompleteOrder(t, token, &customerID, b, 1, "cash")
	assert.Contains(t, reportBody(t, token, "summary?"+reportRange), `"total_revenue":233.33,"total_orders":3,"total_items_sold":3,"avg_order_value":77.78,`)
}

func TestReport_Errors(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	for _, path := range []string{"summary", "sales-trend", "top-products", "category-revenue", "payment-methods", "hourly-sales"} {
		for _, query := range []string{"", "?date_from=2026-01-01", "?date_to=2026-01-01"} {
			resp := doGet(t, "/api/v1/reports/"+path+query, auth.Token)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, path+query)
			assert.JSONEq(t, `{"success":false,"message":"date_from and date_to are required","error_code":"BAD_REQUEST"}`, string(resp.Body))
		}
		// An unparseable date fails in the database.
		resp := doGet(t, "/api/v1/reports/"+path+"?date_from=nope&date_to=2026-01-01", auth.Token)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, path)
		assert.Equal(t, http.StatusUnauthorized, doGet(t, "/api/v1/reports/"+path+"?"+reportRange, "").StatusCode, path)
	}
}
