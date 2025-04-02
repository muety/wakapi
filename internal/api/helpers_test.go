package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           interface{}
		message        string
		err            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success response with data",
			status:         http.StatusOK,
			data:           map[string]string{"key": "value"},
			message:        "Success",
			err:            "",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":{"key":"value"},"message":"Success"}`,
		},
		{
			name:           "Error response with message",
			status:         http.StatusBadRequest,
			data:           nil,
			message:        "",
			err:            "Invalid input",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"data":null,"error":"Invalid input"}`,
		},
		{
			name:           "Empty response",
			status:         http.StatusNoContent,
			data:           nil,
			message:        "",
			err:            "",
			expectedStatus: http.StatusNoContent,
			expectedBody:   `{"data":null}`,
		},
		{
			name:           "HTML escaping disabled",
			status:         http.StatusOK,
			data:           map[string]string{"html": "<div>content</div>"},
			message:        "HTML content",
			err:            "",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":{"html":"<div>content</div>"},"message":"HTML content"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the sendJSON function
			err := sendJSON(rr, tt.status, tt.data, tt.message, tt.err)
			if err != nil {
				t.Fatalf("sendJSON returned an error: %v", err)
			}

			// Check the status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check the response body
			body := strings.TrimSpace(rr.Body.String())
			if body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, body)
			}

			// Check the Content-Type header
			if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}
		})
	}
}

func TestSendJSONEncodingError(t *testing.T) {
	// Create a response recorder
	rr := httptest.NewRecorder()

	// Pass a non-serializable value to trigger an encoding error
	nonSerializableData := make(chan int)

	err := sendJSON(rr, http.StatusOK, nonSerializableData, "This will fail", "")
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}

	expectedError := "error encoding JSON response"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error to contain %q, got %q", expectedError, err.Error())
	}
}
