package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundHandlerMiddleware(t *testing.T) {
	// Create a test router
	mux := http.NewServeMux()
	mux.HandleFunc("/example", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is an example route"))
	})

	// Test a valid route
	req := httptest.NewRequest(http.MethodGet, "/example", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "This is an example route", w.Body.String())

	// Test an invalid route
	req = httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
