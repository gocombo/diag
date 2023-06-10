package server

import (
	"math"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/gocombo/diag"
)

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func runtimeMemMb() float64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return math.Round(float64(memStats.Alloc)/1024.0/1024.0*1000) / 1000
}

// WithHTTPLog log web transaction, it should be placed last in the middleware chain, to measure the latency of route handler logic
func NewHttpLogMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			log := diag.Log(req.Context())

			path := req.URL.Path

			ip, port, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				log.Warn().WithError(err).Msgf("Failed to parse remote addr: %q", req.RemoteAddr)
				ip = req.RemoteAddr
			}

			method := req.Method

			log.Info().
				WithDataFn(func(data diag.MsgData) {
					data.
						Str("method", method).
						Str("host", req.Host).
						Str("url", req.URL.RequestURI()).
						Str("path", req.URL.Path).
						Str("userAgent", req.UserAgent()).
						// Str("headers",       flattenHeaders(req.Header)).
						// Str("query",         flattenQuery(req.URL.Query())).
						Str("remoteAddress", ip).
						Str("remotePort", port).
						Float64("memoryUsageMb", runtimeMemMb())
				}).
				Msgf("BEGIN REQ: %s %s", method, path)

			rw := &responseWrapper{ResponseWriter: w}
			start := time.Now()

			panics := true
			defer func() {
				stop := time.Now()

				status := rw.statusCode

				if panics {
					status = 500
				}

				// Status may not be set by the next chain so we use 200 for such cases
				if status == 0 {
					status = 200
				}

				log.Info().
					WithDataFn(func(data diag.MsgData) {
						data.Int("statusCode", status)
						// data.Str("headers", flattenHeaders(w.Header()))
						data.Float64("duration", stop.Sub(start).Seconds())
						data.Float64("memoryUsageMb", runtimeMemMb())
						data.Str("userAgent", req.UserAgent())
					}).
					Msgf("END REQ: %v - %v", status, path)
			}()

			next.ServeHTTP(rw, req)
			panics = false
		})
	}
}