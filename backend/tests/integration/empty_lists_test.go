package integration_test

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Fields that are legitimately null for a freshly registered company. Any
// other null is treated as a bug.
var emptyCompanyNulls = map[string][]string{
	"/api/v1/branches":         {"$.data[].address", "$.data[].phone"},
	"/api/v1/company-settings": {"$.data.receipt_header", "$.data.receipt_footer"},
}

// GET routes that can't be called bare.
var skippedRoutes = map[string]bool{}

const reportRange = "?date_from=2026-01-01&date_to=2026-12-31"

// GET routes that need query parameters. {branch} is the company's branch.
var routeQuery = map[string]string{
	"/api/v1/products/lookup":          "?code=missing",
	"/api/v1/reports/summary":          reportRange,
	"/api/v1/reports/sales-trend":      reportRange,
	"/api/v1/reports/top-products":     reportRange,
	"/api/v1/reports/category-revenue": reportRange,
	"/api/v1/reports/payment-methods":  reportRange,
	"/api/v1/reports/hourly-sales":     reportRange,
	"/api/v1/daily-settlements/report": "?branch_id={branch}&date=2026-01-01",
}

// Routes that correctly answer 404 for a company with no data.
var routeStatus = map[string]int{
	"/api/v1/products/lookup": http.StatusNotFound,
	"/api/v1/shifts/current":  http.StatusNotFound,
}

// Paginated lists have no next or previous page when they are empty.
var paginationNulls = []string{"$.meta.pagination.next_page", "$.meta.pagination.prev_page"}

// Every list a client renders must be [] rather than null when there is
// nothing in it, or the client crashes calling .map/.forEach on it.
func TestEmptyCompany_GetRoutesReturnNoNullLists(t *testing.T) {
	cleanupDatabase(t)
	a := registerUser(t, "Empty Owner", "empty-lists@store.com", "password123")

	checked := 0
	for _, r := range testEnv.App.Echo().Router().Routes() {
		if r.Method != http.MethodGet || len(r.Parameters) > 0 || !strings.HasPrefix(r.Path, "/api/") || skippedRoutes[r.Path] {
			continue
		}
		checked++
		t.Run(r.Path, func(t *testing.T) {
			query := strings.ReplaceAll(routeQuery[r.Path], "{branch}", strconv.FormatInt(a.BranchID, 10))
			resp := doGet(t, r.Path+query, a.Token)
			want, ok := routeStatus[r.Path]
			if !ok {
				want = http.StatusOK
			}
			require.Equal(t, want, resp.StatusCode, string(resp.Body))
			if want != http.StatusOK {
				return
			}
			allowed := append(emptyCompanyNulls[r.Path], paginationNulls...)
			require.Empty(t, unexpectedNulls(t, resp.Body, allowed),
				"unexpected null (a nil slice encodes as null and breaks clients)")
		})
	}
	require.NotZero(t, checked, "no GET routes found")
}

// unexpectedNulls returns every null in body not covered by allowed. allowed
// holds paths like "$.data[].phone"; "$.data.*" covers every direct field.
func unexpectedNulls(t *testing.T, body []byte, allowed []string) []string {
	t.Helper()
	var v any
	require.NoError(t, json.Unmarshal(body, &v), string(body))
	seen := map[string]bool{}
	var out []string
	for _, p := range nullPaths(v, "$") {
		if !seen[p] && !nullAllowed(p, allowed) {
			out = append(out, p)
		}
		seen[p] = true
	}
	sort.Strings(out)
	return out
}

// nullPaths returns the path of every JSON null in v, with array indexes
// collapsed to [] so one rule covers every row of a list.
func nullPaths(v any, path string) []string {
	var out []string
	switch x := v.(type) {
	case nil:
		out = append(out, path)
	case map[string]any:
		for k, child := range x {
			out = append(out, nullPaths(child, path+"."+k)...)
		}
	case []any:
		for _, child := range x {
			out = append(out, nullPaths(child, path+"[]")...)
		}
	}
	return out
}

func nullAllowed(path string, allowed []string) bool {
	for _, a := range allowed {
		if a == path {
			return true
		}
		if prefix, ok := strings.CutSuffix(a, ".*"); ok {
			rest, found := strings.CutPrefix(path, prefix+".")
			if found && !strings.ContainsAny(rest, ".[") {
				return true
			}
		}
	}
	return false
}
