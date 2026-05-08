package memory

import (
	"testing"

	"github.com/miroslav-matejovsky/pagantic/core"
	"github.com/stretchr/testify/require"
)

func TestWorkingMemorySetAndGetResult(t *testing.T) {
	t.Parallel()

	wm := NewWorkingMemory()
	wm.SetResult("answer", 42)

	value, ok := wm.GetResult("answer")
	require.True(t, ok)
	require.Equal(t, 42, value)
}

func TestWorkingMemoryReset(t *testing.T) {
	t.Parallel()

	wm := NewWorkingMemory()
	wm.SetResult("answer", 42)
	wm.Context = []core.Message{core.NewUserMessage("user")}

	wm.Reset()

	require.Empty(t, wm.StepResults)
	require.Empty(t, wm.Context)
}

func TestWorkingMemoryGetMissingResult(t *testing.T) {
	t.Parallel()

	wm := NewWorkingMemory()

	_, ok := wm.GetResult("missing")
	require.False(t, ok)
}
