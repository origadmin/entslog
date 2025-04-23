// Copyright (c) 2024 OrigAdmin. All rights reserved.

// Package entslog for entgo.io/ent
package entslog

import (
	"context"
	"log/slog"
	"slices"
)

type Handler struct {
	logger *slog.Logger
	filter FilterAttrs
	trace  TraceFunc
	log    func(ctx context.Context, msg string, attrs ...slog.Attr)
	error  func(ctx context.Context, msg string, err error) error
	attrs  []slog.Attr
}

func (h *Handler) init(o *Config) *Handler {
	h.log = func(ctx context.Context, msg string, attrs ...slog.Attr) {
		attrs = h.Filter(ctx, attrs...)
		h.logger.LogAttrs(ctx, o.level.Level(), msg, attrs...)
	}
	if o.handleError {
		h.error = func(ctx context.Context, msg string, err error) error {
			if err != nil {
				attrs := h.Filter(ctx, slog.Any("error", err))
				h.logger.LogAttrs(ctx, o.errorLevel.Level(), msg, attrs...)
			}
			return err
		}
	}
	return h
}

func (h *Handler) with(attrs ...slog.Attr) Handler {
	handlerCopy := *h
	handlerCopy.attrs = attrs
	return handlerCopy
}

func (h *Handler) WithTrace(ctx context.Context) string {
	return h.trace(ctx)
}

func (h *Handler) Filter(ctx context.Context, attrs ...slog.Attr) []slog.Attr {
	return h.filter(ctx, slices.Concat(h.attrs, attrs)...)
}

func (h *Handler) Log(ctx context.Context, msg string, attrs ...slog.Attr) {
	h.log(ctx, msg, attrs...)
}

func (h *Handler) LogError(ctx context.Context, msg string, err error) error {
	return h.error(ctx, msg, err)
}

func logError(h *Handler, o *Config, ctx context.Context, msg string, err error) error {
	attrs := h.Filter(ctx, slog.Any("error", err))
	h.logger.LogAttrs(ctx, o.errorLevel.Level(), msg, attrs...)
	return err
}

func noopError(ctx context.Context, msg string, err error) error {
	return err
}

func makeHandle(o *Config) *Handler {
	// Define a filter function to modify log attributes based on the FilterAttrs option.
	if o.logger == nil {
		o.logger = slog.Default()
	}

	h := Handler{
		logger: o.logger,
		filter: o.filter,
		trace:  o.trace,
		error:  noopError,
	}

	// Return a configured logging handler.
	return h.init(o)
}
