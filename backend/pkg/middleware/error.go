// Package middleware provides HTTP/gRPC middleware implementations
package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
)

// Standard error codes for the application
// Following the MyY Chat error code convention
const (
	// Common errors (1000-1999)
	CodeOK                = 0
	CodeUnknown           = 1000
	CodeInvalidArgument   = 1001
	CodeNotFound          = 1002
	CodeAlreadyExists     = 1003
	CodePermissionDenied  = 1004
	CodeUnauthenticated   = 1005
	CodeResourceExhausted = 1006
	CodeInternal          = 1007
	CodeUnavailable       = 1008
	CodeDeadlineExceeded  = 1009

	// User errors (2000-2999)
	CodeUserNotFound       = 2001
	CodeUserAlreadyExists  = 2002
	CodeInvalidCredentials = 2003
	CodeAccountDeleted     = 2004
	CodeAccountLocked      = 2005

	// Character errors (3000-3999)
	CodeCharacterNotFound      = 3001
	CodeCharacterAlreadyExists = 3002
	CodeMaxCharactersReached   = 3003

	// Conversation errors (4000-4999)
	CodeConversationNotFound = 4001
	CodeMessageTooLong       = 4002
	CodeRateLimited          = 4003

	// Billing errors (5000-5999)
	CodeInsufficientCredits = 5001
	CodePaymentFailed       = 5002
	CodeInvalidPackage      = 5003
)

// AppError represents an application-level error
type AppError struct {
	Code     int         `json:"code"`
	Message  string      `json:"message"`
	Reason   string      `json:"reason,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
	cause    error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.cause != nil {
		return e.Message + ": " + e.cause.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.cause
}

// HTTPStatus returns the appropriate HTTP status code
func (e *AppError) HTTPStatus() int {
	switch {
	case e.Code == CodeOK:
		return http.StatusOK
	case e.Code == CodeInvalidArgument:
		return http.StatusBadRequest
	case e.Code == CodeNotFound, e.Code == CodeUserNotFound,
		e.Code == CodeCharacterNotFound, e.Code == CodeConversationNotFound:
		return http.StatusNotFound
	case e.Code == CodeAlreadyExists, e.Code == CodeUserAlreadyExists,
		e.Code == CodeCharacterAlreadyExists:
		return http.StatusConflict
	case e.Code == CodePermissionDenied:
		return http.StatusForbidden
	case e.Code == CodeUnauthenticated, e.Code == CodeInvalidCredentials:
		return http.StatusUnauthorized
	case e.Code == CodeResourceExhausted, e.Code == CodeRateLimited,
		e.Code == CodeInsufficientCredits:
		return http.StatusTooManyRequests
	case e.Code == CodeDeadlineExceeded:
		return http.StatusGatewayTimeout
	case e.Code == CodeUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// NewAppError creates a new application error
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// WithReason adds a reason to the error
func (e *AppError) WithReason(reason string) *AppError {
	e.Reason = reason
	return e
}

// WithMetadata adds metadata to the error
func (e *AppError) WithMetadata(metadata interface{}) *AppError {
	e.Metadata = metadata
	return e
}

// WithCause adds a cause to the error
func (e *AppError) WithCause(cause error) *AppError {
	e.cause = cause
	return e
}

// Common error constructors

// ErrInvalidArgument creates an invalid argument error
func ErrInvalidArgument(message string) *AppError {
	return NewAppError(CodeInvalidArgument, message)
}

// ErrNotFound creates a not found error
func ErrNotFound(resource string) *AppError {
	return NewAppError(CodeNotFound, resource+" not found")
}

// ErrAlreadyExists creates an already exists error
func ErrAlreadyExists(resource string) *AppError {
	return NewAppError(CodeAlreadyExists, resource+" already exists")
}

// ErrPermissionDenied creates a permission denied error
func ErrPermissionDenied(message string) *AppError {
	return NewAppError(CodePermissionDenied, message)
}

// ErrUnauthenticated creates an unauthenticated error
func ErrUnauthenticated(message string) *AppError {
	return NewAppError(CodeUnauthenticated, message)
}

// ErrInternal creates an internal error
func ErrInternal(message string) *AppError {
	return NewAppError(CodeInternal, message)
}

// ErrInsufficientCredits creates an insufficient credits error
func ErrInsufficientCredits(required, available float64) *AppError {
	return NewAppError(CodeInsufficientCredits, "insufficient credits").
		WithMetadata(map[string]float64{
			"required":  required,
			"available": available,
		})
}

// ErrorHandlerConfig configuration for error handling middleware
type ErrorHandlerConfig struct {
	// Logger for logging errors
	Logger log.Logger
	// LogStackTrace whether to log stack traces
	LogStackTrace bool
	// HideInternalErrors whether to hide internal error details
	HideInternalErrors bool
}

// DefaultErrorHandlerConfig returns default configuration
func DefaultErrorHandlerConfig(logger log.Logger) ErrorHandlerConfig {
	return ErrorHandlerConfig{
		Logger:             logger,
		LogStackTrace:      true,
		HideInternalErrors: true,
	}
}

// ErrorHandler creates an error handling middleware
func ErrorHandler(config ErrorHandlerConfig) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			resp, err := handler(ctx, req)
			if err == nil {
				return resp, nil
			}

			// Log the error
			logError(ctx, err, config)

			// Convert to AppError if not already
			var appErr *AppError
			if !errors.As(err, &appErr) {
				appErr = &AppError{
					Code:    CodeInternal,
					Message: "internal server error",
					cause:   err,
				}

				// Hide internal error details in production
				if !config.HideInternalErrors {
					appErr.Message = err.Error()
				}
			}

			return resp, appErr
		}
	}
}

// logError logs the error with context
func logError(ctx context.Context, err error, config ErrorHandlerConfig) {
	if config.Logger == nil {
		return
	}

	fields := []interface{}{
		"error", err.Error(),
	}

	// Add user info if available
	if userID, ok := UserIDFromContext(ctx); ok {
		fields = append(fields, "user_id", userID)
	}

	// Add request ID if available
	if requestID, ok := RequestIDFromContext(ctx); ok {
		fields = append(fields, "request_id", requestID)
	}

	// Determine log level
	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Code >= CodeInternal {
			_ = config.Logger.Log(log.LevelError, fields...)
		} else {
			_ = config.Logger.Log(log.LevelWarn, fields...)
		}
	} else {
		_ = config.Logger.Log(log.LevelError, fields...)
	}
}

// IsAppError checks if the error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from an error
func GetAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// ErrorCodeIs checks if the error has the given code
func ErrorCodeIs(err error, code int) bool {
	appErr, ok := GetAppError(err)
	if !ok {
		return false
	}
	return appErr.Code == code
}
