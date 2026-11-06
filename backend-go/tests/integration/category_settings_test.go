package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createCategory(t *testing.T, token string, body map[string]any) map[string]any {
	t.Helper()
	resp := doPost(t, "/api/v1/product-categories", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create category: %s", resp.Body)
	return decodeObject(t, decodeResponse(t, resp).Data)
}

var categoryKeys = []string{"id", "code", "name", "parent_id", "sort_order", "is_active", "created_at", "updated_at"}

func categoryNames(t *testing.T, raw json.RawMessage) []string {
	t.Helper()
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(raw, &rows), string(raw))
	names := make([]string, len(rows))
	for i, row := range rows {
		names[i] = row["name"].(string)
	}
	return names
}

func TestCategory_CreateGetUpdateDelete(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	parent := createCategory(t, auth.Token, map[string]any{"code": "DRINK", "name": "Minuman", "sort_order": 2})
	assertKeys(t, parent, categoryKeys...) // no parent_name without a parent
	assert.Nil(t, parent["parent_id"])
	assert.Equal(t, true, parent["is_active"], "is_active defaults to true")
	assert.Equal(t, float64(2), parent["sort_order"])
	parentID := int64(parent["id"].(float64))

	child := createCategory(t, auth.Token, map[string]any{"code": "COFFEE", "name": "Kopi", "parent_id": parentID, "is_active": false})
	assertKeys(t, child, append([]string{"parent_name"}, categoryKeys...)...)
	assert.Equal(t, "Minuman", child["parent_name"])
	assert.Equal(t, false, child["is_active"])
	childPath := fmt.Sprintf("/api/v1/product-categories/%d", int64(child["id"].(float64)))

	r := decodeResponse(t, doGet(t, childPath, auth.Token))
	assert.Equal(t, "Category retrieved successfully", r.Message)
	assert.Equal(t, "Minuman", decodeObject(t, r.Data)["parent_name"])

	// is_active left out keeps the current value; parent_id left out clears the parent.
	resp := doPut(t, childPath, map[string]any{"code": "COFFEE", "name": "Kopi Susu", "sort_order": 5}, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	r = decodeResponse(t, resp)
	assert.Equal(t, "Category updated successfully", r.Message)
	updated := decodeObject(t, r.Data)
	assertKeys(t, updated, categoryKeys...)
	assert.Equal(t, "Kopi Susu", updated["name"])
	assert.Nil(t, updated["parent_id"])
	assert.Equal(t, false, updated["is_active"])
	assert.Equal(t, float64(5), updated["sort_order"])

	resp = doDelete(t, childPath, auth.Token)
	assert.JSONEq(t, `{"success":true,"message":"Category deleted successfully"}`, string(resp.Body))
	assert.Equal(t, http.StatusNotFound, doGet(t, childPath, auth.Token).StatusCode)
}

func TestCategory_ListAndAll(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	createCategory(t, auth.Token, map[string]any{"code": "B", "name": "Beta", "sort_order": 1})
	createCategory(t, auth.Token, map[string]any{"code": "A", "name": "Alpha", "sort_order": 1})
	createCategory(t, auth.Token, map[string]any{"code": "Z", "name": "Zeta", "sort_order": 0, "is_active": false})

	r := decodeResponse(t, doGet(t, "/api/v1/product-categories", auth.Token))
	assert.Equal(t, "Category list retrieved", r.Message)
	// By sort_order, then name.
	assert.Equal(t, []string{"Zeta", "Alpha", "Beta"}, categoryNames(t, r.Data))
	assert.JSONEq(t,
		`{"pagination":{"total_records":3,"total_pages":1,"current_page":1,"per_page":20,"count":0,"next_page":null,"prev_page":null}}`,
		string(r.Meta))

	for path, want := range map[string][]string{
		"/api/v1/product-categories?search=alp":         {"Alpha"},
		"/api/v1/product-categories?search=z":           {"Zeta"},
		"/api/v1/product-categories?is_active=false":    {"Zeta"},
		"/api/v1/product-categories?per_page=1&page=2":  {"Alpha"},
		"/api/v1/product-categories/all":                {"Alpha", "Beta"},
		"/api/v1/product-categories/all?search=ignored": {"Alpha", "Beta"},
	} {
		assert.Equal(t, want, categoryNames(t, decodeResponse(t, doGet(t, path, auth.Token)).Data), path)
	}

	r = decodeResponse(t, doGet(t, "/api/v1/product-categories/all", auth.Token))
	assert.Equal(t, "All categories retrieved", r.Message)
	assert.Empty(t, r.Meta)

	other := registerTestUser(t)
	assert.JSONEq(t, `{"success":true,"message":"All categories retrieved","data":[]}`,
		string(doGet(t, "/api/v1/product-categories/all", other.Token).Body))
}

func TestCategory_Errors(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	parent := createCategory(t, auth.Token, map[string]any{"code": "P", "name": "Parent"})
	parentID := int64(parent["id"].(float64))
	child := createCategory(t, auth.Token, map[string]any{"code": "C", "name": "Child", "parent_id": parentID})
	childID := int64(child["id"].(float64))
	parentPath := fmt.Sprintf("/api/v1/product-categories/%d", parentID)

	badRequest := func(message string, resp int, body string) {
		t.Helper()
		assert.Equal(t, http.StatusBadRequest, resp, body)
		assert.JSONEq(t, fmt.Sprintf(`{"success":false,"message":%q,"error_code":"BAD_REQUEST"}`, message), body)
	}

	r := doPost(t, "/api/v1/product-categories", map[string]any{"code": "P", "name": "Dup"}, auth.Token)
	badRequest("Category code already exists", r.StatusCode, string(r.Body))
	r = doPost(t, "/api/v1/product-categories", map[string]any{"code": "N", "name": "No parent", "parent_id": 999999}, auth.Token)
	badRequest("Parent category not found", r.StatusCode, string(r.Body))
	r = doPut(t, parentPath, map[string]any{"code": "P", "name": "Self", "parent_id": parentID}, auth.Token)
	badRequest("Category cannot be its own parent", r.StatusCode, string(r.Body))
	r = doPut(t, parentPath, map[string]any{"code": "P", "name": "Loop", "parent_id": childID}, auth.Token)
	badRequest("Circular reference detected", r.StatusCode, string(r.Body))
	r = doDelete(t, parentPath, auth.Token)
	badRequest("Cannot delete category with subcategories", r.StatusCode, string(r.Body))
	r = doGet(t, "/api/v1/product-categories/abc", auth.Token)
	badRequest("invalid category ID", r.StatusCode, string(r.Body))

	r = doGet(t, "/api/v1/product-categories/999999", auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"Category not found not found","error_code":"NOT_FOUND"}`, string(r.Body))

	r = doPost(t, "/api/v1/product-categories", map[string]any{"code": "", "name": "", "sort_order": -1}, auth.Token)
	assert.Equal(t, http.StatusUnprocessableEntity, r.StatusCode)
	assert.JSONEq(t, `{"success":false,"message":"Validation failed","error_code":"VALIDATION_ERROR","errors":[
		{"field":"code","message":"code is required"},
		{"field":"name","message":"name is required"},
		{"field":"sort_order","message":"sort_order is invalid"}]}`, string(r.Body))
	r = doPost(t, "/api/v1/product-categories", map[string]any{"code": "X", "name": "X", "sort_order": "first"}, auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"Validation failed","error_code":"VALIDATION_ERROR","errors":[
		{"field":"sort_order","message":"must be a number"}]}`, string(r.Body))

	// A category with products can't be deleted.
	resp := doPost(t, "/api/v1/products", map[string]any{
		"name": "Kopi", "category_id": childID,
		"variants": []map[string]any{{"sku": "SKU-1", "name": "Default", "price": 10000, "standard_cost": 5000, "is_default": true}},
	}, auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "%s", resp.Body)
	r = doDelete(t, fmt.Sprintf("/api/v1/product-categories/%d", childID), auth.Token)
	badRequest("Cannot delete category with products", r.StatusCode, string(r.Body))
}

