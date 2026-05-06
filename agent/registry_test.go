package agent_test

import (
	"testing"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/miroslav-matejovsky/pagantic/agent"
	"github.com/stretchr/testify/require"
)

// fakeTool is a minimal Tool implementation for testing Registry.
type fakeTool struct {
	name      string
	toolType  agent.ToolType
	available bool
	reason    string
}

func (f *fakeTool) Info() agent.ToolInfo {
	return agent.ToolInfo{Name: f.name, Type: f.toolType, Description: "test tool"}
}
func (f *fakeTool) Definition() model.D                      { return model.D{"type": "function"} }
func (f *fakeTool) Execute(_ map[string]any) (string, error) { return "ok", nil }
func (f *fakeTool) Available() (bool, string)                { return f.available, f.reason }

func TestRegistry_CheckAvailability(t *testing.T) {
	r := agent.NewRegistry(
		&fakeTool{name: "alpha", toolType: agent.TypeGo, available: true},
		&fakeTool{name: "bravo", toolType: agent.TypeCLI, available: false, reason: "not installed"},
		&fakeTool{name: "charlie", toolType: agent.TypeGo, available: true},
	)

	statuses := r.CheckAvailability()
	require.Len(t, statuses, 3)

	// Sorted by name.
	require.Equal(t, "alpha", statuses[0].Name)
	require.Equal(t, agent.TypeGo, statuses[0].Type)
	require.True(t, statuses[0].Available)

	require.Equal(t, "bravo", statuses[1].Name)
	require.Equal(t, agent.TypeCLI, statuses[1].Type)
	require.False(t, statuses[1].Available)
	require.Equal(t, "not installed", statuses[1].Reason)

	require.Equal(t, "charlie", statuses[2].Name)
	require.Equal(t, agent.TypeGo, statuses[2].Type)
	require.True(t, statuses[2].Available)
}

func TestRegistry_AvailableDefinitions(t *testing.T) {
	r := agent.NewRegistry(
		&fakeTool{name: "good", toolType: agent.TypeGo, available: true},
		&fakeTool{name: "bad", toolType: agent.TypeCLI, available: false, reason: "missing"},
		&fakeTool{name: "also_good", toolType: agent.TypeGo, available: true},
	)

	all := r.AllDefinitions()
	available := r.AvailableDefinitions()

	require.Len(t, all, 3)
	require.Len(t, available, 2)
}

func TestRegistry_Tools(t *testing.T) {
	r := agent.NewRegistry(
		&fakeTool{name: "zulu", toolType: agent.TypeCLI, available: true},
		&fakeTool{name: "alpha", toolType: agent.TypeGo, available: true},
	)

	tt := r.Tools()
	require.Len(t, tt, 2)
	require.Equal(t, "alpha", tt[0].Info().Name)
	require.Equal(t, agent.TypeGo, tt[0].Info().Type)
	require.Equal(t, "zulu", tt[1].Info().Name)
	require.Equal(t, agent.TypeCLI, tt[1].Info().Type)
}

func TestRegistry_ImplementsToolProvider(t *testing.T) {
	r := agent.NewRegistry(
		&fakeTool{name: "good", toolType: agent.TypeGo, available: true},
		&fakeTool{name: "bad", toolType: agent.TypeCLI, available: false},
	)

	// Definitions() satisfies ToolProvider and returns only available tools.
	defs := r.Definitions()
	require.Len(t, defs, 1, "only available tools should be returned by Definitions()")
}
