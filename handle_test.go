// Copyright (c) 2024 OrigAdmin. All rights reserved.

package entslog

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
)

// TestMakeHandle tests the makeHandle function
func TestMakeHandle(t *testing.T) {
	// Test with nil logger (should use default)
	config := &Config{
		logger:      nil,
		level:       slog.LevelInfo,
		errorLevel:  slog.LevelError,
		handleError: true,
		filter:      emptyFilter,
		trace:       traceUUID,
	}

	handler := makeHandle(config)

	if handler.logger == nil {
		t.Error("Handler logger should not be nil")
	}

	if handler.filter == nil {
		t.Error("Handler filter should not be nil")
	}

	if handler.trace == nil {
		t.Error("Handler trace should not be nil")
	}
}

// TestHandlerInit tests Handler.init method
func TestHandlerInit(t *testing.T) {
	handler := &Handler{
		logger: slog.Default(),
		filter: emptyFilter,
		trace:  traceUUID,
		error:  noopError,
	}

	config := &Config{
		logger:      slog.Default(),
		level:       slog.LevelDebug,
		errorLevel:  slog.LevelWarn,
		handleError: true,
		filter:      emptyFilter,
		trace:       traceUUID,
	}

	initialized := handler.init(config)

	if initialized.log == nil {
		t.Error("Initialized handler should have log function")
	}

	if initialized.error == nil {
		t.Error("Initialized handler should have error function")
	}
}

// TestHandlerLog tests Handler.Log method
func TestHandlerLog(t *testing.T) {
	// Create a test logger
	testHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	handler := &Handler{
		logger: slog.New(testHandler),
		filter: emptyFilter,
		trace:  traceUUID,
	}

	// Manually set log function (simulating init)
	config := &Config{
		logger:      slog.New(testHandler),
		level:       slog.LevelDebug,
		errorLevel:  slog.LevelWarn,
		handleError: true,
		filter:      emptyFilter,
		trace:       traceUUID,
	}
	handler.init(config)

	ctx := context.Background()

	// This should not panic
	handler.Log(ctx, "Test message", slog.String("test", "value"))
}

