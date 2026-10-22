package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/irvanmhndra/nexpos-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests fire the same request from several goroutines at once. Without
// row locks and transactions, each request reads the same stock/order/payment
// state and the last write wins, losing updates or applying them twice.

const concurrency = 8

// concurrently sends one POST per body in parallel and returns the status codes.
func concurrently(t *testing.T, token string, path func(i int) string, body func(i int) any, n int) []int {
	t.Helper()
	codes := make([]int, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := range n {
		wg.Go(func() {
			<-start
			var resp *testutil.Response
			resp, errs[i] = testEnv.Server.POST(path(i), body(i), token)
			if resp != nil {
				codes[i] = resp.StatusCode
			}
		})
	}
	close(start)
	wg.Wait()
	for _, err := range errs {
		require.NoError(t, err)
	}
	return codes
}

func count(codes []int, status int) int {
	n := 0
	for _, c := range codes {
		if c == status {
			n++
		}
	}
	return n
}

// paidDeliveryOrder creates a confirmed, fully paid delivery order (delivery
// orders are not auto-completed on payment) and returns its ID, grand total,
// and payment ID.
func paidDeliveryOrder(t *testing.T, td *orderTestData, quantity int) (orderID int64, total float64, paymentID int64) {
	t.Helper()
	token := td.auth.Token
	resp := doPost(t, "/api/v1/orders", map[string]any{
		"customer_id":      td.customerID,
		"fulfillment_type": "delivery",
		"items":            []map[string]any{{"product_variant_id": td.variantID, "quantity": quantity}},
	}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create order failed: %s", string(resp.Body))
	var order map[string]any
	require.NoError(t, json.Unmarshal(decodeResponse(t, resp).Data, &order))
	orderID = int64(order["id"].(float64))
	total = order["grand_total"].(float64)

	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/confirm", orderID), nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "confirm failed: %s", string(resp.Body))

	resp = doPost(t, fmt.Sprintf("/api/v1/orders/%d/payments", orderID), map[string]any{
		"payments": []map[string]any{{"method": "cash", "amount": total}},
	}, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "payment failed: %s", string(resp.Body))
	var paid map[string]any
	require.NoError(t, json.Unmarshal(decodeResponse(t, resp).Data, &paid))
	paymentID = int64(paid["payments"].([]any)[0].(map[string]any)["id"].(float64))
	return orderID, total, paymentID
}

func stockOf(t *testing.T, td *orderTestData) int {
	t.Helper()
	var qty int
	require.NoError(t, testEnv.DB.Get(&qty,
		`SELECT quantity FROM stocks WHERE product_variant_id = $1 AND branch_id = $2`,
		td.variantID, td.auth.BranchID))
	return qty
}

func TestConcurrency_CompletingAnOrderTwiceDeductsStockOnce(t *testing.T) {
	td := setupOrderTest(t) // 100 units in stock
	orderID, _, _ := paidDeliveryOrder(t, td, 3)

	codes := concurrently(t, td.auth.Token,
		func(int) string { return fmt.Sprintf("/api/v1/orders/%d/complete", orderID) },
		func(int) any { return nil }, concurrency)

	assert.Equal(t, 1, count(codes, http.StatusOK), "exactly one completion must win: %v", codes)
	assert.Equal(t, 97, stockOf(t, td))
	var movements int
	require.NoError(t, testEnv.DB.Get(&movements,
		`SELECT count(*) FROM stock_movements WHERE reference_type = 'order' AND reference_id = $1`, orderID))
	assert.Equal(t, 1, movements)
}

func TestConcurrency_ParallelSalesDoNotLoseStockUpdates(t *testing.T) {
	td := setupOrderTest(t) // 100 units in stock
	orders := make([]int64, concurrency)
	for i := range orders {
		orders[i], _, _ = paidDeliveryOrder(t, td, 1)
	}

	codes := concurrently(t, td.auth.Token,
		func(i int) string { return fmt.Sprintf("/api/v1/orders/%d/complete", orders[i]) },
		func(int) any { return nil }, concurrency)

	assert.Equal(t, concurrency, count(codes, http.StatusOK), "every sale must complete: %v", codes)
	assert.Equal(t, 100-concurrency, stockOf(t, td), "each sale must be deducted exactly once")

	// The movement ledger forms one unbroken chain: each sale starts where the
	// previous one ended, which only holds if they were serialized.
	var chain []struct {
		Before int `db:"stock_before"`
		After  int `db:"stock_after"`
	}
	require.NoError(t, testEnv.DB.Select(&chain,
		`SELECT stock_before, stock_after FROM stock_movements
		 WHERE product_variant_id = $1 AND reference_type = 'order' ORDER BY stock_before DESC`, td.variantID))
	require.Len(t, chain, concurrency)
	for i, m := range chain {
		assert.Equal(t, 100-i, m.Before)
		assert.Equal(t, 99-i, m.After)
	}
}

func TestConcurrency_ParallelRefundsCannotExceedThePayment(t *testing.T) {
	td := setupOrderTest(t)
	orderID, total, paymentID := paidDeliveryOrder(t, td, 1)

	// Each refund is 60% of the payment: any two together would exceed it
	codes := concurrently(t, td.auth.Token,
		func(int) string { return fmt.Sprintf("/api/v1/orders/%d/refund", orderID) },
		func(int) any {
			return map[string]any{"payment_id": paymentID, "amount": total * 0.6, "refund_reason": "concurrency test"}
		}, concurrency)

	assert.Equal(t, 1, count(codes, http.StatusOK), "only one refund fits: %v", codes)
	var refunded float64
	require.NoError(t, testEnv.DB.Get(&refunded, `SELECT refunded_amount FROM payments WHERE id = $1`, paymentID))
	assert.InDelta(t, total*0.6, refunded, 0.001)
}

func TestConcurrency_ParallelStockAdjustmentsAllApply(t *testing.T) {
	td := setupOrderTest(t) // 100 units in stock

	codes := concurrently(t, td.auth.Token,
		func(int) string { return "/api/v1/inventory/adjust" },
		func(int) any {
			return map[string]any{
				"variant_id": td.variantID, "branch_id": td.auth.BranchID,
				"type": "IN", "quantity": 5, "note": "concurrency test",
			}
		}, concurrency)

	assert.Equal(t, concurrency, count(codes, http.StatusOK), "every adjustment must apply: %v", codes)
	assert.Equal(t, 100+5*concurrency, stockOf(t, td))
}

func TestConcurrency_CreateRollsBackWhenAPaymentFails(t *testing.T) {
	td := setupOrderTest(t)
	ctx := context.Background()

	var before int
	require.NoError(t, testEnv.DB.GetContext(ctx, &before, `SELECT count(*) FROM orders`))

	// The reference passes request validation but exceeds payments.reference_no
	// (VARCHAR(100)), so the payment insert fails after the order and its items
	// were written; nothing may be left behind.
	resp := doPost(t, "/api/v1/orders", map[string]any{
		"customer_id": td.customerID,
		"items":       []map[string]any{{"product_variant_id": td.variantID, "quantity": 1}},
		"payments": []map[string]any{{
			"method": "cash", "amount": 100, "reference_no": strings.Repeat("R", 150),
		}},
	}, td.auth.Token)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode, "payment insert must fail: %s", string(resp.Body))

	var after, orphanItems int
	require.NoError(t, testEnv.DB.GetContext(ctx, &after, `SELECT count(*) FROM orders`))
	require.NoError(t, testEnv.DB.GetContext(ctx, &orphanItems,
		`SELECT count(*) FROM order_items oi LEFT JOIN orders o ON o.id = oi.order_id WHERE o.id IS NULL`))
	assert.Equal(t, before, after, "a failed create must not leave a half-written order")
	assert.Zero(t, orphanItems)
}
