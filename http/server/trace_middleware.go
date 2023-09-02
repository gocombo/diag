package server

import (
	"context"
	"net/http"

	"github.com/gocombo/diag"
	"github.com/gofrs/uuid"
)

type httpTraceMiddlewareOpts struct {
	uuidFn func() string
}

type HttpTraceMiddlewareOpt func(opts *httpTraceMiddlewareOpts)

func NewHttpTraceMiddleware(rootCtx context.Context, opts ...HttpTraceMiddlewareOpt) func(http.Handler) http.Handler {
	cfg := httpTraceMiddlewareOpts{
		uuidFn: func() string {
			return uuid.Must(uuid.NewV4()).String()
		},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			correlationID := req.Header.Get("x-correlation-id")
			if correlationID == "" {
				correlationID = cfg.uuidFn()
			}
			reqCtx := diag.DiagifyContext(req.Context(), rootCtx, diag.WithCorrelationID(correlationID))
			next.ServeHTTP(w, req.WithContext(reqCtx))
		})
	}
}
