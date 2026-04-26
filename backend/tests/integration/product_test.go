package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/irvanmhndra/nexpos-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestCategory creates a product category via the API and returns its ID.
func createTestCategory(t *testing.T, token string) int64 {
	t.Helper()
	n := testutil.UniqueCounter()
	resp := doPost(t, "/api/v1/product-categories", map[string]interface{}{
		"code": fmt.Sprintf("CAT-%d", n),
		"name": fmt.Sprintf("Test Category %d", n),
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create category failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	return int64(data["id"].(float64))
}

// createTestProduct creates a product with one variant via the API and returns (productID, variantID).
func createTestProduct(t *testing.T, token string, categoryID int64, name, sku string, price, cost float64) (int64, int64) {
	t.Helper()
	resp := doPost(t, "/api/v1/products", map[string]interface{}{
		"name":                name,
		"product_category_id": categoryID,
		"variants": []map[string]interface{}{
			{
				"sku":           sku,
				"name":          "Default",
				"price":         price,
				"standard_cost": cost,
				"is_default":    true,
			},
		},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create product failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	productID := int64(data["id"].(float64))

	variants := data["variants"].([]interface{})
	variant := variants[0].(map[string]interface{})
	variantID := int64(variant["id"].(float64))

	return productID, variantID
}

func TestProductCategory_CreateAndGet(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	// Create a category
	resp := doPost(t, "/api/v1/product-categories", map[string]interface{}{
		"code":        "ELEC",
		"name":        "Electronics",
		"description": "Electronic devices and accessories",
	}, auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create category failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	assert.True(t, r.Success)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	categoryID := int64(data["id"].(float64))
	assert.NotZero(t, categoryID)
	assert.Equal(t, "Electronics", data["name"])

	// Get the category
	resp = doGet(t, fmt.Sprintf("/api/v1/product-categories/%d", categoryID), auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)

	var getData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &getData))
	assert.Equal(t, float64(categoryID), getData["id"])
}

func TestProductCategory_List(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	// Create multiple categories
	categories := []struct{ code, name string }{
		{"FOOD", "Food"},
		{"BVRG", "Beverages"},
		{"SNCK", "Snacks"},
	}
	for _, c := range categories {
		resp := doPost(t, "/api/v1/product-categories", map[string]interface{}{
			"code": c.code,
			"name": c.name,
		}, auth.Token)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create category failed: %s", string(resp.Body))
	}

	// List categories
	resp := doGet(t, "/api/v1/product-categories", auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r := decodeResponse(t, resp)
	assert.True(t, r.Success)

	var items []map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &items))
	assert.GreaterOrEqual(t, len(items), 3)
}

func TestProduct_CreateAndGet(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	categoryID := createTestCategory(t, auth.Token)

	// Create a product
	resp := doPost(t, "/api/v1/products", map[string]interface{}{
		"name":                "Test Product",
		"description":         "A test product description",
		"product_category_id": categoryID,
		"variants": []map[string]interface{}{
			{
				"sku":           "TEST-001",
				"name":          "Default Variant",
				"price":         99.99,
				"standard_cost": 50.00,
				"is_default":    true,
			},
		},
	}, auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create product failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	assert.True(t, r.Success)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	productID := int64(data["id"].(float64))
	assert.NotZero(t, productID)
	assert.Equal(t, "Test Product", data["name"])

	// Verify variants
	variants := data["variants"].([]interface{})
	assert.Len(t, variants, 1)
	variant := variants[0].(map[string]interface{})
	assert.Equal(t, "TEST-001", variant["sku"])
	assert.Equal(t, 99.99, variant["price"])

	// Get the product
	resp = doGet(t, fmt.Sprintf("/api/v1/products/%d", productID), auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestProduct_List(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	categoryID := createTestCategory(t, auth.Token)

	// Create multiple products
	for i := 0; i < 3; i++ {
		resp := doPost(t, "/api/v1/products", map[string]interface{}{
			"name":                fmt.Sprintf("Product %d", i+1),
			"product_category_id": categoryID,
			"variants": []map[string]interface{}{
				{
					"sku":           fmt.Sprintf("SKU-%d", i+1),
					"name":          "Default",
					"price":         float64(50 + i*10),
					"standard_cost": float64(25 + i*5),
					"is_default":    true,
				},
			},
		}, auth.Token)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create product failed: %s", string(resp.Body))
	}

	// List products
	resp := doGet(t, "/api/v1/products", auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r := decodeResponse(t, resp)
	assert.True(t, r.Success)

	var items []map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &items))
	assert.GreaterOrEqual(t, len(items), 3)
}

func TestProduct_Update(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	categoryID := createTestCategory(t, auth.Token)
	productID, _ := createTestProduct(t, auth.Token, categoryID, "Original Product", "ORIG-001", 100.00, 50.00)

	// Update the product (variants are required for update)
	resp := doPut(t, fmt.Sprintf("/api/v1/products/%d", productID), map[string]interface{}{
		"name":        "Updated Product",
		"description": "Updated description",
		"variants": []map[string]interface{}{
			{
				"sku":           "ORIG-001",
				"name":          "Default",
				"price":         120.00,
				"standard_cost": 60.00,
				"is_default":    true,
			},
		},
	}, auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r := decodeResponse(t, resp)
	assert.True(t, r.Success)

	var updateData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &updateData))
	assert.Equal(t, "Updated Product", updateData["name"])
	assert.Equal(t, "Updated description", updateData["description"])
}

func TestProduct_Delete(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	categoryID := createTestCategory(t, auth.Token)
	productID, _ := createTestProduct(t, auth.Token, categoryID, "To Be Deleted", "DEL-001", 100.00, 50.00)

	// Delete the product
	resp := doDelete(t, fmt.Sprintf("/api/v1/products/%d", productID), auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify deletion
	resp = doGet(t, fmt.Sprintf("/api/v1/products/%d", productID), auth.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProduct_WithMultipleVariants(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	categoryID := createTestCategory(t, auth.Token)

	// Create a product with multiple variants
	resp := doPost(t, "/api/v1/products", map[string]interface{}{
		"name":                "T-Shirt",
		"description":         "Cotton T-Shirt",
		"product_category_id": categoryID,
		"variants": []map[string]interface{}{
			{
				"sku":           "SHIRT-S",
				"name":          "Small",
				"price":         25.00,
				"standard_cost": 10.00,
				"is_default":    true,
				"attributes": map[string]interface{}{
					"size": "S",
				},
			},
			{
				"sku":           "SHIRT-M",
				"name":          "Medium",
				"price":         25.00,
				"standard_cost": 10.00,
				"is_default":    false,
				"attributes": map[string]interface{}{
					"size": "M",
				},
			},
			{
				"sku":           "SHIRT-L",
				"name":          "Large",
				"price":         27.00,
				"standard_cost": 11.00,
				"is_default":    false,
				"attributes": map[string]interface{}{
					"size": "L",
				},
			},
		},
	}, auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create product failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	variants := data["variants"].([]interface{})
	assert.Len(t, variants, 3)

	// Verify variant attributes
	var defaultFound bool
	for _, v := range variants {
		variant := v.(map[string]interface{})
		if variant["is_default"].(bool) {
			defaultFound = true
			assert.Equal(t, "Small", variant["name"])
		}
	}
	assert.True(t, defaultFound, "Should have a default variant")
}

func TestProduct_ValidationErrors(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
	}{
		{
			name: "empty name",
			body: map[string]interface{}{
				"name": "",
				"variants": []map[string]interface{}{
					{
						"sku":           "SKU-001",
						"name":          "Default",
						"price":         100.00,
						"standard_cost": 50.00,
						"is_default":    true,
					},
				},
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "missing variants",
			body: map[string]interface{}{
				"name": "Product Without Variants",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "negative price",
			body: map[string]interface{}{
				"name": "Product",
				"variants": []map[string]interface{}{
					{
						"sku":           "SKU-001",
						"name":          "Default",
						"price":         -10.00,
						"standard_cost": 50.00,
						"is_default":    true,
					},
				},
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := doPost(t, "/api/v1/products", tt.body, auth.Token)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestProduct_Search(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	categoryID := createTestCategory(t, auth.Token)

	// Create products with distinct names
	products := []struct{ name, sku string }{
		{"Apple iPhone", "SEARCH-1"},
		{"Samsung Galaxy", "SEARCH-2"},
		{"Apple Watch", "SEARCH-3"},
	}
	for _, p := range products {
		resp := doPost(t, "/api/v1/products", map[string]interface{}{
			"name":                p.name,
			"product_category_id": categoryID,
			"variants": []map[string]interface{}{
				{
					"sku":           p.sku,
					"name":          "Default",
					"price":         100.00,
					"standard_cost": 50.00,
					"is_default":    true,
				},
			},
		}, auth.Token)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create product failed: %s", string(resp.Body))
	}

	// Search for "Apple"
	resp := doGet(t, "/api/v1/products?search=Apple", auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r := decodeResponse(t, resp)

	var items []map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &items))
	assert.Len(t, items, 2)

	for _, product := range items {
		assert.Contains(t, product["name"], "Apple")
	}
}
