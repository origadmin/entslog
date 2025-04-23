// Copyright (c) 2024 OrigAdmin. All rights reserved.

// Package entslog for entgo.io/ent
package entslog

import (
	"context"
	"fmt"
	"log/slog"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/goexts/generic/settings"
)

// SlogDriver is a init that logs all init operations.
type SlogDriver struct {
	Handler                // log function. defaults to slog.Default()
	dri     dialect.Driver // underlying init.
}

func (d *SlogDriver) Close() error {
	return d.dri.Close()
}

func (d *SlogDriver) Dialect() string {
	return d.dri.Dialect()
}

// New init a dialect.Driver and an optional logging function, and returns
// a new slog-init that prints all outgoing operations.
func New(dri dialect.Driver, options ...Option) dialect.Driver {
	config := settings.ApplyDefault(defaultConfig, options)
	handle := makeHandle(config)
	return &SlogDriver{dri: dri, Handler: handle.with(slog.String("database", "driver"))}
}

// Exec logs its params and calls the underlying init Exec method.
func (d *SlogDriver) Exec(ctx context.Context, query string, args, v any) error {
	d.Log(ctx, "Exec", slog.String("query", query), slog.Any("args", args))
	return d.LogError(ctx, "Exec", d.dri.Exec(ctx, query, args, v))
}

// ExecContext logs its params and calls the underlying init ExecContext method if it is supported.
func (d *SlogDriver) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	drv, ok := d.dri.(interface {
		ExecContext(context.Context, string, ...any) (sql.Result, error)
	})
	if !ok {
		return nil, fmt.Errorf("Driver.ExecContext is not supported")
	}
	d.Log(ctx, "ExecContext", slog.String("query", query), slog.Any("args", args))
	result, err := drv.ExecContext(ctx, query, args...)
	return result, d.LogError(ctx, "ExecContext", err)
}

// Query logs its params and calls the underlying init Query method.
func (d *SlogDriver) Query(ctx context.Context, query string, args, v any) error {
	d.Log(ctx, "Query", slog.String("query", query), slog.Any("args", args))
	return d.LogError(ctx, "Query", d.dri.Query(ctx, query, args, v))
}

// QueryContext logs its params and calls the underlying init QueryContext method if it is supported.
func (d *SlogDriver) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	drv, ok := d.dri.(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	})
	if !ok {
		return nil, fmt.Errorf("Driver.QueryContext is not supported")
	}
	d.Log(ctx, "QueryContext", slog.String("query", query), slog.Any("args", args))
	rows, err := drv.QueryContext(ctx, query, args...)
	return rows, d.LogError(ctx, "QueryContext", err)
}

// Tx adds an log-id for the transaction and calls the underlying init Tx command.
func (d *SlogDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	tx, err := d.dri.Tx(ctx)
	if err != nil {
		return nil, err
	}
	id := d.WithTrace(ctx)
	d.Log(ctx, "Tx started", slog.String("id", id))
	return &SlogTx{tx: tx, Handler: d.Handler.with(slog.String("database", "tx")), id: id, ctx: ctx}, nil
}

// BeginTx adds an log-id for the transaction and calls the underlying init BeginTx command if it is supported.
func (d *SlogDriver) BeginTx(ctx context.Context, opts *sql.TxOptions) (dialect.Tx, error) {
	drv, ok := d.dri.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	})
	if !ok {
		return nil, fmt.Errorf("Driver.BeginTx is not supported")
	}
	tx, err := drv.BeginTx(ctx, opts)
	if err != nil {
		return nil, d.LogError(ctx, "BeginTx", err)
	}
	id := d.WithTrace(ctx)
	d.Log(ctx, "BeginTx started", slog.String("id", id))
	return &SlogTx{tx: tx, Handler: d.Handler.with(slog.String("database", "tx")), id: id, ctx: ctx}, nil
}

// SlogTx is a transaction implementation that logs all transaction operations.
type SlogTx struct {
	Handler
	tx  dialect.Tx      // underlying transaction.
	id  string          // transaction logging id.
	ctx context.Context // underlying transaction context.
}

// Exec logs its params and calls the underlying transaction Exec method.
func (d *SlogTx) Exec(ctx context.Context, query string, args, v any) error {
	d.Log(ctx, "Exec", slog.String("id", d.id), slog.String("query", query), slog.Any("args", args))
	return d.LogError(ctx, "Exec", d.tx.Exec(ctx, query, args, v))
}

// ExecContext logs its params and calls the underlying transaction ExecContext method if it is supported.
func (d *SlogTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	drv, ok := d.tx.(interface {
		ExecContext(context.Context, string, ...any) (sql.Result, error)
	})
	if !ok {
		return nil, fmt.Errorf("Tx.ExecContext is not supported")
	}
	d.Log(ctx, "ExecContext", slog.String("id", d.id), slog.String("query", query), slog.Any("args", args))
	result, err := drv.ExecContext(ctx, query, args...)

	return result, d.LogError(ctx, "ExecContext", err)
}

// Query logs its params and calls the underlying transaction Query method.
func (d *SlogTx) Query(ctx context.Context, query string, args, v any) error {
	d.Log(ctx, "Query", slog.String("id", d.id), slog.String("query", query), slog.Any("args", args))
	return d.LogError(ctx, "Query", d.tx.Query(ctx, query, args, v))
}

// QueryContext logs its params and calls the underlying transaction QueryContext method if it is supported.
func (d *SlogTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	drv, ok := d.tx.(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	})
	if !ok {
		return nil, fmt.Errorf("Tx.QueryContext is not supported")
	}
	d.Log(ctx, "QueryContext", slog.String("id", d.id), slog.String("query", query), slog.Any("args", args))
	rows, err := drv.QueryContext(ctx, query, args...)

	return rows, d.LogError(ctx, "QueryContext", err)
}

// Commit logs this step and calls the underlying transaction Commit method.
func (d *SlogTx) Commit() error {
	d.Log(d.ctx, "Commit", slog.String("id", d.id))
	return d.LogError(d.ctx, "Commit", d.tx.Commit())
}

// Rollback logs this step and calls the underlying transaction Rollback method.
func (d *SlogTx) Rollback() error {
	d.Log(d.ctx, "Rollback", slog.String("id", d.id))
	return d.LogError(d.ctx, "Rollback", d.tx.Rollback())
}
