package memory

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionStateSetAndGet(t *testing.T) {
	t.Parallel()

	ss := NewSessionState()
	ss.Set("name", "pagantic")

	value, ok := ss.Get("name")
	require.True(t, ok)
	require.Equal(t, "pagantic", value)
}

func TestSessionStateGetMissingKey(t *testing.T) {
	t.Parallel()

	ss := NewSessionState()

	_, ok := ss.Get("missing")
	require.False(t, ok)
}

func TestSessionStateDelete(t *testing.T) {
	t.Parallel()

	ss := NewSessionState()
	ss.Set("name", "pagantic")

	ss.Delete("name")

	_, ok := ss.Get("name")
	require.False(t, ok)
}

func TestSessionStateKeysSorted(t *testing.T) {
	t.Parallel()

	ss := NewSessionState()
	ss.Set("zeta", 1)
	ss.Set("alpha", 2)
	ss.Set("beta", 3)

	require.Equal(t, []string{"alpha", "beta", "zeta"}, ss.Keys())
}

func TestSessionStateConcurrentAccess(t *testing.T) {
	t.Parallel()

	ss := NewSessionState()
	const workers = 16
	const perWorker = 32

	var wg sync.WaitGroup
	errs := make(chan string, workers*perWorker)

	for i := range workers {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			for j := range perWorker {
				key := fmt.Sprintf("k-%d-%d", worker, j)
				want := worker*perWorker + j

				ss.Set(key, want)
				got, ok := ss.Get(key)
				if !ok || got != want {
					errs <- key
				}
			}
		}(i)
	}

	for i := range workers {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			for j := range perWorker {
				_, _ = ss.Get(fmt.Sprintf("k-%d-%d", j%workers, worker%perWorker))
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	var failures []string
	for key := range errs {
		failures = append(failures, key)
	}

	require.Empty(t, failures)
	require.Len(t, ss.Keys(), workers*perWorker)
}
