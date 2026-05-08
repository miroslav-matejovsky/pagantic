package validate

import (
	"context"
	"time"
)

// RetryPolicy controls retry behavior for failed operations.
type RetryPolicy struct {
	// MaxRetries is extra tries after first try.
	MaxRetries int
	// Backoff is wait time between tries.
	Backoff time.Duration
}

// Execute runs fn up to MaxRetries+1 times.
// It waits Backoff between retries.
// It returns last error if all tries fail.
func (rp *RetryPolicy) Execute(ctx context.Context, fn func() error) error {
	if ctx == nil {
		ctx = context.Background()
	}

	retries := 0
	backoff := time.Duration(0)
	if rp != nil {
		if rp.MaxRetries > 0 {
			retries = rp.MaxRetries
		}
		backoff = rp.Backoff
	}

	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt == retries {
			break
		}
		if backoff <= 0 {
			continue
		}

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}

	return lastErr
}
