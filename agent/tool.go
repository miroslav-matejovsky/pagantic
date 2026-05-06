package agent

import "github.com/ardanlabs/kronk/sdk/kronk/model"

// ToolType categorizes a tool by its implementation strategy.
type ToolType string

const (
	// TypeGo identifies tools implemented in pure Go (e.g. net/http).
	// These tools have no external dependencies and are always available.
	TypeGo ToolType = "go"
	// TypeCLI identifies tools that delegate to an external binary via exec.
	// Availability depends on the binary being present in PATH.
	TypeCLI ToolType = "cli"
)

// ToolInfo holds metadata about a tool.
type ToolInfo struct {
	Name        string
	Type        ToolType
	Description string // Short one-liner, ~10 words
}

// Tool defines the interface each tool must implement.
// Each tool provides its LLM function definition and execution logic.
type Tool interface {
	// Info returns the tool metadata: name, type, and short description.
	Info() ToolInfo
	// Definition returns the OpenAI-style tool definition for the LLM.
	Definition() model.D
	// Execute runs the tool with parsed arguments and returns output text.
	Execute(args map[string]any) (string, error)
	// Available reports whether this tool is ready to use.
	// Returns (true, "") when ready, (false, reason) when not.
	// Binary tools check exec.LookPath; HTTP tools check env vars.
	Available() (bool, string)
}
