package tool

import "github.com/miroslav-matejovsky/pagantic/core"

// ToolType says how tool runs.
type ToolType string

const (
	// TypeGo marks pure Go tools.
	TypeGo ToolType = "go"
	// TypeCLI marks tools that call external binaries.
	TypeCLI ToolType = "cli"
)

// ToolInfo holds tool metadata.
type ToolInfo struct {
	Name        string
	Type        ToolType
	Description string
}

// Tool is contract each tool must satisfy.
type Tool interface {
	// Info returns tool metadata.
	Info() ToolInfo
	// Definition returns typed tool schema.
	Definition() core.ToolDefinition
	// Execute runs tool with parsed args.
	Execute(args map[string]any) (string, error)
	// Available reports if tool ready now.
	Available() (bool, string)
}
