package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/irvanmhndra/nexpos-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	orderKeysAlways = "id order_no customer_id cashier_id status total_amount total_discount total_tax grand_total notes payment_status fulfillment_type paid_amount balance_due refunded_total items payments created_at updated_at"
	shiftKeys       = "id branch_id cashier_id status opening_float closing_float expected_cash actual_cash cash_difference notes opened_at closed_at created_at updated_at"
	poKeys          = "id branch_id supplier_id po_number status notes total_amount ordered_at received_at cancelled_at created_at updated_at"
	poItemKeys      = "id product_variant_id sku variant_name quantity received_quantity unit_cost subtotal"
	opnameKeys      = "id branch_id opname_number status notes total_items counted_items total_variance_qty total_variance_value started_by completed_by started_at completed_at cancelled_at created_at updated_at"
	opnameItemKeys  = "id stock_opname_id product_variant_id sku product_name variant_name system_stock counted_stock variance_qty unit_cost variance_value notes counted_at"
	settlementKeys  = "id branch_id settlement_date status total_sales total_refunds total_expenses total_expected total_actual total_variance notes recorded_by finalized_by finalized_at created_at updated_at"
	settleItemKeys  = "id payment_method expected_amount actual_amount variance_amount notes"
)

func okData(t *testing.T, resp *testutil.Response) map[string]any {
	t.Helper()
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	return decodeObject(t, decodeResponse(t, resp).Data)
}

// assertFails checks the whole error body, not only the status.
func assertFails(t *testing.T, resp *testutil.Response, status int, code, message string) {
	t.Helper()
	assert.Equal(t, status, resp.StatusCode, "%s", resp.Body)
	want, err := json.Marshal(map[string]any{"success": false, "message": message, "error_code": code})
	require.NoError(t, err)
	assert.JSONEq(t, string(want), string(resp.Body))
}

func assertFieldErrors(t *testing.T, resp *testutil.Response, fields ...[2]string) {
	t.Helper()
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "%s", resp.Body)
	errs := []map[string]string{}
	for _, f := range fields {
		errs = append(errs, map[string]string{"field": f[0], "message": f[1]})
	}
	want, err := json.Marshal(map[string]any{"success": false, "message": "Validation failed", "error_code": "VALIDATION_ERROR", "errors": errs})
	require.NoError(t, err)
	assert.JSONEq(t, string(want), string(resp.Body))
}

func createSupplier(t *testing.T, token, code string) int64 {
	t.Helper()
	return idOf(created(t, "/api/v1/suppliers", token, map[string]any{"name": "Supplier " + code, "code": code}))
}

func stockAt(t *testing.T, variantID, branchID int64) int {
	t.Helper()
	var qty int
	require.NoError(t, testEnv.DB.Get(&qty, `SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2`, variantID, branchID))
	return qty
}

// ============== Inventory ==============

func TestInventory_AdjustListAndStats(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	token, branch := auth.Token, auth.BranchID
	categoryID := createTestCategory(t, token)
	variant := reportProduct(t, token, branch, &categoryID, "INV-1", "Kopi", 10000, 4000.5)
	_, empty := createTestProduct(t, token, categoryID, "Teh", "INV-2", 5000, 2000)

	adjust := func(kind string, qty int) *testutil.Response {
		return doPost(t, "/api/v1/inventory/adjust", map[string]any{
			"variant_id": variant, "branch_id": branch, "type": kind, "quantity": qty, "unit_cost": 4000.5, "note": "Cek",
		}, token)
	}
	resp := adjust("OUT", 100)
	assert.JSONEq(t, `{"success":true,"message":"Stock adjusted successfully"}`, string(resp.Body))
	assert.Equal(t, 900, stockAt(t, variant, branch))
	require.Equal(t, http.StatusOK, adjust("IN", 5).StatusCode)
	require.Equal(t, http.StatusOK, adjust("ADJUST", 8).StatusCode)
	assert.Equal(t, 8, stockAt(t, variant, branch))
	assertFails(t, adjust("OUT", 9), http.StatusBadRequest, "BAD_REQUEST", "Insufficient stock")

	path := fmt.Sprintf("/api/v1/inventory/%d/min-stock?branch_id=%d", variant, branch)
	resp = doPut(t, path, map[string]any{"min_quantity": 8}, token)
	assert.JSONEq(t, `{"success":true,"message":"Min stock updated successfully"}`, string(resp.Body))

	r := decodeResponse(t, doGet(t, fmt.Sprintf("/api/v1/inventory?branch_id=%d", branch), token))
	rows := rowsOf(t, r.Data)
	require.Len(t, rows, 2)
	assertKeys(t, rows[0], strings.Fields("product_variant_id product_id product_name variant_name sku category_name image_data branch_id branch_name current_stock min_stock stock_value status has_variants")...)
	assert.Equal(t, "Kopi", rows[0]["product_name"])
	assert.Equal(t, "low", rows[0]["status"])
	assert.Equal(t, 32004.0, rows[0]["stock_value"])
	assert.Equal(t, "out_of_stock", rows[1]["status"])
	assert.Equal(t, float64(empty), rows[1]["product_variant_id"])
}

func TestInventory_FiltersMovementsAndErrors(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	token, branch := auth.Token, auth.BranchID
	categoryID := createTestCategory(t, token)
	variant := reportProduct(t, token, branch, &categoryID, "INV-3", "Susu", 10000, 4000)
	_, other := createTestProduct(t, token, categoryID, "Gula", "INV-4", 5000, 2000)

	// Without a branch, every variant shows with no stock.
	r := decodeResponse(t, doGet(t, "/api/v1/inventory", token))
	rows := rowsOf(t, r.Data)
	require.Len(t, rows, 2)
	assert.Equal(t, 0.0, rows[0]["branch_id"])
	assert.Equal(t, "", rows[0]["branch_name"])
	assert.Equal(t, 0.0, rows[0]["current_stock"])

	withBranch := fmt.Sprintf("/api/v1/inventory?branch_id=%d", branch)
	assert.Equal(t, []string{"Gula"}, namesOf(t, decodeResponse(t, doGet(t, withBranch+"&status=out_of_stock", token)).Data, "product_name"))
	assert.Equal(t, []string{"Susu"}, namesOf(t, decodeResponse(t, doGet(t, withBranch+"&status=normal", token)).Data, "product_name"))
	assert.Equal(t, []string{}, namesOf(t, decodeResponse(t, doGet(t, withBranch+"&status=low", token)).Data, "product_name"))
	assert.Equal(t, []string{"Gula"}, namesOf(t, decodeResponse(t, doGet(t, withBranch+"&search=INV-4", token)).Data, "product_name"))
	assert.Equal(t, []string{}, namesOf(t, decodeResponse(t, doGet(t, withBranch+"&category=Nope", token)).Data, "product_name"))

	stats := okData(t, doGet(t, fmt.Sprintf("/api/v1/inventory/stats?branch_id=%d", branch), token))
	assert.Equal(t, map[string]any{"total_sku": 2.0, "low_stock": 0.0, "out_of_stock": 1.0, "total_stock_value": 4000000.0}, stats)

	require.Equal(t, http.StatusOK, doPost(t, "/api/v1/inventory/adjust", map[string]any{"variant_id": other, "branch_id": branch, "type": "IN", "quantity": 7, "note": "Masuk"}, token).StatusCode)
	require.Equal(t, http.StatusOK, doPost(t, "/api/v1/inventory/adjust", map[string]any{"variant_id": variant, "branch_id": branch, "type": "ADJUST", "quantity": 1}, token).StatusCode)

	r = decodeResponse(t, doGet(t, "/api/v1/inventory/movements?search=INV-4", token))
	rows = rowsOf(t, r.Data)
	require.Len(t, rows, 1)
	assertKeys(t, rows[0], strings.Fields("id product_variant_id product_name variant_name sku branch_id branch_name type quantity stock_before stock_after unit_cost reference_type reference_id note created_by created_by_name created_at")...)
	assert.Equal(t, "IN", rows[0]["type"])
	assert.Equal(t, "Masuk", rows[0]["note"])
	assert.Equal(t, 0.0, rows[0]["stock_before"])
	assert.Equal(t, 7.0, rows[0]["stock_after"])
	assert.Nil(t, rows[0]["unit_cost"])
	assert.Equal(t, float64(auth.UserID), rows[0]["created_by"])
	assert.Contains(t, rows[0]["created_by_name"], "Test User")
	assert.Len(t, rowsOf(t, decodeResponse(t, doGet(t, "/api/v1/inventory/movements?type=ADJUST", token)).Data), 1)
	assert.Len(t, rowsOf(t, decodeResponse(t, doGet(t, "/api/v1/inventory/movements?start_date=2000-01-01&end_date=2000-01-02", token)).Data), 0)

	stats = okData(t, doGet(t, "/api/v1/inventory/movements/stats", token))
	assert.Equal(t, map[string]any{"total_movements": 2.0, "stock_in": 1.0, "stock_out": 0.0, "adjustments": 1.0}, stats)

	assertFieldErrors(t, doPost(t, "/api/v1/inventory/adjust", map[string]any{"type": "NOPE", "quantity": -1}, token),
		[2]string{"variant_id", "variant_id is required"}, [2]string{"branch_id", "branch_id is required"},
		[2]string{"type", "type must be one of: IN OUT ADJUST"}, [2]string{"quantity", "quantity must be at least 1 characters"})
	assertFails(t, doPost(t, "/api/v1/inventory/adjust", map[string]any{"variant_id": 999999, "branch_id": branch, "type": "IN", "quantity": 1}, token),
		http.StatusNotFound, "NOT_FOUND", "Product variant not found not found")
	assertFails(t, doPost(t, "/api/v1/inventory/adjust", map[string]any{"variant_id": variant, "branch_id": 999999, "type": "IN", "quantity": 1}, token),
		http.StatusNotFound, "NOT_FOUND", "Branch not found not found")
	assertFails(t, doPut(t, fmt.Sprintf("/api/v1/inventory/%d/min-stock", variant), map[string]any{"min_quantity": 1}, token),
		http.StatusBadRequest, "BAD_REQUEST", "branch_id is required")
	assertFails(t, doPut(t, "/api/v1/inventory/abc/min-stock?branch_id=1", map[string]any{"min_quantity": 1}, token),
		http.StatusBadRequest, "BAD_REQUEST", "invalid variant ID")
	assertFieldErrors(t, doPut(t, fmt.Sprintf("/api/v1/inventory/%d/min-stock?branch_id=%d", variant, branch), map[string]any{"min_quantity": -1}, token),
		[2]string{"min_quantity", "min_quantity must be at least 0 characters"})

	other2 := registerTestUser(t)
	assertFails(t, doPost(t, "/api/v1/inventory/adjust", map[string]any{"variant_id": variant, "branch_id": other2.BranchID, "type": "IN", "quantity": 1}, token),
		http.StatusForbidden, "FORBIDDEN", "Branch does not belong to your company")
	// A foreign branch's name is not exposed through the inventory list.
	rows = rowsOf(t, decodeResponse(t, doGet(t, fmt.Sprintf("/api/v1/inventory?branch_id=%d", other2.BranchID), token)).Data)
	require.NotEmpty(t, rows)
	assert.Equal(t, "", rows[0]["branch_name"])
	assert.Equal(t, 0.0, rows[0]["branch_id"])
}

