package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gocombo/diag"
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
	})
}
