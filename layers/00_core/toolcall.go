package core

// ToolCall represents an assistant's request to execute a tool.
type ToolCall struct {
	ID        string
	Name      string
	Arguments map[string]any
}

// ToolDefinition describes a tool available to the model. Parameters uses
// the Schema type to represent the function's input schema.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  Schema
}

// ToolResult carries the output of a tool execution back to the model.
type ToolResult struct {
	CallID  string
	Name    string
	Content string
	IsError bool
}
