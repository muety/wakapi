package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/muety/wakapi/internal/observability"
	"github.com/muety/wakapi/internal/utilities"
)

const InvalidChannelError = "Invalid channel, supported values are 'sms' or 'whatsapp'. 'whatsapp' is only supported if Twilio or Twilio Verify is used as the provider."

type OAuthError struct {
	Err             string `json:"error"`
	Description     string `json:"error_description,omitempty"`
	InternalError   error  `json:"-"`
	InternalMessage string `json:"-"`
}

func (e *OAuthError) Error() string {
	if e.InternalMessage != "" {
		return e.InternalMessage
	}
	return fmt.Sprintf("%s: %s", e.Err, e.Description)
}

// WithInternalError adds internal error information to the error
func (e *OAuthError) WithInternalError(err error) *OAuthError {
	e.InternalError = err
	return e
}

// WithInternalMessage adds internal message information to the error
func (e *OAuthError) WithInternalMessage(fmtString string, args ...interface{}) *OAuthError {
	e.InternalMessage = fmt.Sprintf(fmtString, args...)
	return e
}

// Cause returns the root cause error
func (e *OAuthError) Cause() error {
	if e.InternalError != nil {
		return e.InternalError
	}
	return e
}

// func oauthError(err string, description string) *OAuthError {
// 	return &OAuthError{Err: err, Description: description}
// }

// func badRequestError(errorCode ErrorCode, fmtString string, args ...interface{}) *HTTPError {
// 	return httpError(http.StatusBadRequest, errorCode, fmtString, args...)
// }

// func internalServerError(fmtString string, args ...interface{}) *HTTPError {
// 	return httpError(http.StatusInternalServerError, ErrorCodeUnexpectedFailure, fmtString, args...)
// }

// func notFoundError(errorCode ErrorCode, fmtString string, args ...interface{}) *HTTPError {
// 	return httpError(http.StatusNotFound, errorCode, fmtString, args...)
// }

// func forbiddenError(errorCode ErrorCode, fmtString string, args ...interface{}) *HTTPError {
// 	return httpError(http.StatusForbidden, errorCode, fmtString, args...)
// }

// func unprocessableEntityError(errorCode ErrorCode, fmtString string, args ...interface{}) *HTTPError {
// 	return httpError(http.StatusUnprocessableEntity, errorCode, fmtString, args...)
// }

// func tooManyRequestsError(errorCode ErrorCode, fmtString string, args ...interface{}) *HTTPError {
// 	return httpError(http.StatusTooManyRequests, errorCode, fmtString, args...)
// }

// func conflictError(fmtString string, args ...interface{}) *HTTPError {
// 	return httpError(http.StatusConflict, ErrorCodeConflict, fmtString, args...)
// }

// HTTPError is an error with a message and an HTTP status code.
type HTTPError struct {
	HTTPStatus      int    `json:"code"`                 // do not rename the JSON tags!
	ErrorCode       string `json:"error_code,omitempty"` // do not rename the JSON tags!
	Message         string `json:"msg"`                  // do not rename the JSON tags!
	InternalError   error  `json:"-"`
	InternalMessage string `json:"-"`
	ErrorID         string `json:"error_id,omitempty"`
}

func (e *HTTPError) Error() string {
	if e.InternalMessage != "" {
		return e.InternalMessage
	}
	return fmt.Sprintf("%d: %s", e.HTTPStatus, e.Message)
}

func (e *HTTPError) Is(target error) bool {
	return e.Error() == target.Error()
}

// Cause returns the root cause error
func (e *HTTPError) Cause() error {
	if e.InternalError != nil {
		return e.InternalError
	}
	return e
}

// WithInternalError adds internal error information to the error
func (e *HTTPError) WithInternalError(err error) *HTTPError {
	e.InternalError = err
	return e
}

// WithInternalMessage adds internal message information to the error
func (e *HTTPError) WithInternalMessage(fmtString string, args ...interface{}) *HTTPError {
	e.InternalMessage = fmt.Sprintf(fmtString, args...)
	return e
}

// func httpError(httpStatus int, errorCode ErrorCode, fmtString string, args ...interface{}) *HTTPError {
// 	return &HTTPError{
// 		HTTPStatus: httpStatus,
// 		ErrorCode:  errorCode,
// 		Message:    fmt.Sprintf(fmtString, args...),
// 	}
// }

