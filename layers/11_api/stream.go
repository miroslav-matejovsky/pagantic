package api

import core "github.com/miroslav-matejovsky/pagantic/layers/00_core"

// StreamChunk is single piece of streaming API output.
type StreamChunk struct {
	Content  string
	ToolCall *core.ToolCall
	Usage    *core.TokenUsage
	Done     bool
	Error    *ErrorModel
}

// StreamingInterface sends streaming chunks to client.
type StreamingInterface interface {
	Send(chunk StreamChunk) error
	Close() error
}
