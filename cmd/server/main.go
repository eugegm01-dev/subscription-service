package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	"subscription-service/internal/config"
	"subscription-service/internal/handler"
	"subscription-service/internal/logger"
	"subscription-service/internal/middleware"
	"subscription-service/internal/repository"
	"subscription-service/internal/service"

	_ "subscription-service/docs" // swagger docs (will be generated later)

	_ "github.com/lib/pq"
)

// @title Subscription Service API
// @version 1.0
// @description REST API for managing user subscriptions
// @host localhost:8080
// @BasePath /api
func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 2. Initialize logger
	log := logger.New(cfg.Log.Level, cfg.Log.Format)

	// 3. Connect to PostgreSQL
	db, err := repository.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("database connection established")

	// 4. Initialize layers (dependency injection)
	subRepo := repository.NewSubscriptionRepo(db)
	subSvc := service.NewSubscriptionService(subRepo)
	subHandler := handler.NewSubscriptionHandler(subSvc)

	// 5. Setup Echo
	e := echo.New()
	e.HideBanner = true
	e.HidePort = false
	e.HTTPErrorHandler = customHTTPErrorHandler(log)

	// Custom validator
	e.Validator = &customValidator{validator: validator.New()}

	// Global middleware (order matters)
	e.Use(middleware.RequestIDMiddleware())
	e.Use(middleware.LoggingMiddleware(log))
	e.Use(middleware.RecoveryMiddleware(log))
	e.Use(middleware.CORS()) // optional, but recommended

	// Health check endpoint (useful for orchestration)
	e.GET("/health", handler.HealthCheck)

	// Swagger endpoint
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// API routes
	api := e.Group("/api")
	api.POST("/subscriptions", subHandler.Create)
	api.GET("/subscriptions/:id", subHandler.GetByID)
	api.PUT("/subscriptions/:id", subHandler.Update)
	api.DELETE("/subscriptions/:id", subHandler.Delete)
	api.GET("/subscriptions", subHandler.List)
	api.GET("/subscriptions/total-cost", subHandler.TotalCost)

	// 6. Start server in a goroutine (graceful shutdown)
	go func() {
		addr := ":" + cfg.Server.Port
		log.Info("starting server", "port", cfg.Server.Port)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// 7. Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server...")

	// 8. Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Error("server shutdown error", "error", err)
	}
	log.Info("server exited")
}

// customHTTPErrorHandler handles errors and logs them
func customHTTPErrorHandler(log *slog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		he, ok := err.(*echo.HTTPError)
		if !ok {
			he = &echo.HTTPError{
				Code:    http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			}
		}
		// Log internal errors
		if he.Code >= 500 {
			log.Error("internal server error", "error", err, "path", c.Request().URL.Path)
		}
		// Send JSON response
		c.JSON(he.Code, map[string]interface{}{
			"error": he.Message,
		})
	}
}

// customValidator wraps go-playground/validator
type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
