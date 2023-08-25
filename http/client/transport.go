package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gocombo/diag"
	"github.com/gocombo/diag/http/internal"
)

type roundTripperFn func(*http.Request) (*http.Response, error)

func (fn roundTripperFn) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func writeLogEndMessage(
	log diag.LevelLogger,
	durationSec float64,
	res *http.Response,
) {
	log.Info().
		WithData(
			log.NewData().
				Float64("durationSec", durationSec).
				Interface("headers", internal.FlattenAndObfuscate(res.Header, internal.DefaultObfuscatedHeaders)).
				Int("statusCode", res.StatusCode),
		).Msgf("COMPLETE SENDING REQ: %d - %v", res.StatusCode, res.Request.URL.String())
}

// NewTransport returns a wrapped http.RoundTripper that will produce
// diag like logs for each request
// TODO: Unit test this
func NewTransport(target http.RoundTripper) http.RoundTripper {
	return roundTripperFn(func(req *http.Request) (*http.Response, error) {
		log := diag.Log(req.Context())
		log.Info().WithData(
			log.NewData().
				Interface("headers", internal.FlattenAndObfuscate(req.Header, internal.DefaultObfuscatedHeaders)).
				Str("method", req.Method).
				Str("url", req.URL.String()),
		).Msgf("START SENDING REQ: %s %s", strings.ToUpper(req.Method), req.URL)
		startedAt := time.Now()
		res, err := target.RoundTrip(req)
		reqDuration := time.Since(startedAt).Seconds()
		writeLogEndMessage(log, reqDuration, res)

		// We let consumer decide redirects handling
		if res.StatusCode > 399 {
			var resData []byte
			if res.Body != nil {
				defer res.Body.Close()
				// Read and discard body to avoid memory leaks
				resData, err = io.ReadAll(res.Body)
				if err != nil {
					log.Warn().WithError(err).Msg("Failed to read error body")
				}
			}

			return nil, fmt.Errorf("failed to send request: %v - %s", res.Status, resData)
		}

		return res, err
	})
}
