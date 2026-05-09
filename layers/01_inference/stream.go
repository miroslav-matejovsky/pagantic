package inference

import core "github.com/miroslav-matejovsky/pagantic/layers/00_core"

// StreamHandler receives typed events during streaming inference.
type StreamHandler struct {
	OnContent   func(text string)
	OnReasoning func(text string)
	OnToolCall  func(name, argsJSON string)
}

// EmitContent calls OnContent if handler and callback are non-nil.
func (h *StreamHandler) EmitContent(text string) {
	if h != nil && h.OnContent != nil {
		h.OnContent(text)
	}
}

// EmitReasoning calls OnReasoning if handler and callback are non-nil.
func (h *StreamHandler) EmitReasoning(text string) {
	if h != nil && h.OnReasoning != nil {
		h.OnReasoning(text)
	}
}

// EmitToolCall calls OnToolCall if handler and callback are non-nil.
func (h *StreamHandler) EmitToolCall(name, argsJSON string) {
	if h != nil && h.OnToolCall != nil {
		h.OnToolCall(name, argsJSON)
	}
}

// ChunkType identifies what a StreamChunk carries.
type ChunkType int

const (
	// ChunkContent carries assistant content text.
	ChunkContent ChunkType = iota
	// ChunkReasoning carries reasoning text.
	ChunkReasoning
	// ChunkToolCall carries tool call data.
	ChunkToolCall
	// ChunkDone marks successful stream completion.
	ChunkDone
	// ChunkError marks stream failure.
	ChunkError
)

// StreamChunk is one piece of streaming output.
type StreamChunk struct {
	Type      ChunkType
	Content   string
	Reasoning string
	ToolCall  *core.ToolCall
	Usage     *core.TokenUsage
	Error     error
}
