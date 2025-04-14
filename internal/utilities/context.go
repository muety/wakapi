package utilities

import (
	"context"
	"sync"
)

type contextKey string

func (c contextKey) String() string {
	return "wakana api context key " + string(c)
}

const (
	requestIDKey = contextKey("request_id")
)

// WithRequestID adds the provided request ID to the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// GetRequestID reads the request ID from the context.
func GetRequestID(ctx context.Context) string {
	obj := ctx.Value(requestIDKey)
	if obj == nil {
		return ""
	}

	return obj.(string)
}

// WaitForCleanup waits until all long-running goroutines shut
// down cleanly or until the provided context signals done.
func WaitForCleanup(ctx context.Context, wg *sync.WaitGroup) {
	cleanupDone := make(chan struct{})

	go func() {
		defer close(cleanupDone)

		wg.Wait()
	}()

	select {
	case <-ctx.Done():
		return

	case <-cleanupDone:
		return
	}
}
