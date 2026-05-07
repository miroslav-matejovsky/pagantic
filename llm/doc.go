// Package llm provides general-purpose LLM interaction utilities.
// It is not specific to any application and can be used in any project
// that needs to stream and parse LLM responses.
//
// StreamHandler carries typed callbacks for content, reasoning, and tool-call
// streaming events, replacing the old untyped onToken(kind, text) callback.
// StreamResponse extracts tool calls and content from streaming responses,
// dispatching tokens to a StreamHandler.
// TerminalRenderer returns a ready-made StreamHandler for interactive CLIs.
//
// The Chat interface abstracts the LLM engine so callers remain decoupled
// from the concrete backend implementation.
package llm
