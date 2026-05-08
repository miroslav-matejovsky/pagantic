package tool_test

import (
	"errors"
	"testing"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	"github.com/miroslav-matejovsky/pagantic/tool"
	"github.com/stretchr/testify/require"
)

// fakeTool is small tool for tests.
type fakeTool struct {
	name        string
	toolType    tool.ToolType
	description string
	available   bool
	reason      string
	output      string
	err         error
}

func (f *fakeTool) Info() tool.ToolInfo {
	description := f.description
	if description == "" {
		description = "test tool"
	}
	return tool.ToolInfo{Name: f.name, Type: f.toolType, Description: description}
}

func (f *fakeTool) Definition() core.ToolDefinition {
	return core.ToolDefinition{
		Name:        f.name,
		Description: "test tool",
		Parameters:  core.Schema{Type: "object"},
	}
}

func (f *fakeTool) Execute(_ map[string]any) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if f.output == "" {
		return "ok", nil
	}
	return f.output, nil
}

func (f *fakeTool) Available() (bool, string) { return f.available, f.reason }

func TestRegistry_CheckAvailability(t *testing.T) {
	r := tool.NewRegistry(
		&fakeTool{name: "alpha", toolType: tool.TypeGo, available: true},
		&fakeTool{name: "bravo", toolType: tool.TypeCLI, available: false, reason: "not installed"},
		&fakeTool{name: "charlie", toolType: tool.TypeGo, available: true},
	)

	statuses := r.CheckAvailability()
	require.Len(t, statuses, 3)

	require.Equal(t, "alpha", statuses[0].Name)
	require.Equal(t, tool.TypeGo, statuses[0].Type)
	require.True(t, statuses[0].Available)

	require.Equal(t, "bravo", statuses[1].Name)
	require.Equal(t, tool.TypeCLI, statuses[1].Type)
	require.False(t, statuses[1].Available)
	require.Equal(t, "not installed", statuses[1].Reason)

	require.Equal(t, "charlie", statuses[2].Name)
	require.Equal(t, tool.TypeGo, statuses[2].Type)
	require.True(t, statuses[2].Available)
}

func TestRegistry_AvailableDefinitions(t *testing.T) {
	r := tool.NewRegistry(
		&fakeTool{name: "good", toolType: tool.TypeGo, available: true},
		&fakeTool{name: "bad", toolType: tool.TypeCLI, available: false, reason: "missing"},
		&fakeTool{name: "also_good", toolType: tool.TypeGo, available: true},
	)

	all := r.AllDefinitions()
	available := r.AvailableDefinitions()

	require.Len(t, all, 3)
	require.Len(t, available, 2)
	require.Equal(t, []string{"also_good", "good"}, []string{available[0].Name, available[1].Name})
}

func TestRegistry_Tools(t *testing.T) {
	r := tool.NewRegistry(
		&fakeTool{name: "zulu", toolType: tool.TypeCLI, available: true},
		&fakeTool{name: "alpha", toolType: tool.TypeGo, available: true},
	)

	tt := r.Tools()
	require.Len(t, tt, 2)
	require.Equal(t, "alpha", tt[0].Info().Name)
	require.Equal(t, tool.TypeGo, tt[0].Info().Type)
	require.Equal(t, "zulu", tt[1].Info().Name)
	require.Equal(t, tool.TypeCLI, tt[1].Info().Type)
}

func TestRegistry_Definitions(t *testing.T) {
	r := tool.NewRegistry(
		&fakeTool{name: "good", toolType: tool.TypeGo, available: true},
		&fakeTool{name: "bad", toolType: tool.TypeCLI, available: false},
	)

	defs := r.Definitions()
	require.Len(t, defs, 1)
	require.Equal(t, "good", defs[0].Name)
}

func TestRegistry_Execute(t *testing.T) {
	r := tool.NewRegistry(
		&fakeTool{name: "echo", toolType: tool.TypeGo, available: true, output: "ran"},
	)

	got, err := r.Execute("echo", map[string]any{"x": 1})
	require.NoError(t, err)
	require.Equal(t, "ran", got)
}

func TestRegistry_ExecuteUnknownTool(t *testing.T) {
	r := tool.NewRegistry(&fakeTool{name: "echo", toolType: tool.TypeGo, available: true})

	got, err := r.Execute("missing", nil)
	require.ErrorContains(t, err, "unknown tool: missing")
	require.Empty(t, got)
}

func TestFakeToolErrorPath(t *testing.T) {
	f := &fakeTool{name: "boom", toolType: tool.TypeGo, available: true, err: errors.New("boom")}
	got, err := f.Execute(nil)
	require.ErrorContains(t, err, "boom")
	require.Empty(t, got)
}
