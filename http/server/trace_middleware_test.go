package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocombo/diag"
	"github.com/stretchr/testify/assert"
)

func TestHttpTraceMiddleware(t *testing.T) {
	t.Run("setup correlationId from header", func(t *testing.T) {
		wantCorrelationId := fake.UUID().V4()
		req := httptest.NewRequest("GET", "/something", http.NoBody)
		req.Header.Set("X-Correlation-Id", wantCorrelationId)
		res := httptest.NewRecorder()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			diagData := diag.DiagData(r.Context())
			if err := json.NewEncoder(w).Encode(diagData); err != nil {
				panic(err)
			}
		})
		rootCtx := diag.RootContext(diag.NewRootContextParams())
		wrapped := BuildHandler(h, NewHttpTraceMiddleware(rootCtx))
		wrapped.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
		assert.NotEmpty(t, res.Body.String())
		var gotDiagData diag.ContextDiagData
		if err := json.NewDecoder(res.Body).Decode(&gotDiagData); !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, wantCorrelationId, gotDiagData.CorrelationID)
	})
}
