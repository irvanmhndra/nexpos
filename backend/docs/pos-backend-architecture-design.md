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

```
pos-core-api/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
│
├── docs/                           # Documentation
│   ├── api.md                      # API documentation
│   ├── architecture.md             # Architecture overview
│   ├── database-schema.md          # Database design
│   └── setup.md                    # Setup guide
│
├── internal/
│   ├── handler/                    # HTTP handlers
│   │   ├── auth.go
│   │   ├── product.go
│   │   ├── order.go
│   │   ├── inventory.go
│   │   └── health.go
│   │
│   ├── service/                    # Business logic
│   │   ├── auth.go
│   │   ├── product.go
│   │   ├── order.go
│   │   ├── inventory.go
│   │   └── report.go
│   │
│   ├── repository/                 # Data access layer
│   │   ├── interface.go            # All repository interfaces
│   │   │
│   │   ├── postgres/               # PostgreSQL implementations
│   │   │   ├── user.go
│   │   │   ├── company.go
│   │   │   ├── branch.go
│   │   │   ├── product.go
│   │   │   ├── order.go
│   │   │   ├── stock.go
│   │   │   └── customer.go
│   │   │
│   │   └── mongo/                  # MongoDB implementations
│   │       ├── audit_log.go
│   │       └── activity.go
│   │
│   ├── model/                      # Database/domain models
│   │   ├── user.go
│   │   ├── company.go
│   │   ├── branch.go
│   │   ├── product.go
│   │   ├── order.go
│   │   ├── stock.go
│   │   ├── customer.go
│   │   ├── payment.go
│   │   └── audit_log.go
│   │
│   ├── dto/                        # Data Transfer Objects
│   │   ├── request/                # API input contracts
│   │   │   ├── auth.go
│   │   │   ├── user.go
│   │   │   ├── product.go
│   │   │   ├── order.go
│   │   │   └── inventory.go
│   │   │
│   │   └── response/               # API output contracts
│   │       ├── common.go           # Pagination, etc.
│   │       ├── auth.go
│   │       ├── user.go
│   │       ├── product.go
│   │       ├── order.go
│   │       └── inventory.go
│   │
│   ├── client/                     # Third-party API clients
│   │   ├── payment/
│   │   │   ├── interface.go
│   │   │   ├── midtrans.go
│   │   │   └── xendit.go
│   │   │
│   │   ├── notification/
│   │   │   ├── interface.go
│   │   │   ├── firebase.go
│   │   │   └── twilio.go
│   │   │
│   │   └── shipping/
│   │       ├── interface.go
│   │       └── jne.go
│   │
│   ├── event/                      # Message broker (Kafka)
│   │   ├── publisher/
│   │   │   ├── interface.go
│   │   │   └── kafka.go
│   │   │
│   │   ├── consumer/
│   │   │   ├── order_consumer.go
│   │   │   └── inventory_consumer.go
│   │   │
│   │   └── message/                # Event definitions
│   │       ├── order.go
│   │       └── inventory.go
│   │
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── tenant.go
│   │   ├── rbac.go
│   │   ├── logging.go
│   │   ├── recovery.go
│   │   └── request_id.go
│   │
│   └── router/
│       └── router.go
│
├── pkg/                            # Shared packages
│   ├── apperror/
│   │   └── error.go
│   │
│   ├── httputil/
│   │   └── response.go
│   │
│   ├── logger/
│   │   └── logger.go
│   │
│   ├── pagination/
│   │   └── pagination.go
│   │
│   └── validator/
│       └── validator.go
│
├── migrations/
│   ├── 000001_create_companies_table.up.sql
│   ├── 000001_create_companies_table.down.sql
│   └── ...
│
├── config/
│   └── config.go
│
├── scripts/
│   ├── setup.sh
│   └── seed.sh
│
├── tests/
│   └── integration/
│       ├── auth_test.go
│       ├── order_test.go
│       └── testhelper/
│           ├── database.go
│           ├── fixtures.go
│           └── http.go
│
├── .golangci.yml
├── .env.example
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

## Mandatory Libraries

### Core Dependencies

```go
module github.com/yourusername/pos-core-api

