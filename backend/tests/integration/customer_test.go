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

func TestCustomer_CreateAndGet(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	// Create base test data (company, branch, etc.)
	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)
	_ = testData // Company ID is used internally

	// Create a customer
	customerBody := map[string]interface{}{
		"name":      "John Customer",
		"phone":     "+1234567890",
		"email":     "john.customer@example.com",
		"is_member": true,
	}

	resp, err := testServer.POST("/api/v1/customers", customerBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	assert.True(t, createResp["success"].(bool))
	data := createResp["data"].(map[string]interface{})
	customerID := int64(data["id"].(float64))
	assert.NotZero(t, customerID)
	assert.Equal(t, "John Customer", data["name"])
	assert.NotEmpty(t, data["code"])

	// Get the customer
	resp, err = testServer.GET(fmt.Sprintf("/api/v1/customers/%d", customerID), "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var getResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &getResp)
	require.NoError(t, err)

	assert.True(t, getResp["success"].(bool))
	getData := getResp["data"].(map[string]interface{})
	assert.Equal(t, float64(customerID), getData["id"])
	assert.Equal(t, "John Customer", getData["name"])
}

func TestCustomer_Update(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)
	_ = testData

	// Create a customer
	customerBody := map[string]interface{}{
		"name":      "Original Name",
		"phone":     "+1111111111",
		"is_member": false,
	}

	resp, err := testServer.POST("/api/v1/customers", customerBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	customerID := int64(data["id"].(float64))

	// Update the customer
	updateBody := map[string]interface{}{
		"name":      "Updated Name",
		"phone":     "+2222222222",
		"email":     "updated@example.com",
		"is_member": true,
	}

	resp, err = testServer.PUT(fmt.Sprintf("/api/v1/customers/%d", customerID), updateBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updateResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &updateResp)
	require.NoError(t, err)

	assert.True(t, updateResp["success"].(bool))
	updateData := updateResp["data"].(map[string]interface{})
	assert.Equal(t, "Updated Name", updateData["name"])
	assert.Equal(t, "+2222222222", updateData["phone"])
	assert.Equal(t, "updated@example.com", updateData["email"])
	assert.Equal(t, true, updateData["is_member"])
}

func TestCustomer_Delete(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)
	_ = testData

	// Create a customer
	customerBody := map[string]interface{}{
		"name": "To Be Deleted",
	}

	resp, err := testServer.POST("/api/v1/customers", customerBody, "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	customerID := int64(data["id"].(float64))

	// Delete the customer
	resp, err = testServer.DELETE(fmt.Sprintf("/api/v1/customers/%d", customerID), "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify deletion
	resp, err = testServer.GET(fmt.Sprintf("/api/v1/customers/%d", customerID), "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCustomer_List(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)
	_ = testData

	// Create multiple customers
	for i := 0; i < 5; i++ {
		customerBody := map[string]interface{}{
			"name": fmt.Sprintf("Customer %d", i+1),
		}

		resp, err := testServer.POST("/api/v1/customers", customerBody, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List customers
	resp, err := testServer.GET("/api/v1/customers", "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &listResp)
	require.NoError(t, err)

	assert.True(t, listResp["success"].(bool))
	data := listResp["data"].([]interface{})
	// 5 created + 1 from base test data
	assert.GreaterOrEqual(t, len(data), 5)

	// Verify pagination
	meta := listResp["meta"].(map[string]interface{})
	pagination := meta["pagination"].(map[string]interface{})
	assert.GreaterOrEqual(t, pagination["total_records"].(float64), float64(5))
}

func TestCustomer_ValidationErrors(t *testing.T) {
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
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid email",
			body: map[string]interface{}{
				"name":  "Valid Name",
				"email": "not-an-email",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := testServer.POST("/api/v1/customers", tt.body, "")
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestCustomer_NotFound(t *testing.T) {
	cleanupDatabase(t)

	resp, err := testServer.GET("/api/v1/customers/99999", "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCustomer_ListWithPagination(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)
	_ = testData

	// Create 15 customers
	for i := 0; i < 15; i++ {
		customerBody := map[string]interface{}{
			"name": fmt.Sprintf("Paginated Customer %d", i+1),
		}

		resp, err := testServer.POST("/api/v1/customers", customerBody, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List first page
	resp, err := testServer.GET("/api/v1/customers?page=1&per_page=5", "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &listResp)
	require.NoError(t, err)

	data := listResp["data"].([]interface{})
	assert.Len(t, data, 5)

	meta := listResp["meta"].(map[string]interface{})
	pagination := meta["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["current_page"])
	assert.Equal(t, float64(5), pagination["per_page"])
	assert.NotNil(t, pagination["next_page"])

	// List second page
	resp, err = testServer.GET("/api/v1/customers?page=2&per_page=5", "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.Unmarshal(resp.Body, &listResp)
	require.NoError(t, err)

	meta = listResp["meta"].(map[string]interface{})
	pagination = meta["pagination"].(map[string]interface{})
	assert.Equal(t, float64(2), pagination["current_page"])
	assert.NotNil(t, pagination["prev_page"])
}

func TestCustomer_Search(t *testing.T) {
	cleanupDatabase(t)
	ctx := context.Background()

	testData, err := testFixture.CreateBaseTestData(ctx)
	require.NoError(t, err)
	_ = testData

	// Create customers with distinct names
	customers := []string{"Alice Smith", "Bob Johnson", "Alice Williams"}
	for _, name := range customers {
		customerBody := map[string]interface{}{
			"name": name,
		}

		resp, err := testServer.POST("/api/v1/customers", customerBody, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Search for "Alice"
	resp, err := testServer.GET("/api/v1/customers?search=Alice", "")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &listResp)
	require.NoError(t, err)

	data := listResp["data"].([]interface{})
	assert.Len(t, data, 2)

	for _, item := range data {
		customer := item.(map[string]interface{})
		assert.Contains(t, customer["name"], "Alice")
	}
}
