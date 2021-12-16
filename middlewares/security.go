package middlewares

import (
	"net/http"
)

var securityHeaders = map[string]string{
	"Cross-Origin-Opener-Policy": "same-origin",
	"Content-Security-Policy":    "default-src 'self' 'unsafe-inline' 'unsafe-eval'; img-src 'self' https: data:; form-action 'self'; block-all-mixed-content;",
	"X-Frame-Options":            "DENY",
	"X-Content-Type-Options":     "nosniff",
}

// SecurityMiddleware is a handler to add some basic security headers to responses
type SecurityMiddleware struct {
	handler http.Handler
}

func NewSecurityMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &SecurityMiddleware{h}
	}
}

func (f *SecurityMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for k, v := range securityHeaders {
		if w.Header().Get(k) == "" {
			w.Header().Set(k, v)
		}
	}
	f.handler.ServeHTTP(w, r)
}