go 1.23

require (
    // Web framework
    github.com/labstack/echo/v4 v4.12.0

    // PostgreSQL
    github.com/jmoiron/sqlx v1.4.0
    github.com/lib/pq v1.10.9

    // MongoDB
    go.mongodb.org/mongo-driver v1.17.1

    // Redis
    github.com/redis/go-redis/v9 v9.7.0

    // Kafka
    github.com/segmentio/kafka-go v0.4.47

    // Configuration
    github.com/spf13/viper v1.19.0

    // Validation
    github.com/go-playground/validator/v10 v10.22.1

    // Utilities
    github.com/google/uuid v1.6.0
    github.com/golang-jwt/jwt/v5 v5.2.1
    github.com/rs/zerolog v1.33.0
    golang.org/x/crypto v0.28.0

    // Testing
    github.com/stretchr/testify v1.9.0
    github.com/DATA-DOG/go-sqlmock v1.5.2
    github.com/testcontainers/testcontainers-go v0.34.0

    // Migration
    github.com/golang-migrate/migrate/v4 v4.18.1
)
```

### Installation Commands

```bash
# Web framework
go get github.com/labstack/echo/v4

# Databases
go get github.com/jmoiron/sqlx
go get github.com/lib/pq
go get go.mongodb.org/mongo-driver/mongo
go get github.com/redis/go-redis/v9

# Message broker
go get github.com/segmentio/kafka-go

# Configuration & utilities
go get github.com/spf13/viper
go get github.com/go-playground/validator/v10
go get github.com/google/uuid
go get github.com/golang-jwt/jwt/v5
go get github.com/rs/zerolog
go get golang.org/x/crypto

# Testing
go get github.com/stretchr/testify
go get github.com/DATA-DOG/go-sqlmock
go get github.com/testcontainers/testcontainers-go

# Migration CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

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
    "github.com/labstack/echo/v4"
    "github.com/yourusername/pos-core-api/internal/dto/request"
    "github.com/yourusername/pos-core-api/internal/service"
    "github.com/yourusername/pos-core-api/pkg/apperror"
    "github.com/yourusername/pos-core-api/pkg/httputil"
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
    "github.com/yourusername/pos-core-api/internal/client/payment"
    "github.com/yourusername/pos-core-api/internal/dto/request"
    "github.com/yourusername/pos-core-api/internal/dto/response"
    "github.com/yourusername/pos-core-api/internal/event/message"
    "github.com/yourusername/pos-core-api/internal/event/publisher"
    "github.com/yourusername/pos-core-api/internal/model"
    "github.com/yourusername/pos-core-api/internal/repository"
    "github.com/yourusername/pos-core-api/pkg/apperror"
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
    "github.com/yourusername/pos-core-api/internal/model"
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
    "github.com/yourusername/pos-core-api/internal/model"
    "github.com/yourusername/pos-core-api/internal/repository"
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
    "github.com/yourusername/pos-core-api/internal/model"
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
    "github.com/yourusername/pos-core-api/internal/model"
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

### Client Layer (`internal/client/`)

Third-party API integrations.

```go
// internal/client/payment/interface.go
package payment

import "context"

type Gateway interface {
    Charge(ctx context.Context, req ChargeRequest) (*ChargeResponse, error)
    CheckStatus(ctx context.Context, transactionID string) (*StatusResponse, error)
    Refund(ctx context.Context, transactionID string, amount int64) error
}

type ChargeRequest struct {
    OrderID string
    Amount  int64
    Method  string
}

type ChargeResponse struct {
    TransactionID string
    Status        string
    PaymentURL    string
}

type StatusResponse struct {
    TransactionID string
    Status        string
}
```

