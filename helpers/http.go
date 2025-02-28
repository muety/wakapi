package helpers

import (
    "encoding/json"
    "errors"
    "github.com/muety/wakapi/config"
    "github.com/muety/wakapi/models"
    "net/http"
)

// Refactored ExtractCookieAuth to use an interface for SecureCookie to allow mocking.
type SecureCookieDecoder interface {
    Decode(name, value string, dst interface{}) error
}

func ExtractCookieAuth(r *http.Request, config *config.Config) (username *string, err error) {
    cookie, err := r.Cookie(models.AuthCookieKey)
    if err != nil {
        return nil, errors.New("missing authentication")
    }

    if err := config.Security.SecureCookie.Decode(models.AuthCookieKey, cookie.Value, &username); err != nil {
        return nil, errors.New("cookie is invalid")
    }

    return username, nil
}

func RespondJSON(w http.ResponseWriter, r *http.Request, status int, object interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(object); err != nil {
        config.Log().Request(r).Error("error while writing json response", "error", err)
    }
}

// Mock implementation of SecureCookieDecoder for testing purposes.
type mockSecureCookie struct {
    DecodeFunc func(name, value string, dst interface{}) error
}

func (m *mockSecureCookie) Decode(name, value string, dst interface{}) error {
    return m.DecodeFunc(name, value, dst)
}

// Mock implementation of http.ResponseWriter for testing purposes.
type mockResponseWriter struct {
    header     http.Header
    statusCode int
    body       string
}

func (m *mockResponseWriter) Header() http.Header {
    if m.header == nil {
        m.header = http.Header{}
    }
    return m.header
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
    m.body = string(data)
    return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
    m.statusCode = statusCode
}
