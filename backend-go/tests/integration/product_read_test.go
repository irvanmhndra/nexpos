package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var productKeys = []string{"id", "name", "category_id", "description", "image_data", "image_url", "is_active", "variants", "created_at", "updated_at"}
var variantKeys = []string{"id", "sku", "name", "attributes", "price", "standard_cost", "last_purchase_cost", "is_default", "is_active", "created_at", "updated_at"}

func createProduct(t *testing.T, token string, body map[string]any) map[string]any {
	t.Helper()
	resp := doPost(t, "/api/v1/products", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "create product: %s", resp.Body)
	return decodeObject(t, decodeResponse(t, resp).Data)
}

func productNames(t *testing.T, raw json.RawMessage) []string {
	t.Helper()
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(raw, &rows), string(raw))
	names := make([]string, len(rows))
	for i, row := range rows {
		names[i] = row["name"].(string)
	}
	return names
}

func TestProductRead_ShapeAndVariantOrder(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	category := createCategory(t, auth.Token, map[string]any{"code": "DRINK", "name": "Minuman"})
	categoryID := int64(category["id"].(float64))

	product := createProduct(t, auth.Token, map[string]any{
		"name": "Kopi", "category_id": categoryID, "description": "Enak", "image_url": "https://cdn/x.png",
		"variants": []map[string]any{
			{"sku": "K-L", "name": "Large", "price": 25000.50, "standard_cost": 10000, "last_purchase_cost": 9000.25,
				"barcode": " 8991 ", "attributes": map[string]any{"size": "L", "hot": true, "shots": 2, "ratio": 1.5, "a": []any{"x"}},
				"sale_price": 20000, "sale_start": "2026-01-02T03:04:05Z", "sale_end": "2026-02-02T03:04:05+07:00"},
			{"sku": "K-M", "name": "Medium", "price": 20000, "is_default": true},
			{"sku": "K-A", "name": "Alpha", "price": 1, "is_active": false},
		},
	})
	id := int64(product["id"].(float64))

	resp := doGet(t, fmt.Sprintf("/api/v1/products/%d", id), auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	r := decodeResponse(t, resp)
	assert.Equal(t, "Product retrieved successfully", r.Message)
	got := decodeObject(t, r.Data)
	assertKeys(t, got, append([]string{"category_name"}, productKeys...)...)
	assert.Equal(t, "Minuman", got["category_name"])
	assert.Equal(t, "Enak", got["description"])
	assert.Nil(t, got["image_data"])

	variants := got["variants"].([]any)
	require.Len(t, variants, 3)
	// Default variants first, then by name; the first variant in a request is also made default.
	names := []string{}
	for _, v := range variants {
		names = append(names, v.(map[string]any)["name"].(string))
	}
	assert.Equal(t, []string{"Large", "Medium", "Alpha"}, names)

	large := variants[0].(map[string]any)
	assertKeys(t, large, append([]string{"barcode", "sale_price", "sale_start", "sale_end"}, variantKeys...)...)
	assert.Equal(t, "8991", large["barcode"], "barcodes are trimmed")
	assert.Equal(t, 25000.5, large["price"])
	assert.Equal(t, map[string]any{"size": "L", "hot": true, "shots": float64(2), "ratio": 1.5, "a": []any{"x"}}, large["attributes"])
	assert.Equal(t, float64(20000), large["sale_price"])
	assert.Equal(t, true, large["is_default"])

	medium := variants[1].(map[string]any)
	assertKeys(t, medium, variantKeys...) // no barcode or sale fields
	assert.Nil(t, medium["attributes"])
	assert.Equal(t, false, variants[2].(map[string]any)["is_active"])

	// Bytes matter: attribute keys are sorted and numbers aren't padded.
	assert.Contains(t, string(resp.Body), `"attributes":{"a":["x"],"hot":true,"ratio":1.5,"shots":2,"size":"L"},"price":25000.5,"standard_cost":10000,"last_purchase_cost":9000.25,`)
}

func TestProductRead_ListFilters(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	category := createCategory(t, auth.Token, map[string]any{"code": "C", "name": "Cat"})
	categoryID := int64(category["id"].(float64))
	createProduct(t, auth.Token, map[string]any{"name": "Apel", "category_id": categoryID,
		"variants": []map[string]any{{"sku": "APL-1", "name": "Satu", "barcode": "111222", "price": 1}}})
	createProduct(t, auth.Token, map[string]any{"name": "Bayam", "is_active": false,
		"variants": []map[string]any{{"sku": "BYM-1", "name": "Satu", "price": 1}}})

	r := decodeResponse(t, doGet(t, "/api/v1/products", auth.Token))
	assert.Equal(t, "Product list retrieved", r.Message)
	assert.ElementsMatch(t, []string{"Apel", "Bayam"}, productNames(t, r.Data))
	assert.JSONEq(t,
		`{"pagination":{"total_records":2,"total_pages":1,"current_page":1,"per_page":20,"count":0,"next_page":null,"prev_page":null}}`,
		string(r.Meta))

	for path, want := range map[string][]string{
		"/api/v1/products?search=ape":                              {"Apel"},
		"/api/v1/products?search=%20ape%20":                        {"Apel"},
		"/api/v1/products?search=BYM":                              {"Bayam"},
		"/api/v1/products?search=1222":                             {"Apel"},
		"/api/v1/products?search=nothing":                          {},
		fmt.Sprintf("/api/v1/products?category_id=%d", categoryID): {"Apel"},
		"/api/v1/products?category_id=999999":                      {},
		"/api/v1/products?is_active=false":                         {"Bayam"},
		"/api/v1/products?is_active=true":                          {"Apel"},
	} {
		assert.Equal(t, want, productNames(t, decodeResponse(t, doGet(t, path, auth.Token)).Data), path)
	}
	assert.ElementsMatch(t, []string{"Apel", "Bayam"},
		productNames(t, decodeResponse(t, doGet(t, "/api/v1/products?category_id=abc&is_active=maybe", auth.Token)).Data))

	resp := doGet(t, "/api/v1/products?per_page=1&page=2", auth.Token)
	r = decodeResponse(t, resp)
	assert.Len(t, productNames(t, r.Data), 1)
	assert.JSONEq(t,
		`{"pagination":{"total_records":2,"total_pages":2,"current_page":2,"per_page":1,"count":0,"next_page":null,"prev_page":1}}`,
		string(r.Meta))

	other := registerTestUser(t)
	assert.JSONEq(t, `{"pagination":{"total_records":0,"total_pages":0,"current_page":1,"per_page":20,"count":0,"next_page":null,"prev_page":null}}`,
		string(decodeResponse(t, doGet(t, "/api/v1/products", other.Token)).Meta))
	assert.Contains(t, string(doGet(t, "/api/v1/products", other.Token).Body), `"data":[]`)
}

func TestProductRead_Lookup(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	product := createProduct(t, auth.Token, map[string]any{"name": "Apel",
		"variants": []map[string]any{{"sku": "APL-1", "name": "Satu", "barcode": "111222", "price": 1},
			{"sku": "APL-2", "name": "Dua", "price": 2}}})
	id := product["id"].(float64)
	variantID := func(sku string) float64 {
		for _, v := range product["variants"].([]any) {
			if v.(map[string]any)["sku"] == sku {
				return v.(map[string]any)["id"].(float64)
			}
		}
		t.Fatalf("no variant %s", sku)
		return 0
	}

	r := decodeResponse(t, doGet(t, "/api/v1/products/lookup?code=111222", auth.Token))
	assert.Equal(t, "Product found", r.Message)
	got := decodeObject(t, r.Data)
	assertKeys(t, got, "product", "matched_variant_id", "matched_by")
	assert.Equal(t, "barcode", got["matched_by"])
	assert.Equal(t, variantID("APL-1"), got["matched_variant_id"])
	assert.Equal(t, id, got["product"].(map[string]any)["id"])

	got = decodeObject(t, decodeResponse(t, doGet(t, "/api/v1/products/lookup?code=%20APL-2%20", auth.Token)).Data)
	assert.Equal(t, "sku", got["matched_by"])
	assert.Equal(t, variantID("APL-2"), got["matched_variant_id"])

	resp := doGet(t, "/api/v1/products/lookup", auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"code query parameter is required","error_code":"BAD_REQUEST"}`, string(resp.Body))
	resp = doGet(t, "/api/v1/products/lookup?code=%20%20", auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"code is required","error_code":"BAD_REQUEST"}`, string(resp.Body))
	resp = doGet(t, "/api/v1/products/lookup?code=NOPE", auth.Token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.JSONEq(t, `{"success":false,"message":"No product found for code: NOPE not found","error_code":"NOT_FOUND"}`, string(resp.Body))

	other := registerTestUser(t)
	assert.Equal(t, http.StatusNotFound, doGet(t, "/api/v1/products/lookup?code=111222", other.Token).StatusCode)
}

func TestProductRead_ErrorsTenancyAndWrites(t *testing.T) {
	cleanupDatabase(t)
	auth := registerTestUser(t)
	product := createProduct(t, auth.Token, map[string]any{"name": "Apel",
		"variants": []map[string]any{{"sku": "APL-1", "name": "Satu", "price": 1}}})
	path := fmt.Sprintf("/api/v1/products/%d", int64(product["id"].(float64)))

	resp := doGet(t, "/api/v1/products/abc", auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"invalid product ID","error_code":"BAD_REQUEST"}`, string(resp.Body))
	resp = doGet(t, "/api/v1/products/999999", auth.Token)
	assert.JSONEq(t, `{"success":false,"message":"Product not found not found","error_code":"NOT_FOUND"}`, string(resp.Body))
	assert.Equal(t, http.StatusUnauthorized, doGet(t, "/api/v1/products", "").StatusCode)

	other := registerTestUser(t)
	assert.Equal(t, http.StatusNotFound, doGet(t, path, other.Token).StatusCode)

	// Writes on the same paths keep working.
	resp = doPut(t, path, map[string]any{"name": "Apel Merah",
		"variants": []map[string]any{{"sku": "APL-1", "name": "Satu", "price": 2}}}, auth.Token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", resp.Body)
	assert.Equal(t, "Apel Merah", decodeObject(t, decodeResponse(t, doGet(t, path, auth.Token)).Data)["name"])
	assert.Equal(t, http.StatusOK, doDelete(t, path, auth.Token).StatusCode)
	assert.Equal(t, http.StatusNotFound, doGet(t, path, auth.Token).StatusCode)

	// Strings are HTML-escaped in the encoder, on every endpoint.
	resp = doPost(t, "/api/v1/customers", map[string]any{"code": "E", "name": "A&B <c> 'd' \"e\" é 日本  "}, auth.Token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "%s", resp.Body)
	bs := string(rune(92)) // a backslash
	want := `"name":"A` + bs + `u0026B ` + bs + `u003cc` + bs + `u003e 'd' ` + bs + `"e` + bs + `" é 日本 ` + bs + `u2028"`
	assert.True(t, strings.Contains(string(resp.Body), want), "%s", resp.Body)
}
