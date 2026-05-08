package validate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetryPolicy_SucceedsOnFirstTry(t *testing.T) {
	policy := &RetryPolicy{MaxRetries: 3}
	calls := 0

	err := policy.Execute(context.Background(), func() error {
		calls++
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 1, calls)
}

func TestRetryPolicy_SucceedsAfterRetry(t *testing.T) {
	policy := &RetryPolicy{MaxRetries: 2}
	calls := 0

	err := policy.Execute(context.Background(), func() error {
		calls++
		if calls == 1 {
			return errors.New("first fail")
		}
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 2, calls)
}

func TestRetryPolicy_FailsAfterMaxRetries(t *testing.T) {
	expected := errors.New("still bad")
	policy := &RetryPolicy{MaxRetries: 2}
	calls := 0

	err := policy.Execute(context.Background(), func() error {
		calls++
		return expected
	})

	require.ErrorIs(t, err, expected)
	require.Equal(t, 3, calls)
}

func TestRetryPolicy_ContextCancelledStopsRetries(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	policy := &RetryPolicy{MaxRetries: 5, Backoff: time.Hour}
	calls := 0

	err := policy.Execute(ctx, func() error {
		calls++
		cancel()
		return errors.New("fail")
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, calls)
}

func TestRetryPolicy_ZeroMaxRetriesMeansOneAttempt(t *testing.T) {
	expected := errors.New("fail once")
	policy := &RetryPolicy{}
	calls := 0

	err := policy.Execute(context.Background(), func() error {
		calls++
		return expected
	})

	require.ErrorIs(t, err, expected)
	require.Equal(t, 1, calls)
}
