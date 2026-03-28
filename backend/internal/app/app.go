package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/irvanmhndra/nexpos-api/config"
	authmiddleware "github.com/irvanmhndra/nexpos-api/internal/middleware"
	"github.com/irvanmhndra/nexpos-api/internal/router"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/irvanmhndra/nexpos-api/pkg/validator"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	_ "github.com/lib/pq"
)

type App struct {
	cfg      *config.Config
	db       *sqlx.DB
	echo     *echo.Echo
	Repos    *Repositories
	Services *Services
	Handlers *Handlers
}

func New(cfg *config.Config) (*App, error) {
	// Initialize database
	db, err := sqlx.Connect("postgres", cfg.Postgres.DSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Initialize validator
	v := validator.New()

	// Initialize layers
	repos := initRepositories(db)
	services := initServices(repos, cfg)
	handlers := initHandlers(db, services, v)

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogLatency:      true,
		LogRemoteIP:     true,
		LogHost:         true,
		LogMethod:       true,
		LogURI:          true,
		LogRequestID:    true,
		LogUserAgent:    true,
		LogStatus:       true,
		LogResponseSize: true,
		HandleError:     false,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				slog.Info("request",
					"method", v.Method,
					"uri", v.URI,
					"status", v.Status,
					"latency", v.Latency,
					"remote_ip", v.RemoteIP,
					"request_id", v.RequestID,
				)
			} else {
				slog.Error("request error",
					"method", v.Method,
					"uri", v.URI,
					"status", v.Status,
					"latency", v.Latency,
					"remote_ip", v.RemoteIP,
					"request_id", v.RequestID,
					"error", v.Error,
				)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
	}))
	e.Use(middleware.RequestID())

	// Custom HTTP error handler for consistent error responses
	e.HTTPErrorHandler = customHTTPErrorHandler

	// Setup routes
	router.Setup(e, &router.Handlers{
		Health:          handlers.Health,
		Auth:            handlers.Auth,
		User:            handlers.User,
		Branch:          handlers.Branch,
		Customer:        handlers.Customer,
		ProductCategory: handlers.ProductCategory,
		Product:         handlers.Product,
		Order:           handlers.Order,
		Report:          handlers.Report,
		Promotion:       handlers.Promotion,
		Inventory:       handlers.Inventory,
		CompanySettings: handlers.CompanySettings,
		Supplier:        handlers.Supplier,
		PurchaseOrder:   handlers.PurchaseOrder,
		Shift:           handlers.Shift,
		ExpenseCategory: handlers.ExpenseCategory,
		Expense:         handlers.Expense,
	}, authmiddleware.Auth(repos.UserSession))

	return &App{
		cfg:      cfg,
		db:       db,
		echo:     e,
		Repos:    repos,
		Services: services,
		Handlers: handlers,
	}, nil
}

func (a *App) Run() {
	server := &http.Server{
		Addr:              ":" + a.cfg.Server.Port,
		Handler:           a.echo,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("Starting server", "port", a.cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited gracefully")
}

func (a *App) Close() {
	if a.db != nil {
		_ = a.db.Close()
	}
}

// Echo returns the Echo instance for testing
func (a *App) Echo() *echo.Echo {
	return a.echo
}

// DB returns the database connection for testing
func (a *App) DB() *sqlx.DB {
	return a.db
}

// customHTTPErrorHandler provides consistent error response format
func customHTTPErrorHandler(c *echo.Context, err error) {
	if r, _ := echo.UnwrapResponse(c.Response()); r != nil && r.Committed {
		return
	}

	var sc echo.HTTPStatusCoder
	if errors.As(err, &sc) {
		code := sc.StatusCode()
		_ = c.JSON(code, httputil.Response{
			Success:   false,
			Message:   http.StatusText(code),
			ErrorCode: http.StatusText(code),
		})
		return
	}

	slog.Error("unhandled error", "method", c.Request().Method, "path", c.Request().URL.Path, "error", err)
	_ = c.JSON(http.StatusInternalServerError, httputil.Response{
		Success:   false,
		Message:   "internal server error",
		ErrorCode: "INTERNAL_ERROR",
	})
}
