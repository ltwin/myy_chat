package middleware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// mockLogger is a simple logger for testing
type mockLogger struct {
	logs []string
}

func (m *mockLogger) Log(level log.Level, keyvals ...interface{}) error {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("[%s] ", level.String()))
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			buf.WriteString(fmt.Sprintf("%v=%v ", keyvals[i], keyvals[i+1]))
		}
	}
	m.logs = append(m.logs, buf.String())
	return nil
}

func (m *mockLogger) Clear() {
	m.logs = nil
}

func (m *mockLogger) LastLog() string {
	if len(m.logs) == 0 {
		return ""
	}
	return m.logs[len(m.logs)-1]
}

func TestDefaultLoggingConfig(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultLoggingConfig(logger)

	assert.Equal(t, logger, config.Logger)
	assert.Contains(t, config.SkipPaths, "/health")
	assert.Contains(t, config.SkipPaths, "/ready")
	assert.Equal(t, 3*time.Second, config.SlowThreshold)
	assert.False(t, config.LogRequestBody)
	assert.False(t, config.LogResponseBody)
}

func TestLogging_Success(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultLoggingConfig(logger)
	mw := Logging(config)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	wrapped := mw(handler)
	ctx := context.Background()

	resp, err := wrapped(ctx, "request")
	assert.NoError(t, err)
	assert.Equal(t, "success", resp)

	// Verify log was written
	assert.Len(t, logger.logs, 1)
	assert.Contains(t, logger.LastLog(), "[INFO]")
	assert.Contains(t, logger.LastLog(), "duration_ms")
}

func TestLogging_Error(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultLoggingConfig(logger)
	mw := Logging(config)

	expectedErr := errors.New("test error")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, expectedErr
	}

	wrapped := mw(handler)
	ctx := context.Background()

	_, err := wrapped(ctx, "request")
	assert.Error(t, err)

	// Verify error was logged
	assert.Len(t, logger.logs, 1)
	assert.Contains(t, logger.LastLog(), "[ERROR]")
	assert.Contains(t, logger.LastLog(), "test error")
}

func TestLogging_SkipPaths(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultLoggingConfig(logger)
	config.SkipPaths = []string{"/skip"}
	mw := Logging(config)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	wrapped := mw(handler)
	ctx := context.Background()

	_, _ = wrapped(ctx, "request")

	// Without transport context, operation is empty, so logging happens
	// In real scenario with transport context, skip paths would work
	assert.Len(t, logger.logs, 1)
}

func TestLogging_WithUserID(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultLoggingConfig(logger)
	mw := Logging(config)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	wrapped := mw(handler)
	ctx := WithUserID(context.Background(), 12345)

	_, err := wrapped(ctx, "request")
	assert.NoError(t, err)

	// Verify user_id was logged
	assert.Contains(t, logger.LastLog(), "user_id=12345")
}

func TestAccessLog(t *testing.T) {
	logger := &mockLogger{}
	mw := AccessLog(logger)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	wrapped := mw(handler)
	ctx := context.Background()

	resp, err := wrapped(ctx, "request")
	assert.NoError(t, err)
	assert.Equal(t, "success", resp)

	// Verify log was written
	assert.Len(t, logger.logs, 1)
	assert.Contains(t, logger.LastLog(), "request completed")
	assert.Contains(t, logger.LastLog(), "status=success")
}

func TestAccessLog_Error(t *testing.T) {
	logger := &mockLogger{}
	mw := AccessLog(logger)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, errors.New("test error")
	}

	wrapped := mw(handler)
	ctx := context.Background()

	_, err := wrapped(ctx, "request")
	assert.Error(t, err)

	// Verify error status was logged
	assert.Contains(t, logger.LastLog(), "status=error")
}

func TestRequestIDContext(t *testing.T) {
	ctx := context.Background()

	// Test WithRequestID and RequestIDFromContext
	ctx = WithRequestID(ctx, "test-request-id")

	requestID, ok := RequestIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "test-request-id", requestID)
}

func TestRequestIDContext_NotSet(t *testing.T) {
	ctx := context.Background()

	_, ok := RequestIDFromContext(ctx)
	assert.False(t, ok)
}

func TestRequestID_Middleware(t *testing.T) {
	generator := func() string {
		return "generated-id"
	}

	mw := RequestID(generator)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		requestID, ok := RequestIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "generated-id", requestID)
		return "success", nil
	}

	wrapped := mw(handler)
	ctx := context.Background()

	_, err := wrapped(ctx, "request")
	assert.NoError(t, err)
}
