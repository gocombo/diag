package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildHandler(t *testing.T) {
	t.Run("wrap handler with middleware", func(t *testing.T) {
		calls := []string{}

		req, err := http.NewRequest("GET", "/something", http.NoBody)
		if err != nil {
			panic(err)
		}
		res := httptest.NewRecorder()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "handler")
		})
		makeTestMw := func(name string) MiddlewareFunc {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls = append(calls, name)
					next.ServeHTTP(w, r)
				})
			}
		}
		wrapped := BuildHandler(h, makeTestMw("mw1"), makeTestMw("mw2"), makeTestMw("mw3"))
		wrapped.ServeHTTP(res, req)
		assert.Equal(t, []string{
			"mw1", "mw2", "mw3", "handler",
		}, calls)
	})
}
