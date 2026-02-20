package middleware

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError(CodeNotFound, "resource not found")
	assert.Equal(t, CodeNotFound, err.Code)
	assert.Equal(t, "resource not found", err.Message)
	assert.Empty(t, err.Reason)
	assert.Nil(t, err.Metadata)
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name:     "simple error",
			err:      NewAppError(CodeNotFound, "not found"),
			expected: "not found",
		},
		{
			name:     "error with cause",
			err:      NewAppError(CodeInternal, "internal error").WithCause(errors.New("database error")),
			expected: "internal error: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewAppError(CodeInternal, "wrapped").WithCause(cause)

	assert.Equal(t, cause, err.Unwrap())
	assert.True(t, errors.Is(err, cause))
}

func TestAppError_WithReason(t *testing.T) {
	err := NewAppError(CodeNotFound, "not found").WithReason("user does not exist")
	assert.Equal(t, "user does not exist", err.Reason)
}

func TestAppError_WithMetadata(t *testing.T) {
	metadata := map[string]string{"key": "value"}
	err := NewAppError(CodeNotFound, "not found").WithMetadata(metadata)
	assert.Equal(t, metadata, err.Metadata)
}

func TestAppError_HTTPStatus(t *testing.T) {
	tests := []struct {
		code     int
		expected int
	}{
		{CodeOK, http.StatusOK},
		{CodeInvalidArgument, http.StatusBadRequest},
		{CodeNotFound, http.StatusNotFound},
		{CodeUserNotFound, http.StatusNotFound},
		{CodeCharacterNotFound, http.StatusNotFound},
		{CodeConversationNotFound, http.StatusNotFound},
		{CodeAlreadyExists, http.StatusConflict},
		{CodeUserAlreadyExists, http.StatusConflict},
		{CodeCharacterAlreadyExists, http.StatusConflict},
		{CodePermissionDenied, http.StatusForbidden},
		{CodeUnauthenticated, http.StatusUnauthorized},
		{CodeInvalidCredentials, http.StatusUnauthorized},
		{CodeResourceExhausted, http.StatusTooManyRequests},
		{CodeRateLimited, http.StatusTooManyRequests},
		{CodeInsufficientCredits, http.StatusTooManyRequests},
		{CodeDeadlineExceeded, http.StatusGatewayTimeout},
		{CodeUnavailable, http.StatusServiceUnavailable},
		{CodeInternal, http.StatusInternalServerError},
		{CodeUnknown, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			err := NewAppError(tt.code, "test")
			assert.Equal(t, tt.expected, err.HTTPStatus())
		})
	}
}

func TestErrInvalidArgument(t *testing.T) {
	err := ErrInvalidArgument("invalid email format")
	assert.Equal(t, CodeInvalidArgument, err.Code)
	assert.Equal(t, "invalid email format", err.Message)
}

func TestErrNotFound(t *testing.T) {
	err := ErrNotFound("user")
	assert.Equal(t, CodeNotFound, err.Code)
	assert.Equal(t, "user not found", err.Message)
}

func TestErrAlreadyExists(t *testing.T) {
	err := ErrAlreadyExists("user")
	assert.Equal(t, CodeAlreadyExists, err.Code)
	assert.Equal(t, "user already exists", err.Message)
}

func TestErrPermissionDenied(t *testing.T) {
	err := ErrPermissionDenied("access denied")
	assert.Equal(t, CodePermissionDenied, err.Code)
	assert.Equal(t, "access denied", err.Message)
}

func TestErrUnauthenticated(t *testing.T) {
	err := ErrUnauthenticated("invalid token")
	assert.Equal(t, CodeUnauthenticated, err.Code)
	assert.Equal(t, "invalid token", err.Message)
}

func TestErrInternal(t *testing.T) {
	err := ErrInternal("database error")
	assert.Equal(t, CodeInternal, err.Code)
	assert.Equal(t, "database error", err.Message)
}

func TestErrInsufficientCredits(t *testing.T) {
	err := ErrInsufficientCredits(100, 50)
	assert.Equal(t, CodeInsufficientCredits, err.Code)
	assert.Equal(t, "insufficient credits", err.Message)
	assert.NotNil(t, err.Metadata)

	metadata, ok := err.Metadata.(map[string]float64)
	assert.True(t, ok)
	assert.Equal(t, 100.0, metadata["required"])
	assert.Equal(t, 50.0, metadata["available"])
}

func TestIsAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"app error", NewAppError(CodeNotFound, "not found"), true},
		{"standard error", errors.New("standard error"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsAppError(tt.err))
		})
	}
}

