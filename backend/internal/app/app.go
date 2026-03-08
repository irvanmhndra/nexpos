package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/irvanmhndra/nexpos-api/config"
	"github.com/irvanmhndra/nexpos-api/internal/router"
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
	e.Use(middleware.RequestLogger())
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
	})

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
		Addr:    ":" + a.cfg.Server.Port,
		Handler: a.echo,
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
		a.db.Close()
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
	code := http.StatusInternalServerError
	message := "internal server error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Error()
	} else if err != nil {
		message = err.Error()
	}

	response := map[string]interface{}{
		"success":    false,
		"message":    message,
		"error_code": http.StatusText(code),
		"meta": map[string]interface{}{
			"server_time": time.Now().UTC().Format(time.RFC3339),
		},
	}

	if (*c).Request().Method == http.MethodHead {
		(*c).NoContent(code)
	} else {
		(*c).JSON(code, response)
	}
}
