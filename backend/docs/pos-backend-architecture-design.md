# POS Backend Architecture Design Documentation

## Overview

This document defines the **architectural design** for a company-based, multi-branch Point of Sale (POS) backend system built with **Golang** and **Echo framework**.

The architecture follows **3-Layer Architecture** pattern — simple, pragmatic, and scalable.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Project Structure](#project-structure)
3. [Mandatory Libraries](#mandatory-libraries)
4. [Layer Definitions](#layer-definitions)
5. [Model & DTO](#model--dto)
6. [External Integrations](#external-integrations)
7. [Configuration Management](#configuration-management)
8. [Error Handling Strategy](#error-handling-strategy)
9. [Testing Strategy](#testing-strategy)
10. [Database Migration Strategy](#database-migration-strategy)
11. [Linting & Code Quality](#linting--code-quality)
12. [Step-by-Step Setup Guide](#step-by-step-setup-guide)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         HTTP Layer                               │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                 Handler (Controllers)                    │    │
│  │           Request validation, Response formatting        │    │
│  └────────────────────────┬────────────────────────────────┘    │
└───────────────────────────┼──────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Service Layer                              │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   Business Logic                         │    │
│  │         Orchestrates repositories, clients, events       │    │
│  └────────────────────────┬────────────────────────────────┘    │
└───────────────────────────┼──────────────────────────────────────┘
                            │
            ┌───────────────┼───────────────┬───────────────┐
            ▼               ▼               ▼               ▼
┌───────────────┐  ┌───────────────┐  ┌───────────┐  ┌───────────┐
│  Repository   │  │    Client     │  │   Event   │  │   Cache   │
│  (Database)   │  │ (Third-party) │  │  (Kafka)  │  │  (Redis)  │
├───────────────┤  ├───────────────┤  ├───────────┤  ├───────────┤
│ - PostgreSQL  │  │ - Payment GW  │  │ - Publish │  │ - Session │
│ - MongoDB     │  │ - SMS/Notif   │  │ - Consume │  │ - Cache   │
│               │  │ - Shipping    │  │           │  │           │
└───────────────┘  └───────────────┘  └───────────┘  └───────────┘
```

### Core Principles

1. **Simplicity** — Easy to understand, easy to maintain
2. **Separation of Concerns** — Each layer has clear responsibility
3. **Dependency Injection** — All dependencies injected via constructor
4. **Interface-based** — External dependencies behind interfaces for testability

---

## Project Structure

Actual structure of `nexpos-api`:

```
nexpos-api/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
│
├── docs/
│   ├── api_documentation.md
│   └── pos-backend-architecture-design.md
│
├── internal/
│   ├── app/                        # Application wiring (DI container)
│   │   ├── app.go                  # Setup Echo, middleware, router
│   │   ├── repositories.go         # initRepositories(db)
│   │   ├── services.go             # initServices(repos, cfg)
│   │   └── handlers.go             # initHandlers(db, services, v)
│   │
│   ├── handler/                    # HTTP handlers (one file per resource)
│   │   ├── auth.go        branch.go        customer.go
│   │   ├── product.go     product_category.go  order.go
│   │   ├── inventory.go   report.go        promotion.go
│   │   ├── company_settings.go  supplier.go  purchase_order.go
│   │   ├── shift.go       expense.go       stock_opname.go
│   │   ├── user.go        receipt.go       health.go
│   │   ├── context_helpers.go      # getCompanyID, getBranchID, getUserID
│   │   └── *_test.go               # Selected handler tests (most covered by integration)
│   │
│   ├── service/                    # Business logic
│   │   ├── interface.go            # Service interfaces + var _ assertions
│   │   ├── auth.go        branch.go        customer.go
│   │   ├── product.go     product_category.go  order.go
│   │   ├── inventory.go   report.go        promotion.go
│   │   ├── company_settings.go  purchase_order.go
│   │   ├── shift.go       expense.go       stock_opname.go
│   │   ├── receipt.go     user.go
│   │   ├── *_test.go               # Service-level unit tests (mocked repos)
│   │   └── mocks/                  # Service mocks (for handler tests)
│   │
│   ├── repository/                 # Data access
│   │   ├── interface.go            # All repository interfaces
│   │   ├── types.go                # Shared query result types
│   │   ├── postgres/               # PostgreSQL implementations
│   │   │   └── *.go                # One file per repo
│   │   ├── mongo/                  # MongoDB implementations (receipts only)
│   │   └── mocks/                  # Mockery-generated mocks
│   │
│   ├── model/                      # DB / domain structs (sqlx tags)
│   │   └── *.go                    # One file per entity
│   │
│   ├── dto/                        # Request + response payloads (flat)
│   │   └── *.go                    # One file per resource (auth.go, order.go, ...)
│   │
│   ├── middleware/
│   │   └── auth.go                 # JWT auth middleware (sets company_id, user_id, branch_id)
│   │
│   └── router/
│       └── router.go               # Route table (Echo group definitions)
│
├── pkg/                            # Reusable shared packages
│   ├── apperror/                   # Typed application errors + HTTP mapping
│   ├── httputil/                   # Response builders, pagination meta
│   └── validator/                  # go-playground/validator wrapper
│
├── migrations/                     # golang-migrate SQL files (numbered)
│   ├── 000001_create_companies_table.up.sql
│   ├── 000001_create_companies_table.down.sql
│   └── ...
│
├── config/
│   └── config.go                   # Env-based config loader
│
├── tests/                          # Integration tests
│   ├── integration/
│   │   ├── auth_test.go
│   │   ├── customer_test.go
│   │   ├── order_test.go
│   │   ├── product_test.go
│   │   └── health_test.go
│   └── testutil/                   # Testcontainers helpers (postgres, mongo)
│
├── docker-compose.yml
├── docker-compose.test.yml
├── Dockerfile
├── makefile
├── go.mod
├── go.sum
└── README.md
```

**Differences from earlier drafts of this document:**
- DTOs are **flat** (`internal/dto/*.go`) — not split into `request/` & `response/`.
- No `internal/client/` (no external payment gateways).
- No `internal/event/` (no Kafka).
- `internal/app/` is the DI container wiring repositories → services → handlers.
- `pkg/` is minimal: only `apperror`, `httputil`, `validator`.

---

## Mandatory Libraries

### Core Dependencies

Reflecting the actual `go.mod`:

```go
module github.com/irvanmhndra/nexpos-api

go 1.26.2

require (
    // Web framework
    github.com/labstack/echo/v5 v5.0.1

    // PostgreSQL
    github.com/jmoiron/sqlx v1.4.0
    github.com/lib/pq v1.10.9

    // MongoDB (optional, for receipts)
    go.mongodb.org/mongo-driver v1.17.9

    // Validation
    github.com/go-playground/validator/v10 v10.24.0

    // Utilities
    github.com/google/uuid v1.6.0
    golang.org/x/crypto v0.48.0          // bcrypt password hashing

    // Migration
    github.com/golang-migrate/migrate/v4 v4.19.1

    // Testing
    github.com/stretchr/testify v1.11.1
    github.com/testcontainers/testcontainers-go v0.42.0
    github.com/testcontainers/testcontainers-go/modules/postgres v0.42.0
    github.com/testcontainers/testcontainers-go/modules/mongodb v0.42.0
)
```

**JWT** is handled inside the `auth` service (no third-party JWT library at go.mod root; add one if needed). **Logging** uses Go's standard `log/slog`.

**Not (yet) used** — mentioned in earlier drafts but absent from current code: Redis, Kafka, viper, zerolog, sqlmock, midtrans/xendit/firebase/twilio clients. Add as real needs arise.

### Installation Commands

```bash
# Standard go module dependencies
go mod tidy

# Migration CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Mock generation
go install github.com/vektra/mockery/v2@latest

# Linter
brew install golangci-lint  # macOS
```

---

## Layer Definitions

### 1. Handler Layer (`internal/handler/`)

HTTP request/response handling. Thin layer, validates input and delegates to service.

```go
// internal/handler/order.go
package handler

import (
    "net/http"

    "github.com/google/uuid"
    "github.com/labstack/echo/v5"
    "github.com/irvanmhndra/nexpos-api/internal/dto/request"
    "github.com/irvanmhndra/nexpos-api/internal/service"
    "github.com/irvanmhndra/nexpos-api/pkg/apperror"
    "github.com/irvanmhndra/nexpos-api/pkg/httputil"
)

type OrderHandler struct {
    orderSvc *service.OrderService
}

func NewOrderHandler(orderSvc *service.OrderService) *OrderHandler {
    return &OrderHandler{orderSvc: orderSvc}
}

func (h *OrderHandler) Create(c echo.Context) error {
    var req request.CreateOrder
    if err := c.Bind(&req); err != nil {
        return httputil.Error(c, apperror.BadRequest("invalid request body"))
    }
    if err := c.Validate(&req); err != nil {
        return httputil.ValidationError(c, err)
    }

    ctx := c.Request().Context()
    companyID := getCompanyID(c)
    branchID := getBranchID(c)
    userID := getUserID(c)

    result, err := h.orderSvc.Create(ctx, companyID, branchID, userID, req)
    if err != nil {
        return httputil.Error(c, err)
    }

    return httputil.Success(c, http.StatusCreated, result)
}

func (h *OrderHandler) Get(c echo.Context) error {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        return httputil.Error(c, apperror.BadRequest("invalid order id"))
    }

    ctx := c.Request().Context()
    companyID := getCompanyID(c)

    result, err := h.orderSvc.GetByID(ctx, companyID, id)
    if err != nil {
        return httputil.Error(c, err)
    }

    return httputil.Success(c, http.StatusOK, result)
}

func (h *OrderHandler) List(c echo.Context) error {
    var req request.ListOrders
    if err := c.Bind(&req); err != nil {
        return httputil.Error(c, apperror.BadRequest("invalid query params"))
    }
    req.SetDefaults()

    ctx := c.Request().Context()
    companyID := getCompanyID(c)
    branchID := getBranchID(c)

    result, err := h.orderSvc.List(ctx, companyID, branchID, req)
    if err != nil {
        return httputil.Error(c, err)
    }

    return httputil.Success(c, http.StatusOK, result)
}

func (h *OrderHandler) Void(c echo.Context) error {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        return httputil.Error(c, apperror.BadRequest("invalid order id"))
    }

    var req request.VoidOrder
    if err := c.Bind(&req); err != nil {
        return httputil.Error(c, apperror.BadRequest("invalid request body"))
    }
    if err := c.Validate(&req); err != nil {
        return httputil.ValidationError(c, err)
    }

    ctx := c.Request().Context()
    companyID := getCompanyID(c)
    userID := getUserID(c)

    if err := h.orderSvc.Void(ctx, companyID, id, userID, req.Reason); err != nil {
        return httputil.Error(c, err)
    }

    return httputil.Success(c, http.StatusOK, map[string]string{"message": "order voided"})
}

// Context helpers
func getCompanyID(c echo.Context) int64 { return c.Get("company_id").(int64) }
func getBranchID(c echo.Context) int64  { return c.Get("branch_id").(int64) }
func getUserID(c echo.Context) int64    { return c.Get("user_id").(int64) }
```

### 2. Service Layer (`internal/service/`)

Business logic orchestration. Coordinates repositories, clients, and events.

```go
// internal/service/order.go
package service

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/irvanmhndra/nexpos-api/internal/client/payment"
    "github.com/irvanmhndra/nexpos-api/internal/dto/request"
    "github.com/irvanmhndra/nexpos-api/internal/dto/response"
    "github.com/irvanmhndra/nexpos-api/internal/event/message"
    "github.com/irvanmhndra/nexpos-api/internal/event/publisher"
    "github.com/irvanmhndra/nexpos-api/internal/model"
    "github.com/irvanmhndra/nexpos-api/internal/repository"
    "github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type OrderService struct {
    db          *sqlx.DB
    orderRepo   repository.OrderRepository
    stockRepo   repository.StockRepository
    productRepo repository.ProductRepository
    auditRepo   repository.AuditLogRepository
    paymentGw   payment.Gateway
    publisher   publisher.EventPublisher
}

func NewOrderService(
    db *sqlx.DB,
    orderRepo repository.OrderRepository,
    stockRepo repository.StockRepository,
    productRepo repository.ProductRepository,
    auditRepo repository.AuditLogRepository,
    paymentGw payment.Gateway,
    publisher publisher.EventPublisher,
) *OrderService {
    return &OrderService{
        db:          db,
        orderRepo:   orderRepo,
        stockRepo:   stockRepo,
        productRepo: productRepo,
        auditRepo:   auditRepo,
        paymentGw:   paymentGw,
        publisher:   publisher,
    }
}

func (s *OrderService) Create(ctx context.Context, companyID, branchID, cashierID int64, req request.CreateOrder) (*response.Order, error) {
    // 1. Validate stock availability
    for _, item := range req.Items {
        available, err := s.stockRepo.GetAvailable(ctx, item.ProductVariantID, branchID)
        if err != nil {
            return nil, err
        }
        if available < item.Quantity {
            return nil, apperror.InsufficientStock(item.ProductVariantID, item.Quantity, available)
        }
    }

    // 2. Calculate totals
    var totalAmount int64
    orderItems := make([]model.OrderItem, 0, len(req.Items))
    for _, item := range req.Items {
        variant, err := s.productRepo.GetVariantByID(ctx, companyID, item.ProductVariantID)
        if err != nil {
            return nil, err
        }
        if variant == nil {
            return nil, apperror.NotFound("product variant")
        }

        subtotal := variant.Price * int64(item.Quantity)
        totalAmount += subtotal

        orderItems = append(orderItems, model.OrderItem{
            ProductVariantID: item.ProductVariantID,
            ProductName:      variant.ProductName,
            VariantName:      variant.Name,
            SKU:              variant.SKU,
            Price:            variant.Price,
            Quantity:         item.Quantity,
            Subtotal:         subtotal,
            Notes:            item.Notes,
        })
    }

    // 3. Build order
    order := &model.Order{
        ID:          uuid.New(),
        CompanyID:   companyID,
        BranchID:    branchID,
        OrderNo:     generateOrderNo(branchID),
        CustomerID:  req.CustomerID,
        CashierID:   cashierID,
        Status:      model.OrderStatusPending,
        TotalAmount: totalAmount,
        GrandTotal:  totalAmount,
        Notes:       req.Notes,
        CreatedAt:   time.Now(),
        Items:       orderItems,
    }

    // 4. Start transaction
    tx, err := s.db.BeginTxx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // 5. Create order & items
    if err := s.orderRepo.CreateTx(ctx, tx, order); err != nil {
        return nil, err
    }

    // 6. Deduct stock
    for _, item := range req.Items {
        if err := s.stockRepo.DeductTx(ctx, tx, item.ProductVariantID, branchID, item.Quantity); err != nil {
            return nil, err
        }
    }

    // 7. Process payment
    var paymentRef string
    for _, p := range req.Payments {
        if p.Method != "cash" {
            chargeResp, err := s.paymentGw.Charge(ctx, payment.ChargeRequest{
                OrderID: order.ID.String(),
                Amount:  p.Amount,
                Method:  p.Method,
            })
            if err != nil {
                return nil, apperror.PaymentFailed(err.Error())
            }
            paymentRef = chargeResp.TransactionID
        }
    }
    order.PaymentRef = paymentRef
    order.Status = model.OrderStatusCompleted

    // 8. Commit transaction
    if err := tx.Commit(); err != nil {
        return nil, err
    }

    // 9. Publish event (async)
    go s.publisher.Publish(context.Background(), "order.created", message.OrderCreated{
        OrderID:   order.ID,
        CompanyID: companyID,
        BranchID:  branchID,
        Total:     order.GrandTotal,
        CreatedAt: order.CreatedAt,
    })

    // 10. Audit log (async)
    go s.auditRepo.Insert(context.Background(), &model.AuditLog{
        EntityType: "order",
        EntityID:   order.ID.String(),
        Action:     "created",
        ActorID:    cashierID,
        Data:       order,
        Timestamp:  time.Now(),
    })

    // 11. Load relations & return response DTO
    order, _ = s.orderRepo.GetByIDWithRelations(ctx, companyID, order.ID)
    return response.NewOrderResponse(order), nil
}

func (s *OrderService) GetByID(ctx context.Context, companyID int64, id uuid.UUID) (*response.Order, error) {
    order, err := s.orderRepo.GetByIDWithRelations(ctx, companyID, id)
    if err != nil {
        return nil, err
    }
    if order == nil {
        return nil, apperror.NotFound("order")
    }
    return response.NewOrderResponse(order), nil
}

func (s *OrderService) List(ctx context.Context, companyID, branchID int64, req request.ListOrders) (*response.OrderList, error) {
    params := repository.ListOrderParams{
        Page:     req.Page,
        PageSize: req.PageSize,
        Status:   req.Status,
        DateFrom: req.GetDateFrom(),
        DateTo:   req.GetDateTo(),
    }

    orders, total, err := s.orderRepo.List(ctx, companyID, branchID, params)
    if err != nil {
        return nil, err
    }

    return response.NewOrderListResponse(orders, total, req.Page, req.PageSize), nil
}

func (s *OrderService) Void(ctx context.Context, companyID int64, id uuid.UUID, userID int64, reason string) error {
    order, err := s.orderRepo.GetByIDWithRelations(ctx, companyID, id)
    if err != nil {
        return err
    }
    if order == nil {
        return apperror.NotFound("order")
    }
    if !order.CanBeVoided() {
        return apperror.BadRequest("only completed orders can be voided")
    }

    // Refund if has payment ref
    if order.PaymentRef != "" {
        if err := s.paymentGw.Refund(ctx, order.PaymentRef, order.GrandTotal); err != nil {
            return apperror.RefundFailed(err.Error())
        }
    }

    // Update status
    if err := s.orderRepo.UpdateStatus(ctx, id, model.OrderStatusVoided); err != nil {
        return err
    }

    // Restore stock
    for _, item := range order.Items {
        if err := s.stockRepo.Add(ctx, item.ProductVariantID, order.BranchID, item.Quantity); err != nil {
            return err
        }
    }

    // Audit log
    go s.auditRepo.Insert(context.Background(), &model.AuditLog{
        EntityType: "order",
        EntityID:   order.ID.String(),
        Action:     "voided",
        ActorID:    userID,
        Data:       map[string]string{"reason": reason},
        Timestamp:  time.Now(),
    })

    return nil
}

func generateOrderNo(branchID int64) string {
    return fmt.Sprintf("ORD-%d-%d", branchID, time.Now().UnixNano())
}
```

### 3. Repository Layer (`internal/repository/`)

Data access layer with interface definitions and implementations.

```go
// internal/repository/interface.go
package repository

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/irvanmhndra/nexpos-api/internal/model"
)

// ==================== User ====================

type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    GetByID(ctx context.Context, companyID, id int64) (*model.User, error)
    GetByEmail(ctx context.Context, companyID int64, email string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    List(ctx context.Context, companyID int64, params ListUserParams) ([]model.User, int64, error)
}

type ListUserParams struct {
    Page     int
    PageSize int
    Role     string
    Status   string
    Search   string
}

// ==================== Product ====================

type ProductRepository interface {
    Create(ctx context.Context, product *model.Product) error
    GetByID(ctx context.Context, companyID, id int64) (*model.Product, error)
    GetByIDWithVariants(ctx context.Context, companyID, id int64) (*model.Product, error)
    GetVariantByID(ctx context.Context, companyID, variantID int64) (*model.ProductVariant, error)
    Update(ctx context.Context, product *model.Product) error
    List(ctx context.Context, companyID int64, params ListProductParams) ([]model.Product, int64, error)
}

type ListProductParams struct {
    Page       int
    PageSize   int
    CategoryID *int64
    Search     string
    IsActive   *bool
}

// ==================== Order ====================

type OrderRepository interface {
    Create(ctx context.Context, order *model.Order) error
    CreateTx(ctx context.Context, tx *sqlx.Tx, order *model.Order) error
    GetByID(ctx context.Context, companyID int64, id uuid.UUID) (*model.Order, error)
    GetByIDWithRelations(ctx context.Context, companyID int64, id uuid.UUID) (*model.Order, error)
    GetByOrderNo(ctx context.Context, companyID, branchID int64, orderNo string) (*model.Order, error)
    List(ctx context.Context, companyID, branchID int64, params ListOrderParams) ([]model.Order, int64, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

type ListOrderParams struct {
    Page     int
    PageSize int
    Status   string
    DateFrom *time.Time
    DateTo   *time.Time
}

// ==================== Stock ====================

type StockRepository interface {
    GetAvailable(ctx context.Context, variantID, branchID int64) (int, error)
    Deduct(ctx context.Context, variantID, branchID int64, qty int) error
    DeductTx(ctx context.Context, tx *sqlx.Tx, variantID, branchID int64, qty int) error
    Add(ctx context.Context, variantID, branchID int64, qty int) error
    AddTx(ctx context.Context, tx *sqlx.Tx, variantID, branchID int64, qty int) error
    GetByBranch(ctx context.Context, companyID, branchID int64, params ListStockParams) ([]model.Stock, int64, error)
}

type ListStockParams struct {
    Page       int
    PageSize   int
    CategoryID *int64
    LowStock   bool
    Search     string
}

// ==================== Audit Log (MongoDB) ====================

type AuditLogRepository interface {
    Insert(ctx context.Context, log *model.AuditLog) error
    FindByEntity(ctx context.Context, entityType, entityID string) ([]model.AuditLog, error)
    FindByActor(ctx context.Context, actorID int64, limit int) ([]model.AuditLog, error)
}
```

```go
// internal/repository/postgres/order.go
package postgres

import (
    "context"
    "database/sql"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/irvanmhndra/nexpos-api/internal/model"
    "github.com/irvanmhndra/nexpos-api/internal/repository"
)

type orderRepository struct {
    db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) repository.OrderRepository {
    return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
    return r.createOrder(ctx, r.db, order)
}

func (r *orderRepository) CreateTx(ctx context.Context, tx *sqlx.Tx, order *model.Order) error {
    return r.createOrder(ctx, tx, order)
}

func (r *orderRepository) createOrder(ctx context.Context, db sqlx.ExtContext, order *model.Order) error {
    query := `
        INSERT INTO orders (
            id, company_id, branch_id, order_no, customer_id, cashier_id,
            status, total_amount, total_discount, total_tax, grand_total,
            payment_ref, notes, created_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
        )
    `
    _, err := db.ExecContext(ctx, query,
        order.ID, order.CompanyID, order.BranchID, order.OrderNo, order.CustomerID,
        order.CashierID, order.Status, order.TotalAmount, order.TotalDiscount,
        order.TotalTax, order.GrandTotal, order.PaymentRef, order.Notes, order.CreatedAt,
    )
    if err != nil {
        return err
    }

    // Insert order items
    for i := range order.Items {
        order.Items[i].OrderID = order.ID
        if err := r.createOrderItem(ctx, db, &order.Items[i]); err != nil {
            return err
        }
    }

    return nil
}

func (r *orderRepository) createOrderItem(ctx context.Context, db sqlx.ExtContext, item *model.OrderItem) error {
    query := `
        INSERT INTO order_items (
            order_id, product_variant_id, product_name, variant_name,
            sku, price, quantity, subtotal, notes
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id
    `
    return db.QueryRowxContext(ctx, query,
        item.OrderID, item.ProductVariantID, item.ProductName, item.VariantName,
        item.SKU, item.Price, item.Quantity, item.Subtotal, item.Notes,
    ).Scan(&item.ID)
}

func (r *orderRepository) GetByID(ctx context.Context, companyID int64, id uuid.UUID) (*model.Order, error) {
    query := `SELECT * FROM orders WHERE id = $1 AND company_id = $2`

    var order model.Order
    if err := r.db.GetContext(ctx, &order, query, id, companyID); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &order, nil
}

func (r *orderRepository) GetByIDWithRelations(ctx context.Context, companyID int64, id uuid.UUID) (*model.Order, error) {
    order, err := r.GetByID(ctx, companyID, id)
    if err != nil || order == nil {
        return order, err
    }

    // Load items
    itemsQuery := `SELECT * FROM order_items WHERE order_id = $1`
    if err := r.db.SelectContext(ctx, &order.Items, itemsQuery, id); err != nil {
        return nil, err
    }

    // Load cashier
    cashierQuery := `SELECT id, name, email FROM users WHERE id = $1`
    order.Cashier = &model.User{}
    r.db.GetContext(ctx, order.Cashier, cashierQuery, order.CashierID)

    // Load customer if exists
    if order.CustomerID != nil {
        customerQuery := `SELECT id, name, phone, email FROM customers WHERE id = $1`
        order.Customer = &model.Customer{}
        r.db.GetContext(ctx, order.Customer, customerQuery, *order.CustomerID)
    }

    return order, nil
}

func (r *orderRepository) List(ctx context.Context, companyID, branchID int64, params repository.ListOrderParams) ([]model.Order, int64, error) {
    var orders []model.Order
    var total int64

    // Count query
    countQuery := `
        SELECT COUNT(*) FROM orders
        WHERE company_id = $1 AND branch_id = $2
    `
    args := []any{companyID, branchID}
    argIndex := 3

    if params.Status != "" {
        countQuery += fmt.Sprintf(" AND status = $%d", argIndex)
        args = append(args, params.Status)
        argIndex++
    }
    if params.DateFrom != nil {
        countQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
        args = append(args, params.DateFrom)
        argIndex++
    }
    if params.DateTo != nil {
        countQuery += fmt.Sprintf(" AND created_at <= $%d", argIndex)
        args = append(args, params.DateTo)
        argIndex++
    }

    if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
        return nil, 0, err
    }

    // List query
    listQuery := `
        SELECT o.*, u.name as cashier_name
        FROM orders o
        LEFT JOIN users u ON o.cashier_id = u.id
        WHERE o.company_id = $1 AND o.branch_id = $2
    `
    // Add same filters...
    listQuery += " ORDER BY o.created_at DESC"
    listQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)

    offset := (params.Page - 1) * params.PageSize
    args = append(args, params.PageSize, offset)

    if err := r.db.SelectContext(ctx, &orders, listQuery, args...); err != nil {
        return nil, 0, err
    }

    return orders, total, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
    query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
    _, err := r.db.ExecContext(ctx, query, status, id)
    return err
}
```

---

## Model & DTO

### Model (`internal/model/`)

Database/domain models. Internal representation.

```go
// internal/model/order.go
package model

import (
    "time"
    "github.com/google/uuid"
)

const (
    OrderStatusPending   = "pending"
    OrderStatusCompleted = "completed"
    OrderStatusVoided    = "voided"
)

type Order struct {
    ID            uuid.UUID  `db:"id"`
    CompanyID     int64      `db:"company_id"`
    BranchID      int64      `db:"branch_id"`
    OrderNo       string     `db:"order_no"`
    CustomerID    *int64     `db:"customer_id"`
    CashierID     int64      `db:"cashier_id"`
    Status        string     `db:"status"`
    TotalAmount   int64      `db:"total_amount"`
    TotalDiscount int64      `db:"total_discount"`
    TotalTax      int64      `db:"total_tax"`
    GrandTotal    int64      `db:"grand_total"`
    PaymentRef    string     `db:"payment_ref"`
    Notes         string     `db:"notes"`
    CreatedAt     time.Time  `db:"created_at"`
    UpdatedAt     *time.Time `db:"updated_at"`

    // Relations (loaded when needed)
    Items    []OrderItem `db:"-"`
    Payments []Payment   `db:"-"`
    Customer *Customer   `db:"-"`
    Cashier  *User       `db:"-"`
}

// Domain logic
func (o *Order) CanBeVoided() bool {
    return o.Status == OrderStatusCompleted
}

func (o *Order) CalculateGrandTotal() int64 {
    return o.TotalAmount - o.TotalDiscount + o.TotalTax
}

type OrderItem struct {
    ID               int64     `db:"id"`
    OrderID          uuid.UUID `db:"order_id"`
    ProductVariantID int64     `db:"product_variant_id"`
    ProductName      string    `db:"product_name"`
    VariantName      string    `db:"variant_name"`
    SKU              string    `db:"sku"`
    Price            int64     `db:"price"`
    Quantity         int       `db:"quantity"`
    Subtotal         int64     `db:"subtotal"`
    Notes            string    `db:"notes"`
}
```

```go
// internal/model/user.go
package model

import "time"

const (
    UserStatusActive   = "active"
    UserStatusInactive = "inactive"

    RoleAdmin   = "admin"
    RoleManager = "manager"
    RoleCashier = "cashier"
)

type User struct {
    ID           int64      `db:"id"`
    CompanyID    int64      `db:"company_id"`
    Email        string     `db:"email"`
    PasswordHash string     `db:"password_hash"`
    Name         string     `db:"name"`
    Role         string     `db:"role"`
    Status       string     `db:"status"`
    LastLoginAt  *time.Time `db:"last_login_at"`
    CreatedAt    time.Time  `db:"created_at"`
    UpdatedAt    *time.Time `db:"updated_at"`
}

func (u *User) IsActive() bool {
    return u.Status == UserStatusActive
}

func (u *User) IsAdmin() bool {
    return u.Role == RoleAdmin
}
```

### Request DTO (`internal/dto/request/`)

API input contracts with validation.

```go
// internal/dto/request/order.go
package request

import "time"

type CreateOrder struct {
    CustomerID *int64            `json:"customer_id,omitempty"`
    Items      []CreateOrderItem `json:"items" validate:"required,min=1,dive"`
    Payments   []CreatePayment   `json:"payments" validate:"required,min=1,dive"`
    Notes      string            `json:"notes,omitempty" validate:"max=500"`
}

type CreateOrderItem struct {
    ProductVariantID int64  `json:"product_variant_id" validate:"required"`
    Quantity         int    `json:"quantity" validate:"required,gt=0"`
    Notes            string `json:"notes,omitempty" validate:"max=200"`
}

type CreatePayment struct {
    Method string `json:"method" validate:"required,oneof=cash card qris transfer"`
    Amount int64  `json:"amount" validate:"required,gt=0"`
}

type ListOrders struct {
    Page     int    `query:"page"`
    PageSize int    `query:"page_size"`
    Status   string `query:"status" validate:"omitempty,oneof=pending completed voided"`
    DateFrom string `query:"date_from" validate:"omitempty,datetime=2006-01-02"`
    DateTo   string `query:"date_to" validate:"omitempty,datetime=2006-01-02"`
}

func (r *ListOrders) SetDefaults() {
    if r.Page < 1 {
        r.Page = 1
    }
    if r.PageSize < 1 || r.PageSize > 100 {
        r.PageSize = 20
    }
}

func (r *ListOrders) GetDateFrom() *time.Time {
    if r.DateFrom == "" {
        return nil
    }
    t, _ := time.Parse("2006-01-02", r.DateFrom)
    return &t
}

func (r *ListOrders) GetDateTo() *time.Time {
    if r.DateTo == "" {
        return nil
    }
    t, _ := time.Parse("2006-01-02", r.DateTo)
    // End of day
    t = t.Add(24*time.Hour - time.Second)
    return &t
}

type VoidOrder struct {
    Reason string `json:"reason" validate:"required,min=10,max=500"`
}
```

```go
// internal/dto/request/product.go
package request

type CreateProduct struct {
    CategoryID  int64                  `json:"category_id" validate:"required"`
    Name        string                 `json:"name" validate:"required,min=2,max=200"`
    SKU         string                 `json:"sku" validate:"required,alphanum,min=3,max=50"`
    Description string                 `json:"description,omitempty" validate:"max=1000"`
    Variants    []CreateProductVariant `json:"variants" validate:"required,min=1,dive"`
}

type CreateProductVariant struct {
    Name  string `json:"name" validate:"required,max=100"`
    SKU   string `json:"sku" validate:"required,alphanum,max=50"`
    Price int64  `json:"price" validate:"required,gt=0"`
    Cost  int64  `json:"cost" validate:"gte=0"`
}

type UpdateProduct struct {
    CategoryID  *int64  `json:"category_id,omitempty"`
    Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=200"`
    Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
    IsActive    *bool   `json:"is_active,omitempty"`
}

type ListProducts struct {
    Page       int    `query:"page"`
    PageSize   int    `query:"page_size"`
    CategoryID *int64 `query:"category_id"`
    Search     string `query:"search"`
    IsActive   *bool  `query:"is_active"`
}

func (r *ListProducts) SetDefaults() {
    if r.Page < 1 {
        r.Page = 1
    }
    if r.PageSize < 1 || r.PageSize > 100 {
        r.PageSize = 20
    }
}
```

### Response DTO (`internal/dto/response/`)

API output contracts with conversion constructors.

```go
// internal/dto/response/common.go
package response

type Pagination struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalItems int64 `json:"total_items"`
    TotalPages int64 `json:"total_pages"`
}

func NewPagination(page, pageSize int, totalItems int64) Pagination {
    totalPages := totalItems / int64(pageSize)
    if totalItems%int64(pageSize) > 0 {
        totalPages++
    }
    return Pagination{
        Page:       page,
        PageSize:   pageSize,
        TotalItems: totalItems,
        TotalPages: totalPages,
    }
}

type IDResponse struct {
    ID string `json:"id"`
}

type MessageResponse struct {
    Message string `json:"message"`
}
```

```go
// internal/dto/response/order.go
package response

import (
    "time"
    "github.com/irvanmhndra/nexpos-api/internal/model"
)

type Order struct {
    ID            string        `json:"id"`
    OrderNo       string        `json:"order_no"`
    Status        string        `json:"status"`
    TotalAmount   int64         `json:"total_amount"`
    TotalDiscount int64         `json:"total_discount"`
    TotalTax      int64         `json:"total_tax"`
    GrandTotal    int64         `json:"grand_total"`
    Notes         string        `json:"notes,omitempty"`
    Items         []OrderItem   `json:"items"`
    Customer      *CustomerBrief `json:"customer,omitempty"`
    Cashier       UserBrief     `json:"cashier"`
    CreatedAt     time.Time     `json:"created_at"`
}

type OrderItem struct {
    ID          int64  `json:"id"`
    ProductName string `json:"product_name"`
    VariantName string `json:"variant_name"`
    SKU         string `json:"sku"`
    Price       int64  `json:"price"`
    Quantity    int    `json:"quantity"`
    Subtotal    int64  `json:"subtotal"`
    Notes       string `json:"notes,omitempty"`
}

type CustomerBrief struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Phone string `json:"phone,omitempty"`
}

type UserBrief struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

// Constructor: model -> response
func NewOrderResponse(o *model.Order) *Order {
    resp := &Order{
        ID:            o.ID.String(),
        OrderNo:       o.OrderNo,
        Status:        o.Status,
        TotalAmount:   o.TotalAmount,
        TotalDiscount: o.TotalDiscount,
        TotalTax:      o.TotalTax,
        GrandTotal:    o.GrandTotal,
        Notes:         o.Notes,
        Items:         make([]OrderItem, 0, len(o.Items)),
        CreatedAt:     o.CreatedAt,
    }

    for _, item := range o.Items {
        resp.Items = append(resp.Items, OrderItem{
            ID:          item.ID,
            ProductName: item.ProductName,
            VariantName: item.VariantName,
            SKU:         item.SKU,
            Price:       item.Price,
            Quantity:    item.Quantity,
            Subtotal:    item.Subtotal,
            Notes:       item.Notes,
        })
    }

    if o.Customer != nil {
        resp.Customer = &CustomerBrief{
            ID:    o.Customer.ID,
            Name:  o.Customer.Name,
            Phone: o.Customer.Phone,
        }
    }

    if o.Cashier != nil {
        resp.Cashier = UserBrief{
            ID:   o.Cashier.ID,
            Name: o.Cashier.Name,
        }
    }

    return resp
}

// List response
type OrderList struct {
    Orders []OrderSummary `json:"orders"`
    Meta   Pagination     `json:"meta"`
}

type OrderSummary struct {
    ID         string    `json:"id"`
    OrderNo    string    `json:"order_no"`
    Status     string    `json:"status"`
    GrandTotal int64     `json:"grand_total"`
    ItemCount  int       `json:"item_count"`
    Cashier    string    `json:"cashier"`
    CreatedAt  time.Time `json:"created_at"`
}

func NewOrderListResponse(orders []model.Order, total int64, page, pageSize int) *OrderList {
    summaries := make([]OrderSummary, 0, len(orders))

    for _, o := range orders {
        summary := OrderSummary{
            ID:         o.ID.String(),
            OrderNo:    o.OrderNo,
            Status:     o.Status,
            GrandTotal: o.GrandTotal,
            ItemCount:  len(o.Items),
            CreatedAt:  o.CreatedAt,
        }
        if o.Cashier != nil {
            summary.Cashier = o.Cashier.Name
        }
        summaries = append(summaries, summary)
    }

    return &OrderList{
        Orders: summaries,
        Meta:   NewPagination(page, pageSize, total),
    }
}
```

```go
// internal/dto/response/user.go
package response

import (
    "time"
    "github.com/irvanmhndra/nexpos-api/internal/model"
)

type User struct {
    ID        int64      `json:"id"`
    Email     string     `json:"email"`
    Name      string     `json:"name"`
    Role      string     `json:"role"`
    Status    string     `json:"status"`
    LastLogin *time.Time `json:"last_login,omitempty"`
    CreatedAt time.Time  `json:"created_at"`
}

func NewUserResponse(u *model.User) *User {
    return &User{
        ID:        u.ID,
        Email:     u.Email,
        Name:      u.Name,
        Role:      u.Role,
        Status:    u.Status,
        LastLogin: u.LastLoginAt,
        CreatedAt: u.CreatedAt,
    }
}

type UserList struct {
    Users []User     `json:"users"`
    Meta  Pagination `json:"meta"`
}

func NewUserListResponse(users []model.User, total int64, page, pageSize int) *UserList {
    list := make([]User, 0, len(users))
    for _, u := range users {
        list = append(list, *NewUserResponse(&u))
    }
    return &UserList{
        Users: list,
        Meta:  NewPagination(page, pageSize, total),
    }
}
```

---

## External Integrations

> **Current status:** Nexpos does not integrate any external payment gateway or message broker. This section is kept as guidance **if** these are added later.

### Pattern: Client Layer (`internal/client/`)

If third-party integrations are needed (e.g. Midtrans, Xendit, Firebase, Twilio):

```go
// internal/client/payment/interface.go
package payment

import "context"

type Gateway interface {
    Charge(ctx context.Context, req ChargeRequest) (*ChargeResponse, error)
    CheckStatus(ctx context.Context, transactionID string) (*StatusResponse, error)
    Refund(ctx context.Context, transactionID string, amount int64) error
}
```

Services that need a gateway receive it via constructor injection; the concrete implementation (Midtrans, Xendit, mock) is swapped in `internal/app/services.go`. Goal: services stay testable via a mock gateway without network calls.

### Pattern: Event Layer (`internal/event/`)

If async eventing is needed (e.g. async stock reconcile, push notifications):

```go
type EventPublisher interface {
    Publish(ctx context.Context, topic string, message any) error
    Close() error
}
```

Services publish events at the end of an operation (post-commit). Initial implementation can use `kafka-go` or `nats.go`; start with an in-memory publisher for tests.

**Nexpos currently needs neither** — completed orders are handled synchronously inside the service, and receipt persistence to MongoDB is enough for audit. Add these only when a real need emerges.

---

## Configuration Management

Configuration is loaded from environment variables (no third-party `.env` reader — just `os.Getenv`). See `config/config.go`.

```go
// config/config.go
package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    Server   ServerConfig
    Postgres PostgresConfig
    JWT      JWTConfig
    Mongo    MongoConfig
}

type ServerConfig struct {
    Port string
    Env  string
}

type PostgresConfig struct {
    Host        string
    Port        string
    User        string
    Password    string
    DB          string
    SSLMode     string
    DSNOverride string // used by integration tests
}

func (p PostgresConfig) DSN() string {
    if p.DSNOverride != "" {
        return p.DSNOverride
    }
    return "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + p.Port +
        "/" + p.DB + "?sslmode=" + p.SSLMode
}

type MongoConfig struct {
    URI      string
    Database string
}

type JWTConfig struct {
    Secret             string
    AccessExpiresHours int
    RefreshExpiresDays int
    AccessTokenExpiry  time.Duration
    RefreshTokenExpiry time.Duration
}

func Load() *Config {
    accessHours := getEnvInt("JWT_ACCESS_EXPIRES_HOURS", 2)
    refreshDays := getEnvInt("JWT_REFRESH_EXPIRES_DAYS", 7)
    return &Config{
        Server: ServerConfig{
            Port: getEnv("SERVER_PORT", "8080"),
            Env:  getEnv("SERVER_ENV", "development"),
        },
        Postgres: PostgresConfig{
            Host:     getEnv("POSTGRES_HOST", "localhost"),
            Port:     getEnv("POSTGRES_PORT", "5432"),
            User:     getEnv("POSTGRES_USER", "pos_user"),
            Password: getEnv("POSTGRES_PASSWORD", ""),
            DB:       getEnv("POSTGRES_DB", "pos_db"),
            SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
        },
        Mongo: MongoConfig{
            URI:      getEnv("MONGO_URI", ""),
            Database: getEnv("MONGO_DB", "nexpos"),
        },
        JWT: JWTConfig{
            Secret:             getEnv("JWT_SECRET", "secret"),
            AccessExpiresHours: accessHours,
            RefreshExpiresDays: refreshDays,
            AccessTokenExpiry:  time.Duration(accessHours) * time.Hour,
            RefreshTokenExpiry: time.Duration(refreshDays) * 24 * time.Hour,
        },
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}
```

**MongoDB is optional**: if `MONGO_URI` is empty, the MongoDB client is not created and `GET /orders/:id/receipt` is not registered. The receipt service becomes `nil` and the order service continues to run without receipt persistence.

---

## Error Handling Strategy

```go
// pkg/apperror/error.go
package apperror

import "net/http"

type AppError struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    HTTPStatus int    `json:"-"`
    Details    any    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
    return e.Message
}

// Common errors
func BadRequest(message string) *AppError {
    return &AppError{Code: "BAD_REQUEST", Message: message, HTTPStatus: http.StatusBadRequest}
}

func NotFound(resource string) *AppError {
    return &AppError{Code: "NOT_FOUND", Message: resource + " not found", HTTPStatus: http.StatusNotFound}
}

func Unauthorized(message string) *AppError {
    return &AppError{Code: "UNAUTHORIZED", Message: message, HTTPStatus: http.StatusUnauthorized}
}

func Forbidden(message string) *AppError {
    return &AppError{Code: "FORBIDDEN", Message: message, HTTPStatus: http.StatusForbidden}
}

func Conflict(message string) *AppError {
    return &AppError{Code: "CONFLICT", Message: message, HTTPStatus: http.StatusConflict}
}

func InternalError(err error) *AppError {
    return &AppError{Code: "INTERNAL_ERROR", Message: "internal server error", HTTPStatus: http.StatusInternalServerError}
}

// Domain-specific errors
func InsufficientStock(variantID int64, requested, available int) *AppError {
    return &AppError{
        Code:       "INSUFFICIENT_STOCK",
        Message:    "insufficient stock",
        HTTPStatus: http.StatusUnprocessableEntity,
        Details:    map[string]any{"variant_id": variantID, "requested": requested, "available": available},
    }
}

func PaymentFailed(reason string) *AppError {
    return &AppError{Code: "PAYMENT_FAILED", Message: "payment failed: " + reason, HTTPStatus: http.StatusPaymentRequired}
}

func RefundFailed(reason string) *AppError {
    return &AppError{Code: "REFUND_FAILED", Message: "refund failed: " + reason, HTTPStatus: http.StatusUnprocessableEntity}
}
```

```go
// pkg/httputil/response.go
package httputil

import (
    "github.com/labstack/echo/v5"
    "github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type Response struct {
    Success bool `json:"success"`
    Data    any  `json:"data,omitempty"`
    Error   any  `json:"error,omitempty"`
}

func Success(c echo.Context, status int, data any) error {
    return c.JSON(status, Response{Success: true, Data: data})
}

func Error(c echo.Context, err error) error {
    if appErr, ok := err.(*apperror.AppError); ok {
        return c.JSON(appErr.HTTPStatus, Response{
            Success: false,
            Error:   appErr,
        })
    }
    return c.JSON(500, Response{
        Success: false,
        Error:   apperror.InternalError(err),
    })
}

func ValidationError(c echo.Context, err error) error {
    return c.JSON(400, Response{
        Success: false,
        Error: map[string]any{
            "code":    "VALIDATION_ERROR",
            "message": "validation failed",
            "details": err.Error(),
        },
    })
}
```

---

## Testing Strategy

### Test Directory Structure

```
nexpos-api/
├── internal/
│   ├── handler/
│   │   ├── order.go
│   │   └── order_test.go              # Unit test (same package)
│   │
│   ├── service/
│   │   ├── order.go
│   │   └── order_test.go              # Unit test (same package)
│   │
│   ├── repository/
│   │   ├── postgres/                  # Tidak ada unit test di sini
│   │   │   ├── order.go               # (Postgres repo dicover via integration tests)
│   │   │   └── ...
│   │   │
│   │   └── mocks/                     # Mockery-generated repository mocks
│   │       ├── order.go
│   │       ├── stock.go
│   │       ├── product.go
│   │       └── ...
│   │
│   └── dto/
│       └── *.go                       # Flat DTOs (auth.go, order.go, ...)
│
├── tests/                             # Integration tests
│   ├── integration/
│   │   ├── auth_test.go
│   │   ├── customer_test.go
│   │   ├── order_test.go
│   │   ├── product_test.go
│   │   └── health_test.go
│   │
│   └── testutil/                      # Test helpers
│       ├── db.go                      # Postgres testcontainer setup
│       ├── mongo.go                   # Mongo testcontainer setup
│       └── http.go                    # Test HTTP client + JWT helpers
│
└── makefile
```

### Unit vs Integration Tests

| Aspect | Unit Test | Integration Test |
|--------|-----------|------------------|
| **Location** | `internal/**/*_test.go` | `tests/integration/` |
| **Dependencies** | Mocked | Real (testcontainers) |
| **Speed** | Fast (milliseconds) | Slower (seconds) |
| **Scope** | Single function/method | Full HTTP request flow |
| **Database** | Mocked via repository interface | Real PostgreSQL container |
| **Run** | `make test-unit` | `make test-integration` |

### Mock Generation

Install mockery:

```bash
go install github.com/vektra/mockery/v2@latest
```

Create config file:

```yaml
# .mockery.yaml
with-expecter: true
packages:
  github.com/irvanmhndra/nexpos-api/internal/repository:
    interfaces:
      UserRepository:
      CompanyRepository:
      BranchRepository:
      UserSessionRepository:
      RoleRepository:
      PermissionRepository:
      RolePermissionRepository:
      UserBranchRepository:
      CustomerRepository:
      ProductCategoryRepository:
      ProductRepository:
      ProductVariantRepository:
      OrderRepository:
      OrderItemRepository:
      PaymentRepository:
      CompanySettingsRepository:
      PromotionRepository:
      ReportRepository:
      StockRepository:
      StockMovementRepository:
      SupplierRepository:
      PurchaseOrderRepository:
      ShiftRepository:
      ExpenseCategoryRepository:
      ExpenseRepository:
      StockOpnameRepository:
    config:
      dir: internal/repository/mocks
      outpkg: mocks
```

> Tidak ada `internal/client/` atau `internal/event/` yet. Tambah package + entri mockery saat butuh.

Generate mocks:

```bash
mockery
```

### Unit Test Examples

#### Service Layer Test

Real example from `internal/service/stock_opname_test.go` — pattern: mock all repository dependencies, call service method, assert behavior.

```go
package service

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    "github.com/irvanmhndra/nexpos-api/internal/dto"
    "github.com/irvanmhndra/nexpos-api/internal/model"
    repoMocks "github.com/irvanmhndra/nexpos-api/internal/repository/mocks"
)

type opnameTestSetup struct {
    svc          *StockOpnameService
    opnameRepo   *repoMocks.MockStockOpnameRepository
    stockRepo    *repoMocks.MockStockRepository
    movementRepo *repoMocks.MockStockMovementRepository
    branchRepo   *repoMocks.MockBranchRepository
}

func setupOpnameTest(t *testing.T) *opnameTestSetup {
    t.Helper()
    s := &opnameTestSetup{
        opnameRepo:   repoMocks.NewMockStockOpnameRepository(t),
        stockRepo:    repoMocks.NewMockStockRepository(t),
        movementRepo: repoMocks.NewMockStockMovementRepository(t),
        branchRepo:   repoMocks.NewMockBranchRepository(t),
    }
    s.svc = NewStockOpnameService(s.opnameRepo, s.stockRepo, s.movementRepo, s.branchRepo)
    return s
}

func TestStockOpnameService_Create_Success(t *testing.T) {
    s := setupOpnameTest(t)
    ctx := context.Background()
    companyID, userID, branchID := int64(1), int64(7), int64(2)

    s.branchRepo.EXPECT().GetByID(ctx, branchID).
        Return(&model.Branch{ID: branchID, CompanyID: companyID}, nil).Once()
    s.opnameRepo.EXPECT().GenerateOpnameNumber(ctx, companyID).
        Return("OPN-20260514-0001", nil).Once()
    s.opnameRepo.EXPECT().Create(ctx, mock.MatchedBy(func(op *model.StockOpname) bool {
        return op.Status == model.StockOpnameStatusInProgress
    })).Run(func(args mock.Arguments) {
        args.Get(1).(*model.StockOpname).ID = 100
    }).Return(nil).Once()
    s.opnameRepo.EXPECT().SnapshotItems(ctx, int64(100), companyID, branchID, (*int64)(nil)).
        Return(5, nil).Once()
    s.opnameRepo.EXPECT().GetByID(ctx, companyID, int64(100)).
        Return(&model.StockOpname{ID: 100, Status: model.StockOpnameStatusInProgress}, nil).Once()
    s.opnameRepo.EXPECT().GetItems(ctx, int64(100)).Return([]*model.StockOpnameItem{}, nil).Once()
    s.opnameRepo.EXPECT().GetItemStats(ctx, int64(100)).Return(5, 0, nil).Once()

    resp, err := s.svc.Create(ctx, companyID, userID, dto.CreateStockOpnameRequest{BranchID: branchID})

    require.NoError(t, err)
    assert.Equal(t, int64(100), resp.ID)
    assert.Equal(t, 5, resp.TotalItems)
}
```

Conventions:
- Tests live in the same package as the code under test (no `_test` suffix on package).
- Use mockery-generated mocks with `.EXPECT()` API for type-safe expectations.
- Helper `setupXxxTest(t)` per service to centralize wiring.
- One assertion focus per test; cover happy + failure paths separately.

#### Handler Layer Test

Most handler coverage is delivered by integration tests (real HTTP request flow with a testcontainer Postgres). Selected handlers that need fine-grained unit tests (e.g. auth) use mocked services:

```go
// internal/handler/auth_test.go (excerpt)
package handler

import (
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/labstack/echo/v5"
    "github.com/stretchr/testify/require"

    serviceMock "github.com/irvanmhndra/nexpos-api/internal/service/mocks"
)

func TestAuthHandler_Login_ValidationErrors(t *testing.T) {
    mockSvc := serviceMock.NewMockAuthServiceInterface(t)
    h := NewAuthHandler(mockSvc, testValidator())

    req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{}`))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    c := echo.New().NewContext(req, rec)

    require.NoError(t, h.Login(c))
    require.Equal(t, 422, rec.Code)
}
```

Pattern: same-package tests, mocked service interface, assert HTTP status + response body fields.

#### Repository Layer

Postgres repository implementations are **not** unit-tested with sqlmock. Instead they are covered end-to-end via `tests/integration/*` with a real Postgres testcontainer. Rationale: the SQL is the contract here, and mocking it just tests that strings match, not behaviour. Integration tests give real confidence at small additional cost (containers cached after first run).

### Integration Test Examples

#### Test App Setup

```go
// tests/integration/helper/app.go
package helper

import (
    "context"
    "testing"

    "github.com/jmoiron/sqlx"
    "github.com/labstack/echo/v5"
    _ "github.com/lib/pq"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"

    "github.com/irvanmhndra/nexpos-api/config"
    "github.com/irvanmhndra/nexpos-api/internal/handler"
    "github.com/irvanmhndra/nexpos-api/internal/repository/postgres"
    "github.com/irvanmhndra/nexpos-api/internal/router"
    "github.com/irvanmhndra/nexpos-api/internal/service"
)

type TestApp struct {
    t         *testing.T
    container testcontainers.Container
    DB        *sqlx.DB
    Echo      *echo.Echo
    Config    *config.Config
}

func NewTestApp(t *testing.T) *TestApp {
    ctx := context.Background()

    // Start PostgreSQL container
    container, err := postgres.Run(ctx, "postgres:16-alpine",
        postgres.WithDatabase("pos_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    if err != nil {
        t.Fatalf("failed to start postgres container: %v", err)
    }

    // Get connection string
    connStr, err := container.ConnectionString(ctx, "sslmode=disable")
    if err != nil {
        t.Fatalf("failed to get connection string: %v", err)
    }

    // Connect to database
    db, err := sqlx.Connect("postgres", connStr)
    if err != nil {
        t.Fatalf("failed to connect to database: %v", err)
    }

    // Run migrations
    runMigrations(t, db)

    // Seed base data (company, branch)
    seedBaseData(db)

    // Setup application
    app := &TestApp{
        t:         t,
        container: container,
        DB:        db,
        Config:    loadTestConfig(),
    }

    app.Echo = app.setupEcho()

    return app
}

func (a *TestApp) setupEcho() *echo.Echo {
    e := echo.New()

    // Initialize repositories
    userRepo := pgRepo.NewUserRepository(a.DB)
    productRepo := pgRepo.NewProductRepository(a.DB)
    orderRepo := pgRepo.NewOrderRepository(a.DB)
    stockRepo := pgRepo.NewStockRepository(a.DB)

    // Initialize services (with nil for external dependencies in tests)
    authSvc := service.NewAuthService(userRepo, a.Config.JWT)
    productSvc := service.NewProductService(productRepo)
    orderSvc := service.NewOrderService(a.DB, orderRepo, stockRepo, productRepo, nil, nil, &noopPublisher{})

    // Initialize handlers
    handlers := router.Handlers{
        Auth:    handler.NewAuthHandler(authSvc),
        Product: handler.NewProductHandler(productSvc),
        Order:   handler.NewOrderHandler(orderSvc),
    }

    // Setup routes
    mw := middleware.NewMiddleware(a.Config.JWT)
    router.Setup(e, handlers, mw)

    return e
}

func (a *TestApp) Cleanup() {
    a.DB.Close()
    a.container.Terminate(context.Background())
}

func (a *TestApp) CleanTables(tables ...string) {
    for _, table := range tables {
        a.DB.Exec("TRUNCATE TABLE " + table + " CASCADE")
    }
}

// Noop publisher for tests
type noopPublisher struct{}

func (p *noopPublisher) Publish(ctx context.Context, topic string, message any) error {
    return nil
}

func (p *noopPublisher) Close() error {
    return nil
}
```

#### Database Helpers

```go
// tests/integration/helper/database.go
package helper

import (
    "testing"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    "github.com/jmoiron/sqlx"
)

func runMigrations(t *testing.T, db *sqlx.DB) {
    driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
    if err != nil {
        t.Fatalf("failed to create migrate driver: %v", err)
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://../../migrations",
        "postgres",
        driver,
    )
    if err != nil {
        t.Fatalf("failed to create migrate instance: %v", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        t.Fatalf("failed to run migrations: %v", err)
    }
}

func seedBaseData(db *sqlx.DB) {
    // Create test company
    db.Exec(`
        INSERT INTO companies (id, name, code, is_active)
        VALUES (1, 'Test Company', 'TEST', true)
        ON CONFLICT (id) DO NOTHING
    `)

    // Create test branch
    db.Exec(`
        INSERT INTO branches (id, company_id, name, code, is_active)
        VALUES (1, 1, 'Test Branch', 'BR01', true)
        ON CONFLICT (id) DO NOTHING
    `)

    // Create test category
    db.Exec(`
        INSERT INTO product_categories (id, company_id, name)
        VALUES (1, 1, 'Default Category')
        ON CONFLICT (id) DO NOTHING
    `)
}
```

#### Fixtures

```go
// tests/integration/helper/fixtures.go
package helper

import (
    "github.com/google/uuid"
    "github.com/irvanmhndra/nexpos-api/internal/model"
)

func (a *TestApp) SeedProduct(id int64, name string, price int64) {
    a.DB.Exec(`
        INSERT INTO products (id, company_id, category_id, name, sku, is_active, created_at)
        VALUES ($1, 1, 1, $2, $3, true, NOW())
        ON CONFLICT (id) DO NOTHING
    `, id, name, "SKU-"+name)

    a.DB.Exec(`
        INSERT INTO product_variants (id, product_id, name, sku, price, cost, created_at)
        VALUES ($1, $1, 'Default', $2, $3, 0, NOW())
        ON CONFLICT (id) DO NOTHING
    `, id, "VAR-"+name, price)
}

func (a *TestApp) SeedStock(variantID, branchID int64, qty int) {
    a.DB.Exec(`
        INSERT INTO stocks (product_variant_id, branch_id, quantity, reserved)
        VALUES ($1, $2, $3, 0)
        ON CONFLICT (product_variant_id, branch_id)
        DO UPDATE SET quantity = $3
    `, variantID, branchID, qty)
}

func (a *TestApp) GetOrderByID(id string) *model.Order {
    var order model.Order
    a.DB.Get(&order, "SELECT * FROM orders WHERE id = $1", id)
    return &order
}

func (a *TestApp) GetStock(variantID, branchID int64) int {
    var qty int
    a.DB.Get(&qty, `
        SELECT quantity FROM stocks
        WHERE product_variant_id = $1 AND branch_id = $2
    `, variantID, branchID)
    return qty
}

func (a *TestApp) CountOrders() int {
    var count int
    a.DB.Get(&count, "SELECT COUNT(*) FROM orders WHERE company_id = 1")
    return count
}
```

#### Auth Helpers

```go
// tests/integration/helper/auth.go
package helper

import (
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

func (a *TestApp) CreateUser(email, password, role string) int64 {
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

    var id int64
    a.DB.QueryRow(`
        INSERT INTO users (company_id, email, password_hash, name, role, status, created_at)
        VALUES (1, $1, $2, $3, $4, 'active', NOW())
        RETURNING id
    `, email, string(hash), "Test User", role).Scan(&id)

    // Assign to branch
    a.DB.Exec(`
        INSERT INTO user_branches (user_id, branch_id)
        VALUES ($1, 1)
        ON CONFLICT DO NOTHING
    `, id)

    return id
}

func (a *TestApp) CreateUserAndGetToken(email, password, role string) string {
    userID := a.CreateUser(email, password, role)
    return a.GenerateToken(userID, 1, 1, role)
}

func (a *TestApp) GenerateToken(userID, companyID, branchID int64, role string) string {
    claims := jwt.MapClaims{
        "user_id":    userID,
        "company_id": companyID,
        "branch_id":  branchID,
        "role":       role,
        "exp":        time.Now().Add(time.Hour).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte(a.Config.JWT.Secret))

    return tokenString
}
```

#### Integration Test Suite

```go
// tests/integration/order_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/suite"

    "github.com/irvanmhndra/nexpos-api/tests/integration/helper"
)

type OrderTestSuite struct {
    suite.Suite
    app   *helper.TestApp
    token string
}

func TestOrderSuite(t *testing.T) {
    suite.Run(t, new(OrderTestSuite))
}

func (s *OrderTestSuite) SetupSuite() {
    s.app = helper.NewTestApp(s.T())
    s.token = s.app.CreateUserAndGetToken("cashier@test.com", "password123", "cashier")
}

func (s *OrderTestSuite) TearDownSuite() {
    s.app.Cleanup()
}

func (s *OrderTestSuite) SetupTest() {
    s.app.CleanTables("orders", "order_items", "stocks", "products", "product_variants")
}

func (s *OrderTestSuite) TestCreateOrder_Success() {
    // Seed
    s.app.SeedProduct(1, "Product A", 10000)
    s.app.SeedStock(1, 1, 100)

    // Request
    body, _ := json.Marshal(map[string]any{
        "items": []map[string]any{
            {"product_variant_id": 1, "quantity": 2},
        },
        "payments": []map[string]any{
            {"method": "cash", "amount": 20000},
        },
    })

    req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+s.token)

    rec := httptest.NewRecorder()
    s.app.Echo.ServeHTTP(rec, req)

    // Assert response
    s.Equal(http.StatusCreated, rec.Code)

    var resp map[string]any
    json.Unmarshal(rec.Body.Bytes(), &resp)
    s.True(resp["success"].(bool))

    data := resp["data"].(map[string]any)
    s.NotEmpty(data["id"])
    s.NotEmpty(data["order_no"])
    s.Equal(float64(20000), data["grand_total"])

    // Assert database
    orderID := data["id"].(string)
    order := s.app.GetOrderByID(orderID)
    s.Equal("completed", order.Status)
    s.Equal(int64(20000), order.GrandTotal)

    // Assert stock deducted
    stock := s.app.GetStock(1, 1)
    s.Equal(98, stock) // 100 - 2
}

func (s *OrderTestSuite) TestCreateOrder_InsufficientStock() {
    s.app.SeedProduct(1, "Product A", 10000)
    s.app.SeedStock(1, 1, 1) // Only 1 in stock

    body, _ := json.Marshal(map[string]any{
        "items": []map[string]any{
            {"product_variant_id": 1, "quantity": 10},
        },
        "payments": []map[string]any{
            {"method": "cash", "amount": 100000},
        },
    })

    req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+s.token)

    rec := httptest.NewRecorder()
    s.app.Echo.ServeHTTP(rec, req)

    // Assert
    s.Equal(http.StatusUnprocessableEntity, rec.Code)

    var resp map[string]any
    json.Unmarshal(rec.Body.Bytes(), &resp)
    s.False(resp["success"].(bool))
    s.Equal("INSUFFICIENT_STOCK", resp["error"].(map[string]any)["code"])

    // Assert no order created
    s.Equal(0, s.app.CountOrders())

    // Assert stock unchanged
    stock := s.app.GetStock(1, 1)
    s.Equal(1, stock)
}

func (s *OrderTestSuite) TestCreateOrder_Unauthorized() {
    body, _ := json.Marshal(map[string]any{
        "items": []map[string]any{
            {"product_variant_id": 1, "quantity": 1},
        },
    })

    req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    // No Authorization header

    rec := httptest.NewRecorder()
    s.app.Echo.ServeHTTP(rec, req)

    s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *OrderTestSuite) TestCreateOrder_ValidationError() {
    body, _ := json.Marshal(map[string]any{
        "items": []map[string]any{}, // Empty items
    })

    req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+s.token)

    rec := httptest.NewRecorder()
    s.app.Echo.ServeHTTP(rec, req)

    s.Equal(http.StatusBadRequest, rec.Code)

    var resp map[string]any
    json.Unmarshal(rec.Body.Bytes(), &resp)
    s.Equal("VALIDATION_ERROR", resp["error"].(map[string]any)["code"])
}

func (s *OrderTestSuite) TestGetOrder_Success() {
    // Create order first
    s.app.SeedProduct(1, "Product A", 10000)
    s.app.SeedStock(1, 1, 100)

    createBody, _ := json.Marshal(map[string]any{
        "items":    []map[string]any{{"product_variant_id": 1, "quantity": 2}},
        "payments": []map[string]any{{"method": "cash", "amount": 20000}},
    })

    createReq := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(createBody))
    createReq.Header.Set("Content-Type", "application/json")
    createReq.Header.Set("Authorization", "Bearer "+s.token)

    createRec := httptest.NewRecorder()
    s.app.Echo.ServeHTTP(createRec, createReq)

    var createResp map[string]any
    json.Unmarshal(createRec.Body.Bytes(), &createResp)
    orderID := createResp["data"].(map[string]any)["id"].(string)

    // Get order
    getReq := httptest.NewRequest(http.MethodGet, "/api/v1/orders/"+orderID, nil)
    getReq.Header.Set("Authorization", "Bearer "+s.token)

    getRec := httptest.NewRecorder()
    s.app.Echo.ServeHTTP(getRec, getReq)

    s.Equal(http.StatusOK, getRec.Code)

    var getResp map[string]any
    json.Unmarshal(getRec.Body.Bytes(), &getResp)
    s.True(getResp["success"].(bool))
    s.Equal(orderID, getResp["data"].(map[string]any)["id"])
}

func (s *OrderTestSuite) TestListOrders_Success() {
    // Create multiple orders
    s.app.SeedProduct(1, "Product A", 10000)
    s.app.SeedStock(1, 1, 100)

    for i := 0; i < 3; i++ {
        body, _ := json.Marshal(map[string]any{
            "items":    []map[string]any{{"product_variant_id": 1, "quantity": 1}},
            "payments": []map[string]any{{"method": "cash", "amount": 10000}},
        })

        req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", "Bearer "+s.token)

        rec := httptest.NewRecorder()
        s.app.Echo.ServeHTTP(rec, req)
    }

    // List orders
    req := httptest.NewRequest(http.MethodGet, "/api/v1/orders?page=1&page_size=10", nil)
    req.Header.Set("Authorization", "Bearer "+s.token)

    rec := httptest.NewRecorder()
    s.app.Echo.ServeHTTP(rec, req)

    s.Equal(http.StatusOK, rec.Code)

    var resp map[string]any
    json.Unmarshal(rec.Body.Bytes(), &resp)
    s.True(resp["success"].(bool))

    data := resp["data"].(map[string]any)
    orders := data["orders"].([]any)
    s.Len(orders, 3)

    meta := data["meta"].(map[string]any)
    s.Equal(float64(3), meta["total_items"])
}
```

### Makefile Test Commands

```makefile
# Testing
.PHONY: test test-unit test-integration test-coverage mocks

test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	go test -v -race -short ./internal/...

test-integration:
	@echo "Running integration tests..."
	go test -v -race ./tests/integration/...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-coverage-func:
	@echo "Coverage by function..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Mocks
mocks:
	@echo "Generating mocks..."
	mockery

mocks-clean:
	@echo "Cleaning mocks..."
	rm -rf internal/repository/mocks
	rm -rf internal/client/mocks
	rm -rf internal/event/mocks
	rm -rf internal/service/mocks
```

### CI/CD Test Configuration

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install dependencies
        run: go mod download

      - name: Run linter
        uses: golangci/golangci-lint-action@v4
        with:
          version: v1.62.2

      - name: Run unit tests
        run: go test -v -race -short ./internal/...

      - name: Run integration tests
        run: go test -v -race ./tests/integration/...

      - name: Generate coverage report
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
```

---

## Database Migration Strategy

### Commands

```makefile
MIGRATE=migrate -path migrations -database "$(DATABASE_URL)"

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-version:
	$(MIGRATE) version
```

---

## Linting & Code Quality

```yaml
# .golangci.yml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - bodyclose
    - contextcheck
    - errname
    - errorlint
    - goconst
    - gocritic
    - gofmt
    - goimports
    - gosec
    - misspell
    - noctx
    - prealloc
    - revive
    - sqlclosecheck
    - unconvert

linters-settings:
  goconst:
    min-len: 3
    min-occurrences: 3

  revive:
    rules:
      - name: var-naming
        arguments:
          - ["ID", "URL", "API", "HTTP", "JSON", "SQL", "UUID"]

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - goconst
```

---

## Step-by-Step Setup Guide

### 1. Initialize Project

```bash
mkdir nexpos-api && cd nexpos-api
go mod init github.com/irvanmhndra/nexpos-api

# Create directory structure
mkdir -p cmd/api
mkdir -p docs
mkdir -p internal/{app,handler,service/mocks,repository/{postgres,mongo,mocks},model,dto,middleware,router}
mkdir -p pkg/{apperror,httputil,validator}
mkdir -p migrations config tests/{integration,testutil}
```

### 2. Install Dependencies

```bash
# Core web + DB
go get github.com/labstack/echo/v5
go get github.com/jmoiron/sqlx github.com/lib/pq

# MongoDB (optional — for receipts persistence)
go get go.mongodb.org/mongo-driver/mongo

# Utils
go get github.com/go-playground/validator/v10
go get github.com/google/uuid
go get golang.org/x/crypto

# Testing
go get github.com/stretchr/testify
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
go get github.com/testcontainers/testcontainers-go/modules/mongodb

# Migration CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Mock generator
go install github.com/vektra/mockery/v2@latest

# Hot reload (optional)
go install github.com/air-verse/air@latest

# Linter
brew install golangci-lint  # macOS

go mod tidy
```

### 3. Create Environment File

```bash
cat > .env.example << 'EOF'
SERVER_PORT=8080
SERVER_ENV=development

POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=pos_user
POSTGRES_PASSWORD=pos_password
POSTGRES_DB=pos_db
POSTGRES_SSLMODE=disable

# Optional — set to enable receipt persistence
MONGO_URI=
MONGO_DB=nexpos

JWT_SECRET=replace-me
JWT_ACCESS_EXPIRES_HOURS=2
JWT_REFRESH_EXPIRES_DAYS=7
EOF

cp .env.example .env
```

### 4. Docker Compose

The real `docker-compose.yml` ships Postgres + Mongo + the API service:

```yaml
services:
  postgres:
    image: postgres:18-alpine
    container_name: nexpos_db
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    ports:
      - "5432:5432"
    volumes:
      - nexpos_pgdata:/var/lib/postgresql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 5s
      timeout: 5s
      retries: 5

  mongo:
    image: mongo:7
    container_name: nexpos_mongo
    environment:
      MONGO_INITDB_ROOT_USERNAME: ${MONGO_USER}
      MONGO_INITDB_ROOT_PASSWORD: ${MONGO_PASSWORD}
      MONGO_INITDB_DATABASE: ${MONGO_DB}
    ports:
      - "27017:27017"
    volumes:
      - nexpos_mongodata:/data/db

  nexpos-api:
    build: .
    image: nexpos-api:latest
    container_name: nexpos_api
    env_file: .env
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      mongo:
        condition: service_healthy
    restart: always

volumes:
  nexpos_pgdata:
  nexpos_mongodata:
```

No Redis, Kafka, or external gateway containers are required by current code.

### 5. Makefile

```makefile
.PHONY: all build run test lint mocks

include .env
export

DATABASE_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)
MIGRATE=migrate -path migrations -database "$(DATABASE_URL)"

# ==================== Build & Run ====================

build:
	go build -o bin/api cmd/api/main.go

run:
	go run cmd/api/main.go

dev:
	air

# ==================== Docker ====================

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# ==================== Database ====================

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-version:
	$(MIGRATE) version

migrate-force:
	$(MIGRATE) force $(version)

# ==================== Testing ====================

test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	go test -v -race -short ./internal/...

test-integration:
	@echo "Running integration tests..."
	go test -v -race ./tests/integration/...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-func:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# ==================== Mocks ====================

mocks:
	@echo "Generating mocks..."
	mockery

mocks-clean:
	rm -rf internal/repository/mocks
	rm -rf internal/client/mocks
	rm -rf internal/event/mocks

# ==================== Linting ====================

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

# ==================== Clean ====================

clean:
	rm -rf bin coverage.out coverage.html

# ==================== Help ====================

help:
	@echo "Available commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build              Build the application"
	@echo "  make run                Run the application"
	@echo "  make dev                Run with hot reload (requires air)"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up          Start all containers"
	@echo "  make docker-down        Stop all containers"
	@echo "  make docker-logs        View container logs"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up         Run all migrations"
	@echo "  make migrate-down       Rollback last migration"
	@echo "  make migrate-create name=xxx  Create new migration"
	@echo ""
	@echo "Testing:"
	@echo "  make test               Run all tests"
	@echo "  make test-unit          Run unit tests only"
	@echo "  make test-integration   Run integration tests only"
	@echo "  make test-coverage      Generate coverage report"
	@echo ""
	@echo "Mocks:"
	@echo "  make mocks              Generate mocks (requires mockery)"
	@echo "  make mocks-clean        Remove generated mocks"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint               Run linter"
	@echo "  make lint-fix           Run linter and fix issues"
```

### 6. Create Mockery Config

```yaml
# .mockery.yaml
with-expecter: true
packages:
  github.com/irvanmhndra/nexpos-api/internal/repository:
    interfaces:
      UserRepository:
      ProductRepository:
      OrderRepository:
      StockRepository:
      AuditLogRepository:
    config:
      dir: internal/repository/mocks
      outpkg: mocks

  github.com/irvanmhndra/nexpos-api/internal/client/payment:
    interfaces:
      Gateway:
    config:
      dir: internal/client/mocks
      outpkg: mocks

  github.com/irvanmhndra/nexpos-api/internal/event/publisher:
    interfaces:
      EventPublisher:
    config:
      dir: internal/event/mocks
      outpkg: mocks
```

### 7. Start Development

```bash
# Start infrastructure
make docker-up

# Run migrations
make migrate-up

# Generate mocks (after writing interfaces)
make mocks

# Run linter
make lint

# Run tests
make test

# Start server
make run

# Or with hot reload
make dev
```

---

## Quick Reference

| Task | Command |
|------|---------|
| **Build & Run** | |
| Build | `make build` |
| Run server | `make run` |
| Hot reload | `make dev` |
| **Docker** | |
| Start services | `make docker-up` |
| Stop services | `make docker-down` |
| View logs | `make docker-logs` |
| **Database** | |
| Run migrations | `make migrate-up` |
| Rollback migration | `make migrate-down` |
| Create migration | `make migrate-create name=xxx` |
| **Testing** | |
| Run all tests | `make test` |
| Run unit tests | `make test-unit` |
| Run integration tests | `make test-integration` |
| Coverage report | `make test-coverage` |
| **Mocks** | |
| Generate mocks | `make mocks` |
| Clean mocks | `make mocks-clean` |
| **Code Quality** | |
| Run linter | `make lint` |
| Fix lint issues | `make lint-fix` |
| **Help** | |
| Show all commands | `make help` |

---

## Summary

**Architecture**: 3-Layer (Handler → Service → Repository)

**Key Folders**:
| Folder | Purpose |
|--------|---------|
| `handler/` | HTTP request handling |
| `service/` | Business logic orchestration |
| `repository/` | Data access (postgres/, mongo/) |
| `model/` | Database/domain models |
| `dto/request/` | API input contracts |
| `dto/response/` | API output contracts (with converters) |
| `client/` | Third-party API clients |
| `event/` | Message broker (Kafka) |
| `docs/` | Project documentation |
| `*/mocks/` | Generated mock files for testing |
| `tests/integration/` | Integration tests with real database |

**Testing Strategy**:
| Test Type | Location | Command |
|-----------|----------|---------|
| Unit tests | `internal/**/*_test.go` | `make test-unit` |
| Integration tests | `tests/integration/` | `make test-integration` |
| Mocks | `internal/*/mocks/` | `make mocks` |

**Benefits**:
- Simple & easy to understand
- Clear separation: Model (internal) vs DTO (external)
- Testable with interface-based dependencies
- Comprehensive test coverage (unit + integration)
- Scalable & production-ready
- No over-engineering
