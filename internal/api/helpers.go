package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type StandardResponse struct {
	Data    interface{} `json:"data"`              // Primary response payload
	Message string      `json:"message,omitempty"` // Optional message
	Error   string      `json:"error,omitempty"`   // Optional error message
}

func sendJSON(w http.ResponseWriter, status int, data interface{}, message string, err string) error {
	response := StandardResponse{
		Data:    data,
		Message: message,
		Error:   err,
	}

	w.Header().Set("Content-Type", "application/json")

	// Only call WriteHeader if it hasn't been called yet
	if rw, ok := w.(interface{ Written() bool }); ok && !rw.Written() {
		w.WriteHeader(status)
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false) // Optional: improve performance by disabling HTML escaping

	if err := encoder.Encode(response); err != nil {
		return fmt.Errorf("error encoding JSON response for %T: %w", data, err)
	}

	return nil
}

func sendJSONSuccess(w http.ResponseWriter, status int, data interface{}) error {
	return sendJSON(w, status, data, "", "")
}
