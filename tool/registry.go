package tool

import (
	"fmt"
	"sort"

	"github.com/miroslav-matejovsky/pagantic/core"
)

// ToolStatus is availability state for one tool.
type ToolStatus struct {
	Name        string
	Type        ToolType
	Description string
	Available   bool
	Reason      string
}

// Registry holds tools by name.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry builds registry from tools.
func NewRegistry(tt ...Tool) *Registry {
	r := &Registry{tools: make(map[string]Tool, len(tt))}
	for _, t := range tt {
		r.tools[t.Info().Name] = t
	}
	return r
}

// Definitions returns definitions for available tools.
func (r *Registry) Definitions() []core.ToolDefinition {
	return r.AvailableDefinitions()
}

// AllDefinitions returns definitions for all tools.
func (r *Registry) AllDefinitions() []core.ToolDefinition {
	tt := r.Tools()
	defs := make([]core.ToolDefinition, 0, len(tt))
	for _, t := range tt {
		defs = append(defs, t.Definition())
	}
	return defs
}

// AvailableDefinitions returns definitions for ready tools.
func (r *Registry) AvailableDefinitions() []core.ToolDefinition {
	tt := r.Tools()
	defs := make([]core.ToolDefinition, 0, len(tt))
	for _, t := range tt {
		if ok, _ := t.Available(); ok {
			defs = append(defs, t.Definition())
		}
	}
	return defs
}

// Execute runs named tool.
func (r *Registry) Execute(name string, args map[string]any) (string, error) {
	t, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return t.Execute(args)
}

// CheckAvailability reports state for all tools.
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

// Tools returns all tools sorted by name.
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