```go
// internal/client/payment/midtrans.go
package payment

import (
    "context"
    "github.com/midtrans/midtrans-go"
    "github.com/midtrans/midtrans-go/coreapi"
)

type MidtransClient struct {
    client coreapi.Client
}

func NewMidtransClient(serverKey string, isProduction bool) *MidtransClient {
    env := midtrans.Sandbox
    if isProduction {
        env = midtrans.Production
    }

    c := coreapi.Client{}
    c.New(serverKey, env)

    return &MidtransClient{client: c}
}

func (m *MidtransClient) Charge(ctx context.Context, req ChargeRequest) (*ChargeResponse, error) {
    resp, err := m.client.ChargeTransaction(&coreapi.ChargeReq{
        TransactionDetails: midtrans.TransactionDetails{
            OrderID:  req.OrderID,
            GrossAmt: req.Amount,
        },
    })
    if err != nil {
        return nil, err
    }

    return &ChargeResponse{
        TransactionID: resp.TransactionID,
        Status:        resp.TransactionStatus,
    }, nil
}

func (m *MidtransClient) CheckStatus(ctx context.Context, transactionID string) (*StatusResponse, error) {
    resp, err := m.client.CheckTransaction(transactionID)
    if err != nil {
        return nil, err
    }

    return &StatusResponse{
        TransactionID: resp.TransactionID,
        Status:        resp.TransactionStatus,
    }, nil
}

func (m *MidtransClient) Refund(ctx context.Context, transactionID string, amount int64) error {
    _, err := m.client.RefundTransaction(transactionID, &coreapi.RefundReq{
        Amount: amount,
        Reason: "Order voided",
    })
    return err
}
```

### Event Layer (`internal/event/`)

Message broker integration.

```go
// internal/event/publisher/interface.go
package publisher

import "context"

type EventPublisher interface {
    Publish(ctx context.Context, topic string, message any) error
    Close() error
}
```

```go
// internal/event/publisher/kafka.go
package publisher

import (
    "context"
    "encoding/json"
    "github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
    writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string) *KafkaPublisher {
    return &KafkaPublisher{
        writer: &kafka.Writer{
            Addr:     kafka.TCP(brokers...),
            Balancer: &kafka.LeastBytes{},
        },
    }
}

func (p *KafkaPublisher) Publish(ctx context.Context, topic string, message any) error {
    data, err := json.Marshal(message)
    if err != nil {
        return err
    }

    return p.writer.WriteMessages(ctx, kafka.Message{
        Topic: topic,
        Value: data,
    })
}

func (p *KafkaPublisher) Close() error {
    return p.writer.Close()
}
```

```go
// internal/event/message/order.go
package message

import (
    "time"
    "github.com/google/uuid"
)

type OrderCreated struct {
    OrderID   uuid.UUID `json:"order_id"`
    CompanyID int64     `json:"company_id"`
    BranchID  int64     `json:"branch_id"`
    Total     int64     `json:"total"`
    CreatedAt time.Time `json:"created_at"`
}

type OrderVoided struct {
    OrderID   uuid.UUID `json:"order_id"`
    CompanyID int64     `json:"company_id"`
    VoidedBy  int64     `json:"voided_by"`
    Reason    string    `json:"reason"`
    VoidedAt  time.Time `json:"voided_at"`
}
```

---

## Configuration Management

