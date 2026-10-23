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

var customerKeys = []string{"id", "code", "name", "phone", "email", "is_member", "created_at", "updated_at"}

func TestCustomer_ShapeAndMessages(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	resp := doPost(t, "/api/v1/customers", map[string]any{"code": "C-1", "name": "Ani"}, auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "%s", resp.Body)
	r := decodeResponse(t, resp)
	assert.Equal(t, "Customer created successfully", r.Message)
	created := decodeObject(t, r.Data)
	assertKeys(t, created, customerKeys...)
	assert.Nil(t, created["phone"])
	assert.Nil(t, created["email"])
	assert.Equal(t, false, created["is_member"])
	path := fmt.Sprintf("/api/v1/customers/%d", int64(created["id"].(float64)))

	r = decodeResponse(t, doGet(t, path, auth.Token))
	assert.Equal(t, "Customer retrieved successfully", r.Message)
	assertKeys(t, decodeObject(t, r.Data), customerKeys...)

	// omitempty only skips a missing email; an empty string is still checked.
	resp = doPut(t, path, map[string]any{"code": "C-1", "name": "Ani B", "email": ""}, auth.Token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "%s", resp.Body)
	resp = doPut(t, path, map[string]any{"code": "C-1", "name": "Ani B", "email": "ani@b.id", "phone": "", "is_member": true}, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	updated := decodeObject(t, decodeResponse(t, resp).Data)
	assert.Equal(t, "ani@b.id", updated["email"])
	assert.Equal(t, "", updated["phone"])
	assert.Equal(t, true, updated["is_member"])

	r = decodeResponse(t, doGet(t, "/api/v1/customers", auth.Token))
	assert.Equal(t, "Customer list retrieved", r.Message)
	assert.JSONEq(t,
		`{"pagination":{"total_records":1,"total_pages":1,"current_page":1,"per_page":20,"count":0,"next_page":null,"prev_page":null}}`,
		string(r.Meta))

	resp = doPost(t, "/api/v1/customers", map[string]any{"code": "C-1", "name": "Dup"}, auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"Customer code already exists","error_code":"BAD_REQUEST"}`, string(resp.Body))
	resp = doPost(t, "/api/v1/customers", map[string]any{"code": "C-2", "name": "X", "email": "nope"}, auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"Validation failed","error_code":"VALIDATION_ERROR","errors":[{"field":"email","message":"Invalid email format"}]}`, string(resp.Body))
	resp = doGet(t, "/api/v1/customers/abc", auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"invalid customer ID","error_code":"BAD_REQUEST"}`, string(resp.Body))
	resp = doGet(t, "/api/v1/customers/999999", auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"Customer not found not found","error_code":"NOT_FOUND"}`, string(resp.Body))

	resp = doDelete(t, path, auth.Token)
	assert.JSONEq(t, `{"success":true,"message":"Customer deleted successfully"}`, string(resp.Body))
}

func TestCustomer_OtherCompanyIsNotFound(t *testing.T) {
	cleanupDatabase(t)
	a := registerTestUser(t)
	b := registerTestUser(t)

	resp := doPost(t, "/api/v1/customers", map[string]any{"code": "C-1", "name": "A's customer"}, a.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	path := fmt.Sprintf("/api/v1/customers/%d", int64(decodeObject(t, decodeResponse(t, resp).Data)["id"].(float64)))

	assert.Equal(t, http.StatusNotFound, doGet(t, path, b.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doPut(t, path, map[string]any{"code": "C-1", "name": "Taken"}, b.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doDelete(t, path, b.Token).StatusCode)

	// B may use the same code; B's list doesn't show A's customer.
	resp = doPost(t, "/api/v1/customers", map[string]any{"code": "C-1", "name": "B's customer"}, b.Token)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(decodeResponse(t, doGet(t, "/api/v1/customers", b.Token)).Data, &rows))
	require.Len(t, rows, 1)
	assert.Equal(t, "B's customer", rows[0]["name"])
	assert.Equal(t, http.StatusOK, doGet(t, path, a.Token).StatusCode)
}