func TestCategory_OtherCompanyIsNotFound(t *testing.T) {
	cleanupDatabase(t)
	a := registerTestUser(t)
	b := registerTestUser(t)
	cat := createCategory(t, a.Token, map[string]any{"code": "A", "name": "A's"})
	id := int64(cat["id"].(float64))
	path := fmt.Sprintf("/api/v1/product-categories/%d", id)

	assert.Equal(t, http.StatusNotFound, doGet(t, path, b.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doPut(t, path, map[string]any{"code": "A", "name": "Taken"}, b.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doDelete(t, path, b.Token).StatusCode)

	// A foreign parent counts as not found, and codes are per company.
	resp := doPost(t, "/api/v1/product-categories", map[string]any{"code": "A", "name": "B's", "parent_id": id}, b.Token)
	assert.JSONEq(t, `{"success":false,"message":"Parent category not found","error_code":"BAD_REQUEST"}`, string(resp.Body))
	createCategory(t, b.Token, map[string]any{"code": "A", "name": "B's"})
	assert.Equal(t, []string{"B's"}, categoryNames(t, decodeResponse(t, doGet(t, "/api/v1/product-categories", b.Token)).Data))
}

var settingsKeys = []string{
	"id", "company_id", "tax_enabled", "tax_rate", "tax_inclusive", "rounding_enabled", "rounding_amount",
	"auto_complete_counter_orders", "require_customer_for_delivery", "receipt_header", "receipt_footer",
	"show_tax_on_receipt", "offline_mode_enabled", "max_offline_days", "created_at", "updated_at",
}

func TestCompanySettings_DefaultsUpdateAndTenancy(t *testing.T) {
	cleanupDatabase(t)
	a := registerTestUser(t)
	b := registerTestUser(t)

	// No row yet: defaults, with id 0 and Go's zero time.
	resp := doGet(t, "/api/v1/company-settings", a.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	assert.JSONEq(t, fmt.Sprintf(`{"success":true,"message":"Company settings retrieved successfully","data":{
		"id":0,"company_id":%d,"tax_enabled":false,"tax_rate":0,"tax_inclusive":true,
		"rounding_enabled":false,"rounding_amount":0,"auto_complete_counter_orders":true,
		"require_customer_for_delivery":true,"receipt_header":null,"receipt_footer":null,
		"show_tax_on_receipt":true,"offline_mode_enabled":false,"max_offline_days":7,
		"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}}`, a.CompanyID), string(resp.Body))

	// Partial update; the response echoes the request's decimals, unrounded.
	resp = doPut(t, "/api/v1/company-settings", map[string]any{
		"tax_enabled": true, "tax_rate": 11.555, "rounding_amount": "100.50", "receipt_header": "Halo", "max_offline_days": 3,
	}, a.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	r := decodeResponse(t, resp)
	assert.Equal(t, "Company settings updated successfully", r.Message)
	got := decodeObject(t, r.Data)
	assertKeys(t, got, settingsKeys...)
	assert.NotZero(t, got["id"])
	assert.Equal(t, 11.555, got["tax_rate"])
	assert.Equal(t, 100.5, got["rounding_amount"])
	assert.Equal(t, "Halo", got["receipt_header"])
	assert.Equal(t, float64(3), got["max_offline_days"])
	assert.Equal(t, true, got["tax_inclusive"], "untouched fields keep their value")
	assert.Contains(t, string(r.Data), `"rounding_amount":100.5,`)

	// The stored value is rounded by the column, and trailing zeros are dropped.
	got = decodeObject(t, decodeResponse(t, doGet(t, "/api/v1/company-settings", a.Token)).Data)
	assert.Equal(t, 11.56, got["tax_rate"])
	assert.Equal(t, 100.5, got["rounding_amount"])

	resp = doPut(t, "/api/v1/company-settings", map[string]any{"tax_rate": 11, "receipt_header": nil, "receipt_footer": ""}, a.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	got = decodeObject(t, decodeResponse(t, resp).Data)
	assert.Contains(t, string(resp.Body), `"tax_rate":11,`)
	assert.Equal(t, "Halo", got["receipt_header"], "null leaves a field untouched")
	assert.Equal(t, "", got["receipt_footer"])

	// Validation runs on every field that is present.
	resp = doPut(t, "/api/v1/company-settings", map[string]any{"tax_rate": 101, "rounding_amount": -1, "max_offline_days": 0}, a.Token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.JSONEq(t, `{"success":false,"message":"Validation failed","error_code":"VALIDATION_ERROR","errors":[
		{"field":"tax_rate","message":"tax_rate is invalid"},
		{"field":"rounding_amount","message":"rounding_amount is invalid"},
		{"field":"max_offline_days","message":"max_offline_days is invalid"}]}`, string(resp.Body))

	// Bad decimals are a bad body, not a field error.
	resp = doPut(t, "/api/v1/company-settings", map[string]any{"tax_rate": "abc"}, a.Token)
	assert.JSONEq(t, `{"success":false,"message":"invalid request body","error_code":"BAD_REQUEST"}`, string(resp.Body))
	resp = doPut(t, "/api/v1/company-settings", map[string]any{"tax_rate": true}, a.Token)
	assert.JSONEq(t, `{"success":false,"message":"invalid request body","error_code":"BAD_REQUEST"}`, string(resp.Body))

	// Another company is unaffected.
	other := decodeObject(t, decodeResponse(t, doGet(t, "/api/v1/company-settings", b.Token)).Data)
	assert.Equal(t, float64(0), other["tax_rate"])
	assert.Equal(t, float64(b.CompanyID), other["company_id"])
}