```go
// config/config.go
package config

import (
    "fmt"
    "github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig
    Postgres PostgresConfig
    Mongo    MongoConfig
    Redis    RedisConfig
    Kafka    KafkaConfig
    JWT      JWTConfig
    Midtrans MidtransConfig
}

type ServerConfig struct {
    Port string
    Env  string
}

type PostgresConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

func (c PostgresConfig) DSN() string {
    return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

type MongoConfig struct {
    URI    string
    DBName string
}

type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}

func (c RedisConfig) Addr() string {
    return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type KafkaConfig struct {
    Brokers []string
    GroupID string
}

type JWTConfig struct {
    Secret           string
    AccessExpiresIn  int
    RefreshExpiresIn int
}

type MidtransConfig struct {
    ServerKey    string
    IsProduction bool
}

func Load() *Config {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()
    viper.ReadInConfig()

    return &Config{
        Server: ServerConfig{
            Port: viper.GetString("SERVER_PORT"),
            Env:  viper.GetString("SERVER_ENV"),
        },
        Postgres: PostgresConfig{
            Host:     viper.GetString("POSTGRES_HOST"),
            Port:     viper.GetString("POSTGRES_PORT"),
            User:     viper.GetString("POSTGRES_USER"),
            Password: viper.GetString("POSTGRES_PASSWORD"),
            DBName:   viper.GetString("POSTGRES_DB"),
            SSLMode:  viper.GetString("POSTGRES_SSLMODE"),
        },
        Mongo: MongoConfig{
            URI:    viper.GetString("MONGO_URI"),
            DBName: viper.GetString("MONGO_DB"),
        },
        Redis: RedisConfig{
            Host:     viper.GetString("REDIS_HOST"),
            Port:     viper.GetString("REDIS_PORT"),
            Password: viper.GetString("REDIS_PASSWORD"),
            DB:       viper.GetInt("REDIS_DB"),
        },
        Kafka: KafkaConfig{
            Brokers: viper.GetStringSlice("KAFKA_BROKERS"),
            GroupID: viper.GetString("KAFKA_GROUP_ID"),
        },
        JWT: JWTConfig{
            Secret:           viper.GetString("JWT_SECRET"),
            AccessExpiresIn:  viper.GetInt("JWT_ACCESS_EXPIRES_HOURS"),
            RefreshExpiresIn: viper.GetInt("JWT_REFRESH_EXPIRES_DAYS"),
        },
        Midtrans: MidtransConfig{
            ServerKey:    viper.GetString("MIDTRANS_SERVER_KEY"),
            IsProduction: viper.GetBool("MIDTRANS_IS_PRODUCTION"),
        },
    }
}
```

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
    "github.com/labstack/echo/v4"
    "github.com/yourusername/pos-core-api/pkg/apperror"
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
pos-core-api/
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
│   │   ├── postgres/
│   │   │   ├── order.go
│   │   │   └── order_test.go          # Unit test with sqlmock
│   │   │
│   │   └── mocks/                     # Generated mocks
│   │       ├── order_repository.go
│   │       ├── stock_repository.go
│   │       ├── product_repository.go
│   │       └── user_repository.go
│   │
│   ├── client/
│   │   ├── payment/
│   │   │   ├── midtrans.go
│   │   │   └── midtrans_test.go       # Unit test with HTTP mock
│   │   │
│   │   └── mocks/
│   │       └── payment_gateway.go
│   │
│   ├── event/
│   │   └── mocks/
│   │       └── event_publisher.go
│   │
│   └── dto/
│       └── response/
│           ├── order.go
│           └── order_test.go          # Test converters
│
├── tests/                             # Integration & E2E tests
│   ├── integration/
│   │   ├── auth_test.go
│   │   ├── order_test.go
│   │   ├── product_test.go
│   │   ├── inventory_test.go
│   │   │
│   │   └── helper/
│   │       ├── app.go                 # Test app setup
│   │       ├── database.go            # DB helpers
│   │       ├── fixtures.go            # Test data seeding
│   │       └── auth.go                # Auth helpers
│   │
│   └── e2e/                           # End-to-end tests (optional)
│       └── checkout_flow_test.go
│
└── Makefile
```

### Unit vs Integration Tests

| Aspect | Unit Test | Integration Test |
|--------|-----------|------------------|
| **Location** | `internal/**/*_test.go` | `tests/integration/` |
| **Dependencies** | Mocked | Real (testcontainers) |
| **Speed** | Fast (milliseconds) | Slower (seconds) |
| **Scope** | Single function/method | Full HTTP request flow |
| **Database** | sqlmock / in-memory | Real PostgreSQL container |
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
  github.com/yourusername/pos-core-api/internal/repository:
    interfaces:
      UserRepository:
      ProductRepository:
      OrderRepository:
      StockRepository:
      AuditLogRepository:
    config:
      dir: internal/repository/mocks
      outpkg: mocks

  github.com/yourusername/pos-core-api/internal/client/payment:
    interfaces:
      Gateway:
    config:
      dir: internal/client/mocks
      outpkg: mocks

  github.com/yourusername/pos-core-api/internal/event/publisher:
    interfaces:
      EventPublisher:
    config:
      dir: internal/event/mocks
      outpkg: mocks
```

