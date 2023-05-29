package diag

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/gofrs/uuid"
)

type contextKey string

const (
	contextKeyLogger        = contextKey("gocombo.diag.context-key.root-logger")
	contextKeyDiagData      = contextKey("gocombo.diag.context-key.diag-data")
	contextKeyLoggerFactory = contextKey("gocombo.diag.context-key.logger-factory")
)

type LoggerFactory interface {
	NewLogger(p *RootContextParams) LevelLogger
	ForkLogger(logger LevelLogger, opts ForkOpts) LevelLogger
}

type RootContextParams struct {
	Pretty        bool
	Out           io.Writer
	LogLevel      LogLevel
	LoggerFactory LoggerFactory
	DiagData      ContextDiagData
}

// ContextDiagData is a structure that can be used to hold various
// diagnostic data that will be added to the log entries
type ContextDiagData struct {
	// CorrelationID will be added to each log entry to allow correlating logs
	CorrelationID string

	// Additional application specific entries to include with the log
	Entries map[string]string
}

func NewRootContextParams() *RootContextParams {
	rootContextID := uuid.Must(uuid.NewV4()).String()
	return &RootContextParams{
		Pretty:        false,
		Out:           os.Stderr,
		LogLevel:      LogLevelDebugValue,
		LoggerFactory: zerologLoggerFactory{},
		DiagData: ContextDiagData{
			CorrelationID: rootContextID,
		},
	}
}

func RootContext(p *RootContextParams) context.Context {
	logger := p.LoggerFactory.NewLogger(p)
	ctx := context.WithValue(context.Background(), contextKeyLogger, logger)
	ctx = context.WithValue(ctx, contextKeyDiagData, p.DiagData)
	ctx = context.WithValue(ctx, contextKeyLoggerFactory, p.LoggerFactory)
	return ctx
}

// WithCorrelationID allows setting a predefined root correlation id
// If not set an autogenerated one will be used
func (c *RootContextParams) WithCorrelationID(value string) *RootContextParams {
	c.DiagData.CorrelationID = value
	return c
}

func (c *RootContextParams) WithPretty(value bool) *RootContextParams {
	c.Pretty = value
	return c
}

func (c *RootContextParams) WithOutput(output io.Writer) *RootContextParams {
	c.Out = output
	return c
}

func (c *RootContextParams) WithLogLevel(logLevel LogLevel) *RootContextParams {
	c.LogLevel = logLevel
	return c
}

// WithLoggerFactory allows using a custom logger factory with alternative implementation
func (c *RootContextParams) WithLoggerFactory(factory LoggerFactory) *RootContextParams {
	c.LoggerFactory = factory
	return c
}

// Log returns the logger from the context
// Obtained instance can be used for general purpose logging
func Log(ctx context.Context) LevelLogger {
	logger, ok := ctx.Value(contextKeyLogger).(LevelLogger)
	if !ok {
		// TODO: Better message
		panic(fmt.Errorf("context does not contain a logger"))
	}
	return logger
}

// DiagData returns the diag data from the context
func DiagData(ctx context.Context) ContextDiagData {
	diagData, ok := ctx.Value(contextKeyDiagData).(ContextDiagData)
	if !ok {
		panic(fmt.Errorf("context does not contain diag data"))
	}
	return diagData
}

type ForkOpts struct {
	Level *LogLevel

	// Use a new correlation ID for the child logger
	CorrelationID *string

	// AppendEntries will be added to each log entry in addition to existing entries present
	AppendDiagEntries map[string]string
}

type ForkContextOption func(opts *ForkOpts)

func ForkWithLogLevel(level LogLevel) ForkContextOption {
	return func(opts *ForkOpts) {
		opts.Level = &level
	}
}

func ForkWithCorrelationID(correlationID string) ForkContextOption {
	return func(opts *ForkOpts) {
		opts.CorrelationID = &correlationID
	}
}

func ForkWithAppendDiagEntries(entries map[string]string) ForkContextOption {
	return func(opts *ForkOpts) {
		opts.AppendDiagEntries = entries
	}
}

// ForkContext creates a copy of a given context. Will only copy diag and logger data.
// The context created is not a child of the original context
// so signals will not be propagated.
func ForkContext(ctx context.Context, opts ...ForkContextOption) context.Context {
	var forkOpts ForkOpts
	for _, opt := range opts {
		opt(&forkOpts)
	}

	loggerFactory, ok := ctx.Value(contextKeyLoggerFactory).(LoggerFactory)
	if !ok {
		panic(fmt.Errorf("context does not contain a logger factory"))
	}

	diagData := DiagData(ctx)
	if forkOpts.CorrelationID != nil {
		diagData.CorrelationID = *forkOpts.CorrelationID
	}
	if len(forkOpts.AppendDiagEntries) > 0 {
		if diagData.Entries == nil {
			diagData.Entries = make(map[string]string)
		}
		for k, v := range forkOpts.AppendDiagEntries {
			diagData.Entries[k] = v
		}
	}

	log := loggerFactory.ForkLogger(Log(ctx), forkOpts)

	forkedCtx := context.WithValue(ctx, contextKeyLogger, log)
	forkedCtx = context.WithValue(forkedCtx, contextKeyDiagData, diagData)
	return forkedCtx
}
