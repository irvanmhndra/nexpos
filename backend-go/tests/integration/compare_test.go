package integration_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fetchRaw is a plain GET, so nothing normalizes the bytes before comparing them.
func fetchRaw(t *testing.T, url, token string) (int, string, []byte) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, resp.Header.Get("Content-Type"), body
}

// With TARGET_BASE_URL (Rust) and COMPARE_BASE_URL (Go) both set, every ported
// GET must answer with the same status, Content-Type, and bytes on the same data.
func TestCompare_ReadsAreByteIdentical(t *testing.T) {
	target, compare := os.Getenv("TARGET_BASE_URL"), os.Getenv("COMPARE_BASE_URL")
	if target == "" || compare == "" {
		t.Skip("set TARGET_BASE_URL and COMPARE_BASE_URL to compare two running servers")
	}

	cleanupDatabase(t)
	auth := registerTestUser(t)
	token := auth.Token

	createBranch(t, token, map[string]any{"code": "BDG", "name": "Bandung <&>", "address": "Jl. Asia Afrika", "is_active": true})
	category := createCategory(t, token, map[string]any{"code": "DRINK", "name": "Minuman & Teh", "sort_order": 3})
	categoryID := int64(category["id"].(float64))
	child := createCategory(t, token, map[string]any{"code": "COFFEE", "name": "Kopi", "parent_id": categoryID})

	resp := doPost(t, "/api/v1/customers", map[string]any{"code": "C-1", "name": "Ani \"A\" é 日本", "email": "ani@x.id", "phone": "0812", "is_member": true}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "%s", resp.Body)
	customerID := int64(decodeObject(t, decodeResponse(t, resp).Data)["id"].(float64))

	product := createProduct(t, token, map[string]any{
		"name": "Kopi <Susu>", "category_id": int64(child["id"].(float64)), "description": "Enak & murah",
		"variants": []map[string]any{
			{"sku": "K-L", "name": "Large", "price": 25000.5, "standard_cost": 10000.1, "last_purchase_cost": 9000.25, "barcode": "8991",
				"attributes": map[string]any{"z": "a<b", "a": 2.0, "m": []any{1.5, "x", true, nil}}, "sale_price": 19999.99,
				"sale_start": "2026-01-02T03:04:05Z", "sale_end": "2026-02-02T03:04:05.5+07:00"},
			{"sku": "K-M", "name": "Medium", "price": 20000},
		},
	})
	productID := int64(product["id"].(float64))

	resp = doPut(t, "/api/v1/company-settings", map[string]any{"tax_rate": 11.5, "rounding_amount": 100, "receipt_header": "Halo & selamat", "max_offline_days": 9}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)

	variant := reportProduct(t, token, auth.BranchID, &categoryID, "REP", "Teh", 3333.33, 1111.11)
	reportCompleteOrder(t, token, &customerID, variant, 3, "cash")
	reportCompleteOrder(t, token, &customerID, variant, 1, "qris")

	paths := []string{
		"/health", "/health/live", "/health/ready", "/", "/api/v1/nope",
		"/api/v1/branches", "/api/v1/branches?search=band", "/api/v1/branches?is_active=false", "/api/v1/branches?per_page=1&page=2",
		fmt.Sprintf("/api/v1/branches/%d", auth.BranchID), "/api/v1/branches/999999", "/api/v1/branches/abc",
		"/api/v1/customers", "/api/v1/customers?is_member=true", "/api/v1/customers?search=nothing",
		fmt.Sprintf("/api/v1/customers/%d", customerID), "/api/v1/customers/999999", "/api/v1/customers/abc",
		"/api/v1/product-categories", "/api/v1/product-categories/all", "/api/v1/product-categories?search=kopi",
		fmt.Sprintf("/api/v1/product-categories/%d", int64(child["id"].(float64))), "/api/v1/product-categories/999999", "/api/v1/product-categories/abc",
		"/api/v1/products", "/api/v1/products?search=K-L", "/api/v1/products?is_active=false",
		fmt.Sprintf("/api/v1/products?category_id=%d", categoryID), "/api/v1/products?per_page=1&page=2",
		fmt.Sprintf("/api/v1/products/%d", productID), "/api/v1/products/999999", "/api/v1/products/abc",
		"/api/v1/products/lookup?code=8991", "/api/v1/products/lookup?code=K-M", "/api/v1/products/lookup?code=NOPE", "/api/v1/products/lookup",
		"/api/v1/company-settings",
		"/api/v1/reports/summary?" + reportRange, "/api/v1/reports/sales-trend?" + reportRange,
		"/api/v1/reports/top-products?" + reportRange, "/api/v1/reports/top-products?limit=1&" + reportRange,
		"/api/v1/reports/category-revenue?" + reportRange, "/api/v1/reports/payment-methods?" + reportRange,
		"/api/v1/reports/hourly-sales?" + reportRange, "/api/v1/reports/summary",
	}
	for _, path := range paths {
		goStatus, goType, goBody := fetchRaw(t, compare+path, token)
		rustStatus, rustType, rustBody := fetchRaw(t, target+path, token)
		assert.Equal(t, goStatus, rustStatus, path)
		assert.Equal(t, goType, rustType, path)
		assert.Equal(t, string(goBody), string(rustBody), path)
	}

	// Without a token.
	for _, path := range []string{"/api/v1/branches", "/api/v1/products", "/api/v1/company-settings", "/api/v1/reports/summary"} {
		goStatus, _, goBody := fetchRaw(t, compare+path, "")
		rustStatus, _, rustBody := fetchRaw(t, target+path, "")
		assert.Equal(t, goStatus, rustStatus, path)
		assert.Equal(t, string(goBody), string(rustBody), path)
	}
}
