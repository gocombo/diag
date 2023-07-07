package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gocombo/diag"
	"github.com/stretchr/testify/assert"
)

func TestHttpLogMiddleware(t *testing.T) {
	t.Run("should log start and end of request", func(t *testing.T) {
		var output bytes.Buffer
		outputWriter := bufio.NewWriter(&output)

		rootCtx := diag.RootContext(
			diag.NewRootContextParams().WithOutput(outputWriter),
		)
		method := fake.Internet().HTTPMethod()
		path := "/" + fake.Internet().Slug()
		req := httptest.NewRequest(method, path, http.NoBody).WithContext(rootCtx)
		res := httptest.NewRecorder()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		wrapped := BuildHandler(h, NewHttpLogMiddleware())
		wrapped.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		outputWriter.Flush()
		outputLines := strings.Split(strings.Trim(output.String(), "\n"), "\n")
		assert.Equal(t, 2, len(outputLines))
		var reqStart map[string]interface{}
		if err := json.Unmarshal([]byte(outputLines[0]), &reqStart); !assert.NoError(t, err) {
			return
		}
		data := reqStart["data"].(map[string]interface{})
		assert.Equal(t, method, data["method"])
		assert.Equal(t, "example.com", data["host"])
		assert.Equal(t, path, data["url"])
		assert.Equal(t, path, data["path"])

		assert.Equal(t,
			fmt.Sprintf("BEGIN REQ: %s %s", method, path),
			reqStart["msg"],
		)
	})
}
