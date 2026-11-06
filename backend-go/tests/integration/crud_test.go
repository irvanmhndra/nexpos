package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func created(t *testing.T, path, token string, body map[string]any) map[string]any {
	t.Helper()
	resp := doPost(t, path, body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "POST %s: %s", path, resp.Body)
	return decodeObject(t, decodeResponse(t, resp).Data)
}

func idOf(obj map[string]any) int64 { return int64(obj["id"].(float64)) }

func rowsOf(t *testing.T, raw json.RawMessage) []map[string]any {
	t.Helper()
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(raw, &rows), string(raw))
	return rows
}

func namesOf(t *testing.T, raw json.RawMessage, key string) []string {
	t.Helper()
	names := []string{}
	for _, row := range rowsOf(t, raw) {
		names = append(names, row[key].(string))
	}
	return names
}

const validationPrefix = `{"success":false,"message":"Validation failed","error_code":"VALIDATION_ERROR","errors":`

func assertBadRequest(t *testing.T, message string, status int, body []byte) {
	t.Helper()
	assert.Equal(t, http.StatusBadRequest, status, string(body))
	assert.JSONEq(t, fmt.Sprintf(`{"success":false,"message":%q,"error_code":"BAD_REQUEST"}`, message), string(body))
}

// ============== Suppliers ==============

var supplierKeys = []string{"id", "name", "code", "contact_name", "phone", "email", "address", "notes", "is_active", "created_at", "updated_at"}

