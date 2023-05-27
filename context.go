package diag

import (
	"context"
	"io"
)

type contextKey string

const (
	contextKeyLogger = contextKey("gocombo.diag.context-key.root-logger")
)

type RootContextParams struct {
	correlationID string
	pretty        bool
	out           io.Writer
	logLevel      LogLevel
}

func NewRootContextParams() *RootContextParams {
	return &RootContextParams{
		logLevel: LogLevelDebugValue,
	}
}

func RootContext(p *RootContextParams) context.Context {
	// TODO: Panic if no correlation id
	logger := newZerologLogger(p)
	return context.WithValue(context.Background(), contextKeyLogger, logger)
}

func (c *RootContextParams) WithCorrelationID(value string) *RootContextParams {
	c.correlationID = value
	return c
}

func (c *RootContextParams) WithPretty(value bool) *RootContextParams {
	c.pretty = value
	return c
}

func (c *RootContextParams) WithOutput(output io.Writer) *RootContextParams {
	c.out = output
	return c
}

func (c *RootContextParams) WithLogLevel(logLevel LogLevel) *RootContextParams {
	c.logLevel = logLevel
	return c
}
