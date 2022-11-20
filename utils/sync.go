package utils

import (
	"sync"
	"time"
)

// WaitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
// See // https://stackoverflow.com/a/32843750/3112139.
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
