// Command seed populates a local nexpos database with a demo store — one company
// (tenant), users per role, a branch, a product catalog (shoes/apparel) with
// variants + stock, customers, promotions, suppliers and expense categories — so
// the app can be exercised end-to-end without hand-entering data.
//
// Orders/shifts/purchase-orders are intentionally NOT seeded; create those through
// the real API/UI during testing so the flows themselves get exercised.
//
// Idempotent: refuses to double-seed unless -reset, which wipes the demo company
// (including anything created through the app) and rebuilds.
//
//	go run ./cmd/seed          # seed once (skips if the demo already exists)
//	go run ./cmd/seed -reset   # wipe the demo company and reseed
//
// Reads the same POSTGRES_* env as the API (config.Load).
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/irvanmhndra/nexpos-api/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const (
	demoCompanyCode = "DEMO"
	demoPassword    = "password123"
)

func main() {
	reset := flag.Bool("reset", false, "wipe the existing demo company and reseed")
	flag.Parse()

	cfg := config.Load()
	db, err := sqlx.Connect("postgres", cfg.Postgres.DSN())
	if err != nil {
		log.Fatalf("connect DB (check POSTGRES_* env): %v", err)
	}
	defer func() { _ = db.Close() }()

	var existingID int64
	err = db.Get(&existingID, `SELECT id FROM companies WHERE code = $1`, demoCompanyCode)
	if err == nil {
		if !*reset {
			fmt.Printf("Demo company already seeded (id=%d). Use -reset to rebuild.\n", existingID)
			printCredentials()
			return
		}
		if err := wipeCompany(db, existingID); err != nil {
			log.Fatalf("reset failed: %v", err)
		}
		fmt.Printf("Wiped existing demo company (id=%d).\n", existingID)
	}

	if err := seed(db); err != nil {
		log.Fatalf("seed failed: %v", err)
	}
	fmt.Println("\n✅ Demo store seeded.")
	printCredentials()
}

type variantSpec struct {
	sku, name       string
	price, cost     int
	stock, minStock int
	barcode         string
}
type productSpec struct {
	name, category, desc string
	variants             []variantSpec
}