Generate mocks:

```bash
mockery
```

### Unit Test Examples

#### Service Layer Test

```go
// internal/service/order_test.go
package service_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "github.com/yourusername/pos-core-api/internal/dto/request"
    "github.com/yourusername/pos-core-api/internal/model"
    "github.com/yourusername/pos-core-api/internal/service"
    repoMock "github.com/yourusername/pos-core-api/internal/repository/mocks"
    clientMock "github.com/yourusername/pos-core-api/internal/client/mocks"
    eventMock "github.com/yourusername/pos-core-api/internal/event/mocks"
)

func TestOrderService_Create_Success(t *testing.T) {
    // Arrange
    mockOrderRepo := repoMock.NewMockOrderRepository(t)
    mockStockRepo := repoMock.NewMockStockRepository(t)
    mockProductRepo := repoMock.NewMockProductRepository(t)
    mockAuditRepo := repoMock.NewMockAuditLogRepository(t)
    mockPaymentGw := clientMock.NewMockGateway(t)
    mockPublisher := eventMock.NewMockEventPublisher(t)

    mockStockRepo.On("GetAvailable", mock.Anything, int64(1), int64(1)).Return(100, nil)
    mockProductRepo.On("GetVariantByID", mock.Anything, int64(1), int64(1)).Return(&model.ProductVariant{
        ID: 1, ProductName: "Product A", Name: "Default", SKU: "SKU001", Price: 10000,
    }, nil)
    mockStockRepo.On("DeductTx", mock.Anything, mock.Anything, int64(1), int64(1), 2).Return(nil)
    mockOrderRepo.On("CreateTx", mock.Anything, mock.Anything, mock.AnythingOfType("*model.Order")).Return(nil)
    mockOrderRepo.On("GetByIDWithRelations", mock.Anything, int64(1), mock.Anything).Return(&model.Order{
        GrandTotal: 20000,
        Cashier:    &model.User{ID: 1, Name: "Cashier"},
    }, nil)
    mockPublisher.On("Publish", mock.Anything, "order.created", mock.Anything).Return(nil)
    mockAuditRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)

    svc := service.NewOrderService(nil, mockOrderRepo, mockStockRepo, mockProductRepo, mockAuditRepo, mockPaymentGw, mockPublisher)

    req := request.CreateOrder{
        Items:    []request.CreateOrderItem{{ProductVariantID: 1, Quantity: 2}},
        Payments: []request.CreatePayment{{Method: "cash", Amount: 20000}},
    }

    // Act
    result, err := svc.Create(context.Background(), 1, 1, 1, req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, int64(20000), result.GrandTotal)
}

func TestOrderService_Create_InsufficientStock(t *testing.T) {
    // Arrange
    mockStockRepo := repoMock.NewMockStockRepository(t)
    mockStockRepo.On("GetAvailable", mock.Anything, int64(1), int64(1)).Return(1, nil)

    svc := service.NewOrderService(nil, nil, mockStockRepo, nil, nil, nil, nil)

    req := request.CreateOrder{
        Items: []request.CreateOrderItem{{ProductVariantID: 1, Quantity: 10}},
    }

    // Act
    result, err := svc.Create(context.Background(), 1, 1, 1, req)

    // Assert
    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "insufficient")
}

func TestOrderService_Void_Success(t *testing.T) {
    // Arrange
    orderID := uuid.New()
    mockOrderRepo := repoMock.NewMockOrderRepository(t)
    mockStockRepo := repoMock.NewMockStockRepository(t)
    mockAuditRepo := repoMock.NewMockAuditLogRepository(t)
    mockPaymentGw := clientMock.NewMockGateway(t)

    mockOrderRepo.On("GetByIDWithRelations", mock.Anything, int64(1), orderID).Return(&model.Order{
        ID:         orderID,
        Status:     model.OrderStatusCompleted,
        GrandTotal: 20000,
        PaymentRef: "TXN123",
        Items: []model.OrderItem{
            {ProductVariantID: 1, Quantity: 2},
        },
    }, nil)
    mockPaymentGw.On("Refund", mock.Anything, "TXN123", int64(20000)).Return(nil)
    mockOrderRepo.On("UpdateStatus", mock.Anything, orderID, model.OrderStatusVoided).Return(nil)
    mockStockRepo.On("Add", mock.Anything, int64(1), int64(0), 2).Return(nil)
    mockAuditRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)

    svc := service.NewOrderService(nil, mockOrderRepo, mockStockRepo, nil, mockAuditRepo, mockPaymentGw, nil)

    // Act
    err := svc.Void(context.Background(), 1, orderID, 1, "Customer request")

    // Assert
    assert.NoError(t, err)
    mockPaymentGw.AssertCalled(t, "Refund", mock.Anything, "TXN123", int64(20000))
}
```

