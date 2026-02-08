package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdTokenPayload_Username(t *testing.T) {
	testCases := []struct {
		name     string
		token    IdTokenPayload
		expected string
	}{
		{
			name: "custom claim takes priority",
			token: IdTokenPayload{
				UsernameClaim:     "custom_claim",
				AllClaims:         map[string]interface{}{"custom_claim": "custom_user"},
				PreferredUsername: "preferred_user",
				Nickname:          "nickname_user",
				Subject:           "subject_user",
			},
			expected: "custom_user",
		},
		{
			name: "preferred_username fallback",
			token: IdTokenPayload{
				PreferredUsername: "preferred_user",
				Nickname:          "nickname_user",
				Subject:           "subject_user",
			},
			expected: "preferred_user",
		},
		{
			name: "nickname fallback",
			token: IdTokenPayload{
				PreferredUsername: "",
				Nickname:          "nickname_user",
				Subject:           "subject_user",
			},
			expected: "nickname_user",
		},
		{
			name: "subject fallback",
			token: IdTokenPayload{
				PreferredUsername: "",
				Nickname:          "",
				Subject:           "subject_user",
			},
			expected: "subject_user",
		},
		{
			name: "empty when nothing available",
			token: IdTokenPayload{
				PreferredUsername: "",
				Nickname:          "",
				Subject:           "",
			},
			expected: "",
		},
		{
			name: "custom claim not found falls back to preferred_username",
			token: IdTokenPayload{
				UsernameClaim:     "missing_claim",
				AllClaims:         map[string]interface{}{"other_claim": "other_value"},
				PreferredUsername: "preferred_user",
			},
			expected: "preferred_user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.token.Username()
			assert.Equal(t, tc.expected, result)
		})
	}
}
