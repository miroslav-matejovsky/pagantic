package core

// Schema represents a JSON Schema subset used for structured LLM output
// and tool parameter definitions. Covers the common cases: objects with
// typed properties, arrays, enums, and primitive types.
type Schema struct {
	Type        string            `json:"type,omitempty"`
	Description string            `json:"description,omitempty"`
	Properties  map[string]Schema `json:"properties,omitempty"`
	Required    []string          `json:"required,omitempty"`
	Enum        []string          `json:"enum,omitempty"`
	Items       *Schema           `json:"items,omitempty"`
	Default     any               `json:"default,omitempty"`
}
