// Copyright (c) 2024 OrigAdmin. All rights reserved.

// Package entslog for entgo.io/ent
package entslog

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type (
	// TraceFunc generates a unique identifier for a log entry to string.
	TraceFunc func(context.Context) string
	// FilterAttrs defines a function to filter out attributes from log entries.
	FilterAttrs func(context.Context, ...slog.Attr) []slog.Attr
	// Config defines configuration configs for the logging handler.
	Config struct {
		handleError bool         // HandleError determines whether errors encountered during logging are handled.
		logger      *slog.Logger // Logger specifies the logger to be used for logging.
		level       slog.Leveler // DefaultLevel specifies the default log level for messages.
		errorLevel  slog.Leveler // ErrorLevel specifies the log level for error messages.
		trace       TraceFunc    // GenerateID is a function to generate unique IDs for log entries.
		filter      FilterAttrs  // Filters specifies the set of attributes to filter out from logged messages.
	}
	// Option is a type alias for the settings.Setting type.
	Option func(*Config)
)

// defaultConfig provides the default configuration configs for the logging handler.
var defaultConfig = Config{
	logger:      slog.Default(),  // Defaults to the default logger.
	level:       slog.LevelInfo,  // Defaults to Info level.
	errorLevel:  slog.LevelError, // Defaults to Error level.
	handleError: true,            // Defaults to handling errors.
	filter:      emptyFilter,     // Defaults to no filtering.
	trace:       traceUUID,       // Uses the package-level trace function to generate log entry IDs by default.
}

func emptyFilter(_ context.Context, attrs ...slog.Attr) []slog.Attr {
	return attrs
}

// traceUUID generates a unique identifier for a log entry using UUIDs.
func traceUUID(ctx context.Context) string {
	return uuid.Must(uuid.NewRandom()).String()
}

// WithDefaultLevel sets the default log level for the given logging configs.
//
// - `level`: The log level to be set as the default.
//
// Returns a function that accepts an `*Config` parameter, modifies it by setting the default log level,
// and returns the updated `*Config` pointer.
func WithDefaultLevel(level slog.Leveler) Option {
	return func(o *Config) {
		o.level = level
	}
}

// WithErrorLevel sets the error log level and enables handling of errors for the given logging configs.
//
// - `level`: The log level to be set for error logging.
//
// Returns a function that accepts an `*Config` parameter, modifies it by setting the error log level
// and enabling error handling, then returns the updated `*Config` pointer.
func WithErrorLevel(level slog.Leveler) Option {
	return func(config *Config) {
		config.errorLevel = level
		config.handleError = true
	}
}

// WithError explicitly enables or disables error handling for the given logging configs.
//
// - `handleError`: A boolean indicating whether to enable (true) or disable (false) error handling.
//
// Returns a function that accepts an `*Config` parameter, modifies it by setting the error handling flag,
// and returns the updated `*Config` pointer.
func WithError() Option {
	return func(config *Config) {
		config.handleError = true
	}
}

// WithFilter specifies a list of attribute names to filter out from logged messages for the given logging configs.
//
// - `attrs`: A variadic list of strings representing attribute names to be filtered.
//
// Returns a function that accepts an `*Config` parameter, modifies it by setting the list of filtered attributes,
// and returns the updated `*Config` pointer.
func WithFilter(filter func(context.Context, ...slog.Attr) []slog.Attr) Option {
	return func(config *Config) {
		config.filter = filter
	}
}

// WithTrace assigns a custom ID generation function for the given logging configs.
// This function will be used to generate unique IDs for log entries within a given context.
//
// - `trace`: A function that accepts a `context.Context` and returns a string representing the generated ID.
//
// Returns a function that accepts an `*Config` parameter, modifies it by setting the custom ID generation function,
// and returns the updated `*Config` pointer.
func WithTrace(trace func(context.Context) string) Option {
	return func(config *Config) {
		config.trace = trace
	}
}

// WithLogger specifies the logger to be used for logging.
// If not specified, the default logger will be used.
//
// - `logger`: The logger to be used for logging.
//
// Returns a function that accepts an `*Config` parameter, modifies it by setting the logger,
// and returns the updated `*Config` pointer.
func WithLogger(logger *slog.Logger) Option {
	return func(config *Config) {
		config.logger = logger
	}
}

// make configures and returns a new logging handler based on the provided configs.