func seed(db *sqlx.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Company + settings
	var companyID int64
	if err := tx.Get(&companyID, `
		INSERT INTO companies (code, name, address, phone, email, tax_id, is_active)
		VALUES ($1, 'Toko Sepatu Nusantara (Demo)', 'Jl. Sudirman No. 10, Jakarta', '021-5551234', 'hello@demo.test', '01.234.567.8-901.000', true)
		RETURNING id`, demoCompanyCode,
	); err != nil {
		return fmt.Errorf("company: %w", err)
	}
	if _, err := tx.Exec(`
		INSERT INTO company_settings (company_id, tax_enabled, tax_rate, tax_inclusive, rounding_enabled, rounding_amount,
			auto_complete_counter_orders, require_customer_for_delivery, receipt_header, receipt_footer, show_tax_on_receipt, offline_mode_enabled, max_offline_days)
		VALUES ($1, true, 11, false, false, 0, true, true, 'Toko Sepatu Nusantara', 'Terima kasih telah berbelanja!', true, false, 7)`,
		companyID,
	); err != nil {
		return fmt.Errorf("company_settings: %w", err)
	}

	// System roles (company_id IS NULL, seeded by migration)
	roleID := map[string]int64{}
	for _, code := range []string{"owner", "admin", "staff"} {
		var id int64
		if err := tx.Get(&id, `SELECT id FROM roles WHERE code = $1 AND company_id IS NULL`, code); err != nil {
			return fmt.Errorf("lookup role %q: %w", code, err)
		}
		roleID[code] = id
	}

	// Users
	hash, err := bcrypt.GenerateFromPassword([]byte(demoPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash: %w", err)
	}
	users := []struct{ email, name, role string }{
		{"owner@demo.test", "Budi (Owner)", "owner"},
		{"admin@demo.test", "Sari (Admin)", "admin"},
		{"kasir@demo.test", "Andi (Kasir)", "staff"},
	}
	for _, u := range users {
		if _, err := tx.Exec(`
			INSERT INTO users (company_id, email, password_hash, name, status, role_id)
			VALUES ($1, $2, $3, $4, 'active', $5)`,
			companyID, u.email, string(hash), u.name, roleID[u.role],
		); err != nil {
			return fmt.Errorf("user %s: %w", u.email, err)
		}
	}

	// Branch + assign every user to it as default (so login carries a branch_id)
	var branchID int64
	if err := tx.Get(&branchID, `
		INSERT INTO branches (company_id, code, name, address, phone, is_active)
		VALUES ($1, 'MAIN', 'Toko Pusat', 'Jl. Sudirman No. 10, Jakarta', '021-5551234', true)
		RETURNING id`, companyID,
	); err != nil {
		return fmt.Errorf("branch: %w", err)
	}
	if _, err := tx.Exec(`
		INSERT INTO user_branches (user_id, branch_id, is_default)
		SELECT id, $1, true FROM users WHERE company_id = $2`, branchID, companyID,
	); err != nil {
		return fmt.Errorf("user_branches: %w", err)
	}

	// Categories
	catID := map[string]int64{}
	for i, name := range []string{"Sepatu", "Sandal", "Aksesoris"} {
		var id int64
		if err := tx.Get(&id, `
			INSERT INTO product_categories (company_id, code, name, sort_order, is_active)
			VALUES ($1, $2, $3, $4, true) RETURNING id`,
			companyID, fmt.Sprintf("CAT%02d", i+1), name, i,
		); err != nil {
			return fmt.Errorf("category %s: %w", name, err)
		}
		catID[name] = id
	}

	// Products + variants + stock
	products := []productSpec{
		{"Sepatu Sneakers Putih", "Sepatu", "Sneakers kasual putih", []variantSpec{
			{"SNK-W-39", "Ukuran 39", 350000, 200000, 12, 3, "8991000000391"},
			{"SNK-W-40", "Ukuran 40", 350000, 200000, 15, 3, "8991000000407"},
			{"SNK-W-41", "Ukuran 41", 350000, 200000, 8, 3, "8991000000414"},
		}},
		{"Sepatu Formal Hitam", "Sepatu", "Sepatu pantofel kulit", []variantSpec{
			{"FRM-BLK-42", "Ukuran 42", 450000, 280000, 6, 2, "8991000000421"},
		}},
		{"Sepatu Lari Pro", "Sepatu", "Sepatu running ringan", []variantSpec{
			{"RUN-40", "Ukuran 40", 500000, 300000, 10, 2, "8991000000520"},
			{"RUN-42", "Ukuran 42", 500000, 300000, 2, 3, "8991000000521"}, // low stock
		}},
		{"Sandal Jepit Santai", "Sandal", "Sandal jepit karet", []variantSpec{
			{"SDL-JPT", "Standar", 45000, 20000, 40, 10, "8991000000101"},
		}},
		{"Sandal Gunung", "Sandal", "Sandal outdoor strap", []variantSpec{
			{"SDL-GNG", "Standar", 120000, 60000, 18, 5, "8991000000102"},
		}},
		{"Kaos Kaki (3 pasang)", "Aksesoris", "Kaos kaki katun", []variantSpec{
			{"KK-01", "Standar", 25000, 8000, 60, 15, "8991000000201"},
		}},
		{"Tali Sepatu", "Aksesoris", "Tali sepatu 120cm", []variantSpec{
			{"TL-01", "Hitam", 15000, 5000, 3, 10, "8991000000202"}, // low stock
		}},
		{"Semir Sepatu", "Aksesoris", "Semir kulit netral", []variantSpec{
			{"SMR-01", "Netral", 30000, 12000, 25, 8, "8991000000203"},
		}},
	}
	for _, p := range products {
		var pid int64
		if err := tx.Get(&pid, `
			INSERT INTO products (company_id, product_category_id, name, description, is_active)
			VALUES ($1, $2, $3, $4, true) RETURNING id`,
			companyID, catID[p.category], p.name, p.desc,
		); err != nil {
			return fmt.Errorf("product %s: %w", p.name, err)
		}
		for vi, v := range p.variants {
			var vid int64
			if err := tx.Get(&vid, `
				INSERT INTO product_variants (product_id, sku, name, price, standard_cost, last_purchase_cost, is_default, is_active, barcode)
				VALUES ($1, $2, $3, $4, $5, $5, $6, true, $7) RETURNING id`,
				pid, v.sku, v.name, v.price, v.cost, vi == 0, v.barcode,
			); err != nil {
				return fmt.Errorf("variant %s: %w", v.sku, err)
			}
			if _, err := tx.Exec(`
				INSERT INTO stocks (product_variant_id, branch_id, quantity, min_quantity)
				VALUES ($1, $2, $3, $4)`, vid, branchID, v.stock, v.minStock,
			); err != nil {
				return fmt.Errorf("stock %s: %w", v.sku, err)
			}
		}
	}

	// Customers
	customers := []struct {
		code, name, phone string
		member            bool
	}{
		{"CUST001", "Andi Pratama", "08120001111", true},
		{"CUST002", "Budi Santoso", "08120002222", false},
		{"CUST003", "Citra Dewi", "08120003333", true},
	}
	for _, c := range customers {
		if _, err := tx.Exec(`
			INSERT INTO customers (company_id, code, name, phone, is_member)
			VALUES ($1, $2, $3, $4, $5)`, companyID, c.code, c.name, c.phone, c.member,
		); err != nil {
			return fmt.Errorf("customer %s: %w", c.name, err)
		}
	}

	// Promotions
	if _, err := tx.Exec(`
		INSERT INTO promotions (company_id, code, name, description, type, discount_type, discount_value, min_purchase, max_discount, start_at, end_at, priority, is_active)
		VALUES
		($1, 'DISKON20', 'Diskon 20% (min 300rb)', 'Diskon 20% maks Rp100.000', 'discount', 'percentage', 20, 300000, 100000, NOW() - INTERVAL '7 days', NOW() + INTERVAL '30 days', 1, true),
		($1, 'POTONG50K', 'Potong Rp50.000 (min 500rb)', 'Potongan tetap', 'discount', 'fixed', 50000, 500000, NULL, NOW() - INTERVAL '7 days', NOW() + INTERVAL '30 days', 2, true)`,
		companyID,
	); err != nil {
		return fmt.Errorf("promotions: %w", err)
	}

	// Suppliers
	if _, err := tx.Exec(`
		INSERT INTO suppliers (company_id, code, name, contact_name, phone, email, is_active)
		VALUES
		($1, 'SUP001', 'PT Sepatu Jaya', 'Pak Hendra', '0215559001', 'sales@sepatujaya.test', true),
		($1, 'SUP002', 'CV Aksesoris Makmur', 'Bu Ratna', '0215559002', 'order@aksesorismakmur.test', true)`,
		companyID,
	); err != nil {
		return fmt.Errorf("suppliers: %w", err)
	}

	// Expense categories
	for _, name := range []string{"Operasional", "Gaji", "Sewa"} {
		if _, err := tx.Exec(`
			INSERT INTO expense_categories (company_id, name) VALUES ($1, $2)`, companyID, name,
		); err != nil {
			return fmt.Errorf("expense category %s: %w", name, err)
		}
	}

	return tx.Commit()
}

// wipeCompany removes a company and all of its data (children first).
func wipeCompany(db *sqlx.DB, id int64) error {
	stmts := []string{
		`DELETE FROM receipts WHERE order_id IN (SELECT id FROM orders WHERE company_id=$1)`,
		`DELETE FROM payments WHERE order_id IN (SELECT id FROM orders WHERE company_id=$1)`,
		`DELETE FROM order_items WHERE order_id IN (SELECT id FROM orders WHERE company_id=$1)`,
		`DELETE FROM orders WHERE company_id=$1`,
		`DELETE FROM purchase_order_items WHERE purchase_order_id IN (SELECT id FROM purchase_orders WHERE company_id=$1)`,
		`DELETE FROM purchase_orders WHERE company_id=$1`,
		`DELETE FROM daily_settlements WHERE company_id=$1`,
		`DELETE FROM shifts WHERE company_id=$1`,
		`DELETE FROM expenses WHERE company_id=$1`,
		`DELETE FROM expense_categories WHERE company_id=$1`,
		`DELETE FROM stock_opnames WHERE company_id=$1`,
		`DELETE FROM stock_movements WHERE branch_id IN (SELECT id FROM branches WHERE company_id=$1)`,
		`DELETE FROM stocks WHERE branch_id IN (SELECT id FROM branches WHERE company_id=$1)`,
		`DELETE FROM product_variants WHERE product_id IN (SELECT id FROM products WHERE company_id=$1)`,
		`DELETE FROM products WHERE company_id=$1`,
		`DELETE FROM product_categories WHERE company_id=$1`,
		`DELETE FROM promotions WHERE company_id=$1`,
		`DELETE FROM suppliers WHERE company_id=$1`,
		`DELETE FROM customers WHERE company_id=$1`,
		`DELETE FROM company_settings WHERE company_id=$1`,
		`DELETE FROM user_sessions WHERE user_id IN (SELECT id FROM users WHERE company_id=$1)`,
		`DELETE FROM user_branches WHERE user_id IN (SELECT id FROM users WHERE company_id=$1)`,
		`DELETE FROM users WHERE company_id=$1`,
		`DELETE FROM branches WHERE company_id=$1`,
		`DELETE FROM companies WHERE id=$1`,
	}
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, s := range stmts {
		if _, err := tx.Exec(s, id); err != nil {
			return fmt.Errorf("%s: %w", s, err)
		}
	}
	return tx.Commit()
}

func printCredentials() {
	fmt.Printf("\nLogin (password for all accounts: %s):\n", demoPassword)
	fmt.Println("  Owner : owner@demo.test")
	fmt.Println("  Admin : admin@demo.test")
	fmt.Println("  Kasir : kasir@demo.test")
}
