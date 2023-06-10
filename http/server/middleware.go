package server

import "net/http"

// MiddlewareFunc is an http middleware function
type MiddlewareFunc func(http.Handler) http.Handler

// BuildHandler hook up the http.Handler middleware chain
func BuildHandler(h http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
