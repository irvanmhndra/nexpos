package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// Fixtures provides test data creation utilities
type Fixtures struct {
	db *sqlx.DB
}

// NewFixtures creates a new Fixtures instance
func NewFixtures(db *sqlx.DB) *Fixtures {
	return &Fixtures{db: db}
}

// CompanyFixture represents test company data
type CompanyFixture struct {
	ID       int64
	Code     string
	Name     string
	IsActive bool
}

// BranchFixture represents test branch data
type BranchFixture struct {
	ID        int64
	CompanyID int64
	Code      string
	Name      string
	IsActive  bool
}

// RoleFixture represents test role data
type RoleFixture struct {
	ID        int64
	CompanyID int64
	Name      string
	IsOwner   bool
}

// UserFixture represents test user data
type UserFixture struct {
	ID        int64
	CompanyID int64
	RoleID    *int64
	Email     string
	Name      string
	Password  string
	Status    string
}

// CustomerFixture represents test customer data
type CustomerFixture struct {
	ID        int64
	CompanyID int64
	Code      string
	Name      string
	Phone     *string
	Email     *string
	IsMember  bool
}

// ProductCategoryFixture represents test product category data
type ProductCategoryFixture struct {
	ID        int64
	CompanyID int64
	Name      string
	IsActive  bool
}

// ProductFixture represents test product data
type ProductFixture struct {
	ID                int64
	CompanyID         int64
	ProductCategoryID *int64
	Name              string
	Description       *string
	IsActive          bool
}

// ProductVariantFixture represents test product variant data
type ProductVariantFixture struct {
	ID           int64
	ProductID    int64
	SKU          string
	Name         string
	Price        float64
	StandardCost float64
	IsDefault    bool
	IsActive     bool
}

// CreateCompany creates a test company
func (f *Fixtures) CreateCompany(ctx context.Context, name string) (*CompanyFixture, error) {
	code := fmt.Sprintf("TEST-%d", time.Now().UnixNano())
	query := `
		INSERT INTO companies (code, name, is_active, created_at, updated_at)
		VALUES ($1, $2, true, NOW(), NOW())
		RETURNING id
	`

	var id int64
	err := f.db.QueryRowContext(ctx, query, code, name).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	return &CompanyFixture{
		ID:       id,
		Code:     code,
		Name:     name,
		IsActive: true,
	}, nil
}

// CreateBranch creates a test branch
func (f *Fixtures) CreateBranch(ctx context.Context, companyID int64, name string) (*BranchFixture, error) {
	code := fmt.Sprintf("BR-%d", time.Now().UnixNano())
	query := `
		INSERT INTO branches (company_id, code, name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
		RETURNING id
	`

	var id int64
	err := f.db.QueryRowContext(ctx, query, companyID, code, name).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	return &BranchFixture{
		ID:        id,
		CompanyID: companyID,
		Code:      code,
		Name:      name,
		IsActive:  true,
	}, nil
}

// CreateRole creates a test role
func (f *Fixtures) CreateRole(ctx context.Context, companyID int64, name string, isOwner bool) (*RoleFixture, error) {
	query := `
		INSERT INTO roles (company_id, name, is_owner, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id
	`

	var id int64
	err := f.db.QueryRowContext(ctx, query, companyID, name, isOwner).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return &RoleFixture{
		ID:        id,
		CompanyID: companyID,
		Name:      name,
		IsOwner:   isOwner,
	}, nil
}

