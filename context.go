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
	ChildLogger(logger LevelLogger, diagOpts DiagOpts) LevelLogger
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
			Entries:       map[string]string{},
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

// WithDiagEntries will assign diag entries for the root context
// the root entries will be propagated to all forked/diagified contexts
func (c *RootContextParams) WithDiagEntries(entries map[string]string) *RootContextParams {
	for k, v := range entries {
		c.DiagData.Entries[k] = v
	}
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
		panic(fmt.Errorf("context does not contain a logger"))
	}
	return logger
}

// DiagData returns the diag data from the context
// A copy of a diag data is returned so can not be mutated
func DiagData(ctx context.Context) ContextDiagData {
	diagData, ok := ctx.Value(contextKeyDiagData).(ContextDiagData)
	if !ok {
		panic(fmt.Errorf("context does not contain diag data"))
	}
	sourceEntries := diagData.Entries
	diagData.Entries = make(map[string]string, len(sourceEntries))
	for k, v := range sourceEntries {
		diagData.Entries[k] = v
	}
	return diagData
}

func getLoggerFactory(ctx context.Context) LoggerFactory {
	loggerFactory, ok := ctx.Value(contextKeyLoggerFactory).(LoggerFactory)
	if !ok {
		panic(fmt.Errorf("context does not contain a logger factory"))
	}
	return loggerFactory
}

type DiagOpts struct {
	Level *LogLevel

	DiagData ContextDiagData
}

type DiagContextOption func(opts *DiagOpts)

func WithLogLevel(level LogLevel) DiagContextOption {
	return func(opts *DiagOpts) {
		opts.Level = &level
	}
}

func WithCorrelationID(correlationID string) DiagContextOption {
	return func(opts *DiagOpts) {
		opts.DiagData.CorrelationID = correlationID
	}
}

func WithAppendDiagEntries(entries map[string]string) DiagContextOption {
	return func(opts *DiagOpts) {
		for k, v := range entries {
			opts.DiagData.Entries[k] = v
		}
	}
}

// DiagifyContext creates a child context with diag data taken from
// the diagContext and optionally adjusted via opts.
func DiagifyContext(
	parentCtx context.Context,
	diagContext context.Context,
	opts ...DiagContextOption,
) context.Context {
	// TODO: Check if it's not mutated
	rootDiagData := DiagData(diagContext)
	diagOpts := DiagOpts{
		DiagData: rootDiagData,
	}
	for _, opt := range opts {
		opt(&diagOpts)
	}

	loggerFactory := getLoggerFactory(diagContext)
	log := loggerFactory.ChildLogger(Log(diagContext), diagOpts)

	resultCtx := context.WithValue(parentCtx, contextKeyLogger, log)
	resultCtx = context.WithValue(resultCtx, contextKeyDiagData, diagOpts.DiagData)
	resultCtx = context.WithValue(resultCtx, contextKeyLoggerFactory, loggerFactory)

	return resultCtx
}

// ForkContext creates a copy of a given diag context. Will only copy diag and logger data.
// The context created is not a child of the original context
// so signals will not be propagated.
func ForkContext(ctx context.Context, opts ...DiagContextOption) context.Context {
	return DiagifyContext(context.Background(), ctx, opts...)
}
