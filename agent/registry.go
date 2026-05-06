package agent

import (
	"fmt"
	"sort"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
)

// Registry holds all available tools indexed by name.
// It implements ToolProvider so it can be passed directly to agent.Config.Tools
// or agent.SpecializedConfig.Tools without an adapter.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a Registry from the given tools.
func NewRegistry(tt ...Tool) *Registry {
	r := &Registry{tools: make(map[string]Tool, len(tt))}
	for _, t := range tt {
		r.tools[t.Info().Name] = t
	}
	return r
}

// Definitions returns tool definitions for all tools that report Available() == true,
// sorted by name. This satisfies the ToolProvider interface.
func (r *Registry) Definitions() []model.D {
	return r.AvailableDefinitions()
}

// AllDefinitions returns tool definitions for all registered tools,
// sorted by name. Use for display purposes (e.g. listing all tools).
func (r *Registry) AllDefinitions() []model.D {
	tt := r.Tools()
	defs := model.DocumentArray()
	for _, t := range tt {
		defs = append(defs, t.Definition())
	}
	return defs
}

// AvailableDefinitions returns tool definitions only for tools that
// report Available() == true, sorted by name.
func (r *Registry) AvailableDefinitions() []model.D {
	tt := r.Tools()
	defs := model.DocumentArray()
	for _, t := range tt {
		if ok, _ := t.Available(); ok {
			defs = append(defs, t.Definition())
		}
	}
	return defs
}

// Execute dispatches a tool call by name and returns the result.
func (r *Registry) Execute(name string, args map[string]any) (string, error) {
	t, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return t.Execute(args)
}

// ToolStatus holds the availability result for a single tool.
type ToolStatus struct {
	Name        string
	Type        ToolType
	Description string
	Available   bool
	Reason      string
}

// CheckAvailability checks all registered tools and returns their status,
// sorted by name.
func (r *Registry) CheckAvailability() []ToolStatus {
	statuses := make([]ToolStatus, 0, len(r.tools))
	for _, t := range r.tools {
		info := t.Info()
		ok, reason := t.Available()
		statuses = append(statuses, ToolStatus{
			Name:        info.Name,
			Type:        info.Type,
			Description: info.Description,
			Available:   ok,
			Reason:      reason,
		})
	}
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})
	return statuses
}

// Tools returns all registered tools sorted by name.
func (r *Registry) Tools() []Tool {
	tt := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tt = append(tt, t)
	}
	sort.Slice(tt, func(i, j int) bool {
		return tt[i].Info().Name < tt[j].Info().Name
	})
	return tt
}

// compile-time check: Registry satisfies ToolProvider.
var _ ToolProvider = (*Registry)(nil)
