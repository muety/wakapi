package helpers

import (
    "net/http"
    "testing"
)


// Test generated using Keploy
func TestParseSummaryFilters_MultipleFilters(t *testing.T) {
    req, _ := http.NewRequest("GET", "/?project=TestProject&language=Go&editor=VSCode", nil)

    filters := ParseSummaryFilters(req)

    if filters == nil {
        t.Fatalf("Expected Filters, got nil")
    }
    if filters.Count() != 3 {
        t.Errorf("Expected 3 filters, got %d", filters.Count())
    }
}


// Test generated using Keploy
func TestExtractUser_NoPrincipal(t *testing.T) {
    req, _ := http.NewRequest("GET", "/", nil)

    user := extractUser(req)

    if user != nil {
        t.Errorf("Expected nil user, got %v", user)
    }
}
