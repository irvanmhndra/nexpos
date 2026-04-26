package integration

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
	categoryBody := map[string]interface{}{
		"code": fmt.Sprintf("CAT-%d", n),
		"name": fmt.Sprintf("Test Category %d", n),
	}
	resp, err := testEnv.Server.POST("/api/v1/product-categories", categoryBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create category failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &createResp))
	data := createResp["data"].(map[string]interface{})
	return int64(data["id"].(float64))
}

// createTestProduct creates a product with one variant via the API and returns (productID, variantID).
func createTestProduct(t *testing.T, token string, categoryID int64, name, sku string, price, cost float64) (int64, int64) {
	t.Helper()
	productBody := map[string]interface{}{
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
	}

	resp, err := testEnv.Server.POST("/api/v1/products", productBody, token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create product failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	require.NoError(t, json.Unmarshal(resp.Body, &createResp))
	data := createResp["data"].(map[string]interface{})
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
	categoryBody := map[string]interface{}{
		"code":        "ELEC",
		"name":        "Electronics",
		"description": "Electronic devices and accessories",
	}

	resp, err := testEnv.Server.POST("/api/v1/product-categories", categoryBody, auth.Token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create category failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	assert.True(t, createResp["success"].(bool))
	data := createResp["data"].(map[string]interface{})
	categoryID := int64(data["id"].(float64))
	assert.NotZero(t, categoryID)
	assert.Equal(t, "Electronics", data["name"])

	// Get the category
	resp, err = testEnv.Server.GET(fmt.Sprintf("/api/v1/product-categories/%d", categoryID), auth.Token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var getResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &getResp)
	require.NoError(t, err)

	assert.True(t, getResp["success"].(bool))
	getData := getResp["data"].(map[string]interface{})
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
		categoryBody := map[string]interface{}{
			"code": c.code,
			"name": c.name,
		}

		resp, err := testEnv.Server.POST("/api/v1/product-categories", categoryBody, auth.Token)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create category failed: %s", string(resp.Body))
	}

	// List categories
	resp, err := testEnv.Server.GET("/api/v1/product-categories", auth.Token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &listResp)
	require.NoError(t, err)

	assert.True(t, listResp["success"].(bool))
	data := listResp["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 3)
}

func TestProduct_CreateAndGet(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	categoryID := createTestCategory(t, auth.Token)

	// Create a product
	productBody := map[string]interface{}{
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
	}

	resp, err := testEnv.Server.POST("/api/v1/products", productBody, auth.Token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create product failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	assert.True(t, createResp["success"].(bool))
	data := createResp["data"].(map[string]interface{})
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
	resp, err = testEnv.Server.GET(fmt.Sprintf("/api/v1/products/%d", productID), auth.Token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestProduct_List(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	categoryID := createTestCategory(t, auth.Token)

	// Create multiple products
	for i := 0; i < 3; i++ {
		productBody := map[string]interface{}{
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
		}

		resp, err := testEnv.Server.POST("/api/v1/products", productBody, auth.Token)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create product failed: %s", string(resp.Body))
	}

	// List products
	resp, err := testEnv.Server.GET("/api/v1/products", auth.Token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &listResp)
	require.NoError(t, err)

	assert.True(t, listResp["success"].(bool))
	data := listResp["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 3)
}

func TestProduct_Update(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	categoryID := createTestCategory(t, auth.Token)
	productID, _ := createTestProduct(t, auth.Token, categoryID, "Original Product", "ORIG-001", 100.00, 50.00)

	// Update the product (variants are required for update)
	updateBody := map[string]interface{}{
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
	}

	resp, err := testEnv.Server.PUT(fmt.Sprintf("/api/v1/products/%d", productID), updateBody, auth.Token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updateResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &updateResp)
	require.NoError(t, err)

	assert.True(t, updateResp["success"].(bool))
	updateData := updateResp["data"].(map[string]interface{})
	assert.Equal(t, "Updated Product", updateData["name"])
	assert.Equal(t, "Updated description", updateData["description"])
}

func TestProduct_Delete(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	categoryID := createTestCategory(t, auth.Token)
	productID, _ := createTestProduct(t, auth.Token, categoryID, "To Be Deleted", "DEL-001", 100.00, 50.00)

	// Delete the product
	resp, err := testEnv.Server.DELETE(fmt.Sprintf("/api/v1/products/%d", productID), auth.Token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify deletion
	resp, err = testEnv.Server.GET(fmt.Sprintf("/api/v1/products/%d", productID), auth.Token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProduct_WithMultipleVariants(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	categoryID := createTestCategory(t, auth.Token)

	// Create a product with multiple variants
	productBody := map[string]interface{}{
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
	}

	resp, err := testEnv.Server.POST("/api/v1/products", productBody, auth.Token)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create product failed: %s", string(resp.Body))

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
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
			resp, err := testEnv.Server.POST("/api/v1/products", tt.body, auth.Token)
			require.NoError(t, err)
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
		productBody := map[string]interface{}{
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
		}

		resp, err := testEnv.Server.POST("/api/v1/products", productBody, auth.Token)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create product failed: %s", string(resp.Body))
	}

	// Search for "Apple"
	resp, err := testEnv.Server.GET("/api/v1/products?search=Apple", auth.Token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &listResp)
	require.NoError(t, err)

	data := listResp["data"].([]interface{})
	assert.Len(t, data, 2)

	for _, item := range data {
		product := item.(map[string]interface{})
		assert.Contains(t, product["name"], "Apple")
	}
}