func TestSupplier_CRUD(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	s := created(t, "/api/v1/suppliers", auth.Token, map[string]any{
		"name": "Maju Jaya", "code": "SUP-1", "contact_name": "Budi", "phone": "021", "email": "a@b.co",
		"address": "Jl. X", "notes": "n", "is_active": true,
	})
	assertKeys(t, s, supplierKeys...)
	path := fmt.Sprintf("/api/v1/suppliers/%d", idOf(s))

	r := decodeResponse(t, doGet(t, path, auth.Token))
	assert.Equal(t, "Supplier retrieved successfully", r.Message)
	assert.Equal(t, "Budi", decodeObject(t, r.Data)["contact_name"])

	resp := doPut(t, path, map[string]any{"name": "Maju", "code": "SUP-1"}, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	r = decodeResponse(t, resp)
	assert.Equal(t, "Supplier updated successfully", r.Message)
	u := decodeObject(t, r.Data)
	assert.Nil(t, u["contact_name"])
	assert.Equal(t, false, u["is_active"])

	resp = doDelete(t, path, auth.Token)
	assert.JSONEq(t, `{"success":true,"message":"Supplier deleted successfully"}`, string(resp.Body))
	assert.Equal(t, http.StatusNotFound, doGet(t, path, auth.Token).StatusCode)
}

func TestSupplier_ListAndErrors(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	created(t, "/api/v1/suppliers", auth.Token, map[string]any{"name": "Zebra", "code": "Z", "contact_name": "Anto", "is_active": true})
	created(t, "/api/v1/suppliers", auth.Token, map[string]any{"name": "Alfa", "code": "A"})

	r := decodeResponse(t, doGet(t, "/api/v1/suppliers", auth.Token))
	assert.Equal(t, "Supplier list retrieved", r.Message)
	assert.Equal(t, []string{"Alfa", "Zebra"}, namesOf(t, r.Data, "name"))
	assert.JSONEq(t, `{"pagination":{"total_records":2,"total_pages":1,"current_page":1,"per_page":20,"count":0,"next_page":null,"prev_page":null}}`, string(r.Meta))
	for path, want := range map[string][]string{
		"/api/v1/suppliers?search=ant":          {"Zebra"},
		"/api/v1/suppliers?search=a&per_page=1": {"Alfa"},
		"/api/v1/suppliers?is_active=true":      {"Zebra"},
		"/api/v1/suppliers?is_active=false":     {"Alfa"},
	} {
		assert.Equal(t, want, namesOf(t, decodeResponse(t, doGet(t, path, auth.Token)).Data, "name"), path)
	}

	resp := doPost(t, "/api/v1/suppliers", map[string]any{"name": "Dup", "code": "A"}, auth.Token)
	assertBadRequest(t, "Supplier code already exists", resp.StatusCode, resp.Body)
	resp = doGet(t, "/api/v1/suppliers/abc", auth.Token)
	assertBadRequest(t, "invalid supplier ID", resp.StatusCode, resp.Body)
	resp = doGet(t, "/api/v1/suppliers/999999", auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"Supplier not found not found","error_code":"NOT_FOUND"}`, string(resp.Body))
	resp = doPost(t, "/api/v1/suppliers", map[string]any{"email": "nope"}, auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"name","message":"name is required"},{"field":"code","message":"code is required"},{"field":"email","message":"Invalid email format"}]}`, string(resp.Body))

	// Another company: not found, may reuse codes.
	other := registerTestUser(t)
	rows := rowsOf(t, decodeResponse(t, doGet(t, "/api/v1/suppliers", auth.Token)).Data)
	foreign := fmt.Sprintf("/api/v1/suppliers/%d", idOf(rows[0]))
	assert.Equal(t, http.StatusNotFound, doGet(t, foreign, other.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doPut(t, foreign, map[string]any{"name": "X", "code": "X"}, other.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doDelete(t, foreign, other.Token).StatusCode)
	created(t, "/api/v1/suppliers", other.Token, map[string]any{"name": "Other", "code": "A"})
}

// ============== Users ==============

var userKeys = []string{"id", "company_id", "role_id", "email", "name", "status", "created_at", "updated_at"}

func TestUser_CRUD(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	// role_id is accepted but not stored.
	u := created(t, "/api/v1/users", auth.Token, map[string]any{"email": "kasir@x.id", "password": "password123", "name": "Kasir", "role_id": 1})
	assertKeys(t, u, userKeys...)
	assert.Nil(t, u["role_id"])
	assert.Equal(t, "active", u["status"])
	assert.Equal(t, float64(auth.CompanyID), u["company_id"])
	path := fmt.Sprintf("/api/v1/users/%d", idOf(u))

	r := decodeResponse(t, doGet(t, path, auth.Token))
	assert.Equal(t, "User retrieved successfully", r.Message)
	assertKeys(t, decodeObject(t, r.Data), userKeys...)

	// Empty fields are left alone; updated_at in the response isn't refreshed.
	resp := doPut(t, path, map[string]any{"name": "Kasir Satu"}, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	r = decodeResponse(t, resp)
	assert.Equal(t, "User updated successfully", r.Message)
	got := decodeObject(t, r.Data)
	assert.Equal(t, "Kasir Satu", got["name"])
	assert.Equal(t, "kasir@x.id", got["email"])
	assert.Equal(t, u["updated_at"], got["updated_at"])

	resp = doPut(t, path, map[string]any{"email": "baru@x.id"}, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	assert.Equal(t, "baru@x.id", decodeObject(t, decodeResponse(t, resp).Data)["email"])

	resp = doRequest(t, http.MethodPatch, path+"/status", map[string]any{"status": "inactive"}, auth.Token)
	assert.JSONEq(t, `{"success":true,"message":"User status updated successfully"}`, string(resp.Body))
	assert.Equal(t, "inactive", decodeObject(t, decodeResponse(t, doGet(t, path, auth.Token)).Data)["status"])

	// The new user can log in, unless inactive.
	resp = doPost(t, "/api/v1/auth/login", map[string]any{"email": "baru@x.id", "password": "password123"}, "")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "%s", resp.Body)
	doRequest(t, http.MethodPatch, path+"/status", map[string]any{"status": "active"}, auth.Token)
	resp = doPost(t, "/api/v1/auth/login", map[string]any{"email": "baru@x.id", "password": "password123"}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)

	resp = doDelete(t, path, auth.Token)
	assert.JSONEq(t, `{"success":true,"message":"User deleted successfully"}`, string(resp.Body))
	assert.Equal(t, http.StatusNotFound, doGet(t, path, auth.Token).StatusCode)
}

func TestUser_ListAndErrors(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	u := created(t, "/api/v1/users", auth.Token, map[string]any{"email": "a@x.id", "password": "password123", "name": "Ani"})
	path := fmt.Sprintf("/api/v1/users/%d", idOf(u))

	r := decodeResponse(t, doGet(t, "/api/v1/users", auth.Token))
	assert.Equal(t, "User list retrieved", r.Message)
	require.Len(t, rowsOf(t, r.Data), 2)
	assert.JSONEq(t, `{"pagination":{"total_records":2,"total_pages":1,"current_page":1,"per_page":10,"count":2,"next_page":null,"prev_page":null}}`, string(r.Meta))
	assert.JSONEq(t, `{"pagination":{"total_records":2,"total_pages":2,"current_page":2,"per_page":1,"count":1,"next_page":null,"prev_page":1}}`,
		string(decodeResponse(t, doGet(t, "/api/v1/users?limit=1&page=2", auth.Token)).Meta))
	assert.Contains(t, string(decodeResponse(t, doGet(t, "/api/v1/users?limit=500&page=0", auth.Token)).Meta), `"current_page":1,"per_page":10`)

	// Unlike other lists, these parameters are bound strictly.
	for _, q := range []string{"page=abc", "limit=1.5"} {
		resp := doGet(t, "/api/v1/users?"+q, auth.Token)
		assertBadRequest(t, "invalid query params", resp.StatusCode, resp.Body)
	}
	assert.Equal(t, http.StatusOK, doGet(t, "/api/v1/users?page=&limit=", auth.Token).StatusCode)

	resp := doGet(t, "/api/v1/users/abc", auth.Token)
	assertBadRequest(t, "invalid user id", resp.StatusCode, resp.Body)
	resp = doGet(t, "/api/v1/users/999999", auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"user not found","error_code":"NOT_FOUND"}`, string(resp.Body))

	resp = doPost(t, "/api/v1/users", map[string]any{"email": "a@x.id", "password": "password123", "name": "Dup"}, auth.Token)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	assert.JSONEq(t, `{"success":false,"message":"email already exists","error_code":"EMAIL_ALREADY_EXISTS"}`, string(resp.Body))
	resp = doPost(t, "/api/v1/users", map[string]any{"email": "nope", "password": "short", "name": "A"}, auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"email","message":"Invalid email format"},{"field":"password","message":"password must be at least 8 characters"},{"field":"name","message":"name must be at least 2 characters"}]}`, string(resp.Body))
	resp = doPost(t, "/api/v1/users", map[string]any{"email": "b@x.id", "password": "password123", "name": "Bob", "role_id": "x"}, auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"role_id","message":"must be a number"}]}`, string(resp.Body))

	resp = doPut(t, path, map[string]any{"email": "nope", "name": "A"}, auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"email","message":"Invalid email format"},{"field":"name","message":"name must be at least 2 characters"}]}`, string(resp.Body))
	created(t, "/api/v1/users", auth.Token, map[string]any{"email": "c@x.id", "password": "password123", "name": "Cici"})
	resp = doPut(t, path, map[string]any{"email": "c@x.id"}, auth.Token)
	assert.Equal(t, http.StatusConflict, resp.StatusCode, "%s", resp.Body)

	resp = doRequest(t, http.MethodPatch, path+"/status", map[string]any{"status": "banned"}, auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"status","message":"status must be one of: active inactive"}]}`, string(resp.Body))
	resp = doRequest(t, http.MethodPatch, path+"/status", map[string]any{}, auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"status","message":"status is required"}]}`, string(resp.Body))

	other := registerTestUser(t)
	assert.Equal(t, http.StatusNotFound, doGet(t, path, other.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doPut(t, path, map[string]any{"name": "Taken"}, other.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doRequest(t, http.MethodPatch, path+"/status", map[string]any{"status": "inactive"}, other.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doDelete(t, path, other.Token).StatusCode)
	assert.Len(t, rowsOf(t, decodeResponse(t, doGet(t, "/api/v1/users", other.Token)).Data), 1)
}

// ============== Expenses ==============

var expenseKeys = []string{"id", "branch_id", "category_id", "amount", "description", "reference_no", "expense_date", "recorded_by", "notes", "created_at", "updated_at"}

func TestExpenseCategory_CRUD(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	c := created(t, "/api/v1/expense-categories", auth.Token, map[string]any{"name": "Listrik", "description": "PLN"})
	assertKeys(t, c, "id", "name", "description", "created_at", "updated_at")
	created(t, "/api/v1/expense-categories", auth.Token, map[string]any{"name": "Air"})
	path := fmt.Sprintf("/api/v1/expense-categories/%d", idOf(c))

	r := decodeResponse(t, doGet(t, "/api/v1/expense-categories", auth.Token))
	assert.Equal(t, "Expense categories retrieved", r.Message)
	assert.Equal(t, []string{"Air", "Listrik"}, namesOf(t, r.Data, "name"))
	assert.JSONEq(t, `{"pagination":{"total_records":2,"total_pages":1,"current_page":1,"per_page":20,"count":0,"next_page":null,"prev_page":null}}`, string(r.Meta))
	assert.Equal(t, []string{"Listrik"}, namesOf(t, decodeResponse(t, doGet(t, "/api/v1/expense-categories?search=lis", auth.Token)).Data, "name"))

	resp := doPut(t, path, map[string]any{"name": "Listrik 2"}, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	r = decodeResponse(t, resp)
	assert.Equal(t, "Expense category updated successfully", r.Message)
	assert.Nil(t, decodeObject(t, r.Data)["description"])

	resp = doPut(t, "/api/v1/expense-categories/abc", map[string]any{"name": "X"}, auth.Token)
	assertBadRequest(t, "invalid expense category ID", resp.StatusCode, resp.Body)
	resp = doDelete(t, "/api/v1/expense-categories/999999", auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"Expense category not found not found","error_code":"NOT_FOUND"}`, string(resp.Body))
	resp = doPost(t, "/api/v1/expense-categories", map[string]any{}, auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"name","message":"name is required"}]}`, string(resp.Body))

	other := registerTestUser(t)
	assert.Equal(t, http.StatusNotFound, doPut(t, path, map[string]any{"name": "Taken"}, other.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doDelete(t, path, other.Token).StatusCode)

	resp = doDelete(t, path, auth.Token)
	assert.JSONEq(t, `{"success":true,"message":"Expense category deleted successfully"}`, string(resp.Body))
}

func TestExpense_CRUD(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	cat := created(t, "/api/v1/expense-categories", auth.Token, map[string]any{"name": "Listrik"})

	// On create, the date and amount echo the request as sent.
	e := created(t, "/api/v1/expenses", auth.Token, map[string]any{
		"branch_id": auth.BranchID, "category_id": idOf(cat), "amount": "15000.555", "description": "Token", "reference_no": "R-1",
		"expense_date": "2026-10-01", "notes": "n",
	})
	assertKeys(t, e, expenseKeys...)
	assert.Equal(t, "2026-10-01", e["expense_date"])
	assert.Equal(t, 15000.555, e["amount"])
	assert.Equal(t, float64(auth.UserID), e["recorded_by"])
	path := fmt.Sprintf("/api/v1/expenses/%d", idOf(e))

	// Reads show what the columns hold: a rounded amount and the date as a time.
	r := decodeResponse(t, doGet(t, path, auth.Token))
	assert.Equal(t, "Expense retrieved successfully", r.Message)
	got := decodeObject(t, r.Data)
	assert.Equal(t, "2026-10-01T00:00:00Z", got["expense_date"])
	assert.Equal(t, 15000.56, got["amount"])

	resp := doPut(t, path, map[string]any{"amount": 2000, "description": "Beda", "expense_date": "2026-10-02"}, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	r = decodeResponse(t, resp)
	assert.Equal(t, "Expense updated successfully", r.Message)
	got = decodeObject(t, r.Data)
	assert.Equal(t, "2026-10-02T00:00:00Z", got["expense_date"])
	assert.Nil(t, got["branch_id"])
	assert.Nil(t, got["category_id"])
	assert.Nil(t, got["notes"])
	assert.Equal(t, float64(auth.UserID), got["recorded_by"], "update keeps who recorded it")

	resp = doDelete(t, path, auth.Token)
	assert.JSONEq(t, `{"success":true,"message":"Expense deleted successfully"}`, string(resp.Body))
	assert.Equal(t, http.StatusNotFound, doGet(t, path, auth.Token).StatusCode)
}

func TestExpense_ListSummaryAndErrors(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	listrik := created(t, "/api/v1/expense-categories", auth.Token, map[string]any{"name": "Listrik"})
	air := created(t, "/api/v1/expense-categories", auth.Token, map[string]any{"name": "Air"})
	created(t, "/api/v1/expense-categories", auth.Token, map[string]any{"name": "Kosong"})
	add := func(cat any, branch any, amount any, day, desc string) {
		created(t, "/api/v1/expenses", auth.Token, map[string]any{"category_id": cat, "branch_id": branch, "amount": amount, "expense_date": day, "description": desc})
	}
	add(idOf(listrik), auth.BranchID, 1000.5, "2026-10-01", "A")
	add(idOf(listrik), nil, 2000, "2026-10-03", "B")
	add(idOf(air), auth.BranchID, 500.25, "2026-10-02", "C")
	add(nil, nil, 100, "2026-10-04", "D")

	r := decodeResponse(t, doGet(t, "/api/v1/expenses", auth.Token))
	assert.Equal(t, "Expenses retrieved", r.Message)
	assert.Equal(t, []string{"D", "B", "C", "A"}, namesOf(t, r.Data, "description"), "newest expense date first")
	assert.JSONEq(t, `{"pagination":{"total_records":4,"total_pages":1,"current_page":1,"per_page":20,"count":0,"next_page":null,"prev_page":null}}`, string(r.Meta))
	for path, want := range map[string][]string{
		fmt.Sprintf("/api/v1/expenses?category_id=%d", idOf(listrik)): {"B", "A"},
		fmt.Sprintf("/api/v1/expenses?branch_id=%d", auth.BranchID):   {"C", "A"},
		"/api/v1/expenses?date_from=2026-10-03":                       {"D", "B"},
		"/api/v1/expenses?date_to=2026-10-01":                         {"A"},
		"/api/v1/expenses?date_from=2026-10-02&date_to=2026-10-03":    {"B", "C"},
		"/api/v1/expenses?branch_id=abc&per_page=2&page=2":            {"C", "A"},
	} {
		assert.Equal(t, want, namesOf(t, decodeResponse(t, doGet(t, path, auth.Token)).Data, "description"), path)
	}

	resp := doGet(t, "/api/v1/expenses/summary", auth.Token)
	assert.JSONEq(t, fmt.Sprintf(`{"success":true,"message":"Expense summary retrieved","data":{"total_amount":3600.75,"date_from":"","date_to":"","by_category":[
		{"category_id":%d,"category_name":"Air","total_amount":500.25,"count":1},
		{"category_id":%d,"category_name":"Listrik","total_amount":3000.5,"count":2},
		{"category_id":null,"category_name":"Uncategorized (1)","total_amount":100,"count":1}]}}`, idOf(air), idOf(listrik)), string(resp.Body))

	resp = doGet(t, fmt.Sprintf("/api/v1/expenses/summary?branch_id=%d&date_from=2026-10-02", auth.BranchID), auth.Token)
	assert.JSONEq(t, fmt.Sprintf(`{"success":true,"message":"Expense summary retrieved","data":{"total_amount":500.25,"date_from":"2026-10-02","date_to":"","by_category":[
		{"category_id":%d,"category_name":"Air","total_amount":500.25,"count":1}]}}`, idOf(air)), string(resp.Body))

	other := registerTestUser(t)
	resp = doGet(t, "/api/v1/expenses/summary", other.Token)
	assert.JSONEq(t, `{"success":true,"message":"Expense summary retrieved","data":{"total_amount":0,"date_from":"","date_to":"","by_category":[]}}`, string(resp.Body))

	// Errors, in the order the Go service checks them.
	base := func(extra map[string]any) map[string]any {
		m := map[string]any{"amount": 1000, "description": "X", "expense_date": "2026-10-01"}
		for k, v := range extra {
			m[k] = v
		}
		return m
	}
	resp = doPost(t, "/api/v1/expenses", map[string]any{}, auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"amount","message":"amount is required"},{"field":"description","message":"description is required"},{"field":"expense_date","message":"expense_date is required"}]}`, string(resp.Body))
	resp = doPost(t, "/api/v1/expenses", base(map[string]any{"amount": -5}), auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"amount","message":"amount must be at least 0.01 characters"}]}`, string(resp.Body))
	resp = doPost(t, "/api/v1/expenses", base(map[string]any{"amount": "abc"}), auth.Token)
	assertBadRequest(t, "invalid request body", resp.StatusCode, resp.Body)
	resp = doPost(t, "/api/v1/expenses", base(map[string]any{"category_id": 999999, "branch_id": 999999}), auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"Expense category not found not found","error_code":"NOT_FOUND"}`, string(resp.Body))
	resp = doPost(t, "/api/v1/expenses", base(map[string]any{"branch_id": 999999}), auth.Token)
	assertBadRequest(t, "Branch not found", resp.StatusCode, resp.Body)
	resp = doPost(t, "/api/v1/expenses", base(map[string]any{"category_id": idOf(listrik)}), other.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "another company's category")
	resp = doPost(t, "/api/v1/expenses", base(map[string]any{"expense_date": "soon"}), auth.Token)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "an unparseable date fails in the database")
	assert.Equal(t, http.StatusInternalServerError, doGet(t, "/api/v1/expenses?date_from=soon", auth.Token).StatusCode)
	resp = doGet(t, "/api/v1/expenses/abc", auth.Token)
	assertBadRequest(t, "invalid expense ID", resp.StatusCode, resp.Body)
	resp = doGet(t, "/api/v1/expenses/999999", auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"Expense not found not found","error_code":"NOT_FOUND"}`, string(resp.Body))

	rows := rowsOf(t, decodeResponse(t, doGet(t, "/api/v1/expenses", auth.Token)).Data)
	foreign := fmt.Sprintf("/api/v1/expenses/%d", idOf(rows[0]))
	assert.Equal(t, http.StatusNotFound, doGet(t, foreign, other.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doPut(t, foreign, base(nil), other.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doDelete(t, foreign, other.Token).StatusCode)
	resp = doPut(t, foreign, base(map[string]any{"branch_id": other.BranchID}), auth.Token)
	assertBadRequest(t, "Branch not found", resp.StatusCode, resp.Body)
}

// ============== Promotions ==============

var promotionKeys = []string{"id", "code", "name", "description", "type", "discount_type", "discount_value", "min_purchase", "max_discount", "start_at", "end_at", "priority", "is_active", "status", "created_at", "updated_at"}

func TestPromotion_CRUD(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	// Times keep the offset they were sent with until read back from the database.
	p := created(t, "/api/v1/promotions", auth.Token, map[string]any{
		"code": "P1", "name": "Promo", "description": "d", "type": "discount", "discount_type": "percentage",
		"discount_value": "10.50", "min_purchase": 5000, "max_discount": "2500.5",
		"start_at": "2020-01-02T03:04:05+07:00", "end_at": "2099-01-02T03:04:05.25Z", "priority": 3, "is_active": true,
	})
	assertKeys(t, p, promotionKeys...)
	assert.Equal(t, "2020-01-02T03:04:05+07:00", p["start_at"])
	assert.Equal(t, "2099-01-02T03:04:05.25Z", p["end_at"])
	assert.Equal(t, 10.5, p["discount_value"])
	assert.Equal(t, "active", p["status"])
	path := fmt.Sprintf("/api/v1/promotions/%d", idOf(p))

	r := decodeResponse(t, doGet(t, path, auth.Token))
	assert.Equal(t, "Promotion retrieved successfully", r.Message)
	assertKeys(t, decodeObject(t, r.Data), promotionKeys...)

	resp := doPut(t, path, map[string]any{"code": "P1", "name": "Baru", "type": "bundle", "start_at": "2098-01-01T00:00:00Z", "is_active": true}, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	r = decodeResponse(t, resp)
	assert.Equal(t, "Promotion updated successfully", r.Message)
	u := decodeObject(t, r.Data)
	assert.Equal(t, "scheduled", u["status"])
	assert.Nil(t, u["end_at"])
	assert.Nil(t, u["discount_type"])
	assert.Nil(t, u["discount_value"])
	assert.Equal(t, p["created_at"], u["created_at"])

	resp = doDelete(t, path, auth.Token)
	assert.JSONEq(t, `{"success":true,"message":"Promotion deleted successfully"}`, string(resp.Body))
	assert.Equal(t, http.StatusNotFound, doGet(t, path, auth.Token).StatusCode)
}

func TestPromotion_ListAndErrors(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	add := func(code, kind string, priority int, start string, end any, active bool) map[string]any {
		return created(t, "/api/v1/promotions", auth.Token, map[string]any{
			"code": code, "name": "Promo " + code, "type": kind, "priority": priority, "start_at": start, "end_at": end, "is_active": active,
		})
	}
	running := add("RUN", "discount", 1, "2020-01-01T00:00:00Z", nil, true)
	add("LATER", "bundle", 5, "2098-01-01T00:00:00Z", nil, true)
	add("OLD", "discount", 3, "2020-01-01T00:00:00Z", "2020-02-01T00:00:00Z", true)
	add("OFF", "conditional", 2, "2020-01-01T00:00:00Z", nil, false)

	r := decodeResponse(t, doGet(t, "/api/v1/promotions", auth.Token))
	assert.Equal(t, "Promotion list retrieved", r.Message)
	assert.Equal(t, []string{"LATER", "OLD", "OFF", "RUN"}, namesOf(t, r.Data, "code"), "highest priority first")
	statuses := map[string]string{}
	for _, row := range rowsOf(t, r.Data) {
		statuses[row["code"].(string)] = row["status"].(string)
	}
	assert.Equal(t, map[string]string{"RUN": "active", "LATER": "scheduled", "OLD": "expired", "OFF": "inactive"}, statuses)
	for path, want := range map[string][]string{
		"/api/v1/promotions?status=active":     {"RUN"},
		"/api/v1/promotions?status=scheduled":  {"LATER"},
		"/api/v1/promotions?status=expired":    {"OLD"},
		"/api/v1/promotions?status=inactive":   {"OFF"},
		"/api/v1/promotions?status=whatever":   {"LATER", "OLD", "OFF", "RUN"},
		"/api/v1/promotions?type=bundle":       {"LATER"},
		"/api/v1/promotions?is_active=false":   {"OFF"},
		"/api/v1/promotions?search=promo%20ru": {"RUN"},
		"/api/v1/promotions?per_page=1&page=2": {"OLD"},
	} {
		assert.Equal(t, want, namesOf(t, decodeResponse(t, doGet(t, path, auth.Token)).Data, "code"), path)
	}

	valid := func(extra map[string]any) map[string]any {
		m := map[string]any{"code": "NEW", "name": "New", "type": "discount", "start_at": "2020-01-01T00:00:00Z"}
		for k, v := range extra {
			m[k] = v
		}
		return m
	}
	resp := doPost(t, "/api/v1/promotions", map[string]any{"discount_type": "half", "discount_value": -1, "min_purchase": -1, "max_discount": -1, "priority": -1}, auth.Token)
	assert.JSONEq(t, validationPrefix+`[
		{"field":"code","message":"code is required"},{"field":"name","message":"name is required"},
		{"field":"type","message":"type is required"},
		{"field":"discount_type","message":"discount_type must be one of: percentage fixed"},
		{"field":"discount_value","message":"discount_value is invalid"},{"field":"min_purchase","message":"min_purchase is invalid"},
		{"field":"max_discount","message":"max_discount is invalid"},{"field":"start_at","message":"start_at is required"},
		{"field":"priority","message":"priority is invalid"}]}`, string(resp.Body))
	resp = doPost(t, "/api/v1/promotions", valid(map[string]any{"type": "gift"}), auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"type","message":"type must be one of: discount bundle conditional"}]}`, string(resp.Body))
	resp = doPost(t, "/api/v1/promotions", valid(map[string]any{"code": "RUN"}), auth.Token)
	assertBadRequest(t, "Promotion code already exists", resp.StatusCode, resp.Body)
	resp = doPost(t, "/api/v1/promotions", valid(map[string]any{"start_at": "2020-01-01"}), auth.Token)
	assertBadRequest(t, "Invalid start_at format", resp.StatusCode, resp.Body)
	resp = doPost(t, "/api/v1/promotions", valid(map[string]any{"end_at": ""}), auth.Token)
	assertBadRequest(t, "Invalid end_at format", resp.StatusCode, resp.Body)
	resp = doGet(t, "/api/v1/promotions/abc", auth.Token)
	assertBadRequest(t, "invalid promotion ID", resp.StatusCode, resp.Body)
	resp = doGet(t, "/api/v1/promotions/999999", auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"Promotion not found not found","error_code":"NOT_FOUND"}`, string(resp.Body))

	path := fmt.Sprintf("/api/v1/promotions/%d", idOf(running))
	other := registerTestUser(t)
	assert.Equal(t, http.StatusNotFound, doGet(t, path, other.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doPut(t, path, valid(nil), other.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doDelete(t, path, other.Token).StatusCode)
	assert.Empty(t, rowsOf(t, decodeResponse(t, doGet(t, "/api/v1/promotions", other.Token)).Data))
}

// ============== Product writes ==============

func TestProductWrite_CreateRules(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	cat := created(t, "/api/v1/product-categories", auth.Token, map[string]any{"code": "C", "name": "Cat"})
	v := func(sku string) map[string]any { return map[string]any{"sku": sku, "name": "V " + sku, "price": 1000} }

	p := created(t, "/api/v1/products", auth.Token, map[string]any{
		"name": "A", "category_id": idOf(cat), "description": "d", "image_url": "https://cdn/a.png", "image_data": "b64", "is_active": false,
		"variants": []map[string]any{v("A-1"), {"sku": "A-2", "name": "Dua", "price": "2500.50", "barcode": "  ", "sale_start": "not-a-time", "sale_end": "", "attributes": map[string]any{}}},
	})
	assert.Equal(t, false, p["is_active"])
	assert.Equal(t, "b64", p["image_data"])
	vs := p["variants"].([]any)
	require.Len(t, vs, 2)
	assert.Equal(t, true, vs[0].(map[string]any)["is_default"], "the first variant becomes the default")
	second := vs[1].(map[string]any)
	assert.Equal(t, false, second["is_default"])
	assert.Equal(t, map[string]any{}, second["attributes"])
	assert.NotContains(t, second, "barcode", "a blank barcode is no barcode")
	assert.NotContains(t, second, "sale_start", "an unparseable time is dropped")

	resp := doPost(t, "/api/v1/products", map[string]any{"name": "B", "variants": []map[string]any{v("A-1")}}, auth.Token)
	assertBadRequest(t, "SKU already exists: A-1", resp.StatusCode, resp.Body)
	resp = doPost(t, "/api/v1/products", map[string]any{"name": "B", "category_id": 999999, "variants": []map[string]any{v("B-1")}}, auth.Token)
	assertBadRequest(t, "Category not found", resp.StatusCode, resp.Body)

	bc := func(sku, code string) map[string]any { m := v(sku); m["barcode"] = code; return m }
	created(t, "/api/v1/products", auth.Token, map[string]any{"name": "C", "variants": []map[string]any{bc("C-1", "777")}})
	resp = doPost(t, "/api/v1/products", map[string]any{"name": "D", "variants": []map[string]any{bc("D-1", "777")}}, auth.Token)
	assertBadRequest(t, "Barcode already registered: 777", resp.StatusCode, resp.Body)
	resp = doPost(t, "/api/v1/products", map[string]any{"name": "D", "variants": []map[string]any{bc("D-1", "1"), bc("D-2", " 1 ")}}, auth.Token)
	assertBadRequest(t, "Duplicate barcode in request: 1", resp.StatusCode, resp.Body)

	resp = doPost(t, "/api/v1/products", map[string]any{}, auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"name","message":"name is required"},{"field":"variants","message":"variants is required"}]}`, string(resp.Body))
	resp = doPost(t, "/api/v1/products", map[string]any{"name": "E", "variants": []map[string]any{}}, auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"variants","message":"variants must be at least 1 characters"}]}`, string(resp.Body))
	resp = doPost(t, "/api/v1/products", map[string]any{"name": "E", "variants": []map[string]any{
		{"price": -1, "standard_cost": -1, "last_purchase_cost": -1}, {"sku": "OK", "name": "OK"},
	}}, auth.Token)
	assert.JSONEq(t, validationPrefix+`[{"field":"sku","message":"sku is required"},{"field":"name","message":"name is required"},
		{"field":"price","message":"price is invalid"},{"field":"standard_cost","message":"standard_cost is invalid"},
		{"field":"last_purchase_cost","message":"last_purchase_cost is invalid"}]}`, string(resp.Body))

	// SKUs are unique across companies; barcodes only within one.
	other := registerTestUser(t)
	resp = doPost(t, "/api/v1/products", map[string]any{"name": "O", "variants": []map[string]any{v("A-1")}}, other.Token)
	assertBadRequest(t, "SKU already exists: A-1", resp.StatusCode, resp.Body)
	created(t, "/api/v1/products", other.Token, map[string]any{"name": "O", "variants": []map[string]any{bc("O-1", "777")}})
	resp = doPost(t, "/api/v1/products", map[string]any{"name": "O", "category_id": idOf(cat), "variants": []map[string]any{v("O-2")}}, other.Token)
	assertBadRequest(t, "Category not found", resp.StatusCode, resp.Body)
}

