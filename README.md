# entslog: slog ent handler

[![tag](https://img.shields.io/github/tag/godcong/entslog.svg)](https://github.com/godcong/entslog/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-%23007d9c)
[![GoDoc](https://godoc.org/github.com/godcong/entslog?status.svg)](https://pkg.go.dev/github.com/godcong/entslog)
![Build Status](https://github.com/godcong/entslog/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/godcong/entslog)](https://goreportcard.com/report/github.com/godcong/entslog)
[![Coverage](https://img.shields.io/codecov/c/github.com/godcong/entslog)](https://codecov.io/gh/godcong/entslog)
[![Contributors](https://img.shields.io/github/contributors/godcong/entslog)](https://github.com/godcong/entslog/graphs/contributors)
[![License](https://img.shields.io/github/license/godcong/entslog)](./LICENSE)

A [ent](https://entgo.io/ent) Driver wrapper that logs all database operations using [slog](https://pkg.go.dev/log/slog).

## Installation

```sh
go get github.com/origadmin/entslog/v3@latest
```

**Compatibility**: go >= 1.21

No breaking changes will be made to exported APIs before v1.0.0.

## Usage

### Basic Usage

```go
import (
    "entgo.io/ent/dialect/sql"
    "github.com/godcong/entslog"
)

// Wrap your driver with entslog
driver := entslog.New(sql.Open("mysql", "user:password@tcp(localhost:3306)/db?parseTime=true"))

// Use the driver with ent client
client := ent.NewClient(ent.Driver(driver))
```

### Custom Log Level

```go
import "log/slog"

driver := entslog.New(
    sql.Open("mysql", dsn),
    entslog.WithDefaultLevel(slog.LevelDebug),    // Set default log level
    entslog.WithErrorLevel(slog.LevelWarn),       // Set error log level
)
```

### Custom Logger

```go
import (
    "log/slog"
    "os"
)

// Create a custom logger
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

driver := entslog.New(
    sql.Open("mysql", dsn),
    entslog.WithLogger(logger),
)
```

### Filter Sensitive Data

```go
import "log/slog"

driver := entslog.New(
    sql.Open("mysql", dsn),
    entslog.WithFilter(func(ctx context.Context, attrs ...slog.Attr) []slog.Attr {
        filtered := make([]slog.Attr, 0, len(attrs))
        for _, attr := range attrs {
            // Filter out sensitive information like passwords
            if attr.Key == "password" || attr.Key == "token" {
                continue
            }
            filtered = append(filtered, attr)
        }
        return filtered
    }),
)
```

### Custom Trace Function

```go
driver := entslog.New(
    sql.Open("mysql", dsn),
    entslog.WithTrace(func(ctx context.Context) string {
        // Use your own trace ID generation logic
        if traceID := getTraceIDFromContext(ctx); traceID != "" {
            return traceID
        }
        return uuid.New().String()
    }),
)
```

### Transaction Logging

```go
func transactionExample(ctx context.Context, client *ent.Client) error {
    // Start transaction (logs transaction start with trace ID)
    tx, err := client.Tx(ctx)
    if err != nil {
        return err
    }

    // All operations within the transaction are logged with the same trace ID
    if err := tx.User.Create().SetName("John").Exec(ctx); err != nil {
        tx.Rollback()
        return err
    }

    // Commit (logs commit with the same trace ID)
    return tx.Commit()
}
```

### Output Example

```json
{
  "time": "2024-01-01T12:00:00.000Z",
  "level": "INFO",
  "msg": "Query",
  "database": "driver",
  "query": "SELECT * FROM users WHERE id = ?",
  "args": "[1]"
}

{
  "time": "2024-01-01T12:00:01.000Z",
  "level": "INFO",
  "msg": "Tx started",
  "database": "tx",
  "id": "550e8400-e29b-41d4-a716-446655440000"
}

{
  "time": "2024-01-01T12:00:02.000Z",
  "level": "INFO",
  "msg": "Exec",
  "database": "tx",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "query": "INSERT INTO users (name) VALUES (?)",
  "args": "[\"John\"]"
}

{
  "time": "2024-01-01T12:00:03.000Z",
  "level": "INFO",
  "msg": "Commit",
  "database": "tx",
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Features

- **Structured Logging**: Uses Go 1.21's `log/slog` for structured, leveled logging
- **Transaction Tracing**: Assigns unique trace IDs to transactions for complete operation tracking
- **Configurable**: Customize log levels, logger, filters, and trace generation
- **Ent Compatible**: Works seamlessly with ent's driver interface
- **Production Ready**: Efficient implementation with minimal overhead

## API Reference

GoDoc: [https://pkg.go.dev/github.com/godcong/entslog](https://pkg.go.dev/github.com/godcong/entslog)
