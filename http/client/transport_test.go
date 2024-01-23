package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gocombo/diag"
	"github.com/gocombo/diag/http/internal"
	"github.com/gocombo/diag/http/internal/testing/httptst"
	"github.com/gocombo/diag/http/internal/testing/testrand"
	"github.com/stretchr/testify/assert"
)

func unmarshalLogLines(
	t *testing.T,
	outputWriter *bufio.Writer,
	output *bytes.Buffer,
) ([]map[string]interface{}, bool) {
	outputWriter.Flush()
	outputLines := strings.Split(strings.Trim(output.String(), "\n"), "\n")

	result := make([]map[string]interface{}, len(outputLines))
	for i, line := range outputLines {
		var logLine map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logLine); !assert.NoError(t, err) {
			return nil, false
		}
		result[i] = logLine
	}
	return result, true
}

func TestTransport(t *testing.T) {
	fake := testrand.Faker()

	t.Run("should log start and end of request", func(t *testing.T) {
		var output bytes.Buffer
		outputWriter := bufio.NewWriter(&output)

		rootCtx := diag.RootContext(
			diag.NewRootContextParams().WithOutput(outputWriter),
		)

		req := httptst.RandomHttpReq(testrand.Faker(), rootCtx)
		wantRes := &http.Response{
			StatusCode: 200,
			Body:       http.NoBody,
			Request:    req,
		}
		transport := NewTransport(roundTripperFn(func(r *http.Request) (*http.Response, error) {
			return wantRes, nil
		}))
		res, err := transport.RoundTrip(req)
		if !assert.NoError(t, err) {
			return
		}
		defer res.Body.Close()
		assert.Equal(t, wantRes, res)

		logLines, ok := unmarshalLogLines(t, outputWriter, &output)
		if !ok {
			return
		}
		assert.Equal(t, 2, len(logLines))

		reqStart := logLines[0]
		assert.Equal(t,
			fmt.Sprintf("START SENDING REQ: %v %v", strings.ToUpper(req.Method), req.URL),
			reqStart["msg"],
		)
		reqStartData := reqStart["data"].(map[string]interface{})
		assert.Equal(t, req.Method, reqStartData["method"])
		assert.Equal(t, req.URL.Path+"?"+req.URL.RawQuery, reqStartData["url"])
		gotStartHeaders := reqStartData["headers"].(map[string]interface{})
		for k, v := range internal.FlattenAndObfuscate(req.Header, internal.DefaultObfuscatedHeaders) {
			assert.Equal(t, v, gotStartHeaders[k])
		}

		reqEnd := logLines[1]
		assert.Equal(t,
			fmt.Sprintf("COMPLETE SENDING REQ: %d - %v", res.StatusCode, res.Request.URL.String()),
			reqEnd["msg"],
		)
		reqEndData := reqEnd["data"].(map[string]interface{})
		assert.Equal(t, float64(res.StatusCode), reqEndData["statusCode"])
		assert.NotZero(t, reqEndData["durationSec"])
		gotEndHeaders := reqEndData["headers"].(map[string]interface{})
		for k, v := range internal.FlattenAndObfuscate(res.Header, nil) {
			assert.Equal(t, v, gotEndHeaders[k])
		}
	})
	t.Run("should log non success responses with warn", func(t *testing.T) {
		var output bytes.Buffer
		outputWriter := bufio.NewWriter(&output)

		rootCtx := diag.RootContext(
			diag.NewRootContextParams().WithOutput(outputWriter),
		)

		req := httptst.RandomHttpReq(testrand.Faker(), rootCtx)
		wantRes := &http.Response{
			StatusCode: fake.IntBetween(400, 599),
			Body:       http.NoBody,
			Request:    req,
		}
		transport := NewTransport(roundTripperFn(func(r *http.Request) (*http.Response, error) {
			return wantRes, nil
		}))
		res, _ := transport.RoundTrip(req)
		defer res.Body.Close()

		logLines, ok := unmarshalLogLines(t, outputWriter, &output)
		if !ok {
			return
		}
		assert.Equal(t, 2, len(logLines))

		reqStart := logLines[0]
		assert.Equal(t,
			fmt.Sprintf("START SENDING REQ: %v %v", strings.ToUpper(req.Method), req.URL),
			reqStart["msg"],
		)
		assert.Equal(t, "info", reqStart["level"])

		reqEnd := logLines[1]
		assert.Equal(t, "warn", reqEnd["level"])
	})
	t.Run("should handle request errors", func(t *testing.T) {
		var output bytes.Buffer
		outputWriter := bufio.NewWriter(&output)

		rootCtx := diag.RootContext(
			diag.NewRootContextParams().WithOutput(outputWriter),
		)

		req := httptst.RandomHttpReq(testrand.Faker(), rootCtx)
		wantErr := errors.New(fake.Lorem().Word())
		transport := NewTransport(roundTripperFn(func(r *http.Request) (*http.Response, error) {
			return nil, wantErr
		}))
		_, err := transport.RoundTrip(req)
		assert.Equal(t, wantErr, err)

		logLines, ok := unmarshalLogLines(t, outputWriter, &output)
		if !ok {
			return
		}
		assert.Equal(t, 2, len(logLines))

		reqStart := logLines[0]
		assert.Equal(t,
			fmt.Sprintf("START SENDING REQ: %v %v", strings.ToUpper(req.Method), req.URL),
			reqStart["msg"],
		)
		assert.Equal(t, "info", reqStart["level"])

		reqEnd := logLines[1]
		assert.Equal(t,
			fmt.Sprintf("COMPLETE SENDING REQ: 599 - %v", req.URL.String()),
			reqEnd["msg"],
		)
		reqEndData := reqEnd["data"].(map[string]interface{})
		assert.NotZero(t, reqEndData["durationSec"])
	})
}