// ============== Purchase orders ==============

func TestPurchaseOrder_CreateAndReceiveInParts(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	token, branch := auth.Token, auth.BranchID
	categoryID := createTestCategory(t, token)
	_, variant := createTestProduct(t, token, categoryID, "Kopi", "PO-1", 10000, 4000)
	supplier := createSupplier(t, token, "SUP-1")

	po := created(t, "/api/v1/purchase-orders", token, map[string]any{
		"branch_id": branch, "supplier_id": supplier, "notes": "Bulanan",
		"items": []map[string]any{
			{"product_variant_id": variant, "sku": "PO-1", "variant_name": "Default", "quantity": 10, "unit_cost": 1500.5},
			{"sku": "CUSTOM", "variant_name": "Lepas", "quantity": 2, "unit_cost": 1000},
		},
	})
	assertKeys(t, po, append(strings.Fields(poKeys), "items")...)
	assert.Equal(t, "draft", po["status"])
	assert.Equal(t, 17005.0, po["total_amount"])
	assert.Regexp(t, `^PO-\d{8}-0001$`, po["po_number"])
	assert.NotNil(t, po["ordered_at"])
	items := po["items"].([]any)
	require.Len(t, items, 2)
	first := items[0].(map[string]any)
	assertKeys(t, first, strings.Fields(poItemKeys)...)
	assert.Equal(t, 15005.0, first["subtotal"])
	id := idOf(po)
	firstID, secondID := int64(first["id"].(float64)), int64(items[1].(map[string]any)["id"].(float64))

	receive := func(body ...map[string]any) *testutil.Response {
		return doPost(t, fmt.Sprintf("/api/v1/purchase-orders/%d/receive", id), map[string]any{"items": body}, token)
	}
	got := okData(t, receive(map[string]any{"item_id": firstID, "received_quantity": 4}))
	assert.Equal(t, "partial", got["status"])
	assert.Nil(t, got["received_at"])
	assert.Equal(t, 4, stockAt(t, variant, branch))

	// A line without a variant has no stock to move; over-receiving still stocks all of it.
	got = okData(t, receive(map[string]any{"item_id": firstID, "received_quantity": 9}, map[string]any{"item_id": secondID, "received_quantity": 2}))
	assert.Equal(t, "received", got["status"])
	assert.NotNil(t, got["received_at"])
	assert.Equal(t, 10.0, got["items"].([]any)[0].(map[string]any)["received_quantity"], "received quantity is capped at what was ordered")
	assert.Equal(t, 13, stockAt(t, variant, branch))

	var cost float64
	require.NoError(t, testEnv.DB.Get(&cost, `SELECT last_purchase_cost FROM product_variants WHERE id = $1`, variant))
	assert.Equal(t, 1500.5, cost)
	var movements int
	require.NoError(t, testEnv.DB.Get(&movements, `SELECT count(*) FROM stock_movements WHERE reference_type = 'purchase_order' AND reference_id = $1 AND unit_cost = 1500.5`, id))
	assert.Equal(t, 2, movements)

	assertFails(t, receive(map[string]any{"item_id": firstID, "received_quantity": 1}), http.StatusBadRequest, "BAD_REQUEST", "Purchase order cannot be received in its current status")

	detail := okData(t, doGet(t, fmt.Sprintf("/api/v1/purchase-orders/%d", id), token))
	assert.Equal(t, "received", detail["status"])
	assert.Len(t, detail["items"], 2)

	r := decodeResponse(t, doGet(t, "/api/v1/purchase-orders?status=received", token))
	rows := rowsOf(t, r.Data)
	require.Len(t, rows, 1)
	assertKeys(t, rows[0], strings.Fields(poKeys)...)
	assert.Len(t, rowsOf(t, decodeResponse(t, doGet(t, "/api/v1/purchase-orders?status=draft", token)).Data), 0)
	assert.Len(t, rowsOf(t, decodeResponse(t, doGet(t, fmt.Sprintf("/api/v1/purchase-orders?supplier_id=%d", supplier+1), token)).Data), 0)
}

func TestPurchaseOrder_Errors(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	token := auth.Token
	categoryID := createTestCategory(t, token)
	_, variant := createTestProduct(t, token, categoryID, "Kopi", "PO-2", 10000, 4000)
	supplier := createSupplier(t, token, "SUP-2")

	assertFieldErrors(t, doPost(t, "/api/v1/purchase-orders", map[string]any{}, token),
		[2]string{"branch_id", "branch_id is required"}, [2]string{"supplier_id", "supplier_id is required"}, [2]string{"items", "items is required"})
	assertFieldErrors(t, doPost(t, "/api/v1/purchase-orders", map[string]any{"branch_id": 1, "supplier_id": 1, "items": []any{}}, token),
		[2]string{"items", "items must be at least 1 characters"})
	assertFieldErrors(t, doPost(t, "/api/v1/purchase-orders", map[string]any{"branch_id": 1, "supplier_id": 1, "items": []map[string]any{{"unit_cost": -1, "quantity": -2}}}, token),
		[2]string{"sku", "sku is required"}, [2]string{"variant_name", "variant_name is required"},
		[2]string{"quantity", "quantity must be at least 1 characters"}, [2]string{"unit_cost", "unit_cost must be at least 0 characters"})
	line := []map[string]any{{"sku": "X", "variant_name": "Y", "quantity": 1, "unit_cost": 1}}
	assertFails(t, doPost(t, "/api/v1/purchase-orders", map[string]any{"branch_id": 1, "supplier_id": 999999, "items": line}, token),
		http.StatusNotFound, "NOT_FOUND", "Supplier not found not found")

	po := created(t, "/api/v1/purchase-orders", token, map[string]any{"branch_id": 1, "supplier_id": supplier, "items": []map[string]any{{"product_variant_id": variant, "sku": "X", "variant_name": "Y", "quantity": 1, "unit_cost": 1}}})
	path := fmt.Sprintf("/api/v1/purchase-orders/%d", idOf(po))
	assertFieldErrors(t, doPost(t, path+"/receive", map[string]any{"items": []map[string]any{{"received_quantity": -1}}}, token),
		[2]string{"item_id", "item_id is required"}, [2]string{"received_quantity", "received_quantity must be at least 0 characters"})
	assertFails(t, doPost(t, path+"/receive", map[string]any{"items": []map[string]any{{"item_id": 999999, "received_quantity": 1}}}, token),
		http.StatusBadRequest, "BAD_REQUEST", "Purchase order item not found")
	assertFails(t, doGet(t, "/api/v1/purchase-orders/abc", token), http.StatusBadRequest, "BAD_REQUEST", "invalid purchase order ID")
	assertFails(t, doGet(t, "/api/v1/purchase-orders/999999", token), http.StatusNotFound, "NOT_FOUND", "Purchase order not found not found")

	other := registerTestUser(t)
	assertFails(t, doGet(t, path, other.Token), http.StatusNotFound, "NOT_FOUND", "Purchase order not found not found")
	assertFails(t, doPost(t, path+"/receive", map[string]any{"items": []map[string]any{{"item_id": 1, "received_quantity": 1}}}, other.Token),
		http.StatusNotFound, "NOT_FOUND", "Purchase order not found not found")
}

// ============== Shifts ==============

func TestShift_OpenPayCloseFlow(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	token, branch := auth.Token, auth.BranchID
	categoryID := createTestCategory(t, token)
	variant := reportProduct(t, token, branch, &categoryID, "SH-1", "Kopi", 25000, 1000)

	assertFails(t, doGet(t, "/api/v1/shifts/current", token), http.StatusNotFound, "NOT_FOUND", "No open shift found not found")

	shift := created(t, "/api/v1/shifts", token, map[string]any{"branch_id": branch, "opening_float": 100000.5})
	assertKeys(t, shift, strings.Fields(shiftKeys)...)
	assert.Equal(t, "open", shift["status"])
	assert.Equal(t, 100000.5, shift["opening_float"])
	assert.Equal(t, 0.0, shift["closing_float"])
	assert.Nil(t, shift["actual_cash"])
	assert.Nil(t, shift["closed_at"])

	assertFails(t, doPost(t, "/api/v1/shifts", map[string]any{"branch_id": branch, "opening_float": 1}, token),
		http.StatusBadRequest, "BAD_REQUEST", "There is already an open shift for this cashier at this branch")
	assert.Equal(t, shift["id"], okData(t, doGet(t, "/api/v1/shifts/current", token))["id"])

	reportCompleteOrder(t, token, deliveryCustomer(t, token), variant, 2, "cash")
	reportCompleteOrder(t, token, deliveryCustomer(t, token), variant, 1, "qris")

	id := idOf(shift)
	closed := okData(t, doPost(t, fmt.Sprintf("/api/v1/shifts/%d/close", id), map[string]any{"actual_cash": 150000, "notes": "Pas"}, token))
	assert.Equal(t, "closed", closed["status"])
	assert.Equal(t, 150000.5, closed["expected_cash"])
	assert.Equal(t, 150000.0, closed["actual_cash"])
	assert.Equal(t, 150000.0, closed["closing_float"])
	assert.Equal(t, -0.5, closed["cash_difference"])
	assert.Equal(t, "Pas", closed["notes"])
	assert.NotNil(t, closed["closed_at"])

	assertFails(t, doPost(t, fmt.Sprintf("/api/v1/shifts/%d/close", id), map[string]any{"actual_cash": 1}, token),
		http.StatusBadRequest, "BAD_REQUEST", "Shift is not open")
	assert.Equal(t, "closed", okData(t, doGet(t, fmt.Sprintf("/api/v1/shifts/%d", id), token))["status"])

	for query, want := range map[string]int{
		"": 1, "?status=closed": 1, "?status=open": 0,
		fmt.Sprintf("?branch_id=%d", branch): 1, "?branch_id=999999": 0,
		fmt.Sprintf("?cashier_id=%d", auth.UserID): 1, "?cashier_id=999999": 0,
		"?date_from=2000-01-01&date_to=2000-01-02": 0,
	} {
		assert.Len(t, rowsOf(t, decodeResponse(t, doGet(t, "/api/v1/shifts"+query, token)).Data), want, query)
	}
}

