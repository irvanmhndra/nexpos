package storage

import "testing"

func testClient() *Client {
	return New(Config{
		AccessKeyID:   "ak",
		SecretKey:     "sk",
		Bucket:        "nexpos-prod",
		Endpoint:      "https://acct.r2.cloudflarestorage.com",
		PublicBaseURL: "https://cdn.nexpos.irvanmahendra.com/", // trailing slash trimmed by New
	})
}

func TestPublicURLKeyRoundTrip(t *testing.T) {
	c := testClient()
	key := "products/6/abc-123.webp"

	url := c.PublicURL(key)
	if want := "https://cdn.nexpos.irvanmahendra.com/products/6/abc-123.webp"; url != want {
		t.Fatalf("PublicURL = %q, want %q", url, want)
	}
	if got := c.KeyFromURL(url); got != key {
		t.Fatalf("KeyFromURL round-trip = %q, want %q", got, key)
	}
}

func TestKeyFromURLIgnoresForeignAndEmpty(t *testing.T) {
	c := testClient()
	cases := []string{
		"",
		"https://example.com/products/6/abc.webp",                 // different host
		"http://cdn.nexpos.irvanmahendra.com/products/6/abc.webp", // different scheme
		"cdn.nexpos.irvanmahendra.com/products/6/abc.webp",        // no scheme
	}
	for _, url := range cases {
		if got := c.KeyFromURL(url); got != "" {
			t.Errorf("KeyFromURL(%q) = %q, want \"\" (skip)", url, got)
		}
	}
}
