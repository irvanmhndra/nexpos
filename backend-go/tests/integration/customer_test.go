package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomer_CreateAndGet(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	// Create a customer
	resp := doPost(t, "/api/v1/customers", map[string]interface{}{
		"code":      "CUST-001",
		"name":      "John Customer",
		"phone":     "+1234567890",
		"email":     "john.customer@example.com",
		"is_member": true,
	}, auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create customer failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	assert.True(t, r.Success)

	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	customerID := int64(data["id"].(float64))
	assert.NotZero(t, customerID)
	assert.Equal(t, "John Customer", data["name"])
	assert.NotEmpty(t, data["code"])

	// Get the customer
	resp = doGet(t, fmt.Sprintf("/api/v1/customers/%d", customerID), auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)

	var getData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &getData))
	assert.Equal(t, float64(customerID), getData["id"])
	assert.Equal(t, "John Customer", getData["name"])
}

func TestCustomer_Update(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	// Create a customer
	resp := doPost(t, "/api/v1/customers", map[string]interface{}{
		"code":      "CUST-UPD",
		"name":      "Original Name",
		"phone":     "+1111111111",
		"is_member": false,
	}, auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create customer failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	customerID := int64(data["id"].(float64))

	// Update the customer
	resp = doPut(t, fmt.Sprintf("/api/v1/customers/%d", customerID), map[string]interface{}{
		"code":      "CUST-UPD",
		"name":      "Updated Name",
		"phone":     "+2222222222",
		"email":     "updated@example.com",
		"is_member": true,
	}, auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	assert.True(t, r.Success)

	var updateData map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &updateData))
	assert.Equal(t, "Updated Name", updateData["name"])
	assert.Equal(t, "+2222222222", updateData["phone"])
	assert.Equal(t, "updated@example.com", updateData["email"])
	assert.Equal(t, true, updateData["is_member"])
}

func TestCustomer_Delete(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	// Create a customer
	resp := doPost(t, "/api/v1/customers", map[string]interface{}{
		"code": "CUST-DEL",
		"name": "To Be Deleted",
	}, auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create customer failed: %s", string(resp.Body))

	r := decodeResponse(t, resp)
	var data map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &data))
	customerID := int64(data["id"].(float64))

	// Delete the customer
	resp = doDelete(t, fmt.Sprintf("/api/v1/customers/%d", customerID), auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify deletion
	resp = doGet(t, fmt.Sprintf("/api/v1/customers/%d", customerID), auth.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCustomer_List(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	// Create multiple customers
	for i := 0; i < 5; i++ {
		resp := doPost(t, "/api/v1/customers", map[string]interface{}{
			"code": fmt.Sprintf("CUST-L%d", i+1),
			"name": fmt.Sprintf("Customer %d", i+1),
		}, auth.Token)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create customer failed: %s", string(resp.Body))
	}

	// List customers
	resp := doGet(t, "/api/v1/customers", auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r := decodeResponse(t, resp)
	assert.True(t, r.Success)

	var items []map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &items))
	assert.GreaterOrEqual(t, len(items), 5)

	// Verify pagination
	var meta struct {
		Pagination struct {
			TotalRecords float64 `json:"total_records"`
		} `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(r.Meta, &meta))
	assert.GreaterOrEqual(t, meta.Pagination.TotalRecords, float64(5))
}

func TestCustomer_ValidationErrors(t *testing.T) {
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
				"code": "CUST-VAL1",
				"name": "",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid email",
			body: map[string]interface{}{
				"code":  "CUST-VAL2",
				"name":  "Valid Name",
				"email": "not-an-email",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := doPost(t, "/api/v1/customers", tt.body, auth.Token)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestCustomer_NotFound(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	resp := doGet(t, "/api/v1/customers/99999", auth.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCustomer_ListWithPagination(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	// Create 15 customers
	for i := 0; i < 15; i++ {
		resp := doPost(t, "/api/v1/customers", map[string]interface{}{
			"code": fmt.Sprintf("CUST-P%02d", i+1),
			"name": fmt.Sprintf("Paginated Customer %d", i+1),
		}, auth.Token)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create customer failed: %s", string(resp.Body))
	}

	// List first page
	resp := doGet(t, "/api/v1/customers?page=1&per_page=5", auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r := decodeResponse(t, resp)

	var items []map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &items))
	assert.Len(t, items, 5)

	var meta struct {
		Pagination struct {
			TotalRecords float64 `json:"total_records"`
			CurrentPage  float64 `json:"current_page"`
			PerPage      float64 `json:"per_page"`
			NextPage     *int    `json:"next_page"`
			PrevPage     *int    `json:"prev_page"`
		} `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(r.Meta, &meta))
	assert.Equal(t, float64(1), meta.Pagination.CurrentPage)
	assert.Equal(t, float64(5), meta.Pagination.PerPage)
	assert.NotNil(t, meta.Pagination.NextPage)

	// List second page
	resp = doGet(t, "/api/v1/customers?page=2&per_page=5", auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r = decodeResponse(t, resp)
	require.NoError(t, json.Unmarshal(r.Meta, &meta))
	assert.Equal(t, float64(2), meta.Pagination.CurrentPage)
	assert.NotNil(t, meta.Pagination.PrevPage)
}

func TestCustomer_Search(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	// Create customers with distinct names
	customers := []struct{ code, name string }{
		{"CUST-AS", "Alice Smith"},
		{"CUST-BJ", "Bob Johnson"},
		{"CUST-AW", "Alice Williams"},
	}
	for _, c := range customers {
		resp := doPost(t, "/api/v1/customers", map[string]interface{}{
			"code": c.code,
			"name": c.name,
		}, auth.Token)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create customer failed: %s", string(resp.Body))
	}

	// Search for "Alice"
	resp := doGet(t, "/api/v1/customers?search=Alice", auth.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	r := decodeResponse(t, resp)

	var items []map[string]interface{}
	require.NoError(t, json.Unmarshal(r.Data, &items))
	assert.Len(t, items, 2)

	for _, customer := range items {
		assert.Contains(t, customer["name"], "Alice")
	}
}
