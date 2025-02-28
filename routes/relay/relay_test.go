package relay

import (
    "net/http"
    "net/http/httptest"
    "testing"
)


// Test generated using Keploy
func TestAny_ValidTargetUrlHeader_ProxiesRequest(t *testing.T) {
    handler := NewRelayHandler()
    targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("proxied response"))
    }))
    defer targetServer.Close()

    req, err := http.NewRequest("GET", "/", nil)
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }
    req.Header.Set(targetUrlHeader, targetServer.URL)

    rr := httptest.NewRecorder()
    handler.Any(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("Expected status code %v, got %v", http.StatusOK, rr.Code)
    }
    if rr.Body.String() != "proxied response" {
        t.Errorf("Expected body %v, got %v", "proxied response", rr.Body.String())
    }
}


// Test generated using Keploy
func TestFilteringMiddleware_InvalidPath_ReturnsForbidden(t *testing.T) {
    middleware := newFilteringMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req, err := http.NewRequest("GET", "/invalid/path", nil)
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }
    req.Header.Set(targetUrlHeader, "http://example.com/invalid/path")

    rr := httptest.NewRecorder()
    middleware.ServeHTTP(rr, req)

    if rr.Code != http.StatusForbidden {
        t.Errorf("Expected status code %v, got %v", http.StatusForbidden, rr.Code)
    }
}


// Test generated using Keploy
func TestFilteringMiddleware_ValidPath_AllowsRequest(t *testing.T) {
    middleware := newFilteringMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req, err := http.NewRequest("GET", "/api/heartbeats", nil)
    if err != nil {
        t.Fatalf("Failed to create request: %v", err)
    }
    req.Header.Set(targetUrlHeader, "http://example.com/api/heartbeats")

    rr := httptest.NewRecorder()
    middleware.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("Expected status code %v, got %v", http.StatusOK, rr.Code)
    }
}
