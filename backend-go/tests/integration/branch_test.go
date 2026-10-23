package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertKeys checks that a JSON object has exactly these keys, so any shape
// drift between backends (a missing, extra, or renamed field) fails a test.
func assertKeys(t *testing.T, obj map[string]any, keys ...string) {
	t.Helper()
	got := make([]string, 0, len(obj))
	for k := range obj {
		got = append(got, k)
	}
	sort.Strings(got)
	sort.Strings(keys)
	assert.Equal(t, keys, got)
}

func decodeObject(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()
	var obj map[string]any
	require.NoError(t, json.Unmarshal(raw, &obj), string(raw))
	return obj
}

func createBranch(t *testing.T, token string, body map[string]any) map[string]any {
	t.Helper()
	resp := doPost(t, "/api/v1/branches", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create branch: %s", resp.Body)
	return decodeObject(t, decodeResponse(t, resp).Data)
}

var branchKeys = []string{"id", "code", "name", "address", "phone", "is_active", "created_at", "updated_at"}

func TestBranch_CreateGetUpdateDelete(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	created := createBranch(t, auth.Token, map[string]any{
		"code": "JKT", "name": "Jakarta", "address": "Jl. Sudirman 1", "phone": "021555", "is_active": true,
	})
	assertKeys(t, created, branchKeys...)
	assert.Equal(t, "JKT", created["code"])
	assert.Equal(t, "Jl. Sudirman 1", created["address"])
	assert.Equal(t, true, created["is_active"])
	id := int64(created["id"].(float64))

	resp := doGet(t, fmt.Sprintf("/api/v1/branches/%d", id), auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	r := decodeResponse(t, resp)
	assert.Equal(t, "Branch retrieved successfully", r.Message)
	got := decodeObject(t, r.Data)
	assertKeys(t, got, branchKeys...)
	assert.Equal(t, created["created_at"], got["created_at"])

	// Optional fields left out are stored as null, and is_active defaults to false.
	resp = doPut(t, fmt.Sprintf("/api/v1/branches/%d", id), map[string]any{"code": "JKT-1", "name": "Jakarta Pusat"}, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	r = decodeResponse(t, resp)
	assert.Equal(t, "Branch updated successfully", r.Message)
	updated := decodeObject(t, r.Data)
	assert.Equal(t, "JKT-1", updated["code"])
	assert.Nil(t, updated["address"])
	assert.Nil(t, updated["phone"])
	assert.Equal(t, false, updated["is_active"])

	resp = doDelete(t, fmt.Sprintf("/api/v1/branches/%d", id), auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	assert.JSONEq(t, `{"success":true,"message":"Branch deleted successfully"}`, string(resp.Body))

	resp = doGet(t, fmt.Sprintf("/api/v1/branches/%d", id), auth.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestBranch_ListFiltersAndPagination(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t) // comes with the HQ branch
	createBranch(t, auth.Token, map[string]any{"code": "BDG", "name": "Bandung", "is_active": true})
	createBranch(t, auth.Token, map[string]any{"code": "SBY", "name": "Surabaya", "phone": "031777", "is_active": false})

	resp := doGet(t, "/api/v1/branches", auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	r := decodeResponse(t, resp)
	assert.Equal(t, "Branch list retrieved", r.Message)
	var list []map[string]any
	require.NoError(t, json.Unmarshal(r.Data, &list))
	require.Len(t, list, 3)
	// Ordered by name.
	assert.Equal(t, []any{"Bandung", "Headquarters", "Surabaya"}, []any{list[0]["name"], list[1]["name"], list[2]["name"]})
	assertKeys(t, list[0], branchKeys...)
	assert.JSONEq(t,
		`{"pagination":{"total_records":3,"total_pages":1,"current_page":1,"per_page":20,"count":0,"next_page":null,"prev_page":null}}`,
		string(r.Meta))

	cases := map[string][]string{
		"/api/v1/branches?search=sura":           {"Surabaya"},
		"/api/v1/branches?search=031":            {"Surabaya"},
		"/api/v1/branches?is_active=false":       {"Surabaya"},
		"/api/v1/branches?is_active=true":        {"Bandung", "Headquarters"},
		"/api/v1/branches?is_active=nope":        {"Bandung", "Headquarters", "Surabaya"},
		"/api/v1/branches?per_page=2&page=2":     {"Surabaya"},
		"/api/v1/branches?per_page=abc&page=xyz": {"Bandung", "Headquarters", "Surabaya"},
	}
	for path, want := range cases {
		resp := doGet(t, path, auth.Token)
		require.Equal(t, http.StatusOK, resp.StatusCode, path)
		var rows []map[string]any
		require.NoError(t, json.Unmarshal(decodeResponse(t, resp).Data, &rows), path)
		names := make([]string, len(rows))
		for i, row := range rows {
			names[i] = row["name"].(string)
		}
		assert.Equal(t, want, names, path)
	}

	resp = doGet(t, "/api/v1/branches?per_page=2&page=2", auth.Token)
	assert.JSONEq(t,
		`{"pagination":{"total_records":3,"total_pages":2,"current_page":2,"per_page":2,"count":0,"next_page":null,"prev_page":1}}`,
		string(decodeResponse(t, resp).Meta))
	resp = doGet(t, "/api/v1/branches?per_page=500", auth.Token)
	assert.Contains(t, string(decodeResponse(t, resp).Meta), `"per_page":100`)
}

func TestBranch_Errors(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)

	resp := doPost(t, "/api/v1/branches", map[string]any{"code": "HQ", "name": "Duplicate"}, auth.Token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.JSONEq(t, `{"success":false,"message":"Branch code already exists","error_code":"BAD_REQUEST"}`, string(resp.Body))

	resp = doPost(t, "/api/v1/branches", map[string]any{"code": "", "name": "", "phone": "123456789012345678901234567890123456789012345678901"}, auth.Token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.JSONEq(t, `{"success":false,"message":"Validation failed","error_code":"VALIDATION_ERROR","errors":[
		{"field":"code","message":"code is required"},
		{"field":"name","message":"name is required"},
		{"field":"phone","message":"phone must be at most 50 characters"}]}`, string(resp.Body))

	resp = doGet(t, "/api/v1/branches/abc", auth.Token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.JSONEq(t, `{"success":false,"message":"invalid branch ID","error_code":"BAD_REQUEST"}`, string(resp.Body))

	resp = doGet(t, "/api/v1/branches/999999", auth.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.JSONEq(t, `{"success":false,"message":"Branch not found not found","error_code":"NOT_FOUND"}`, string(resp.Body))

	resp = doGet(t, "/api/v1/branches", "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestBranch_OtherCompanyIsNotFound(t *testing.T) {
	cleanupDatabase(t)
	a := registerTestUser(t)
	b := registerTestUser(t)
	path := fmt.Sprintf("/api/v1/branches/%d", a.BranchID)

	assert.Equal(t, http.StatusNotFound, doGet(t, path, b.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doPut(t, path, map[string]any{"code": "X", "name": "Taken"}, b.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doDelete(t, path, b.Token).StatusCode)

	// Company B may reuse company A's codes, and only sees its own branches.
	createBranch(t, b.Token, map[string]any{"code": "JKT", "name": "B Jakarta"})
	createBranch(t, a.Token, map[string]any{"code": "JKT", "name": "A Jakarta"})
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(decodeResponse(t, doGet(t, "/api/v1/branches", b.Token)).Data, &rows))
	assert.Len(t, rows, 2)

	resp := doGet(t, path, a.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "A still has its branch")
}
