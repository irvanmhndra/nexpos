package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A company must not be able to use another company's product variants:
// variant IDs arrive in request bodies, so every lookup is company-scoped.
func TestTenancy_ForeignVariantIsNotFound(t *testing.T) {
	cleanupDatabase(t)
	owner := registerTestUser(t)
	categoryID := createTestCategory(t, owner.Token)
	_, foreignVariant := createTestProduct(t, owner.Token, categoryID, "Owner Product", "OWN-SKU-1", 100.00, 50.00)

	other := registerTestUser(t)
	item := []map[string]any{{"product_variant_id": foreignVariant, "quantity": 1}}

	resp := doPost(t, "/api/v1/orders", map[string]any{"items": item}, other.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "order with a foreign variant: %s", string(resp.Body))

	resp = doPost(t, "/api/v1/orders/preview", map[string]any{"items": item}, other.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "preview with a foreign variant: %s", string(resp.Body))

	resp = doPost(t, "/api/v1/inventory/adjust", map[string]any{
		"variant_id": foreignVariant, "branch_id": other.BranchID, "type": "IN", "quantity": 5,
	}, other.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "stock adjust of a foreign variant: %s", string(resp.Body))

	resp = doPost(t, "/api/v1/suppliers", map[string]any{"name": "Other Supplier", "code": "SUP-OTHER"}, other.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create supplier: %s", string(resp.Body))
	var supplier map[string]any
	require.NoError(t, json.Unmarshal(decodeResponse(t, resp).Data, &supplier))
	resp = doPost(t, "/api/v1/purchase-orders", map[string]any{
		"branch_id": other.BranchID, "supplier_id": supplier["id"],
		"items": []map[string]any{{
			"product_variant_id": foreignVariant, "sku": "OWN-SKU-1", "variant_name": "Default",
			"quantity": 1, "unit_cost": 1,
		}},
	}, other.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "PO with a foreign variant: %s", string(resp.Body))

	var foreignRows int
	require.NoError(t, testEnv.DB.Get(&foreignRows,
		`SELECT count(*) FROM stocks WHERE product_variant_id = $1 AND branch_id = $2`, foreignVariant, other.BranchID))
	assert.Zero(t, foreignRows, "no stock row may be created for another company's variant")

	// The owner still uses its own variant
	resp = doPost(t, "/api/v1/orders/preview", map[string]any{"items": item}, owner.Token)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "owner preview: %s", string(resp.Body))
}

// An unknown variant ID is a 404, not a server error.
func TestTenancy_UnknownVariantIsNotFound(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	resp := doPost(t, "/api/v1/orders", map[string]any{
		"items": []map[string]any{{"product_variant_id": 999999, "quantity": 1}},
	}, auth.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "unknown variant: %s", string(resp.Body))
}

// The database itself rejects a cross-company reference, even from code that
// skips the service checks (migration 000037).
func TestTenancy_DatabaseRejectsCrossCompanyReference(t *testing.T) {
	cleanupDatabase(t)
	a := registerTestUser(t)
	b := registerTestUser(t)
	categoryA := createTestCategory(t, a.Token)
	productB, _ := createTestProduct(t, b.Token, createTestCategory(t, b.Token), "B Product", "B-SKU-1", 10.00, 5.00)

	_, err := testEnv.DB.Exec(`UPDATE products SET product_category_id = $1 WHERE id = $2`, categoryA, productB)
	var pqErr *pq.Error
	require.ErrorAs(t, err, &pqErr)
	assert.Equal(t, pq.ErrorCode("23503"), pqErr.Code)
	assert.Equal(t, "fk_products_product_category_id_same_company", pqErr.Constraint)
}
