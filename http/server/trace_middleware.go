package server

import (
	"context"
	"net/http"

	"github.com/gocombo/diag"
	"github.com/gofrs/uuid"
)

func NewHttpTraceMiddleware(rootCtx context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			correlationID := req.Header.Get("x-correlation-id")
			if correlationID == "" {
				correlationID = uuid.Must(uuid.NewV4()).String()
			}
			reqCtx := diag.DiagifyContext(req.Context(), rootCtx, diag.WithCorrelationID(correlationID))
			next.ServeHTTP(w, req.WithContext(reqCtx))
		})
	}
}
