// Package middleware provides HTTP/gRPC middleware implementations
package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// LoggingConfig logging middleware configuration
type LoggingConfig struct {
	// Logger is the logger instance
	Logger log.Logger
	// SkipPaths paths to skip logging
	SkipPaths []string
	// SlowThreshold is the threshold for slow request logging
	SlowThreshold time.Duration
	// LogRequestBody whether to log request body
	LogRequestBody bool
	// LogResponseBody whether to log response body
	LogResponseBody bool
}

// DefaultLoggingConfig returns default logging configuration
func DefaultLoggingConfig(logger log.Logger) LoggingConfig {
	return LoggingConfig{
		Logger:          logger,
		SkipPaths:       []string{"/health", "/ready"},
		SlowThreshold:   3 * time.Second,
		LogRequestBody:  false,
		LogResponseBody: false,
	}
}

// Logging creates a logging middleware
// Log format follows English logging standard as per project requirements
func Logging(config LoggingConfig) middleware.Middleware {
	skipPaths := make(map[string]bool, len(config.SkipPaths))
	for _, p := range config.SkipPaths {
		skipPaths[p] = true
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			startTime := time.Now()

			// Get transport info
			var (
				operation string
				kind      string
			)
			if tr, ok := transport.FromServerContext(ctx); ok {
				operation = tr.Operation()
				kind = string(tr.Kind())
			}

			// Skip logging for specified paths
			if skipPaths[operation] {
				return handler(ctx, req)
			}

			// Get user info from context
			var userID int64
			if id, ok := UserIDFromContext(ctx); ok {
				userID = id
			}

			// Execute handler
			resp, err := handler(ctx, req)

			// Calculate duration
			duration := time.Since(startTime)
			durationMs := float64(duration.Nanoseconds()) / float64(time.Millisecond)

			// Build log fields
			logFields := []interface{}{
				"operation", operation,
				"kind", kind,
				"duration_ms", fmt.Sprintf("%.2f", durationMs),
			}

			if userID > 0 {
				logFields = append(logFields, "user_id", userID)
			}

			// Log based on result
			if err != nil {
				logFields = append(logFields, "error", err.Error())
				_ = config.Logger.Log(log.LevelError, logFields...)
			} else if duration > config.SlowThreshold {
				logFields = append(logFields, "slow_request", true)
				_ = config.Logger.Log(log.LevelWarn, logFields...)
			} else {
				_ = config.Logger.Log(log.LevelInfo, logFields...)
			}

			return resp, err
		}
	}
}

// AccessLog creates a simple access log middleware
// Logs: timestamp, method, path, status, duration
func AccessLog(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			startTime := time.Now()

			// Get transport info
			var operation string
			if tr, ok := transport.FromServerContext(ctx); ok {
				operation = tr.Operation()
			}

			// Execute handler
			resp, err := handler(ctx, req)

			// Log access
			duration := time.Since(startTime)
			status := "success"
			if err != nil {
				status = "error"
			}

			_ = logger.Log(log.LevelInfo,
				"msg", "request completed",
				"operation", operation,
				"status", status,
				"duration_ms", fmt.Sprintf("%.2f", float64(duration.Nanoseconds())/float64(time.Millisecond)),
			)

			return resp, err
		}
	}
}

// RequestIDKey is the context key for request ID
type requestIDKey struct{}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestIDFromContext gets request ID from context
func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey{}).(string)
	return requestID, ok
}

// RequestID creates a middleware that adds request ID to context
func RequestID(generator func() string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Try to get request ID from header
			var requestID string
			if tr, ok := transport.FromServerContext(ctx); ok {
				requestID = tr.RequestHeader().Get("X-Request-ID")
			}

			// Generate new request ID if not provided
			if requestID == "" {
				requestID = generator()
			}

			// Add to context
			ctx = WithRequestID(ctx, requestID)

			return handler(ctx, req)
		}
	}
}
