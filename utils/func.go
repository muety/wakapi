package utils

import (
	"fmt"
	"log/slog"
	"runtime"
)

// WithRecovery executes a given function while recovering from panics and returning them as an errors
func WithRecovery(fn func(...interface{}), args ...interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stackbuf := make([]byte, 1<<16)
			err = fmt.Errorf("got panic: %v\n%s", r, stackbuf[:runtime.Stack(stackbuf, false)])
			slog.Error(err.Error())
		}
	}()
	fn(args...)
	return err
}

// WithRecovery1 executes a given function with one parameter while recovering from panics and returning them as an errors
func WithRecovery1[T any](fn func(T), a1 T) (err error) {
	return WithRecovery(func(args ...interface{}) {
		fn(args[0].(T))
	}, a1)
}