// deliveryCustomer returns a customer for the delivery orders the report helpers create.
func deliveryCustomer(t *testing.T, token string) *int64 {
	t.Helper()
	id := idOf(created(t, "/api/v1/customers", token, map[string]any{"code": fmt.Sprintf("C-%d", testutil.UniqueCounter()), "name": "Pelanggan"}))
	return &id
}

func TestShift_ErrorsAndTenancy(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	other := registerTestUser(t)
	token := auth.Token

	assertFieldErrors(t, doPost(t, "/api/v1/shifts", map[string]any{"opening_float": -1}, token),
		[2]string{"branch_id", "branch_id is required"}, [2]string{"opening_float", "opening_float must be at least 0 characters"})
	assertFails(t, doPost(t, "/api/v1/shifts", map[string]any{"branch_id": 999999}, token), http.StatusNotFound, "NOT_FOUND", "Branch not found not found")
	assertFails(t, doPost(t, "/api/v1/shifts", map[string]any{"branch_id": other.BranchID}, token), http.StatusForbidden, "FORBIDDEN", "Branch does not belong to your company")
	assertFails(t, doGet(t, "/api/v1/shifts/abc", token), http.StatusBadRequest, "BAD_REQUEST", "invalid shift ID")
	assertFails(t, doGet(t, "/api/v1/shifts/999999", token), http.StatusNotFound, "NOT_FOUND", "Shift not found not found")

	shift := created(t, "/api/v1/shifts", token, map[string]any{"branch_id": auth.BranchID})
	path := fmt.Sprintf("/api/v1/shifts/%d", idOf(shift))
	assertFails(t, doGet(t, path, other.Token), http.StatusNotFound, "NOT_FOUND", "Shift not found not found")
	assertFails(t, doPost(t, path+"/close", map[string]any{"actual_cash": 1}, other.Token), http.StatusNotFound, "NOT_FOUND", "Shift not found not found")
	assertFieldErrors(t, doPost(t, path+"/close", map[string]any{"actual_cash": -1}, token), [2]string{"actual_cash", "actual_cash must be at least 0 characters"})
}

// ============== Stock opnames ==============

func TestStockOpname_CountAndComplete(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	token, branch := auth.Token, auth.BranchID
	categoryID := createTestCategory(t, token)
	kopi := reportProduct(t, token, branch, &categoryID, "OP-1", "Kopi", 10000, 4000)
	teh := reportProduct(t, token, branch, &categoryID, "OP-2", "Teh", 5000, 2000.5)
	_, gula := createTestProduct(t, token, categoryID, "Gula", "OP-3", 3000, 1000)

	opname := created(t, "/api/v1/stock-opnames", token, map[string]any{"branch_id": branch, "notes": "Bulanan"})
	assertKeys(t, opname, append(strings.Fields(opnameKeys), "items")...)
	assert.Equal(t, "in_progress", opname["status"])
	assert.Regexp(t, `^OPN-\d{8}-0001$`, opname["opname_number"])
	assert.Equal(t, 3.0, opname["total_items"])
	assert.Equal(t, 0.0, opname["counted_items"])
	assert.Equal(t, float64(auth.UserID), opname["started_by"])
	id := idOf(opname)
	path := fmt.Sprintf("/api/v1/stock-opnames/%d", id)

	items := map[string]map[string]any{}
	for _, raw := range opname["items"].([]any) {
		item := raw.(map[string]any)
		assertKeys(t, item, strings.Fields(opnameItemKeys)...)
		items[item["sku"].(string)] = item
	}
	assert.Equal(t, 1000.0, items["OP-2"]["system_stock"])
	assert.Equal(t, 2000.5, items["OP-2"]["unit_cost"])
	assert.Nil(t, items["OP-2"]["counted_stock"])
	assert.Equal(t, 0.0, items["OP-3"]["system_stock"])

	itemID := func(sku string) int64 { return int64(items[sku]["id"].(float64)) }
	one := okData(t, doRequest(t, http.MethodPatch, fmt.Sprintf("%s/items/%d", path, itemID("OP-2")), map[string]any{"counted_stock": 990, "notes": "Rusak"}, token))
	assertKeys(t, one, strings.Fields(opnameItemKeys)...)
	assert.Equal(t, 990.0, one["counted_stock"])
	assert.Equal(t, -10.0, one["variance_qty"])
	assert.Equal(t, -20005.0, one["variance_value"])
	assert.Equal(t, "Rusak", one["notes"])
	assert.NotNil(t, one["counted_at"])

	bulk := okData(t, doRequest(t, http.MethodPatch, path+"/items", map[string]any{"items": []map[string]any{
		{"item_id": itemID("OP-1"), "counted_stock": 1000}, {"item_id": itemID("OP-3"), "counted_stock": 4},
	}}, token))
	assert.Equal(t, 3.0, bulk["counted_items"])
	assertKeys(t, bulk, append(strings.Fields(opnameKeys), "items")...)

	// A sale during the count: completion sets the counted quantity regardless.
	require.Equal(t, http.StatusOK, doPost(t, "/api/v1/inventory/adjust", map[string]any{"variant_id": kopi, "branch_id": branch, "type": "OUT", "quantity": 5}, token).StatusCode)

	done := okData(t, doPost(t, path+"/complete", map[string]any{"notes": "Selesai"}, token))
	assert.Equal(t, "completed", done["status"])
	assert.Equal(t, "Selesai", done["notes"])
	assert.Equal(t, -6.0, done["total_variance_qty"])
	assert.Equal(t, -16005.0, done["total_variance_value"])
	assert.Equal(t, float64(auth.UserID), done["completed_by"])
	assert.NotNil(t, done["completed_at"])
	assert.Equal(t, 1000, stockAt(t, kopi, branch))
	assert.Equal(t, 990, stockAt(t, teh, branch))
	assert.Equal(t, 4, stockAt(t, gula, branch))

	var quantity, before, after int
	require.NoError(t, testEnv.DB.QueryRow(`SELECT quantity, stock_before, stock_after FROM stock_movements WHERE type = 'ADJUST' AND reference_type = 'stock_opname' AND reference_id = $1 AND product_variant_id = $2`, id, kopi).Scan(&quantity, &before, &after))
	assert.Equal(t, [3]int{5, 995, 1000}, [3]int{quantity, before, after})

	for _, resp := range []*testutil.Response{doPost(t, path+"/complete", nil, token)} {
		assertFails(t, resp, http.StatusBadRequest, "BAD_REQUEST", "Stock opname is not in progress")
	}
	assertFails(t, doPost(t, path+"/cancel", nil, token), http.StatusBadRequest, "BAD_REQUEST", "Only in-progress opnames can be cancelled")
	assertFails(t, doRequest(t, http.MethodPatch, fmt.Sprintf("%s/items/%d", path, itemID("OP-1")), map[string]any{"counted_stock": 1}, token),
		http.StatusBadRequest, "BAD_REQUEST", "Cannot edit items on a completed or cancelled opname")
	assertFails(t, doRequest(t, http.MethodPatch, path+"/items", map[string]any{"items": []map[string]any{{"item_id": 1, "counted_stock": 1}}}, token),
		http.StatusBadRequest, "BAD_REQUEST", "Cannot edit items on a completed or cancelled opname")

	r := decodeResponse(t, doGet(t, "/api/v1/stock-opnames?status=completed", token))
	rows := rowsOf(t, r.Data)
	require.Len(t, rows, 1)
	assertKeys(t, rows[0], strings.Fields(opnameKeys)...)
	assert.Equal(t, 3.0, rows[0]["counted_items"])
	assert.Len(t, rowsOf(t, decodeResponse(t, doGet(t, "/api/v1/stock-opnames?status=cancelled", token)).Data), 0)

	second := created(t, "/api/v1/stock-opnames", token, map[string]any{"branch_id": branch, "category_id": categoryID})
	assert.Regexp(t, `-0002$`, second["opname_number"])
	cancelled := okData(t, doPost(t, fmt.Sprintf("/api/v1/stock-opnames/%d/cancel", idOf(second)), nil, token))
	assert.Equal(t, "cancelled", cancelled["status"])
	assert.NotNil(t, cancelled["cancelled_at"])
	assert.Len(t, rowsOf(t, decodeResponse(t, doGet(t, fmt.Sprintf("/api/v1/stock-opnames?branch_id=%d", branch), token)).Data), 2)
}

