package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func sendJSON(w http.ResponseWriter, status int, obj interface{}) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    
    encoder := json.NewEncoder(w)
    encoder.SetEscapeHTML(false) // Optional: improve performance by disabling HTML escaping
    
    if err := encoder.Encode(obj); err != nil {
        return fmt.Errorf("error encoding JSON response for %T: %w", obj, err)
    }
    
    return nil
}