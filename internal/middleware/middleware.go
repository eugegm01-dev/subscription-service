package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func LoggingMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			reqID := c.Request().Context().Value("request_id").(string)
			logger.Info("request started",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"request_id", reqID,
			)
			err := next(c)
			logger.Info("request completed",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", c.Response().Status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", reqID,
			)
			return err
		}
	}
}
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := uuid.New().String()
			ctx := context.WithValue(c.Request().Context(), "request_id", reqID)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Response().Header().Set("X-Request-Id", reqID)
			return next(c)
		}
	}
}

func RecoveryMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic recovered", "error", r)
					c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
				}
			}()
			return next(c)
		}
	}
}
func CORS() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if c.Request().Method == http.MethodOptions {
				return c.NoContent(http.StatusNoContent)
			}
			return next(c)
		}
	}
}
