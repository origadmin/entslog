// Copyright (c) 2024 OrigAdmin. All rights reserved.

package entslog

import (
	"context"
	"log/slog"
	"testing"

	"entgo.io/ent/dialect"
)

// mockDriver is a mock implementation of dialect.Driver for testing
type mockDriver struct {
	dialect.Driver
	queryFunc  func(context.Context, string, any, any) error
	execFunc   func(context.Context, string, any, any) error
	txFunc     func(context.Context) (dialect.Tx, error)
	closeFunc  func() error
	dialFunc   func() string
}

func (m *mockDriver) Query(ctx context.Context, query string, args, v any) error {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, query, args, v)
	}
	return nil
}

func (m *mockDriver) Exec(ctx context.Context, query string, args, v any) error {
	if m.execFunc != nil {
		return m.execFunc(ctx, query, args, v)
	}
	return nil
}

func (m *mockDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	if m.txFunc != nil {
		return m.txFunc(ctx)
	}
	return &mockTx{}, nil
}

func (m *mockDriver) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockDriver) Dialect() string {
	if m.dialFunc != nil {
		return m.dialFunc()
	}
	return "mock"
}

// mockTx is a mock implementation of dialect.Tx for testing
type mockTx struct {
	dialect.Tx
	queryFunc    func(context.Context, string, any, any) error
	execFunc     func(context.Context, string, any, any) error
	commitFunc   func() error
	rollbackFunc func() error
}

func (m *mockTx) Query(ctx context.Context, query string, args, v any) error {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, query, args, v)
	}
	return nil
}

func (m *mockTx) Exec(ctx context.Context, query string, args, v any) error {
	if m.execFunc != nil {
		return m.execFunc(ctx, query, args, v)
	}
	return nil
}

func (m *mockTx) Commit() error {
	if m.commitFunc != nil {
		return m.commitFunc()
	}
	return nil
}

func (m *mockTx) Rollback() error {
	if m.rollbackFunc != nil {
		return m.rollbackFunc()
	}
	return nil
}