// CreateUser creates a test user
func (f *Fixtures) CreateUser(ctx context.Context, companyID int64, roleID *int64, email, name, password string) (*UserFixture, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (company_id, role_id, email, name, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'active', NOW(), NOW())
		RETURNING id
	`

	var id int64
	err = f.db.QueryRowContext(ctx, query, companyID, roleID, email, name, string(hashedPassword)).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &UserFixture{
		ID:        id,
		CompanyID: companyID,
		RoleID:    roleID,
		Email:     email,
		Name:      name,
		Password:  password,
		Status:    "active",
	}, nil
}

// CreateCustomer creates a test customer
func (f *Fixtures) CreateCustomer(ctx context.Context, companyID int64, name string) (*CustomerFixture, error) {
	code := fmt.Sprintf("CUST-%d", time.Now().UnixNano())
	query := `
		INSERT INTO customers (company_id, code, name, is_member, created_at, updated_at)
		VALUES ($1, $2, $3, false, NOW(), NOW())
		RETURNING id
	`

	var id int64
	err := f.db.QueryRowContext(ctx, query, companyID, code, name).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return &CustomerFixture{
		ID:        id,
		CompanyID: companyID,
		Code:      code,
		Name:      name,
		IsMember:  false,
	}, nil
}

// CreateProductCategory creates a test product category
func (f *Fixtures) CreateProductCategory(ctx context.Context, companyID int64, name string) (*ProductCategoryFixture, error) {
	query := `
		INSERT INTO product_categories (company_id, name, is_active, created_at, updated_at)
		VALUES ($1, $2, true, NOW(), NOW())
		RETURNING id
	`

	var id int64
	err := f.db.QueryRowContext(ctx, query, companyID, name).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create product category: %w", err)
	}

	return &ProductCategoryFixture{
		ID:        id,
		CompanyID: companyID,
		Name:      name,
		IsActive:  true,
	}, nil
}

// CreateProduct creates a test product
func (f *Fixtures) CreateProduct(ctx context.Context, companyID int64, categoryID *int64, name string) (*ProductFixture, error) {
	query := `
		INSERT INTO products (company_id, product_category_id, name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
		RETURNING id
	`

	var id int64
	err := f.db.QueryRowContext(ctx, query, companyID, categoryID, name).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &ProductFixture{
		ID:                id,
		CompanyID:         companyID,
		ProductCategoryID: categoryID,
		Name:              name,
		IsActive:          true,
	}, nil
}

// CreateProductVariant creates a test product variant
func (f *Fixtures) CreateProductVariant(ctx context.Context, productID int64, sku, name string, price, cost float64) (*ProductVariantFixture, error) {
	query := `
		INSERT INTO product_variants (product_id, sku, name, attributes, price, standard_cost, last_purchase_cost, is_default, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, '{}', $4, $5, $5, true, true, NOW(), NOW())
		RETURNING id
	`

	var id int64
	err := f.db.QueryRowContext(ctx, query, productID, sku, name, price, cost).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create product variant: %w", err)
	}

	return &ProductVariantFixture{
		ID:           id,
		ProductID:    productID,
		SKU:          sku,
		Name:         name,
		Price:        price,
		StandardCost: cost,
		IsDefault:    true,
		IsActive:     true,
	}, nil
}

// TestData holds all created test fixtures
type TestData struct {
	Company         *CompanyFixture
	Branch          *BranchFixture
	Role            *RoleFixture
	User            *UserFixture
	Customer        *CustomerFixture
	ProductCategory *ProductCategoryFixture
	Product         *ProductFixture
	ProductVariant  *ProductVariantFixture
}

// CreateBaseTestData creates a complete set of test data for integration tests
func (f *Fixtures) CreateBaseTestData(ctx context.Context) (*TestData, error) {
	// Create company
	company, err := f.CreateCompany(ctx, "Test Company")
	if err != nil {
		return nil, err
	}

	// Create branch
	branch, err := f.CreateBranch(ctx, company.ID, "Main Branch")
	if err != nil {
		return nil, err
	}

	// Create owner role
	role, err := f.CreateRole(ctx, company.ID, "Owner", true)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := f.CreateUser(ctx, company.ID, &role.ID, "test@example.com", "Test User", "password123")
	if err != nil {
		return nil, err
	}

	// Create customer
	customer, err := f.CreateCustomer(ctx, company.ID, "Test Customer")
	if err != nil {
		return nil, err
	}

	// Create product category
	category, err := f.CreateProductCategory(ctx, company.ID, "Test Category")
	if err != nil {
		return nil, err
	}

	// Create product
	product, err := f.CreateProduct(ctx, company.ID, &category.ID, "Test Product")
	if err != nil {
		return nil, err
	}

	// Create product variant
	variant, err := f.CreateProductVariant(ctx, product.ID, "SKU-001", "Default", 100.00, 50.00)
	if err != nil {
		return nil, err
	}

	return &TestData{
		Company:         company,
		Branch:          branch,
		Role:            role,
		User:            user,
		Customer:        customer,
		ProductCategory: category,
		Product:         product,
		ProductVariant:  variant,
	}, nil
}