func TestProductWrite_UpdateVariants(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	p := created(t, "/api/v1/products", auth.Token, map[string]any{"name": "A", "image_url": "https://cdn/a.png", "description": "d", "is_active": false, "variants": []map[string]any{
		{"sku": "A-1", "name": "Satu", "price": 1000}, {"sku": "A-2", "name": "Dua", "price": 2000}, {"sku": "A-3", "name": "Tiga", "price": 3000},
	}})
	path := fmt.Sprintf("/api/v1/products/%d", idOf(p))
	ids := map[string]float64{}
	for _, v := range p["variants"].([]any) {
		ids[v.(map[string]any)["sku"].(string)] = v.(map[string]any)["id"].(float64)
	}
	put := func(body map[string]any) map[string]any {
		t.Helper()
		resp := doPut(t, path, body, auth.Token)
		require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
		r := decodeResponse(t, resp)
		assert.Equal(t, "Product updated successfully", r.Message)
		return decodeObject(t, r.Data)
	}
	bySKU := func(p map[string]any) map[string]map[string]any {
		out := map[string]map[string]any{}
		for _, v := range p["variants"].([]any) {
			out[v.(map[string]any)["sku"].(string)] = v.(map[string]any)
		}
		return out
	}

	// A variant with a known id is updated in place; omitted is_active keeps the product's value,
	// and omitted description/image clear them.
	got := put(map[string]any{"name": "B", "variants": []map[string]any{
		{"id": ids["A-2"], "sku": "A-2", "name": "Dua", "price": 2222}, {"id": ids["A-1"], "sku": "A-1", "name": "Satu", "price": 1111},
		{"id": ids["A-3"], "sku": "A-3", "name": "Tiga", "price": 3333, "is_active": false},
	}})
	assert.Equal(t, "B", got["name"])
	assert.Equal(t, false, got["is_active"])
	assert.Nil(t, got["description"])
	assert.Nil(t, got["image_url"])
	vs := bySKU(got)
	assert.Equal(t, ids["A-2"], vs["A-2"]["id"])
	assert.Equal(t, float64(2222), vs["A-2"]["price"])
	assert.Equal(t, true, vs["A-2"]["is_default"], "the first in the request is the default")
	assert.Equal(t, false, vs["A-1"]["is_default"])
	assert.Equal(t, false, vs["A-3"]["is_active"])

	// A variant without a known id reuses a row that was left out, instead of inserting.
	got = put(map[string]any{"name": "B", "is_active": true, "variants": []map[string]any{
		{"id": ids["A-1"], "sku": "A-1", "name": "Satu", "price": 1},
		{"sku": "A-9", "name": "Baru", "price": 9},
	}})
	assert.Equal(t, true, got["is_active"])
	vs = bySKU(got)
	require.Len(t, vs, 2, "the third row is deleted")
	assert.Contains(t, []float64{ids["A-2"], ids["A-3"]}, vs["A-9"]["id"])

	// More variants than rows: the extra one is inserted.
	got = put(map[string]any{"name": "B", "variants": []map[string]any{
		{"id": ids["A-1"], "sku": "A-1", "name": "Satu", "price": 1},
		{"sku": "A-9", "name": "Baru", "price": 9, "barcode": "555", "attributes": map[string]any{"k": "v"}, "sale_price": 5, "sale_start": "2026-01-01T00:00:00Z"},
		{"sku": "A-10", "name": "Lagi", "price": 10},
	}})
	vs = bySKU(got)
	require.Len(t, vs, 3)
	assert.Equal(t, "555", vs["A-9"]["barcode"])
	assert.Equal(t, map[string]any{"k": "v"}, vs["A-9"]["attributes"])
	assert.Equal(t, float64(5), vs["A-9"]["sale_price"])

	// Conflicts with other products.
	created(t, "/api/v1/products", auth.Token, map[string]any{"name": "Z", "variants": []map[string]any{{"sku": "Z-1", "name": "Z", "barcode": "999", "price": 1}}})
	resp := doPut(t, path, map[string]any{"name": "B", "variants": []map[string]any{{"sku": "Z-1", "name": "X", "price": 1}}}, auth.Token)
	assertBadRequest(t, "SKU already exists: Z-1", resp.StatusCode, resp.Body)
	resp = doPut(t, path, map[string]any{"name": "B", "variants": []map[string]any{{"sku": "A-1", "name": "X", "price": 1, "barcode": "999"}}}, auth.Token)
	assertBadRequest(t, "Barcode already registered: 999", resp.StatusCode, resp.Body)
	resp = doPut(t, path, map[string]any{"name": "B", "category_id": 999999, "variants": []map[string]any{{"sku": "A-1", "name": "X", "price": 1}}}, auth.Token)
	assertBadRequest(t, "Category not found", resp.StatusCode, resp.Body)
	resp = doPut(t, "/api/v1/products/abc", map[string]any{"name": "B", "variants": []map[string]any{{"sku": "A-1", "name": "X", "price": 1}}}, auth.Token)
	assertBadRequest(t, "invalid product ID", resp.StatusCode, resp.Body)
	resp = doPut(t, "/api/v1/products/999999", map[string]any{"name": "B", "variants": []map[string]any{{"sku": "A-1", "name": "X", "price": 1}}}, auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"Product not found not found","error_code":"NOT_FOUND"}`, string(resp.Body))

	other := registerTestUser(t)
	assert.Equal(t, http.StatusNotFound, doPut(t, path, map[string]any{"name": "T", "variants": []map[string]any{{"sku": "T-1", "name": "T", "price": 1}}}, other.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doDelete(t, path, other.Token).StatusCode)

	resp = doDelete(t, path, auth.Token)
	assert.JSONEq(t, `{"success":true,"message":"Product deleted successfully"}`, string(resp.Body))
	assert.Equal(t, http.StatusNotFound, doGet(t, path, auth.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doDelete(t, path, auth.Token).StatusCode)
}

// A variant with transaction history can't be removed from a product.
func TestProductWrite_VariantWithHistoryIsKept(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	p := created(t, "/api/v1/products", auth.Token, map[string]any{"name": "A", "variants": []map[string]any{
		{"sku": "A-1", "name": "Satu", "price": 1000}, {"sku": "A-2", "name": "Dua", "price": 2000},
	}})
	var keep, history float64
	for _, v := range p["variants"].([]any) {
		m := v.(map[string]any)
		if m["sku"] == "A-1" {
			keep = m["id"].(float64)
		} else {
			history = m["id"].(float64)
		}
	}
	_, err := testEnv.DB.Exec(`INSERT INTO stock_movements (product_variant_id, branch_id, type, quantity, stock_before, stock_after)
		VALUES ($1, $2, 'IN', 1, 0, 1)`, int64(history), auth.BranchID)
	require.NoError(t, err)

	resp := doPut(t, fmt.Sprintf("/api/v1/products/%d", idOf(p)), map[string]any{"name": "A", "variants": []map[string]any{
		{"id": keep, "sku": "A-1", "name": "Satu", "price": 1000},
	}}, auth.Token)
	assertBadRequest(t, "Variant 'Dua' (SKU: A-2) tidak dapat dihapus karena sudah memiliki riwayat transaksi", resp.StatusCode, resp.Body)
}

// ============== Uploads ==============

// Without object storage configured, the upload endpoint refuses plainly.
func TestUpload_NotConfigured(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	resp := doPost(t, "/api/v1/uploads", map[string]any{}, auth.Token)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "%s", resp.Body)
	assert.JSONEq(t, `{"success":false,"message":"object storage is not configured","error_code":"INTERNAL_ERROR"}`, string(resp.Body))
	assert.Equal(t, http.StatusUnauthorized, doPost(t, "/api/v1/uploads", map[string]any{}, "").StatusCode)
}