// TestNewDriver tests the New function
func TestNewDriver(t *testing.T) {
	mock := &mockDriver{
		dialFunc: func() string { return "test" },
	}

	driver := New(mock)
	if driver == nil {
		t.Fatal("New() returned nil")
	}

	if driver.Dialect() != "test" {
		t.Errorf("Expected dialect 'test', got '%s'", driver.Dialect())
	}

	if err := driver.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

// TestNewDriverWithOptions tests the New function with options
func TestNewDriverWithOptions(t *testing.T) {
	mock := &mockDriver{
		dialFunc: func() string { return "test" },
	}

	logger := slog.Default()

	driver := New(mock,
		WithLogger(logger),
		WithDefaultLevel(slog.LevelDebug),
		WithErrorLevel(slog.LevelWarn),
		WithError(),
		WithFilter(emptyFilter),
		WithTrace(traceUUID),
	)

	if driver == nil {
		t.Fatal("New() with options returned nil")
	}
}

// TestDriverQuery tests the Query method
func TestDriverQuery(t *testing.T) {
	queryCalled := false
	mock := &mockDriver{
		dialFunc: func() string { return "test" },
		queryFunc: func(ctx context.Context, query string, args, v any) error {
			queryCalled = true
			return nil
		},
	}

	driver := New(mock)
	ctx := context.Background()

	err := driver.Query(ctx, "SELECT * FROM users", []any{}, nil)
	if err != nil {
		t.Errorf("Query() returned error: %v", err)
	}

	if !queryCalled {
		t.Error("Underlying driver Query was not called")
	}
}

// TestDriverExec tests the Exec method
func TestDriverExec(t *testing.T) {
	execCalled := false
	mock := &mockDriver{
		dialFunc: func() string { return "test" },
		execFunc: func(ctx context.Context, query string, args, v any) error {
			execCalled = true
			return nil
		},
	}

	driver := New(mock)
	ctx := context.Background()

	err := driver.Exec(ctx, "INSERT INTO users VALUES (?)", []any{1}, nil)
	if err != nil {
		t.Errorf("Exec() returned error: %v", err)
	}

	if !execCalled {
		t.Error("Underlying driver Exec was not called")
	}
}

// TestDriverTx tests the Tx method
func TestDriverTx(t *testing.T) {
	txCalled := false
	mock := &mockDriver{
		dialFunc: func() string { return "test" },
		txFunc: func(ctx context.Context) (dialect.Tx, error) {
			txCalled = true
			return &mockTx{}, nil
		},
	}

	driver := New(mock)
	ctx := context.Background()

	tx, err := driver.Tx(ctx)
	if err != nil {
		t.Errorf("Tx() returned error: %v", err)
	}

	if tx == nil {
		t.Fatal("Tx() returned nil transaction")
	}

	if !txCalled {
		t.Error("Underlying driver Tx was not called")
	}

	// Test commit
	if err := tx.Commit(); err != nil {
		t.Errorf("Commit() returned error: %v", err)
	}
}

// TestTxQuery tests transaction Query method
func TestTxQuery(t *testing.T) {
	queryCalled := false
	mockTx := &mockTx{
		queryFunc: func(ctx context.Context, query string, args, v any) error {
			queryCalled = true
			return nil
		},
	}

	driver := New(&mockDriver{dialFunc: func() string { return "test" }})
	slogDriver, ok := driver.(*SlogDriver)
	if !ok {
		t.Fatal("Could not convert to *SlogDriver")
	}
	driverWithAttrs := slogDriver.Handler.with(slog.String("test", "value"))
	tx := &SlogTx{
		Handler: driverWithAttrs,
		tx:      mockTx,
		id:      "test-tx-id",
		ctx:     context.Background(),
	}

	ctx := context.Background()
	err := tx.Query(ctx, "SELECT * FROM users", []any{}, nil)
	if err != nil {
		t.Errorf("Tx.Query() returned error: %v", err)
	}

	if !queryCalled {
		t.Error("Underlying transaction Query was not called")
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("Tx.Commit() returned error: %v", err)
	}
}

// TestTxExec tests transaction Exec method
func TestTxExec(t *testing.T) {
	execCalled := false
	mockTx := &mockTx{
		execFunc: func(ctx context.Context, query string, args, v any) error {
			execCalled = true
			return nil
		},
	}

	driver := New(&mockDriver{dialFunc: func() string { return "test" }})
	slogDriver, ok := driver.(*SlogDriver)
	if !ok {
		t.Fatal("Could not convert to *SlogDriver")
	}
	driverWithAttrs := slogDriver.Handler.with(slog.String("test", "value"))
	tx := &SlogTx{
		Handler: driverWithAttrs,
		tx:      mockTx,
		id:      "test-tx-id",
		ctx:     context.Background(),
	}

	ctx := context.Background()
	err := tx.Exec(ctx, "INSERT INTO users VALUES (?)", []any{1}, nil)
	if err != nil {
		t.Errorf("Tx.Exec() returned error: %v", err)
	}

	if !execCalled {
		t.Error("Underlying transaction Exec was not called")
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("Tx.Commit() returned error: %v", err)
	}
}

// TestTxRollback tests transaction Rollback method
func TestTxRollback(t *testing.T) {
	rollbackCalled := false
	mockTx := &mockTx{
		rollbackFunc: func() error {
			rollbackCalled = true
			return nil
		},
	}

	driver := New(&mockDriver{dialFunc: func() string { return "test" }})
	slogDriver, ok := driver.(*SlogDriver)
	if !ok {
		t.Fatal("Could not convert to *SlogDriver")
	}
	driverWithAttrs := slogDriver.Handler.with(slog.String("test", "value"))
	tx := &SlogTx{
		Handler: driverWithAttrs,
		tx:      mockTx,
		id:      "test-tx-id",
		ctx:     context.Background(),
	}

	if err := tx.Rollback(); err != nil {
		t.Errorf("Tx.Rollback() returned error: %v", err)
	}

	if !rollbackCalled {
		t.Error("Underlying transaction Rollback was not called")
	}
}

// TestTraceUUID tests traceUUID function
func TestTraceUUID(t *testing.T) {
	ctx := context.Background()

	// Test without trace in context
	traceID := traceUUID(ctx)
	if traceID == "" {
		t.Error("traceUUID returned empty string")
	}

	// Test with trace in context
	ctxWithTrace := TraceContext(ctx, "custom-trace-id")
	traceID = traceUUID(ctxWithTrace)
	if traceID != "custom-trace-id" {
		t.Errorf("Expected 'custom-trace-id', got '%s'", traceID)
	}
}

// TestEmptyFilter tests emptyFilter function
func TestEmptyFilter(t *testing.T) {
	ctx := context.Background()
	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}

	result := emptyFilter(ctx, attrs...)
	if len(result) != len(attrs) {
		t.Errorf("Expected %d attrs, got %d", len(attrs), len(result))
	}
}