// TestHandlerLogError tests Handler.LogError method
func TestHandlerLogError(t *testing.T) {
	testHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	handler := &Handler{
		logger: slog.New(testHandler),
		filter: emptyFilter,
		trace:  traceUUID,
	}

	config := &Config{
		logger:      slog.New(testHandler),
		level:       slog.LevelDebug,
		errorLevel:  slog.LevelWarn,
		handleError: true,
		filter:      emptyFilter,
		trace:       traceUUID,
	}
	handler.init(config)

	ctx := context.Background()
	testErr := errors.New("test error")

	// Test with error (should log error)
	result := handler.LogError(ctx, "Test error", testErr)
	if result == nil || result.Error() != testErr.Error() {
		t.Errorf("Expected error to be returned")
	}

	// Test with nil error
	result = handler.LogError(ctx, "Test error", nil)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

// TestHandlerWithTrace tests Handler.WithTrace method
func TestHandlerWithTrace(t *testing.T) {
	customTrace := func(ctx context.Context) string {
		return "custom-trace-id"
	}

	handler := &Handler{
		trace: customTrace,
	}

	ctx := context.Background()
	traceID := handler.WithTrace(ctx)

	if traceID != "custom-trace-id" {
		t.Errorf("Expected 'custom-trace-id', got '%s'", traceID)
	}
}

// TestHandlerFilter tests Handler.Filter method
func TestHandlerFilter(t *testing.T) {
	filterCalled := false

	customFilter := func(ctx context.Context, attrs ...slog.Attr) []slog.Attr {
		filterCalled = true
		return attrs
	}

	handler := &Handler{
		filter: customFilter,
		attrs:  []slog.Attr{slog.String("base", "value")},
	}

	ctx := context.Background()
	attrs := []slog.Attr{slog.String("test", "value")}
	result := handler.Filter(ctx, attrs...)

	if !filterCalled {
		t.Error("Filter was not called")
	}

	// Filter should combine base attrs with passed attrs
	if len(result) != 2 {
		t.Errorf("Expected 2 attrs (base + passed), got %d", len(result))
	}
}

// TestNoopError tests noopError function
func TestNoopError(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("test error")

	result := noopError(ctx, "message", testErr)

	if result == nil || result.Error() != testErr.Error() {
		t.Errorf("Expected error to be returned unchanged")
	}

	// Test with nil error
	result = noopError(ctx, "message", nil)

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

// TestHandlerWithErrorHandlingDisabled tests handler with error handling disabled
func TestHandlerWithErrorHandlingDisabled(t *testing.T) {
	testHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	handler := &Handler{
		logger: slog.New(testHandler),
		filter: emptyFilter,
		trace:  traceUUID,
		error:  noopError,
	}

	config := &Config{
		logger:      slog.New(testHandler),
		level:       slog.LevelDebug,
		errorLevel:  slog.LevelWarn,
		handleError: false, // Disable error handling
		filter:      emptyFilter,
		trace:       traceUUID,
	}
	handler.init(config)

	ctx := context.Background()
	testErr := errors.New("test error")

	// Error should still be returned even if logging is disabled
	result := handler.LogError(ctx, "Test error", testErr)

	if result == nil || result.Error() != testErr.Error() {
		t.Errorf("Expected error to be returned even with error handling disabled")
	}
}

// TestHandlerWithMultipleAttrs tests Handler.with with multiple attributes
func TestHandlerWithMultipleAttrs(t *testing.T) {
	handler := &Handler{
		attrs: []slog.Attr{
			slog.String("base1", "value1"),
			slog.String("base2", "value2"),
		},
	}

	newHandler := handler.with(
		slog.String("new1", "value3"),
		slog.String("new2", "value4"),
	)

	// The new handler should have base attrs appended with new attrs (total 4)
	if len(newHandler.attrs) != 4 {
		t.Errorf("Expected 4 attrs, got %d", len(newHandler.attrs))
	}

	// Verify the new handler has the correct base attrs
	if newHandler.attrs[0].Key != "base1" {
		t.Errorf("Expected attr key 'base1', got '%s'", newHandler.attrs[0].Key)
	}
}

// TestFilterRemovesSensitiveData tests filtering sensitive data
func TestFilterRemovesSensitiveData(t *testing.T) {
	sensitiveFilter := func(ctx context.Context, attrs ...slog.Attr) []slog.Attr {
		filtered := make([]slog.Attr, 0, len(attrs))
		for _, attr := range attrs {
			// Filter out sensitive keys
			if attr.Key == "password" || attr.Key == "token" {
				continue
			}
			filtered = append(filtered, attr)
		}
		return filtered
	}

	handler := &Handler{
		filter: sensitiveFilter,
	}

	ctx := context.Background()
	attrs := []slog.Attr{
		slog.String("username", "john"),
		slog.String("password", "secret"),
		slog.String("token", "abc123"),
		slog.String("email", "john@example.com"),
	}

	result := handler.Filter(ctx, attrs...)

	// Should have 2 non-sensitive attrs
	if len(result) != 2 {
		t.Errorf("Expected 2 non-sensitive attrs, got %d", len(result))
	}

	// Verify sensitive data was removed
	for _, attr := range result {
		if attr.Key == "password" || attr.Key == "token" {
			t.Errorf("Sensitive data '%s' was not filtered", attr.Key)
		}
	}
}

// TestHandlerLogLevels tests that different log levels work correctly
func TestHandlerLogLevels(t *testing.T) {
	for _, level := range []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	} {
		t.Run(level.String(), func(t *testing.T) {
			testHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: level,
			})

			handler := &Handler{
				logger: slog.New(testHandler),
				filter: emptyFilter,
				trace:  traceUUID,
			}

			config := &Config{
				logger:      slog.New(testHandler),
				level:       level,
				errorLevel:  level,
				handleError: true,
				filter:      emptyFilter,
				trace:       traceUUID,
			}
			handler.init(config)

			ctx := context.Background()
			handler.Log(ctx, "Test message")
		})
	}
}

// BenchmarkHandlerLog benchmarks the Handler.Log method
func BenchmarkHandlerLog(b *testing.B) {
	testHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	handler := &Handler{
		logger: slog.New(testHandler),
		filter: emptyFilter,
		trace:  traceUUID,
	}

	config := &Config{
		logger:      slog.New(testHandler),
		level:       slog.LevelDebug,
		errorLevel:  slog.LevelError,
		handleError: true,
		filter:      emptyFilter,
		trace:       traceUUID,
	}
	handler.init(config)

	ctx := context.Background()
	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.String("key2", "value2"),
		slog.Int("key3", 42),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Log(ctx, "Benchmark message", attrs...)
	}
}
