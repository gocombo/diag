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
		req.Header.Add("X-Test-Header-1", fake.Internet().Slug())
		req.Header.Add("X-Test-Header", fake.Internet().Slug())
		query := req.URL.Query()
		query.Add("test", fake.Internet().Slug())
		query.Add("test2", fake.Internet().Slug())
		req.URL.RawQuery = query.Encode()
		res := httptest.NewRecorder()

		var wantReqHeaders http.Header
		wantStatus := fake.IntBetween(200, 500)
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wantReqHeaders = r.Header
			w.WriteHeader(wantStatus)
		})
		wrapped := BuildHandler(h, NewHttpLogMiddleware())
		wrapped.ServeHTTP(res, req)
		assert.Equal(t, wantStatus, res.Code)

		outputWriter.Flush()
		outputLines := strings.Split(strings.Trim(output.String(), "\n"), "\n")
		assert.Equal(t, 2, len(outputLines))

		var reqStart map[string]interface{}
		if err := json.Unmarshal([]byte(outputLines[0]), &reqStart); !assert.NoError(t, err) {
			return
		}
		data := reqStart["data"].(map[string]interface{})
		assert.Equal(t, method, data["method"])
		assert.Equal(t, req.URL.Path+"?"+req.URL.RawQuery, data["url"])
		gotHeaders := data["headers"].(map[string]interface{})
		for k, v := range flattenAndObfuscate(wantReqHeaders, nil) {
			assert.Equal(t, v, gotHeaders[k])
		}
		gotQuery := data["query"].(map[string]interface{})
		for k, v := range flattenAndObfuscate(query, nil) {
			assert.Equal(t, v, gotQuery[k])
		}
		assert.NotEmpty(t, data["memoryUsageMb"])

		assert.Equal(t,
			fmt.Sprintf("BEGIN REQ: %s %s", method, path),
			reqStart["msg"],
		)
	})
}
