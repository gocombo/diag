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
		logData = logData.Interface("headers", internal.FlattenAndObfuscate(res.Header, internal.DefaultObfuscatedHeaders))
	}

	levelLog.
		WithData(logData).
		Msgf("COMPLETE SENDING REQ: %d - %v", resCode, req.URL.String())
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
		writeLogEndMessage(log, reqDuration, req, res)
		return res, err
	})
}
