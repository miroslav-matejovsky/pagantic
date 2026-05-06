// Package llm provides general-purpose LLM interaction utilities.
// It is not specific to any application and can be used in any project
// that needs to stream and parse LLM responses.
//
// It separates streaming response parsing from terminal rendering,
// giving callers full control over how LLM output is displayed.
// StreamResponse extracts tool calls and content from streaming responses.
// TerminalRenderer provides a ready-made ANSI callback for interactive CLIs.
//
// The Chat interface abstracts the LLM engine so callers remain decoupled
// from the concrete backend implementation.
package llm
