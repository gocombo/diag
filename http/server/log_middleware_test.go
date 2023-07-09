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
		userAgent := fake.UserAgent().UserAgent()
		method := fake.Internet().HTTPMethod()
		path := "/" + fake.Internet().Slug()
		req := httptest.NewRequest(method, path, http.NoBody).WithContext(rootCtx)
		req.Header.Add("User-Agent", userAgent)
		req.Header.Add("X-Test-Header-1", fake.Internet().Slug())
		req.Header.Add("X-Test-Header", fake.Internet().Slug())
		query := req.URL.Query()
		query.Add("test", fake.Internet().Slug())
		query.Add("test2", fake.Internet().Slug())
		req.URL.RawQuery = query.Encode()
		res := httptest.NewRecorder()

		var wantReqHeaders http.Header
		wantStatus := fake.IntBetween(200, 500)
		wantResHeaders := map[string]string{
			"X-Test-Header-1": fake.Internet().Slug(),
			"X-Test-Header-2": fake.Internet().Slug(),
			"X-Test-Header-3": fake.Internet().Slug(),
		}
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wantReqHeaders = r.Header
			for k, v := range wantResHeaders {
				w.Header().Set(k, v)
			}
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
		startData := reqStart["data"].(map[string]interface{})
		assert.Equal(t, method, startData["method"])
		assert.Equal(t, req.URL.Path+"?"+req.URL.RawQuery, startData["url"])
		gotStartHeaders := startData["headers"].(map[string]interface{})
		for k, v := range flattenAndObfuscate(wantReqHeaders, nil) {
			assert.Equal(t, v, gotStartHeaders[k])
		}
		gotQuery := startData["query"].(map[string]interface{})
		for k, v := range flattenAndObfuscate(query, nil) {
			assert.Equal(t, v, gotQuery[k])
		}
		assert.NotEmpty(t, startData["memoryUsageMb"])

		assert.Equal(t,
			fmt.Sprintf("BEGIN REQ: %s %s", method, path),
			reqStart["msg"],
		)

		var reqEnd map[string]interface{}
		if err := json.Unmarshal([]byte(outputLines[1]), &reqEnd); !assert.NoError(t, err) {
			return
		}

		endData := reqEnd["data"].(map[string]interface{})
		assert.Equal(t, float64(wantStatus), endData["statusCode"])
		assert.NotZero(t, endData["durationSec"])
		assert.NotZero(t, endData["memoryUsageMb"])
		assert.Equal(t, userAgent, endData["userAgent"])
		gotEndHeaders := endData["headers"].(map[string]interface{})
		for k, v := range wantResHeaders {
			assert.Equal(t, v, gotEndHeaders[k])
		}

		assert.Equal(t,
			fmt.Sprintf("END REQ: %v - %s", wantStatus, path),
			reqEnd["msg"],
		)
	})

	t.Run("should obfuscate sensitive headers", func(t *testing.T) {
		var output bytes.Buffer
		outputWriter := bufio.NewWriter(&output)

		rootCtx := diag.RootContext(
			diag.NewRootContextParams().WithOutput(outputWriter),
		)
		req := httptest.NewRequest("GET", "/", http.NoBody).WithContext(rootCtx)
		req.Header.Add("Authorization", fake.Lorem().Sentence(20))
		req.Header.Add("Proxy-Authorization", fake.Lorem().Sentence(20))
		res := httptest.NewRecorder()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})
		wrapped := BuildHandler(h, NewHttpLogMiddleware())
		wrapped.ServeHTTP(res, req)
		assert.Equal(t, 200, res.Code)

		outputWriter.Flush()
		outputLines := strings.Split(strings.Trim(output.String(), "\n"), "\n")
		assert.Equal(t, 2, len(outputLines))

		var reqStart map[string]interface{}
		if err := json.Unmarshal([]byte(outputLines[0]), &reqStart); !assert.NoError(t, err) {
			return
		}
		startData := reqStart["data"].(map[string]interface{})
		gotStartHeaders := startData["headers"].(map[string]interface{})
		assert.Contains(t, gotStartHeaders["Authorization"], "*obfuscated, length=")
		assert.Contains(t, gotStartHeaders["Proxy-Authorization"], "*obfuscated, length=")
	})
}