#### Handler Layer Test

```go
// internal/handler/order_test.go
package handler_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "github.com/yourusername/pos-core-api/internal/dto/response"
    "github.com/yourusername/pos-core-api/internal/handler"
    serviceMock "github.com/yourusername/pos-core-api/internal/service/mocks"
)

func TestOrderHandler_Create_Success(t *testing.T) {
    // Arrange
    mockOrderSvc := serviceMock.NewMockOrderService(t)
    mockOrderSvc.On("Create", mock.Anything, int64(1), int64(1), int64(1), mock.Anything).
        Return(&response.Order{
            ID:         "uuid-123",
            OrderNo:    "ORD-1-123",
            GrandTotal: 20000,
        }, nil)

    h := handler.NewOrderHandler(mockOrderSvc)

    e := echo.New()
    body, _ := json.Marshal(map[string]any{
        "items":    []map[string]any{{"product_variant_id": 1, "quantity": 2}},
        "payments": []map[string]any{{"method": "cash", "amount": 20000}},
    })

    req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    c := e.NewContext(req, rec)
    c.Set("company_id", int64(1))
    c.Set("branch_id", int64(1))
    c.Set("user_id", int64(1))

    // Act
    err := h.Create(c)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, rec.Code)

    var resp map[string]any
    json.Unmarshal(rec.Body.Bytes(), &resp)
    assert.True(t, resp["success"].(bool))
    assert.Equal(t, "ORD-1-123", resp["data"].(map[string]any)["order_no"])
}
```

#### Repository Layer Test (with sqlmock)

```go
// internal/repository/postgres/order_test.go
package postgres_test

import (
    "context"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/stretchr/testify/assert"

    "github.com/yourusername/pos-core-api/internal/model"
    "github.com/yourusername/pos-core-api/internal/repository/postgres"
)

func TestOrderRepository_GetByID_Found(t *testing.T) {
    // Arrange
    db, mock, _ := sqlmock.New()
    defer db.Close()
    sqlxDB := sqlx.NewDb(db, "postgres")

    orderID := uuid.New()
    rows := sqlmock.NewRows([]string{"id", "company_id", "branch_id", "order_no", "status", "grand_total", "created_at"}).
        AddRow(orderID, 1, 1, "ORD-1-123", "completed", 20000, time.Now())

    mock.ExpectQuery("SELECT \\* FROM orders WHERE").
        WithArgs(orderID, int64(1)).
        WillReturnRows(rows)

    repo := postgres.NewOrderRepository(sqlxDB)

    // Act
    order, err := repo.GetByID(context.Background(), 1, orderID)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, order)
    assert.Equal(t, "ORD-1-123", order.OrderNo)
    assert.Equal(t, int64(20000), order.GrandTotal)
}

func TestOrderRepository_GetByID_NotFound(t *testing.T) {
    // Arrange
    db, mock, _ := sqlmock.New()
    defer db.Close()
    sqlxDB := sqlx.NewDb(db, "postgres")

    orderID := uuid.New()
    mock.ExpectQuery("SELECT \\* FROM orders WHERE").
        WithArgs(orderID, int64(1)).
        WillReturnRows(sqlmock.NewRows(nil)) // Empty result

    repo := postgres.NewOrderRepository(sqlxDB)

    // Act
    order, err := repo.GetByID(context.Background(), 1, orderID)

    // Assert
    assert.NoError(t, err)
    assert.Nil(t, order)
}
```

