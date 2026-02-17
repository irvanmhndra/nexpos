# POS Core API

A multi-tenant Point of Sale (POS) backend API built with Go, designed for retail businesses with support for order management, payment processing, product catalogs, and sales analytics.

## Tech Stack

- **Language:** Go 1.25
- **Framework:** Echo v5
- **Database:** PostgreSQL 18
- **Additional Services:** MongoDB, Redis (optional)
- **Authentication:** JWT with refresh tokens

## Features

- **Multi-Tenant Architecture** - Company and branch-based data isolation
- **Order Management** - Full lifecycle: Draft → Confirmed → Completed (or Cancelled/Voided)
- **Payment Processing** - Multiple payment methods, split payments, refunds
- **Product Catalog** - Products with variants, categories, SKUs, and cost tracking
- **Customer Management** - Customer profiles and membership tracking
- **Sales Reports** - Summary metrics, trends, top products, category revenue
- **User Management** - Role-based access control with permissions

## Prerequisites

- Go 1.25+
- Docker and Docker Compose
- Make

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/irvanmhndra/pos-core-api.git
cd pos-core-api
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Server
SERVER_PORT=8080
SERVER_ENV=development

# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_password
POSTGRES_DB=pos_db
POSTGRES_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key
JWT_ACCESS_EXPIRES_HOURS=2
JWT_REFRESH_EXPIRES_DAYS=7
```

### 3. Start services

```bash
make docker-up
```

### 4. Run database migrations

```bash
make migrate-up
```

### 5. Start the application

```bash
make run
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health status |
| GET | `/health/live` | Liveness probe |
| GET | `/health/ready` | Readiness probe |

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/auth/register` | Register user and company |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Logout and revoke session |

### Orders

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/orders` | Create order (draft) |
| GET | `/api/v1/orders` | List orders (paginated, filterable) |
| GET | `/api/v1/orders/:id` | Get order details |
| PUT | `/api/v1/orders/:id` | Update draft order |
| POST | `/api/v1/orders/:id/confirm` | Confirm order |
| POST | `/api/v1/orders/:id/payments` | Add payment(s) |
| POST | `/api/v1/orders/:id/complete` | Complete order |
| POST | `/api/v1/orders/:id/cancel` | Cancel order |
| POST | `/api/v1/orders/:id/void` | Void completed order |
| POST | `/api/v1/orders/:id/refund` | Process refund |

### Products

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/products` | Create product with variants |
| GET | `/api/v1/products` | List products (paginated) |
| GET | `/api/v1/products/:id` | Get product with variants |
| PUT | `/api/v1/products/:id` | Update product |
| DELETE | `/api/v1/products/:id` | Delete product |

### Product Categories

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/product-categories` | Create category |
| GET | `/api/v1/product-categories` | List categories (paginated) |
| GET | `/api/v1/product-categories/all` | List all categories |
| GET | `/api/v1/product-categories/:id` | Get category details |
| PUT | `/api/v1/product-categories/:id` | Update category |
| DELETE | `/api/v1/product-categories/:id` | Delete category |

### Customers

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/customers` | Create customer |
| GET | `/api/v1/customers` | List customers (paginated) |
| GET | `/api/v1/customers/:id` | Get customer details |
| PUT | `/api/v1/customers/:id` | Update customer |
| DELETE | `/api/v1/customers/:id` | Delete customer |

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/users` | Create user |
| GET | `/api/v1/users` | List users (paginated) |
| GET | `/api/v1/users/:id` | Get user details |
| PUT | `/api/v1/users/:id` | Update user |
| PATCH | `/api/v1/users/:id/status` | Update user status |
| DELETE | `/api/v1/users/:id` | Delete user |

### Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/reports/summary` | Sales summary metrics |
| GET | `/api/v1/reports/sales-trend` | Sales trend analysis |
| GET | `/api/v1/reports/top-products` | Top performing products |
| GET | `/api/v1/reports/category-revenue` | Revenue by category |
| GET | `/api/v1/reports/payment-methods` | Payment method breakdown |
| GET | `/api/v1/reports/hourly-sales` | Sales by hour |

## Project Structure

```
pos-core-api/
├── cmd/api/                 # Application entry point
├── config/                  # Configuration management
├── internal/
│   ├── app/                 # Application initialization
│   ├── router/              # Route definitions
│   ├── handler/             # HTTP request handlers
│   ├── service/             # Business logic
│   ├── repository/          # Data access layer
│   │   ├── postgres/        # PostgreSQL implementations
│   │   └── mongo/           # MongoDB implementations
│   ├── model/               # Domain models
│   ├── dto/                 # Data transfer objects
│   ├── middleware/          # Custom middleware
│   ├── event/               # Event system
│   └── client/              # External integrations
├── pkg/                     # Reusable utilities
│   ├── apperror/            # Custom error types
│   ├── httputil/            # HTTP response utilities
│   ├── validator/           # Request validation
│   ├── logger/              # Logging
│   └── pagination/          # Pagination helpers
├── migrations/              # Database migrations
├── tests/                   # Integration tests
├── docker-compose.yml
├── makefile
└── .env.example
```

## Development Commands

```bash
# Docker
make docker-up          # Start all services
make docker-down        # Stop all services

# Database
make migrate-up         # Run migrations
make migrate-down       # Rollback last migration
make migrate-create name=xyz  # Create new migration

# Development
make run                # Run the application
make dev                # Run with hot reload (air)
make build              # Build binary

# Testing
make test               # Run all tests
make test-coverage      # Generate coverage report
make mocks              # Generate test mocks

# Code Quality
make lint               # Run linter
make lint-fix           # Auto-fix linting issues
```

## Architecture

The application follows a layered architecture:

1. **Handler Layer** - HTTP request/response handling
2. **Service Layer** - Business logic and orchestration
3. **Repository Layer** - Data access with interface-based design
4. **Model Layer** - Domain entities

Key patterns:
- Dependency injection for testability
- Multi-tenancy via company_id/branch_id scoping
- Interface-based repositories for database abstraction
- Custom error types with HTTP status mapping

## Documentation

Additional documentation is available in the `docs/` directory:

- [API Documentation](docs/api_documentation.md)
- [Architecture Design](docs/pos-backend-architecture-design.md)
- [Database Schema](docs/pos-database-schema-design-early-phase.md)
- [Testing Guide](TESTING.md)

## License

This project is proprietary software. All rights reserved.