func TestStockOpname_ErrorsAndTenancy(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	other := registerTestUser(t)
	token, branch := auth.Token, auth.BranchID

	assertFieldErrors(t, doPost(t, "/api/v1/stock-opnames", map[string]any{}, token), [2]string{"branch_id", "branch_id is required"})
	assertFails(t, doPost(t, "/api/v1/stock-opnames", map[string]any{"branch_id": 999999}, token), http.StatusNotFound, "NOT_FOUND", "Branch not found not found")
	assertFails(t, doPost(t, "/api/v1/stock-opnames", map[string]any{"branch_id": other.BranchID}, token), http.StatusForbidden, "FORBIDDEN", "Branch does not belong to your company")
	assertFails(t, doPost(t, "/api/v1/stock-opnames", map[string]any{"branch_id": branch}, token), http.StatusBadRequest, "BAD_REQUEST", "No active products found to snapshot for this branch")

	categoryID := createTestCategory(t, token)
	reportProduct(t, token, branch, &categoryID, "OP-9", "Kopi", 10000, 4000)
	opname := created(t, "/api/v1/stock-opnames", token, map[string]any{"branch_id": branch})
	path := fmt.Sprintf("/api/v1/stock-opnames/%d", idOf(opname))

	assertFails(t, doGet(t, "/api/v1/stock-opnames/abc", token), http.StatusBadRequest, "BAD_REQUEST", "invalid stock opname ID")
	assertFails(t, doGet(t, "/api/v1/stock-opnames/999999", token), http.StatusNotFound, "NOT_FOUND", "Stock opname not found not found")
	assertFails(t, doRequest(t, http.MethodPatch, path+"/items/abc", map[string]any{"counted_stock": 1}, token), http.StatusBadRequest, "BAD_REQUEST", "invalid item ID")
	assertFails(t, doRequest(t, http.MethodPatch, path+"/items/999999", map[string]any{"counted_stock": 1}, token), http.StatusNotFound, "NOT_FOUND", "Stock opname item not found not found")
	assertFieldErrors(t, doRequest(t, http.MethodPatch, path+"/items/1", map[string]any{"counted_stock": -1}, token), [2]string{"counted_stock", "counted_stock must be at least 0 characters"})
	assertFails(t, doRequest(t, http.MethodPatch, path+"/items", map[string]any{"items": []map[string]any{{"item_id": 999999, "counted_stock": 1}}}, token), http.StatusBadRequest, "BAD_REQUEST", "Stock opname item not found")
	assertFieldErrors(t, doRequest(t, http.MethodPatch, path+"/items", map[string]any{"items": []any{}}, token), [2]string{"items", "items must be at least 1 characters"})

	for _, resp := range []*testutil.Response{
		doGet(t, path, other.Token),
		doPost(t, path+"/complete", nil, other.Token),
		doPost(t, path+"/cancel", nil, other.Token),
		doRequest(t, http.MethodPatch, path+"/items", map[string]any{"items": []map[string]any{{"item_id": 1, "counted_stock": 1}}}, other.Token),
	} {
		assertFails(t, resp, http.StatusNotFound, "NOT_FOUND", "Stock opname not found not found")
	}
}

// ============== Daily settlements ==============

func TestDailySettlement_ReportCreateAndFinalize(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	token, branch := auth.Token, auth.BranchID
	categoryID := createTestCategory(t, token)
	variant := reportProduct(t, token, branch, &categoryID, "DS-1", "Kopi", 10000, 4000)

	var today string
	require.NoError(t, testEnv.DB.Get(&today, `SELECT CURRENT_DATE::text`))

	customer := deliveryCustomer(t, token)
	reportCompleteOrder(t, token, customer, variant, 3, "cash")
	reportCompleteOrder(t, token, customer, variant, 1, "qris")
	created(t, "/api/v1/expenses", token, map[string]any{"branch_id": branch, "amount": 5000.5, "description": "Es batu", "expense_date": today})

	orderID, total := reportOrder(t, token, customer, variant, 1)
	require.Equal(t, http.StatusOK, doPost(t, fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token).StatusCode)
	paid := okData(t, doPost(t, fmt.Sprintf("/api/v1/orders/%d/payments", orderID), map[string]any{"payments": []map[string]any{{"method": "e_wallet", "amount": total}}}, token))
	payment := paid["payments"].([]any)[0].(map[string]any)
	require.Equal(t, http.StatusOK, doPost(t, fmt.Sprintf("/api/v1/orders/%d/refund", orderID), map[string]any{"payment_id": payment["id"], "amount": 2500, "refund_reason": "Salah"}, token).StatusCode)

	report := okData(t, doGet(t, fmt.Sprintf("/api/v1/daily-settlements/report?branch_id=%d&date=%s", branch, today), token))
	assertKeys(t, report, strings.Fields("branch_id settlement_date total_sales total_refunds total_expenses total_expected by_method")...)
	assert.Equal(t, today, report["settlement_date"])
	assert.Equal(t, 50000.0, report["total_sales"])
	assert.Equal(t, 2500.0, report["total_refunds"])
	assert.Equal(t, 5000.5, report["total_expenses"])
	assert.Equal(t, 42499.5, report["total_expected"])
	methods := report["by_method"].([]any)
	require.Len(t, methods, 6)
	cash := methods[0].(map[string]any)
	assert.Equal(t, map[string]any{"payment_method": "cash", "gross_sales": 30000.0, "refunds": 0.0, "expenses_out": 5000.5, "expected_amount": 24999.5}, cash)
	order := []string{}
	for _, m := range methods {
		order = append(order, m.(map[string]any)["payment_method"].(string))
	}
	assert.Equal(t, []string{"cash", "debit_card", "credit_card", "e_wallet", "bank_transfer", "qris"}, order)

	settlement := created(t, "/api/v1/daily-settlements", token, map[string]any{"branch_id": branch, "settlement_date": today, "notes": "Harian"})
	assertKeys(t, settlement, append(strings.Fields(settlementKeys), "items")...)
	assert.Equal(t, today+"T00:00:00Z", settlement["settlement_date"])
	assert.Equal(t, "draft", settlement["status"])
	assert.Equal(t, 42499.5, settlement["total_expected"])
	assert.Equal(t, 0.0, settlement["total_actual"])
	assert.Equal(t, -42499.5, settlement["total_variance"])
	assert.Equal(t, float64(auth.UserID), settlement["recorded_by"])
	id := idOf(settlement)
	path := fmt.Sprintf("/api/v1/daily-settlements/%d", id)
	items := settlement["items"].([]any)
	require.Len(t, items, 6)
	firstItem := items[0].(map[string]any)
	assertKeys(t, firstItem, strings.Fields(settleItemKeys)...)
	assert.Equal(t, "cash", firstItem["payment_method"])
	assert.Nil(t, firstItem["actual_amount"])
	assert.Equal(t, -24999.5, firstItem["variance_amount"])

	assertFails(t, doPost(t, "/api/v1/daily-settlements", map[string]any{"branch_id": branch, "settlement_date": today}, token),
		http.StatusBadRequest, "BAD_REQUEST", "Settlement already exists for this branch and date")

	item := okData(t, doRequest(t, http.MethodPatch, fmt.Sprintf("%s/items/%d", path, int64(firstItem["id"].(float64))), map[string]any{"actual_amount": 25000, "notes": "Lebih"}, token))
	assertKeys(t, item, strings.Fields(settleItemKeys)...)
	assert.Equal(t, 25000.0, item["actual_amount"])
	assert.Equal(t, 0.5, item["variance_amount"])
	assert.Equal(t, "Lebih", item["notes"])

	second := items[1].(map[string]any)
	bulk := okData(t, doRequest(t, http.MethodPatch, path+"/items", map[string]any{"items": []map[string]any{{"item_id": second["id"], "actual_amount": 10}}}, token))
	assert.Equal(t, 25010.0, bulk["total_actual"])

	done := okData(t, doPost(t, path+"/finalize", map[string]any{"notes": "Tutup"}, token))
	assert.Equal(t, "finalized", done["status"])
	assert.Equal(t, "Tutup", done["notes"])
	assert.Equal(t, float64(auth.UserID), done["finalized_by"])
	assert.NotNil(t, done["finalized_at"])

	assertFails(t, doPost(t, path+"/finalize", nil, token), http.StatusBadRequest, "BAD_REQUEST", "Settlement is not in draft status")
	assertFails(t, doRequest(t, http.MethodPatch, path+"/items/1", map[string]any{"actual_amount": 1}, token), http.StatusBadRequest, "BAD_REQUEST", "Cannot edit a finalized settlement")
	assertFails(t, doRequest(t, http.MethodPatch, path+"/items", map[string]any{"items": []map[string]any{{"item_id": 1, "actual_amount": 1}}}, token), http.StatusBadRequest, "BAD_REQUEST", "Cannot edit a finalized settlement")

	r := decodeResponse(t, doGet(t, "/api/v1/daily-settlements?status=finalized", token))
	rows := rowsOf(t, r.Data)
	require.Len(t, rows, 1)
	assertKeys(t, rows[0], append(strings.Fields(settlementKeys), "items")...)
	assert.Equal(t, 25010.0, rows[0]["total_actual"])
	for query, want := range map[string]int{
		"?status=draft": 0, fmt.Sprintf("?branch_id=%d", branch): 1, "?branch_id=999999": 0,
		"?date_from=2000-01-01&date_to=2000-01-02": 0, "?date_from=" + today + "&date_to=" + today: 1,
	} {
		assert.Len(t, rowsOf(t, decodeResponse(t, doGet(t, "/api/v1/daily-settlements"+query, token)).Data), want, query)
	}
}

func TestDailySettlement_ErrorsAndTenancy(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	other := registerTestUser(t)
	token, branch := auth.Token, auth.BranchID

	base := "/api/v1/daily-settlements"
	assertFails(t, doGet(t, base+"/report?date=2026-01-01", token), http.StatusBadRequest, "BAD_REQUEST", "branch_id is required")
	assertFails(t, doGet(t, fmt.Sprintf("%s/report?branch_id=%d", base, branch), token), http.StatusBadRequest, "BAD_REQUEST", "date is required")
	assertFails(t, doGet(t, base+"/report?branch_id=999999&date=2026-01-01", token), http.StatusNotFound, "NOT_FOUND", "Branch not found not found")
	assertFails(t, doGet(t, fmt.Sprintf("%s/report?branch_id=%d&date=2026-01-01", base, other.BranchID), token), http.StatusForbidden, "FORBIDDEN", "Branch does not belong to your company")

	assertFieldErrors(t, doPost(t, base, map[string]any{}, token), [2]string{"branch_id", "branch_id is required"}, [2]string{"settlement_date", "settlement_date is required"})
	for _, bad := range []string{"2026-1-1", "01-01-2026", "2026-13-01", "soon"} {
		assertFails(t, doPost(t, base, map[string]any{"branch_id": branch, "settlement_date": bad}, token), http.StatusBadRequest, "BAD_REQUEST", "settlement_date must be YYYY-MM-DD")
	}
	assertFails(t, doPost(t, base, map[string]any{"branch_id": other.BranchID, "settlement_date": "2026-01-01"}, token), http.StatusForbidden, "FORBIDDEN", "Branch does not belong to your company")

	settlement := created(t, base, token, map[string]any{"branch_id": branch, "settlement_date": "2026-01-01"})
	path := fmt.Sprintf("%s/%d", base, idOf(settlement))
	assertFails(t, doGet(t, base+"/abc", token), http.StatusBadRequest, "BAD_REQUEST", "invalid settlement ID")
	assertFails(t, doGet(t, base+"/999999", token), http.StatusNotFound, "NOT_FOUND", "Daily settlement not found not found")
	assertFails(t, doRequest(t, http.MethodPatch, path+"/items/abc", map[string]any{"actual_amount": 1}, token), http.StatusBadRequest, "BAD_REQUEST", "invalid item ID")
	assertFails(t, doRequest(t, http.MethodPatch, path+"/items/999999", map[string]any{"actual_amount": 1}, token), http.StatusNotFound, "NOT_FOUND", "Settlement item not found not found")
	assertFieldErrors(t, doRequest(t, http.MethodPatch, path+"/items/1", map[string]any{"actual_amount": -1}, token), [2]string{"actual_amount", "actual_amount is invalid"})
	assertFails(t, doRequest(t, http.MethodPatch, path+"/items", map[string]any{"items": []map[string]any{{"item_id": 999999, "actual_amount": 1}}}, token), http.StatusBadRequest, "BAD_REQUEST", "Settlement item not found")
	assertFieldErrors(t, doRequest(t, http.MethodPatch, path+"/items", map[string]any{"items": []map[string]any{{"actual_amount": -1}}}, token),
		[2]string{"item_id", "item_id is required"}, [2]string{"actual_amount", "actual_amount is invalid"})

	for _, resp := range []*testutil.Response{
		doGet(t, path, other.Token),
		doPost(t, path+"/finalize", nil, other.Token),
		doRequest(t, http.MethodPatch, path+"/items", map[string]any{"items": []map[string]any{{"item_id": 1, "actual_amount": 1}}}, other.Token),
	} {
		assertFails(t, resp, http.StatusNotFound, "NOT_FOUND", "Daily settlement not found not found")
	}
}

