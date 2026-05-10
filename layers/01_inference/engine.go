package inference

import (
	"context"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
)

// Engine is core inference abstraction.
// Implementations convert typed requests into model-specific formats and stream
// results back.
type Engine interface {
	Infer(ctx context.Context, req Request) (*Result, error)
	ModelInfo() ModelInfo
}

// Request carries typed inference input.
type Request struct {
	Messages    []core.Message
	Tools       []core.ToolDefinition
	Schema      *core.Schema
	Grammar     string // GBNF grammar for decoder-level output constraint; empty means none
	MaxTokens   int
	Temperature *float64 // nil means use model default
	Options     map[string]any
}

// Result carries typed inference output.
type Result struct {
	Content   string
	ToolCalls []core.ToolCall
	Messages  []core.Message
	Usage     core.TokenUsage
}

// ModelInfo describes model limits and identity.
type ModelInfo struct {
	Name          string
	ContextWindow int
}
