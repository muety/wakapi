package test

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
)

func AssertContainsHeaderMatching(t assert.TestingT, headers http.Header, headerName string, predicate func(string) bool, msgAndArgs ...interface{}) bool {
	values := headers.Values(headerName)
	if len(values) == 0 {
		return assert.Fail(t, fmt.Sprintf("header %s not found in response", headerName), msgAndArgs...)
	}
	for _, value := range values {
		if predicate(value) {
			return true
		}
	}
	return assert.Fail(t, fmt.Sprintf("header matching predicate not found for %s", headerName), msgAndArgs...)
}