// ============== Orders: rules the older flow tests don't pin down ==============

func TestOrderRules_PricingTaxRoundingAndPromotion(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token
	item := []map[string]any{{"product_variant_id": td.variantID, "quantity": 3, "discount_amount": 10.5}}

	preview := okData(t, doPost(t, "/api/v1/orders/preview", map[string]any{"items": item}, token))
	assert.Equal(t, map[string]any{"subtotal": 289.5, "item_discount": 10.5, "promo_discount": 0.0, "tax": 0.0, "grand_total": 289.5}, preview)

	// Exclusive tax is added, rounded per line.
	require.Equal(t, http.StatusOK, doPut(t, "/api/v1/company-settings", map[string]any{"tax_enabled": true, "tax_rate": 11, "tax_inclusive": false}, token).StatusCode)
	preview = okData(t, doPost(t, "/api/v1/orders/preview", map[string]any{"items": item}, token))
	assert.Equal(t, 31.85, preview["tax"])
	assert.Equal(t, 321.35, preview["grand_total"])

	// Inclusive tax is reported but not charged again; cash rounding applies last.
	require.Equal(t, http.StatusOK, doPut(t, "/api/v1/company-settings", map[string]any{"tax_inclusive": true, "rounding_enabled": true, "rounding_amount": 100}, token).StatusCode)
	preview = okData(t, doPost(t, "/api/v1/orders/preview", map[string]any{"items": item}, token))
	assert.Equal(t, 28.69, preview["tax"])
	assert.Equal(t, 300.0, preview["grand_total"])
	require.Equal(t, http.StatusOK, doPut(t, "/api/v1/company-settings", map[string]any{"tax_enabled": false, "rounding_enabled": false}, token).StatusCode)

	// The best active promotion applies on its own; a code picks exactly one.
	created(t, "/api/v1/promotions", token, map[string]any{"code": "P10", "name": "Sepuluh", "type": "discount", "discount_type": "percentage", "discount_value": 10, "max_discount": 20, "start_at": "2020-01-01T00:00:00Z", "is_active": true})
	created(t, "/api/v1/promotions", token, map[string]any{"code": "F5", "name": "Lima", "type": "discount", "discount_type": "fixed", "discount_value": 5, "start_at": "2020-01-01T00:00:00Z", "is_active": true})
	created(t, "/api/v1/promotions", token, map[string]any{"code": "MIN", "name": "Minimal", "type": "discount", "discount_type": "fixed", "discount_value": 250, "min_purchase": 100000, "start_at": "2020-01-01T00:00:00Z", "is_active": true})

	preview = okData(t, doPost(t, "/api/v1/orders/preview", map[string]any{"items": item}, token))
	assert.Equal(t, 20.0, preview["promo_discount"], "10%% capped at max_discount beats the fixed 5")
	assert.Equal(t, 269.5, preview["grand_total"])
	applied := preview["applied_promotion"].(map[string]any)
	assertKeys(t, applied, "id", "code", "name", "discount_amount")
	assert.Equal(t, "P10", applied["code"])

	preview = okData(t, doPost(t, "/api/v1/orders/preview", map[string]any{"items": item, "promo_code": "F5"}, token))
	assert.Equal(t, 5.0, preview["promo_discount"])
	preview = okData(t, doPost(t, "/api/v1/orders/preview", map[string]any{"items": item, "promo_code": "NOPE"}, token))
	assert.Equal(t, 0.0, preview["promo_discount"])
	assert.NotContains(t, preview, "applied_promotion")

	order := created(t, "/api/v1/orders", token, map[string]any{"items": item, "promo_code": "F5", "notes": "Catatan"})
	assertKeys(t, order, append(strings.Fields(orderKeysAlways), "cashier_name", "applied_promotion")...)
	assert.Equal(t, 300.0, order["total_amount"])
	assert.Equal(t, 15.5, order["total_discount"])
	assert.Equal(t, 284.5, order["grand_total"])
	assert.Equal(t, 284.5, order["balance_due"])
	assert.Equal(t, "F5", order["applied_promotion"].(map[string]any)["code"])
	assert.Equal(t, 5.0, order["applied_promotion"].(map[string]any)["discount_amount"])
	assert.Equal(t, []any{}, order["payments"])
	line := order["items"].([]any)[0].(map[string]any)
	assertKeys(t, line, strings.Fields("id product_id product_variant_id sku product_name variant_name variant_attributes unit_price unit_cost quantity discount_amount tax_amount subtotal cogs_amount")...)
	assert.Equal(t, 289.5, line["subtotal"])
	assert.Equal(t, 150.0, line["cogs_amount"])
	assert.Nil(t, line["variant_attributes"])
	assert.Regexp(t, `^ORD-\d{8}-0001$`, order["order_no"])

	second := created(t, "/api/v1/orders", token, map[string]any{"items": item, "offline_id": "6F9619FF-8B86-D011-B42D-00C04FC964FF"})
	assert.Regexp(t, `-0002$`, second["order_no"])
	assert.Equal(t, "6f9619ff-8b86-d011-b42d-00c04fc964ff", second["offline_id"])
	assertFails(t, doPost(t, "/api/v1/orders", map[string]any{"items": item, "offline_id": "nope"}, token), http.StatusBadRequest, "BAD_REQUEST", "invalid request body")

	// An update replaces the items and totals, without re-applying the promotion.
	path := fmt.Sprintf("/api/v1/orders/%d", idOf(order))
	updated := okData(t, doPut(t, path, map[string]any{
		"customer_id": td.customerID, "fulfillment_type": "delivery", "shipping_address": "Jl. Baru", "notes": "Ganti",
		"items": []map[string]any{{"product_variant_id": td.variantID, "quantity": 1}},
	}, token))
	assert.Equal(t, 100.0, updated["grand_total"])
	assert.Equal(t, "delivery", updated["fulfillment_type"])
	assert.Equal(t, "Jl. Baru", updated["shipping_address"])
	assert.Equal(t, "Order Test Customer", updated["customer_name"])
	assert.Len(t, updated["items"], 1)

	// Selling more than the stock is refused when the order is created.
	resp := doPost(t, "/api/v1/orders", map[string]any{"items": []map[string]any{{"product_variant_id": td.variantID, "quantity": 101}}}, token)
	assertFails(t, resp, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Stok ORD-SKU-001 tidak cukup (tersedia: 100, diminta: 101)")

	listed := decodeResponse(t, doGet(t, "/api/v1/orders?payment_status=unpaid&per_page=1", token))
	assert.Len(t, rowsOf(t, listed.Data), 1)
	assertFails(t, doGet(t, "/api/v1/orders?page=abc", token), http.StatusBadRequest, "BAD_REQUEST", "invalid query parameters")
	assertFails(t, doGet(t, "/api/v1/orders/abc", token), http.StatusBadRequest, "BAD_REQUEST", "invalid order ID")
	assertFails(t, doGet(t, "/api/v1/orders/999999", token), http.StatusNotFound, "NOT_FOUND", "Order not found not found")
	assertFails(t, doGet(t, "/api/v1/orders/999999/receipt", token), http.StatusNotFound, "NOT_FOUND", "receipt not found")
}

func TestOrderRules_AutoCompleteAndTransitions(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token
	item := []map[string]any{{"product_variant_id": td.variantID, "quantity": 2}}

	// A counter order paid in full on creation completes itself and moves stock.
	sale := created(t, "/api/v1/orders", token, map[string]any{"items": item, "payments": []map[string]any{{"method": "cash", "amount": 200, "reference_no": "R1"}}})
	assert.Equal(t, "completed", sale["status"])
	assert.Equal(t, "paid", sale["payment_status"])
	assert.NotNil(t, sale["confirmed_at"])
	assert.NotNil(t, sale["completed_at"])
	assert.NotNil(t, sale["paid_at"])
	assert.Equal(t, 0.0, sale["balance_due"])
	assert.Equal(t, 98, stockOf(t, td))
	payment := sale["payments"].([]any)[0].(map[string]any)
	assertKeys(t, payment, strings.Fields("id method amount reference_no status refunded_amount paid_at")...)
	assert.Equal(t, "R1", payment["reference_no"])

	// Auto-completing doesn't snapshot a receipt; only the complete endpoint does.
	assertFails(t, doGet(t, fmt.Sprintf("/api/v1/orders/%d/receipt", idOf(sale)), token), http.StatusNotFound, "NOT_FOUND", "receipt not found")

	path := fmt.Sprintf("/api/v1/orders/%d", idOf(sale))
	assertFails(t, doPost(t, path+"/confirm", nil, token), http.StatusBadRequest, "BAD_REQUEST", "Only draft orders can be confirmed")
	assertFails(t, doPost(t, path+"/complete", nil, token), http.StatusBadRequest, "BAD_REQUEST", "Only confirmed orders can be completed")
	assertFails(t, doPost(t, path+"/cancel", map[string]any{"reason": "x"}, token), http.StatusBadRequest, "BAD_REQUEST", "Cannot cancel this order")
	assertFails(t, doPut(t, path, map[string]any{"notes": "x"}, token), http.StatusBadRequest, "BAD_REQUEST", "Only draft orders can be updated")
	assertFails(t, doPost(t, path+"/payments", map[string]any{"payments": []map[string]any{{"method": "cash", "amount": 1}}}, token), http.StatusBadRequest, "BAD_REQUEST", "Cannot add payment to this order")

	// Refund: partial, then over the remainder, then a payment of another order.
	refunded := okData(t, doPost(t, path+"/refund", map[string]any{"payment_id": payment["id"], "amount": 50.5, "refund_reason": "Rusak"}, token))
	assert.Equal(t, 149.5, refunded["paid_amount"])
	assert.Equal(t, 50.5, refunded["refunded_total"])
	assert.Equal(t, 50.5, refunded["balance_due"])
	assert.Equal(t, "paid", refunded["payment_status"], "the order's payment status doesn't change on a refund")
	p := refunded["payments"].([]any)[0].(map[string]any)
	assert.Equal(t, "partially_refunded", p["status"])
	assert.Equal(t, "Rusak", p["refund_reason"])
	assert.NotNil(t, p["refunded_at"])
	assertFails(t, doPost(t, path+"/refund", map[string]any{"payment_id": payment["id"], "amount": 150, "refund_reason": "x"}, token), http.StatusBadRequest, "BAD_REQUEST", "Refund amount exceeds available amount")
	assertFails(t, doPost(t, path+"/refund", map[string]any{"payment_id": 999999, "amount": 1, "refund_reason": "x"}, token), http.StatusNotFound, "NOT_FOUND", "Payment not found not found")

	draft := created(t, "/api/v1/orders", token, map[string]any{"items": item})
	draftPath := fmt.Sprintf("/api/v1/orders/%d", idOf(draft))
	assertFails(t, doPost(t, draftPath+"/refund", map[string]any{"payment_id": payment["id"], "amount": 1, "refund_reason": "x"}, token), http.StatusBadRequest, "BAD_REQUEST", "Payment does not belong to this order")
	assertFails(t, doPost(t, draftPath+"/void", map[string]any{"reason": "x"}, token), http.StatusBadRequest, "BAD_REQUEST", "Cannot void this order")
	cancelled := okData(t, doPost(t, draftPath+"/cancel", map[string]any{"reason": "Batal"}, token))
	assert.Equal(t, "cancelled", cancelled["status"])
	assert.Equal(t, "Batal", cancelled["cancel_reason"])

	voided := okData(t, doPost(t, path+"/void", map[string]any{"reason": "Salah"}, token))
	assert.Equal(t, "voided", voided["status"])
	assert.Equal(t, "refunded", voided["payment_status"])
	assert.Equal(t, "Salah", voided["void_reason"])
	assert.Equal(t, 100, stockOf(t, td), "voiding a completed order restores its stock")

	// With auto-complete off a fully paid counter order stays open.
	require.Equal(t, http.StatusOK, doPut(t, "/api/v1/company-settings", map[string]any{"auto_complete_counter_orders": false}, token).StatusCode)
	kept := created(t, "/api/v1/orders", token, map[string]any{"items": item, "payments": []map[string]any{{"method": "qris", "amount": 200}}})
	assert.Equal(t, "draft", kept["status"])
	assert.Equal(t, "paid", kept["payment_status"])
	keptPath := fmt.Sprintf("/api/v1/orders/%d", idOf(kept))
	assertFails(t, doPost(t, keptPath+"/cancel", map[string]any{"reason": "x"}, token), http.StatusBadRequest, "BAD_REQUEST", "Cannot cancel a fully paid order. Use void with refund instead.")

	require.Equal(t, http.StatusOK, doPost(t, keptPath+"/confirm", nil, token).StatusCode)
	done := okData(t, doPost(t, keptPath+"/complete", map[string]any{"fulfillment_status": "delivered"}, token))
	assert.Equal(t, "completed", done["status"])
	assert.NotContains(t, done, "fulfillment_status", "only a delivery order takes a fulfillment status")
	assert.Equal(t, 98, stockOf(t, td))
	receipt := okData(t, doGet(t, keptPath+"/receipt", token))
	assertKeys(t, receipt, strings.Fields("id company_id order_id order_no cashier_id cashier_name customer_id customer_name items payments total_amount total_discount total_tax grand_total notes completed_at created_at")...)
	assert.Equal(t, 200.0, receipt["grand_total"])
	assertKeys(t, receipt["items"].([]any)[0].(map[string]any), strings.Fields("product_name variant_name sku quantity unit_price discount tax subtotal")...)
	assertKeys(t, receipt["payments"].([]any)[0].(map[string]any), "method", "amount", "reference_no")

	partial := created(t, "/api/v1/orders", token, map[string]any{"items": item, "payments": []map[string]any{{"method": "cash", "amount": 50}}})
	assert.Equal(t, "partial", partial["payment_status"])
	assert.NotContains(t, partial, "paid_at")
	require.Equal(t, http.StatusOK, doPost(t, fmt.Sprintf("/api/v1/orders/%d/confirm", idOf(partial)), nil, token).StatusCode)
	assertFails(t, doPost(t, fmt.Sprintf("/api/v1/orders/%d/complete", idOf(partial)), nil, token), http.StatusBadRequest, "BAD_REQUEST", "Order must be fully paid before completion")
}

func TestOrderRules_DeliveryAndValidation(t *testing.T) {
	td := setupOrderTest(t)
	token := td.auth.Token
	item := []map[string]any{{"product_variant_id": td.variantID, "quantity": 1}}

	assertFails(t, doPost(t, "/api/v1/orders", map[string]any{"fulfillment_type": "delivery", "items": item}, token),
		http.StatusBadRequest, "BAD_REQUEST", "Customer is required for delivery orders")
	assertFails(t, doPost(t, "/api/v1/orders", map[string]any{"items": []map[string]any{{"quantity": 1}}}, token),
		http.StatusBadRequest, "BAD_REQUEST", "Product variant is required")
	assertFieldErrors(t, doPost(t, "/api/v1/orders", map[string]any{"fulfillment_type": "x", "items": []map[string]any{{"quantity": -1, "discount_amount": -1}}}, token),
		[2]string{"fulfillment_type", "fulfillment_type must be one of: counter delivery"},
		[2]string{"quantity", "quantity must be at least 1 characters"}, [2]string{"discount_amount", "discount_amount is invalid"})
	assertFieldErrors(t, doPost(t, "/api/v1/orders/preview", map[string]any{}, token), [2]string{"items", "items is required"})

	delivery := created(t, "/api/v1/orders", token, map[string]any{"customer_id": td.customerID, "fulfillment_type": "delivery", "shipping_address": "Jl. Asia", "items": item})
	assert.Equal(t, "pending", delivery["fulfillment_status"])
	path := fmt.Sprintf("/api/v1/orders/%d", idOf(delivery))

	assertFieldErrors(t, doPut(t, path, map[string]any{"fulfillment_type": "", "items": []any{}}, token),
		[2]string{"fulfillment_type", "fulfillment_type must be one of: counter delivery"}, [2]string{"items", "items must be at least 1 characters"})
	assertFieldErrors(t, doPost(t, path+"/payments", map[string]any{}, token), [2]string{"payments", "payments is required"})
	assertFieldErrors(t, doPost(t, path+"/payments", map[string]any{"payments": []map[string]any{{"method": "gold", "amount": -1}, {}}}, token),
		[2]string{"method", "method must be one of: cash debit_card credit_card e_wallet bank_transfer qris"}, [2]string{"amount", "amount is invalid"},
		[2]string{"method", "method is required"}, [2]string{"amount", "amount is required"})
	assertFieldErrors(t, doPost(t, path+"/cancel", map[string]any{}, token), [2]string{"reason", "reason is required"})
	assertFieldErrors(t, doPost(t, path+"/void", map[string]any{}, token), [2]string{"reason", "reason is required"})
	assertFieldErrors(t, doPost(t, path+"/refund", map[string]any{"amount": -1}, token),
		[2]string{"payment_id", "payment_id is required"}, [2]string{"amount", "amount is invalid"}, [2]string{"refund_reason", "refund_reason is required"})

	require.Equal(t, http.StatusOK, doPost(t, path+"/confirm", nil, token).StatusCode)
	paid := okData(t, doPost(t, path+"/payments", map[string]any{"payments": []map[string]any{{"method": "cash", "amount": 60}, {"method": "qris", "amount": 40}}}, token))
	assert.Equal(t, "paid", paid["payment_status"])
	assert.Equal(t, "confirmed", paid["status"], "a delivery order isn't auto-completed")
	done := okData(t, doPost(t, path+"/complete", map[string]any{"fulfillment_status": "delivered"}, token))
	assert.Equal(t, "delivered", done["fulfillment_status"])

	for _, tc := range []struct {
		method, path string
		body         any
	}{
		{http.MethodGet, path, nil},
		{http.MethodPut, path, map[string]any{"notes": "x"}},
		{http.MethodPost, path + "/confirm", nil},
		{http.MethodPost, path + "/payments", map[string]any{"payments": []map[string]any{{"method": "cash", "amount": 1}}}},
		{http.MethodPost, path + "/complete", nil},
		{http.MethodPost, path + "/cancel", map[string]any{"reason": "x"}},
		{http.MethodPost, path + "/void", map[string]any{"reason": "x"}},
		{http.MethodPost, path + "/refund", map[string]any{"payment_id": 1, "amount": 1, "refund_reason": "x"}},
	} {
		other := otherCompanyToken(t)
		assertFails(t, doRequest(t, tc.method, tc.path, tc.body, other), http.StatusNotFound, "NOT_FOUND", "Order not found not found")
	}
	assertFails(t, doGet(t, path+"/receipt", otherCompanyToken(t)), http.StatusNotFound, "NOT_FOUND", "receipt not found")
}

func otherCompanyToken(t *testing.T) string {
	t.Helper()
	return registerTestUser(t).Token
}

// A customer of another company is refused before anything is written.
func TestTenancy_OrderRejectsForeignCustomer(t *testing.T) {
	td := setupOrderTest(t)
	other := registerTestUser(t)
	foreign := idOf(created(t, "/api/v1/customers", other.Token, map[string]any{"code": "F-1", "name": "Asing"}))
	item := []map[string]any{{"product_variant_id": td.variantID, "quantity": 1}}

	assertFails(t, doPost(t, "/api/v1/orders", map[string]any{"customer_id": foreign, "items": item}, td.auth.Token),
		http.StatusNotFound, "NOT_FOUND", "Customer not found not found")
	draft := created(t, "/api/v1/orders", td.auth.Token, map[string]any{"items": item})
	assertFails(t, doPut(t, fmt.Sprintf("/api/v1/orders/%d", idOf(draft)), map[string]any{"customer_id": foreign}, td.auth.Token),
		http.StatusNotFound, "NOT_FOUND", "Customer not found not found")
}

// ============== Rust against Go ==============

// sendRaw posts without normalizing anything, for comparing two servers.
func sendRaw(t *testing.T, method, url, token, body string) (int, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, raw
}

// With TARGET_BASE_URL (Rust) and COMPARE_BASE_URL (Go) both set, reads and
// refused writes of the money core answer with the same bytes on the same data.
func TestCompare_MoneyCoreIsByteIdentical(t *testing.T) {
	target, compare := os.Getenv("TARGET_BASE_URL"), os.Getenv("COMPARE_BASE_URL")
	if target == "" || compare == "" {
		t.Skip("set TARGET_BASE_URL and COMPARE_BASE_URL to compare two running servers")
	}

	td := setupOrderTest(t)
	token, branch := td.auth.Token, td.auth.BranchID
	categoryID := td.categoryID
	attributed := createProduct(t, token, map[string]any{"name": "Kaos <L>", "category_id": categoryID, "variants": []map[string]any{{
		"sku": "K-1", "name": "L & XL", "price": 150000.5, "standard_cost": 70000.25, "sale_price": 120000.75, "attributes": map[string]any{"size": "L", "n": 2.0},
	}}})
	kaos := int64(attributed["variants"].([]any)[0].(map[string]any)["id"].(float64))
	require.NoError(t, testEnv.Fixtures.CreateStock(t.Context(), kaos, branch, 40, 5))
	require.Equal(t, http.StatusOK, doPut(t, "/api/v1/company-settings", map[string]any{"tax_enabled": true, "tax_rate": 11.5, "tax_inclusive": false, "rounding_enabled": true, "rounding_amount": 50}, token).StatusCode)
	created(t, "/api/v1/promotions", token, map[string]any{"code": "PROMO", "name": "Promo <&>", "type": "discount", "discount_type": "percentage", "discount_value": 7.5, "start_at": "2020-01-01T00:00:00Z", "is_active": true})

	item := []map[string]any{{"product_variant_id": kaos, "quantity": 2, "discount_amount": 1000.25, "notes": "ignored"}, {"product_variant_id": td.variantID, "quantity": 1}}
	draft := created(t, "/api/v1/orders", token, map[string]any{"customer_id": td.customerID, "items": item, "notes": "Pesan <&>", "offline_id": "6f9619ff-8b86-d011-b42d-00c04fc964ff"})
	sale := created(t, "/api/v1/orders", token, map[string]any{"customer_id": td.customerID, "fulfillment_type": "delivery", "shipping_address": "Jl. <A>", "items": item})
	salePath := fmt.Sprintf("/api/v1/orders/%d", idOf(sale))
	total := sale["grand_total"].(float64)
	require.Equal(t, http.StatusOK, doPost(t, salePath+"/confirm", nil, token).StatusCode)
	paid := okData(t, doPost(t, salePath+"/payments", map[string]any{"payments": []map[string]any{{"method": "cash", "amount": total - 1000, "reference_no": "R<1>"}, {"method": "qris", "amount": 1000}}}, token))
	payment := paid["payments"].([]any)[0].(map[string]any)["id"]
	require.Equal(t, http.StatusOK, doPost(t, salePath+"/complete", map[string]any{"fulfillment_status": "shipped"}, token).StatusCode)
	require.Equal(t, http.StatusOK, doPost(t, salePath+"/refund", map[string]any{"payment_id": payment, "amount": 500.5, "refund_reason": "Rusak <&>"}, token).StatusCode)

	require.Equal(t, http.StatusOK, doPost(t, "/api/v1/inventory/adjust", map[string]any{"variant_id": kaos, "branch_id": branch, "type": "IN", "quantity": 3, "unit_cost": 70000.25, "note": "Masuk <&>"}, token).StatusCode)

	supplier := createSupplier(t, token, "CMP-1")
	po := created(t, "/api/v1/purchase-orders", token, map[string]any{"branch_id": branch, "supplier_id": supplier, "notes": "PO <&>", "items": []map[string]any{{"product_variant_id": kaos, "sku": "K-1", "variant_name": "L", "quantity": 5, "unit_cost": 1000.5}}})
	poPath := fmt.Sprintf("/api/v1/purchase-orders/%d", idOf(po))
	poItem := po["items"].([]any)[0].(map[string]any)["id"]
	require.Equal(t, http.StatusOK, doPost(t, poPath+"/receive", map[string]any{"items": []map[string]any{{"item_id": poItem, "received_quantity": 2}}}, token).StatusCode)

	closed := created(t, "/api/v1/shifts", token, map[string]any{"branch_id": branch, "opening_float": 1000.5})
	closedPath := fmt.Sprintf("/api/v1/shifts/%d", idOf(closed))
	require.Equal(t, http.StatusOK, doPost(t, closedPath+"/close", map[string]any{"actual_cash": 900, "notes": "Kurang <&>"}, token).StatusCode)
	created(t, "/api/v1/shifts", token, map[string]any{"branch_id": branch})

	opname := created(t, "/api/v1/stock-opnames", token, map[string]any{"branch_id": branch, "notes": "Hitung <&>"})
	opnamePath := fmt.Sprintf("/api/v1/stock-opnames/%d", idOf(opname))
	opnameItem := opname["items"].([]any)[0].(map[string]any)["id"]
	require.Equal(t, http.StatusOK, doRequest(t, http.MethodPatch, fmt.Sprintf("%s/items/%v", opnamePath, opnameItem), map[string]any{"counted_stock": 3, "notes": "Pecah"}, token).StatusCode)

	var today string
	require.NoError(t, testEnv.DB.Get(&today, `SELECT CURRENT_DATE::text`))
	settlement := created(t, "/api/v1/daily-settlements", token, map[string]any{"branch_id": branch, "settlement_date": today, "notes": "Harian <&>"})
	settlementPath := fmt.Sprintf("/api/v1/daily-settlements/%d", idOf(settlement))
	settlementItem := settlement["items"].([]any)[0].(map[string]any)["id"]
	require.Equal(t, http.StatusOK, doRequest(t, http.MethodPatch, fmt.Sprintf("%s/items/%v", settlementPath, settlementItem), map[string]any{"actual_amount": 123.45}, token).StatusCode)
	other := registerTestUser(t)

	reads := []string{
		"/api/v1/orders", "/api/v1/orders?status=draft", "/api/v1/orders?payment_status=paid&per_page=1", "/api/v1/orders?status=", "/api/v1/orders?customer_id=",
		fmt.Sprintf("/api/v1/orders?customer_id=%d&search=ORD", td.customerID), "/api/v1/orders?fulfillment_type=delivery&fulfillment_status=shipped",
		"/api/v1/orders?date_from=2000-01-01&date_to=2000-01-02", "/api/v1/orders?date_to=2999-01-01&page=2&per_page=1", "/api/v1/orders?per_page=abc", "/api/v1/orders?customer_id=x",
		salePath, fmt.Sprintf("/api/v1/orders/%d", idOf(draft)), salePath + "/receipt", fmt.Sprintf("/api/v1/orders/%d/receipt", idOf(draft)), "/api/v1/orders/abc", "/api/v1/orders/abc/receipt", "/api/v1/orders/999999",
		"/api/v1/inventory", fmt.Sprintf("/api/v1/inventory?branch_id=%d", branch), fmt.Sprintf("/api/v1/inventory?branch_id=%d&status=low&search=k", branch), "/api/v1/inventory?category=Nope&per_page=1", fmt.Sprintf("/api/v1/inventory?branch_id=%d", other.BranchID),
		"/api/v1/inventory/stats", fmt.Sprintf("/api/v1/inventory/stats?branch_id=%d", branch),
		"/api/v1/inventory/movements", fmt.Sprintf("/api/v1/inventory/movements?branch_id=%d&type=IN&search=K-1", branch), "/api/v1/inventory/movements?start_date=2000-01-01&end_date=2999-01-01&per_page=1", "/api/v1/inventory/movements/stats", fmt.Sprintf("/api/v1/inventory/movements/stats?branch_id=%d", branch),
		"/api/v1/purchase-orders", "/api/v1/purchase-orders?status=partial", fmt.Sprintf("/api/v1/purchase-orders?supplier_id=%d", supplier), poPath, "/api/v1/purchase-orders/abc", "/api/v1/purchase-orders/999999",
		"/api/v1/shifts", "/api/v1/shifts?status=open", fmt.Sprintf("/api/v1/shifts?branch_id=%d&cashier_id=%d", branch, td.auth.UserID), "/api/v1/shifts?date_from=2000-01-01&date_to=2999-01-01", "/api/v1/shifts/current", closedPath, "/api/v1/shifts/abc", "/api/v1/shifts/999999",
		"/api/v1/stock-opnames", "/api/v1/stock-opnames?status=in_progress&per_page=1", fmt.Sprintf("/api/v1/stock-opnames?branch_id=%d", branch), opnamePath, "/api/v1/stock-opnames/abc", "/api/v1/stock-opnames/999999",
		"/api/v1/daily-settlements", "/api/v1/daily-settlements?status=draft", fmt.Sprintf("/api/v1/daily-settlements?branch_id=%d&date_from=%s&date_to=%s", branch, today, today), settlementPath, "/api/v1/daily-settlements/abc", "/api/v1/daily-settlements/999999",
		fmt.Sprintf("/api/v1/daily-settlements/report?branch_id=%d&date=%s", branch, today), fmt.Sprintf("/api/v1/daily-settlements/report?branch_id=%d&date=2000-01-01", branch), "/api/v1/daily-settlements/report", "/api/v1/daily-settlements/report?branch_id=999999&date=2026-01-01",
	}
	for _, path := range reads {
		goStatus, goType, goBody := fetchRaw(t, compare+path, token)
		rustStatus, rustType, rustBody := fetchRaw(t, target+path, token)
		assert.Equal(t, goStatus, rustStatus, path)
		assert.Equal(t, goType, rustType, path)
		assert.Equal(t, string(goBody), string(rustBody), path)
	}

	line := fmt.Sprintf(`{"product_variant_id":%d,"quantity":1}`, td.variantID)
	refused := []struct{ method, path, body string }{
		// Reads that happen to be POSTs: nothing is written.
		{"POST", "/api/v1/orders/preview", fmt.Sprintf(`{"items":[%s]}`, line)},
		{"POST", "/api/v1/orders/preview", fmt.Sprintf(`{"items":[{"product_variant_id":%d,"quantity":3,"discount_amount":"12.345"}],"promo_code":"PROMO"}`, kaos)},
		{"POST", "/api/v1/orders/preview", fmt.Sprintf(`{"items":[%s],"promo_code":"NOPE"}`, line)},
		{"POST", "/api/v1/orders/preview", `{"items":[{"product_variant_id":999999,"quantity":1}]}`},
		{"POST", "/api/v1/orders/preview", `{"items":[{"quantity":1}]}`},
		{"POST", "/api/v1/orders/preview", `{"items":null}`},
		{"POST", "/api/v1/orders/preview", `{"items":[]}`},
		{"POST", "/api/v1/orders/preview", `{"items":"x"}`},
		{"POST", "/api/v1/orders/preview", `{"items":[{"quantity":"x"}]}`},
		{"POST", "/api/v1/orders/preview", `{"items":[{"quantity":1.5}]}`},
		{"POST", "/api/v1/orders/preview", `{"items":[{"quantity":1,"discount_amount":"abc"}]}`},
		{"POST", "/api/v1/orders/preview", `{bad`},
		{"POST", "/api/v1/orders", ``},
		{"POST", "/api/v1/orders", `{"fulfillment_type":"x","items":[{"quantity":0,"discount_amount":-1}]}`},
		{"POST", "/api/v1/orders", fmt.Sprintf(`{"fulfillment_type":"delivery","items":[%s]}`, line)},
		{"POST", "/api/v1/orders", fmt.Sprintf(`{"items":[{"product_variant_id":%d,"quantity":100000}]}`, td.variantID)},
		{"POST", "/api/v1/orders", fmt.Sprintf(`{"customer_id":999999,"items":[%s]}`, line)},
		{"POST", "/api/v1/orders", fmt.Sprintf(`{"offline_id":"nope","items":[%s]}`, line)},
		{"PUT", "/api/v1/orders/abc", `{}`},
		{"PUT", "/api/v1/orders/999999", `{}`},
		{"PUT", salePath, `{"fulfillment_type":"","items":[]}`},
		{"PUT", salePath, `{"notes":"x"}`},
		{"POST", salePath + "/confirm", ``},
		{"POST", "/api/v1/orders/abc/confirm", ``},
		{"POST", "/api/v1/orders/999999/confirm", ``},
		{"POST", salePath + "/payments", `{}`},
		{"POST", salePath + "/payments", `{"payments":[]}`},
		{"POST", salePath + "/payments", `{"payments":[{"method":"gold","amount":0},{"amount":-5}]}`},
		{"POST", salePath + "/payments", `{"payments":[{"method":"cash","amount":1}]}`},
		{"POST", salePath + "/complete", ``},
		{"POST", salePath + "/cancel", `{}`},
		{"POST", salePath + "/cancel", `{"reason":"x"}`},
		{"POST", salePath + "/void", `{}`},
		{"POST", salePath + "/refund", `{}`},
		{"POST", salePath + "/refund", `{"payment_id":999999,"amount":1,"refund_reason":"x"}`},
		{"POST", salePath + "/refund", fmt.Sprintf(`{"payment_id":%v,"amount":99999999,"refund_reason":"x"}`, payment)},
		{"POST", "/api/v1/inventory/adjust", `{}`},
		{"POST", "/api/v1/inventory/adjust", `{"variant_id":1,"branch_id":1,"type":"NOPE","quantity":0}`},
		{"POST", "/api/v1/inventory/adjust", fmt.Sprintf(`{"variant_id":%d,"branch_id":999999,"type":"IN","quantity":1}`, kaos)},
		{"POST", "/api/v1/inventory/adjust", fmt.Sprintf(`{"variant_id":%d,"branch_id":%d,"type":"IN","quantity":1}`, kaos, other.BranchID)},
		{"POST", "/api/v1/inventory/adjust", fmt.Sprintf(`{"variant_id":%d,"branch_id":%d,"type":"OUT","quantity":100000}`, kaos, branch)},
		{"PUT", fmt.Sprintf("/api/v1/inventory/%d/min-stock", kaos), `{"min_quantity":1}`},
		{"PUT", "/api/v1/inventory/abc/min-stock?branch_id=1", `{"min_quantity":1}`},
		{"PUT", fmt.Sprintf("/api/v1/inventory/%d/min-stock?branch_id=%d", kaos, branch), `{"min_quantity":-1}`},
		{"PUT", fmt.Sprintf("/api/v1/inventory/999999/min-stock?branch_id=%d", branch), `{"min_quantity":1}`},
		{"PUT", fmt.Sprintf("/api/v1/inventory/%d/min-stock?branch_id=%d", kaos, other.BranchID), `{"min_quantity":1}`},
		{"POST", "/api/v1/purchase-orders", `{}`},
		{"POST", "/api/v1/purchase-orders", `{"branch_id":1,"supplier_id":1,"items":[]}`},
		{"POST", "/api/v1/purchase-orders", `{"branch_id":1,"supplier_id":1,"items":[{"quantity":0,"unit_cost":-1}]}`},
		{"POST", "/api/v1/purchase-orders", `{"branch_id":1,"supplier_id":999999,"items":[{"sku":"a","variant_name":"b","quantity":1,"unit_cost":1}]}`},
		{"POST", "/api/v1/purchase-orders", fmt.Sprintf(`{"branch_id":1,"supplier_id":%d,"items":[{"product_variant_id":999999,"sku":"a","variant_name":"b","quantity":1,"unit_cost":1}]}`, supplier)},
		{"POST", poPath + "/receive", `{}`},
		{"POST", poPath + "/receive", `{"items":[{"received_quantity":-1}]}`},
		{"POST", poPath + "/receive", `{"items":[{"item_id":999999,"received_quantity":1}]}`},
		{"POST", "/api/v1/purchase-orders/abc/receive", `{}`},
		{"POST", "/api/v1/purchase-orders/999999/receive", `{"items":[{"item_id":1}]}`},
		{"POST", "/api/v1/shifts", `{}`},
		{"POST", "/api/v1/shifts", `{"branch_id":999999}`},
		{"POST", "/api/v1/shifts", fmt.Sprintf(`{"branch_id":%d}`, other.BranchID)},
		{"POST", "/api/v1/shifts", fmt.Sprintf(`{"branch_id":%d,"opening_float":-1}`, branch)},
		{"POST", "/api/v1/shifts", fmt.Sprintf(`{"branch_id":%d}`, branch)},
		{"POST", closedPath + "/close", `{"actual_cash":1}`},
		{"POST", closedPath + "/close", `{"actual_cash":-1}`},
		{"POST", "/api/v1/shifts/abc/close", `{}`},
		{"POST", "/api/v1/shifts/999999/close", `{}`},
		{"POST", "/api/v1/stock-opnames", `{}`},
		{"POST", "/api/v1/stock-opnames", `{"branch_id":999999}`},
		{"POST", "/api/v1/stock-opnames", fmt.Sprintf(`{"branch_id":%d}`, other.BranchID)},
		{"PATCH", opnamePath + "/items/abc", `{}`},
		{"PATCH", "/api/v1/stock-opnames/abc/items/1", `{}`},
		{"PATCH", opnamePath + "/items/999999", `{"counted_stock":1}`},
		{"PATCH", opnamePath + "/items/1", `{"counted_stock":-1}`},
		{"PATCH", opnamePath + "/items", `{}`},
		{"PATCH", opnamePath + "/items", `{"items":[{"counted_stock":-1}]}`},
		{"PATCH", opnamePath + "/items", `{"items":[{"item_id":999999,"counted_stock":1}]}`},
		{"POST", "/api/v1/stock-opnames/abc/complete", `{}`},
		{"POST", "/api/v1/stock-opnames/999999/complete", `{}`},
		{"POST", "/api/v1/stock-opnames/999999/cancel", ``},
		{"POST", "/api/v1/daily-settlements", `{}`},
		{"POST", "/api/v1/daily-settlements", `{"branch_id":999999,"settlement_date":"2026-01-01"}`},
		{"POST", "/api/v1/daily-settlements", fmt.Sprintf(`{"branch_id":%d,"settlement_date":"2026-1-1"}`, branch)},
		{"POST", "/api/v1/daily-settlements", fmt.Sprintf(`{"branch_id":%d,"settlement_date":"%s"}`, branch, today)},
		{"PATCH", settlementPath + "/items/abc", `{}`},
		{"PATCH", settlementPath + "/items/999999", `{"actual_amount":1}`},
		{"PATCH", settlementPath + "/items/1", `{"actual_amount":-1}`},
		{"PATCH", settlementPath + "/items", `{"items":[{"actual_amount":-1}]}`},
		{"PATCH", settlementPath + "/items", `{"items":[{"item_id":999999,"actual_amount":1}]}`},
		{"POST", "/api/v1/daily-settlements/abc/finalize", ``},
		{"POST", "/api/v1/daily-settlements/999999/finalize", ``},
		// Methods and paths around the static segments.
		{"GET", "/api/v1/orders/preview", ``},
		{"POST", "/api/v1/shifts/current", ``},
		{"DELETE", salePath, ``},
		{"POST", "/api/v1/inventory/stats", ``},
		{"GET", "/api/v1/inventory/adjust", ``},
		{"PATCH", "/api/v1/daily-settlements/report", ``},
	}
	for _, c := range refused {
		name := c.method + " " + c.path + " " + c.body
		goStatus, goBody := sendRaw(t, c.method, compare+c.path, token, c.body)
		rustStatus, rustBody := sendRaw(t, c.method, target+c.path, token, c.body)
		assert.Equal(t, goStatus, rustStatus, name)
		assert.Equal(t, string(goBody), string(rustBody), name)
	}
}
