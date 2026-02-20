// Package middleware provides HTTP/gRPC middleware implementations
package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig configuration for tracing middleware
type TracingConfig struct {
	// TracerName is the name of the tracer
	TracerName string
	// Propagator is the propagator for context propagation
	Propagator propagation.TextMapPropagator
	// TracerProvider is the provider for creating tracers
	TracerProvider trace.TracerProvider
}

// DefaultTracingConfig returns default tracing configuration
func DefaultTracingConfig() TracingConfig {
	return TracingConfig{
		TracerName:     "myy-chat",
		Propagator:     otel.GetTextMapPropagator(),
		TracerProvider: otel.GetTracerProvider(),
	}
}

// Tracing creates a distributed tracing middleware using OpenTelemetry
func Tracing(config TracingConfig) middleware.Middleware {
	tracer := config.TracerProvider.Tracer(config.TracerName)

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Get transport info
			var (
				operation string
				kind      trace.SpanKind
			)
			if tr, ok := transport.FromServerContext(ctx); ok {
				operation = tr.Operation()
				kind = trace.SpanKindServer
			} else {
				operation = "unknown"
				kind = trace.SpanKindInternal
			}

			// Extract trace context from headers
			if tr, ok := transport.FromServerContext(ctx); ok {
				ctx = config.Propagator.Extract(ctx, newHeaderCarrier(tr.RequestHeader()))
			}

			// Start span
			ctx, span := tracer.Start(ctx, operation, trace.WithSpanKind(kind))
			defer span.End()

			// Add common attributes
			span.SetAttributes(
				attribute.String("operation", operation),
			)

			// Add user info if available
			if userID, ok := UserIDFromContext(ctx); ok {
				span.SetAttributes(attribute.Int64("user_id", userID))
			}

			// Add request ID if available
			if requestID, ok := RequestIDFromContext(ctx); ok {
				span.SetAttributes(attribute.String("request_id", requestID))
			}

			// Execute handler
			resp, err := handler(ctx, req)

			// Record error if any
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				// Add error attributes
				if appErr, ok := GetAppError(err); ok {
					span.SetAttributes(
						attribute.Int("error.code", appErr.Code),
						attribute.String("error.message", appErr.Message),
					)
				}
			} else {
				span.SetStatus(codes.Ok, "success")
			}

			return resp, err
		}
	}
}

// headerCarrier implements propagation.TextMapCarrier for transport.Header
type headerCarrier struct {
	header transport.Header
}

func newHeaderCarrier(h transport.Header) *headerCarrier {
	return &headerCarrier{header: h}
}

func (h *headerCarrier) Get(key string) string {
	return h.header.Get(key)
}

func (h *headerCarrier) Set(key, value string) {
	h.header.Set(key, value)
}

func (h *headerCarrier) Keys() []string {
	return h.header.Keys()
}

// ClientTracing creates a client-side tracing middleware
func ClientTracing(config TracingConfig) middleware.Middleware {
	tracer := config.TracerProvider.Tracer(config.TracerName)

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Get transport info
			var operation string
			if tr, ok := transport.FromClientContext(ctx); ok {
				operation = tr.Operation()
			} else {
				operation = "unknown"
			}

			// Start span
			ctx, span := tracer.Start(ctx, operation, trace.WithSpanKind(trace.SpanKindClient))
			defer span.End()

			// Inject trace context into headers
			if tr, ok := transport.FromClientContext(ctx); ok {
				config.Propagator.Inject(ctx, newHeaderCarrier(tr.RequestHeader()))
			}

			// Execute handler
			resp, err := handler(ctx, req)

			// Record error if any
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "success")
			}

			return resp, err
		}
	}
}

// SpanFromContext extracts the current span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// TraceIDFromContext extracts the trace ID from context
func TraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// SpanIDFromContext extracts the span ID from context
func SpanIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetSpanAttributes sets attributes on the current span
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordSpanError records an error on the current span
func RecordSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}
