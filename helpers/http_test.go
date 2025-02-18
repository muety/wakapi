package helpers

import (
    "net/http"
    "testing"
)


// Test generated using Keploy
func TestRespondJSON_ValidResponse(t *testing.T) {
    w := &mockResponseWriter{}
    req := &http.Request{}
    status := http.StatusOK
    object := map[string]string{"key": "value"}

    RespondJSON(w, req, status, object)

    if w.statusCode != status {
        t.Errorf("Expected status code %d, got %d", status, w.statusCode)
    }
    if w.header.Get("Content-Type") != "application/json" {
        t.Errorf("Expected Content-Type 'application/json', got %v", w.header.Get("Content-Type"))
    }
    if w.body != `{"key":"value"}`+"\n" {
        t.Errorf("Expected body '{\"key\":\"value\"}', got %v", w.body)
    }
}


// Test generated using Keploy
func TestRespondJSON_EncodingError(t *testing.T) {
    w := &mockResponseWriter{}
    req := &http.Request{}
    status := http.StatusInternalServerError
    object := make(chan int) // Invalid type for JSON encoding

    RespondJSON(w, req, status, object)

    if w.statusCode != status {
        t.Errorf("Expected status code %d, got %d", status, w.statusCode)
    }
    if w.header.Get("Content-Type") != "application/json" {
        t.Errorf("Expected Content-Type 'application/json', got %v", w.header.Get("Content-Type"))
    }
    // Check logs manually or mock the logger to verify the error message.
}
