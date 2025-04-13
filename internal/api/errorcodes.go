package api

type ErrorCode = string

const (
	// ErrorCodeUnknown should not be used directly, it only indicates a failure in the error handling system in such a way that an error code was not assigned properly.
	ErrorCodeUnknown ErrorCode = "unknown"

	// ErrorCodeUnexpectedFailure signals an unexpected failure such as a 500 Internal Server Error.
	ErrorCodeUnexpectedFailure ErrorCode = "unexpected_failure"

	ErrorCodeValidationFailed   ErrorCode = "validation_failed"
	ErrorCodeBadJSON            ErrorCode = "bad_json"
	ErrorCodeEmailExists        ErrorCode = "email_exists"
	ErrorCodePhoneExists        ErrorCode = "phone_exists"
	ErrorCodeBadJWT             ErrorCode = "bad_jwt"
	ErrorCodeNotAdmin           ErrorCode = "not_admin"
	ErrorCodeNoAuthorization    ErrorCode = "no_authorization"
	ErrorCodeUserNotFound       ErrorCode = "user_not_found"
	ErrorCodeSessionExpired     ErrorCode = "session_expired"
	ErrorCodeSignupDisabled     ErrorCode = "signup_disabled"
	ErrorCodeUserBanned         ErrorCode = "user_banned"
	ErrorCodeBadOAuthState      ErrorCode = "bad_oauth_state"
	ErrorCodeBadOAuthCallback   ErrorCode = "bad_oauth_callback"
	ErrorCodeConflict           ErrorCode = "conflict"
	ErrorCodeInvalidCredentials ErrorCode = "invalid_credentials"
)
