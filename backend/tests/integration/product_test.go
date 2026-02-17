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

func TestProductCategory_CreateAndGet(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)
	_ = testData

	// Create a category
	categoryBody := map[string]interface{}{
		"name":        "Electronics",
		"description": "Electronic devices and accessories",
	}

	resp, err := testServer.POST("/api/v1/product-categories", categoryBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	assert.True(t, createResp["success"].(bool))
	data := createResp["data"].(map[string]interface{})
	categoryID := int64(data["id"].(float64))
	assert.NotZero(t, categoryID)
	assert.Equal(t, "Electronics", data["name"])

	// Get the category
	resp, err = testServer.GET(fmt.Sprintf("/api/v1/product-categories/%d", categoryID), "")
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
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)
	_ = testData

	// Create multiple categories
	categories := []string{"Food", "Beverages", "Snacks"}
	for _, name := range categories {
		categoryBody := map[string]interface{}{
			"name": name,
		}

		resp, err := testServer.POST("/api/v1/product-categories", categoryBody, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List categories
	resp, err := testServer.GET("/api/v1/product-categories", "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &listResp)
	require.NoError(t, err)

	assert.True(t, listResp["success"].(bool))
	data := listResp["data"].([]interface{})
	// 3 created + 1 from base test data
	assert.GreaterOrEqual(t, len(data), 3)
}

func TestProduct_CreateAndGet(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create a product
	productBody := map[string]interface{}{
		"name":                "Test Product",
		"description":         "A test product description",
		"product_category_id": testData.ProductCategory.ID,
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

	resp, err := testServer.POST("/api/v1/products", productBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

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
	resp, err = testServer.GET(fmt.Sprintf("/api/v1/products/%d", productID), "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestProduct_List(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create multiple products
	for i := 0; i < 3; i++ {
		productBody := map[string]interface{}{
			"name":                fmt.Sprintf("Product %d", i+1),
			"product_category_id": testData.ProductCategory.ID,
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

		resp, err := testServer.POST("/api/v1/products", productBody, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List products
	resp, err := testServer.GET("/api/v1/products", "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &listResp)
	require.NoError(t, err)

	assert.True(t, listResp["success"].(bool))
	data := listResp["data"].([]interface{})
	// 3 created + 1 from base test data
	assert.GreaterOrEqual(t, len(data), 3)
}

func TestProduct_Update(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create a product
	productBody := map[string]interface{}{
		"name":                "Original Product",
		"product_category_id": testData.ProductCategory.ID,
		"variants": []map[string]interface{}{
			{
				"sku":           "ORIG-001",
				"name":          "Original Variant",
				"price":         100.00,
				"standard_cost": 50.00,
				"is_default":    true,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/products", productBody, "")
	require.NoError(t, err)

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	productID := int64(data["id"].(float64))

	// Update the product
	updateBody := map[string]interface{}{
		"name":        "Updated Product",
		"description": "Updated description",
	}

	resp, err = testServer.PUT(fmt.Sprintf("/api/v1/products/%d", productID), updateBody, "")
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
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create a product
	productBody := map[string]interface{}{
		"name":                "To Be Deleted",
		"product_category_id": testData.ProductCategory.ID,
		"variants": []map[string]interface{}{
			{
				"sku":           "DEL-001",
				"name":          "Default",
				"price":         100.00,
				"standard_cost": 50.00,
				"is_default":    true,
			},
		},
	}

	resp, err := testServer.POST("/api/v1/products", productBody, "")
	require.NoError(t, err)

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	productID := int64(data["id"].(float64))

	// Delete the product
	resp, err = testServer.DELETE(fmt.Sprintf("/api/v1/products/%d", productID), "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify deletion
	resp, err = testServer.GET(fmt.Sprintf("/api/v1/products/%d", productID), "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestProduct_WithMultipleVariants(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create a product with multiple variants
	productBody := map[string]interface{}{
		"name":                "T-Shirt",
		"description":         "Cotton T-Shirt",
		"product_category_id": testData.ProductCategory.ID,
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

	resp, err := testServer.POST("/api/v1/products", productBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

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
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)
	_ = testData

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
			resp, err := testServer.POST("/api/v1/products", tt.body, "")
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestProduct_Search(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)

	// Create products with distinct names
	products := []string{"Apple iPhone", "Samsung Galaxy", "Apple Watch"}
	for i, name := range products {
		productBody := map[string]interface{}{
			"name":                name,
			"product_category_id": testData.ProductCategory.ID,
			"variants": []map[string]interface{}{
				{
					"sku":           fmt.Sprintf("SEARCH-%d", i+1),
					"name":          "Default",
					"price":         100.00,
					"standard_cost": 50.00,
					"is_default":    true,
				},
			},
		}

		resp, err := testServer.POST("/api/v1/products", productBody, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Search for "Apple"
	resp, err := testServer.GET("/api/v1/products?search=Apple", "")
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