// Recoverer is a middleware that recovers from panics, logs the panic (and a
// backtrace), and returns a HTTP 500 (Internal Server Error) status if
// possible. Recoverer prints a request ID if one is provided.
func recoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				logEntry := observability.GetLogEntry(r)
				if logEntry != nil {
					logEntry.Panic(rvr, debug.Stack())
				} else {
					fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
					debug.PrintStack()
				}

				se := &HTTPError{
					HTTPStatus: http.StatusInternalServerError,
					Message:    http.StatusText(http.StatusInternalServerError),
				}
				HandleResponseError(se, w, r)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// ErrorCause is an error interface that contains the method Cause() for returning root cause errors
type ErrorCause interface {
	Cause() error
}

func NotFoundHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a response recorder to capture the response
		recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)

		// If the status code is 404, return a JSON response
		if recorder.statusCode == http.StatusNotFound {
			sendJSON(w, http.StatusNotFound, nil, "Resource not found", "The resource you're looking for cannot be found")
		}
	})
}

// responseRecorder is a wrapper for http.ResponseWriter to capture the status code.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func HandleResponseError(err error, w http.ResponseWriter, r *http.Request) {
	log := observability.GetLogEntry(r).Entry
	errorID := utilities.GetRequestID(r.Context())

	switch e := err.(type) {
	case *HTTPError:
		switch {
		case e.HTTPStatus >= http.StatusInternalServerError:
			e.ErrorID = errorID
			// this will get us the stack trace too
			log.WithError(e.Cause()).Error(e.Error())
		case e.HTTPStatus == http.StatusTooManyRequests:
			log.WithError(e.Cause()).Warn(e.Error())
		default:
			log.WithError(e.Cause()).Info(e.Error())
		}

		if e.ErrorCode != "" {
			w.Header().Set("x-sb-error-code", e.ErrorCode)
		} else {
			if e.ErrorCode == "" {
				if e.HTTPStatus == http.StatusInternalServerError {
					e.ErrorCode = ErrorCodeUnexpectedFailure
				} else {
					e.ErrorCode = ErrorCodeUnknown
				}
			}

			// Provide better error messages for certain user-triggered Postgres errors.
			if pgErr := utilities.NewPostgresError(e.InternalError); pgErr != nil {
				if jsonErr := sendJSON(w, pgErr.HttpStatusCode, nil, pgErr.Message, pgErr.Detail); jsonErr != nil && jsonErr != context.DeadlineExceeded {
					log.WithError(jsonErr).Warn("Failed to send JSON on ResponseWriter")
				}
				return
			}

			if jsonErr := sendJSON(w, e.HTTPStatus, nil, e.Message, e.Error()); jsonErr != nil && jsonErr != context.DeadlineExceeded {
				log.WithError(jsonErr).Warn("Failed to send JSON on ResponseWriter")
			}
		}

	case *OAuthError:
		log.WithError(e.Cause()).Info(e.Error())
		if jsonErr := sendJSON(w, http.StatusBadRequest, nil, e.InternalMessage, e.Error()); jsonErr != nil && jsonErr != context.DeadlineExceeded {
			log.WithError(jsonErr).Warn("Failed to send JSON on ResponseWriter")
		}

	case ErrorCause:
		HandleResponseError(e.Cause(), w, r)

	default:
		log.WithError(e).Errorf("Unhandled server error: %s", e.Error())

		httpError := HTTPError{
			HTTPStatus: http.StatusInternalServerError,
			ErrorCode:  ErrorCodeUnexpectedFailure,
			Message:    "Unexpected failure, please check server logs for more information",
		}

		if jsonErr := sendJSON(w, http.StatusInternalServerError, httpError, httpError.Message, ""); jsonErr != nil && jsonErr != context.DeadlineExceeded {
			log.WithError(jsonErr).Warn("Failed to send JSON on ResponseWriter")
		}
	}
}

// func generateFrequencyLimitErrorMessage(timeStamp *time.Time, maxFrequency time.Duration) string {
// 	now := time.Now()
// 	left := timeStamp.Add(maxFrequency).Sub(now) / time.Second
// 	return fmt.Sprintf("For security purposes, you can only request this after %d seconds.", left)
// }
