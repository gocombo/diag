package client

import (
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
	req *http.Request,
	res *http.Response,
	obfuscateHeaders []string,
) {
	var levelLog diag.LogLevelEvent
	var resCode int
	if res != nil {
		resCode = res.StatusCode
	} else {
		// This is a special case where the request was never sent
		// 599 is a closest unofficial code to indicate this
		resCode = 599
	}
	if resCode >= 400 {
		levelLog = log.Warn()
	} else {
		levelLog = log.Info()
	}

	logData := log.NewData().
		Float64("durationSec", durationSec).
		Int("statusCode", resCode)
	if res != nil {
		logData = logData.Interface("headers", internal.FlattenAndObfuscate(res.Header, obfuscateHeaders))
	}

	levelLog.
		WithData(logData).
		Msgf("COMPLETE SENDING REQ: %d - %v", resCode, req.URL.String())
}

type transportCfg struct {
	obfuscateHeaders []string
}

// TransportOption is a functional option for configuring the transport
type TransportOption func(*transportCfg)

// WithObfuscateHeaders sets the headers that should be obfuscated in the logs
func WithObfuscateHeaders(headers []string) TransportOption {
	return func(cfg *transportCfg) {
		lowercaseHeaders := make([]string, len(headers))
		for i, header := range headers {
			lowercaseHeaders[i] = strings.ToLower(header)
		}
		cfg.obfuscateHeaders = append(cfg.obfuscateHeaders, lowercaseHeaders...)
	}
}

// NewTransport returns a wrapped http.RoundTripper that will produce
// diag like logs for each request
// TODO: Unit test this
func NewTransport(target http.RoundTripper, opts ...TransportOption) http.RoundTripper {
	cfg := &transportCfg{
		obfuscateHeaders: internal.DefaultObfuscatedHeaders,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return roundTripperFn(func(req *http.Request) (*http.Response, error) {
		log := diag.Log(req.Context())
		log.Info().WithData(
			log.NewData().
				Interface("headers", internal.FlattenAndObfuscate(req.Header, cfg.obfuscateHeaders)).
				Str("method", req.Method).
				Str("url", req.URL.String()),
		).Msgf("START SENDING REQ: %s %s", strings.ToUpper(req.Method), req.URL)
		startedAt := time.Now()
		res, err := target.RoundTrip(req)
		reqDuration := time.Since(startedAt).Seconds()
		writeLogEndMessage(log, reqDuration, req, res, cfg.obfuscateHeaders)
		return res, err
	})
}