func TestGetAppError(t *testing.T) {
	appErr := NewAppError(CodeNotFound, "not found")

	tests := []struct {
		name     string
		err      error
		expected *AppError
		ok       bool
	}{
		{"app error", appErr, appErr, true},
		{"standard error", errors.New("standard error"), nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := GetAppError(tt.err)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestErrorCodeIs(t *testing.T) {
	err := NewAppError(CodeNotFound, "not found")

	assert.True(t, ErrorCodeIs(err, CodeNotFound))
	assert.False(t, ErrorCodeIs(err, CodeInternal))
	assert.False(t, ErrorCodeIs(errors.New("not app error"), CodeNotFound))
}

func TestErrorHandler_AppError(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultErrorHandlerConfig(logger)
	mw := ErrorHandler(config)

	appErr := NewAppError(CodeNotFound, "user not found")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, appErr
	}

	wrapped := mw(handler)
	ctx := context.Background()

	_, err := wrapped(ctx, "request")
	assert.Error(t, err)

	// Should return the original AppError
	var resultErr *AppError
	assert.True(t, errors.As(err, &resultErr))
	assert.Equal(t, CodeNotFound, resultErr.Code)
}

func TestErrorHandler_StandardError(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultErrorHandlerConfig(logger)
	mw := ErrorHandler(config)

	standardErr := errors.New("database connection failed")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, standardErr
	}

	wrapped := mw(handler)
	ctx := context.Background()

	_, err := wrapped(ctx, "request")
	assert.Error(t, err)

	// Should be converted to AppError with internal code
	var resultErr *AppError
	assert.True(t, errors.As(err, &resultErr))
	assert.Equal(t, CodeInternal, resultErr.Code)
	// Message should be hidden in production mode
	assert.Equal(t, "internal server error", resultErr.Message)
}

func TestErrorHandler_StandardError_ShowDetails(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultErrorHandlerConfig(logger)
	config.HideInternalErrors = false
	mw := ErrorHandler(config)

	standardErr := errors.New("database connection failed")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, standardErr
	}

	wrapped := mw(handler)
	ctx := context.Background()

	_, err := wrapped(ctx, "request")
	assert.Error(t, err)

	// Should show original error message
	var resultErr *AppError
	assert.True(t, errors.As(err, &resultErr))
	assert.Equal(t, "database connection failed", resultErr.Message)
}

func TestErrorHandler_NoError(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultErrorHandlerConfig(logger)
	mw := ErrorHandler(config)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	wrapped := mw(handler)
	ctx := context.Background()

	resp, err := wrapped(ctx, "request")
	assert.NoError(t, err)
	assert.Equal(t, "success", resp)
}

func TestErrorHandler_WithContext(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultErrorHandlerConfig(logger)
	mw := ErrorHandler(config)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, NewAppError(CodeNotFound, "not found")
	}

	wrapped := mw(handler)
	ctx := context.Background()
	ctx = WithUserID(ctx, 12345)
	ctx = WithRequestID(ctx, "req-123")

	_, _ = wrapped(ctx, "request")

	// Verify context info was logged
	assert.Len(t, logger.logs, 1)
	assert.Contains(t, logger.LastLog(), "user_id=12345")
	assert.Contains(t, logger.LastLog(), "request_id=req-123")
}

func TestDefaultErrorHandlerConfig(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultErrorHandlerConfig(logger)

	assert.Equal(t, logger, config.Logger)
	assert.True(t, config.LogStackTrace)
	assert.True(t, config.HideInternalErrors)
}

func TestErrorCodes(t *testing.T) {
	// Verify error codes are in correct ranges
	assert.Equal(t, 0, CodeOK)

	// Common errors (1000-1999)
	assert.True(t, CodeUnknown >= 1000 && CodeUnknown < 2000)
	assert.True(t, CodeInvalidArgument >= 1000 && CodeInvalidArgument < 2000)
	assert.True(t, CodeInternal >= 1000 && CodeInternal < 2000)

	// User errors (2000-2999)
	assert.True(t, CodeUserNotFound >= 2000 && CodeUserNotFound < 3000)
	assert.True(t, CodeInvalidCredentials >= 2000 && CodeInvalidCredentials < 3000)

	// Character errors (3000-3999)
	assert.True(t, CodeCharacterNotFound >= 3000 && CodeCharacterNotFound < 4000)

	// Conversation errors (4000-4999)
	assert.True(t, CodeConversationNotFound >= 4000 && CodeConversationNotFound < 5000)

	// Billing errors (5000-5999)
	assert.True(t, CodeInsufficientCredits >= 5000 && CodeInsufficientCredits < 6000)
}