### Integration Test Examples

#### Test App Setup

```go
// tests/integration/helper/app.go
package helper

import (
    "context"
    "testing"

    "github.com/jmoiron/sqlx"
    "github.com/labstack/echo/v4"
    _ "github.com/lib/pq"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"

    "github.com/yourusername/pos-core-api/config"
    "github.com/yourusername/pos-core-api/internal/handler"
    "github.com/yourusername/pos-core-api/internal/repository/postgres"
    "github.com/yourusername/pos-core-api/internal/router"
    "github.com/yourusername/pos-core-api/internal/service"
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
    "github.com/yourusername/pos-core-api/internal/model"
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

    "github.com/yourusername/pos-core-api/tests/integration/helper"
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
mkdir pos-core-api && cd pos-core-api
go mod init github.com/irvanmhndra/pos-core-api

# Create directory structure
mkdir -p cmd/api
mkdir -p docs
mkdir -p internal/{handler,service,repository/{postgres,mongo},model,dto/{request,response},client/{payment,notification},event/{publisher,consumer,message},middleware,router}
mkdir -p pkg/{apperror,httputil,logger,pagination,validator}
mkdir -p migrations config scripts tests/integration/testhelper
```

### 2. Install Dependencies

```bash
# Core
go get github.com/labstack/echo/v5
go get github.com/jmoiron/sqlx github.com/lib/pq
go get go.mongodb.org/mongo-driver/mongo
go get github.com/redis/go-redis/v9
go get github.com/segmentio/kafka-go

# Utils
go get github.com/spf13/viper
go get github.com/go-playground/validator/v10
go get github.com/google/uuid
go get github.com/golang-jwt/jwt/v5
go get github.com/rs/zerolog
go get golang.org/x/crypto

# Testing
go get github.com/stretchr/testify
go get github.com/DATA-DOG/go-sqlmock
go get github.com/testcontainers/testcontainers-go

# Migration CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Mock generator
go install github.com/vektra/mockery/v3@latest

# Hot reload (optional)
go install github.com/air-verse/air@latest

# Linter
brew install golangci-lint  # macOS
# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.62.2

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

MONGO_URI=mongodb://localhost:27017
MONGO_DB=pos_audit

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

KAFKA_BROKERS=localhost:9092
KAFKA_GROUP_ID=pos-core-api

JWT_SECRET=your-secret-key
JWT_ACCESS_EXPIRES_HOURS=2
JWT_REFRESH_EXPIRES_DAYS=7

MIDTRANS_SERVER_KEY=your-server-key
MIDTRANS_IS_PRODUCTION=false
EOF

cp .env.example .env
```

### 4. Docker Compose

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: pos_user
      POSTGRES_PASSWORD: pos_password
      POSTGRES_DB: pos_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  mongo:
    image: mongo:7
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  kafka:
    image: bitnami/kafka:latest
    environment:
      - KAFKA_CFG_NODE_ID=0
      - KAFKA_CFG_PROCESS_ROLES=controller,broker
      - KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093
      - KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      - KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@kafka:9093
      - KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER
    ports:
      - "9092:9092"

volumes:
  postgres_data:
  mongo_data:
```

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
  github.com/yourusername/pos-core-api/internal/repository:
    interfaces:
      UserRepository:
      ProductRepository:
      OrderRepository:
      StockRepository:
      AuditLogRepository:
    config:
      dir: internal/repository/mocks
      outpkg: mocks

  github.com/yourusername/pos-core-api/internal/client/payment:
    interfaces:
      Gateway:
    config:
      dir: internal/client/mocks
      outpkg: mocks

  github.com/yourusername/pos-core-api/internal/event/publisher:
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