// TestCustomFilter tests custom filter function
func TestCustomFilter(t *testing.T) {
	filterCalled := false

	customFilter := func(ctx context.Context, attrs ...slog.Attr) []slog.Attr {
		filterCalled = true
		return attrs
	}

	driver := New(&mockDriver{dialFunc: func() string { return "test" }},
		WithFilter(customFilter),
	)

	ctx := context.Background()
	driver.Query(ctx, "SELECT 1", []any{}, nil)

	if !filterCalled {
		t.Error("Custom filter was not called")
	}
}

// TestWithDefaultLevel tests WithDefaultLevel option
func TestWithDefaultLevel(t *testing.T) {
	config := defaultConfig()
	WithDefaultLevel(slog.LevelDebug)(config)

	if config.level.Level() != slog.LevelDebug {
		t.Errorf("Expected LevelDebug, got %v", config.level.Level())
	}
}

// TestWithErrorLevel tests WithErrorLevel option
func TestWithErrorLevel(t *testing.T) {
	config := defaultConfig()
	WithErrorLevel(slog.LevelWarn)(config)

	if config.errorLevel.Level() != slog.LevelWarn {
		t.Errorf("Expected LevelWarn, got %v", config.errorLevel.Level())
	}

	if !config.handleError {
		t.Error("Expected handleError to be true")
	}
}

// TestWithLogger tests WithLogger option
func TestWithLogger(t *testing.T) {
	logger := slog.Default()
	config := defaultConfig()
	WithLogger(logger)(config)

	if config.logger != logger {
		t.Error("Logger was not set correctly")
	}
}

// TestWithTrace tests WithTrace option
func TestWithTrace(t *testing.T) {
	customTrace := func(ctx context.Context) string {
		return "custom-id"
	}

	config := defaultConfig()
	WithTrace(customTrace)(config)

	if config.trace == nil {
		t.Error("Trace function was not set")
	}

	result := config.trace(context.Background())
	if result != "custom-id" {
		t.Errorf("Expected 'custom-id', got '%s'", result)
	}
}

// TestHandlerWith tests Handler.with method
func TestHandlerWith(t *testing.T) {
	handler := &Handler{
		attrs: []slog.Attr{slog.String("key1", "value1")},
	}

	newHandler := handler.with(slog.String("key2", "value2"))

	if len(newHandler.attrs) != 2 {
		t.Errorf("Expected 2 attrs, got %d", len(newHandler.attrs))
	}
}

// TestSQLDriverWrapper tests integration with dialect.Driver
func TestSQLDriverWrapper(t *testing.T) {
	// Create a simple dialect driver wrapper
	mockDriver := &mockDriver{
		dialFunc: func() string { return "test" },
	}

	// Wrap with entslog
	loggedDriver := New(mockDriver)

	// Test that it implements all required interfaces
	if _, ok := loggedDriver.(dialect.Driver); !ok {
		t.Error("Logged driver does not implement dialect.Driver")
	}

	if _, ok := loggedDriver.(dialect.ExecQuerier); !ok {
		t.Error("Logged driver does not implement dialect.ExecQuerier")
	}
}
