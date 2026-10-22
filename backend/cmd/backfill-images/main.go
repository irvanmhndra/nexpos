// Command backfill-images migrates existing product images from base64
// (products.image_data) to Cloudflare R2, populating products.image_url.
//
// It is Phase 3 of the asset-storage migration (see
// docs/asset-storage-migration-plan.md). It is idempotent and resumable: it
// only touches rows where image_data is present and image_url is still NULL,
// so re-running it picks up where it left off and never re-uploads a migrated
// image. It never clears image_data — that column is dropped only in Phase 4,
// after this backfill is verified, so a rollback stays possible.
//
// Usage (run on the VPS, with the same R2_* / POSTGRES_* env as the API):
//
//	go run ./cmd/backfill-images            # migrate everything
//	go run ./cmd/backfill-images -dry-run   # report only, no upload/DB writes
//	go run ./cmd/backfill-images -limit 20  # process at most 20 (batch/test)
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/irvanmhndra/nexpos-api/config"
	"github.com/irvanmhndra/nexpos-api/internal/storage"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "decode and report what would happen, but do not upload or write to the DB")
	limit := flag.Int("limit", 0, "process at most N products (0 = all)")
	flag.Parse()

	cfg := config.Load()

	store := storage.New(storage.Config{
		AccountID:     cfg.R2.AccountID,
		AccessKeyID:   cfg.R2.AccessKeyID,
		SecretKey:     cfg.R2.SecretKey,
		Bucket:        cfg.R2.Bucket,
		Endpoint:      cfg.R2.Endpoint,
		PublicBaseURL: cfg.R2.PublicBaseURL,
	})
	if store == nil && !*dryRun {
		log.Fatal("R2 is not configured — set R2_ACCOUNT_ID / R2_ACCESS_KEY_ID / R2_SECRET_ACCESS_KEY / R2_BUCKET / R2_ENDPOINT / R2_PUBLIC_BASE_URL (or use -dry-run)")
	}

	db, err := sqlx.Connect("postgres", cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("connect DB: %v", err)
	}
	defer func() { _ = db.Close() }()

	query := `
		SELECT id, company_id, image_data
		FROM products
		WHERE image_data IS NOT NULL AND image_data <> '' AND image_url IS NULL
		ORDER BY id`
	if *limit > 0 {
		query += fmt.Sprintf("\n\t\tLIMIT %d", *limit)
	}

	var rows []struct {
		ID        int64  `db:"id"`
		CompanyID int64  `db:"company_id"`
		ImageData string `db:"image_data"`
	}
	if err := db.Select(&rows, query); err != nil {
		log.Fatalf("query products: %v", err)
	}

	log.Printf("found %d product(s) to migrate (dry-run=%v)", len(rows), *dryRun)

	var migrated, skipped, failed int
	for _, r := range rows {
		raw, ext, contentType, err := decodeImage(r.ImageData)
		if err != nil {
			log.Printf("  product %d: SKIP — %v", r.ID, err)
			skipped++
			continue
		}

		key := fmt.Sprintf("products/%d/%s%s", r.CompanyID, uuid.NewString(), ext)

		if *dryRun {
			log.Printf("  product %d: would upload %d bytes (%s) -> %s", r.ID, len(raw), contentType, key)
			migrated++
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		url, err := store.Upload(ctx, key, contentType, bytes.NewReader(raw), int64(len(raw)))
		cancel()
		if err != nil {
			log.Printf("  product %d: FAILED upload — %v", r.ID, err)
			failed++
			continue
		}

		// Guard on image_url IS NULL so a concurrent run can't double-write.
		res, err := db.Exec(
			`UPDATE products SET image_url = $1 WHERE id = $2 AND image_url IS NULL`,
			url, r.ID)
		if err != nil {
			// Uploaded but DB not updated: the object is orphaned but harmless
			// (a later run re-uploads under a new key). Report it loudly.
			log.Printf("  product %d: FAILED db update (object uploaded at %s, orphaned) — %v", r.ID, key, err)
			failed++
			continue
		}
		if n, _ := res.RowsAffected(); n == 0 {
			log.Printf("  product %d: already had image_url (raced) — object %s orphaned", r.ID, key)
			skipped++
			continue
		}

		log.Printf("  product %d: OK -> %s", r.ID, url)
		migrated++
	}

	fmt.Printf("\nDone. migrated=%d skipped=%d failed=%d (of %d)\n", migrated, skipped, failed, len(rows))
	if failed > 0 {
		os.Exit(1)
	}
}

// decodeImage turns a stored image_data value into raw image bytes plus the
// file extension and content type, verified by magic bytes (not by any
// client-supplied MIME). It accepts both a full data URI
// ("data:image/png;base64,...") and a bare base64 string.
func decodeImage(s string) (raw []byte, ext, contentType string, err error) {
	// Strip a data-URI prefix if present: data:[<mime>][;base64],<payload>
	if strings.HasPrefix(s, "data:") {
		if i := strings.IndexByte(s, ','); i >= 0 {
			s = s[i+1:]
		}
	}
	// Remove any whitespace/newlines the base64 payload may carry.
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, s)

	raw, err = base64.StdEncoding.DecodeString(s)
	if err != nil {
		if raw, err = base64.RawStdEncoding.DecodeString(s); err != nil {
			return nil, "", "", fmt.Errorf("not valid base64: %w", err)
		}
	}

	head := raw
	if len(head) > 512 {
		head = head[:512]
	}
	ext, contentType, ok := sniffImage(head)
	if !ok {
		return nil, "", "", fmt.Errorf("not a JPG/PNG/WEBP image (%d bytes)", len(raw))
	}
	return raw, ext, contentType, nil
}

// sniffImage mirrors internal/handler.sniffImage: it returns the extension
// (with leading dot) and content type from an image's magic bytes.
func sniffImage(head []byte) (ext, contentType string, ok bool) {
	switch {
	case len(head) >= 3 && head[0] == 0xFF && head[1] == 0xD8 && head[2] == 0xFF:
		return ".jpg", "image/jpeg", true
	case len(head) >= 8 && bytes.Equal(head[:8], []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}):
		return ".png", "image/png", true
	case len(head) >= 12 && string(head[0:4]) == "RIFF" && string(head[8:12]) == "WEBP":
		return ".webp", "image/webp", true
	}
	return "", "", false
}
