// Package tui provides generic terminal UI primitives and an agent-harness
// REPL reusable across CLI agent applications.
//
// # Colors
//
// Bold, Dim, Green, Red, Yellow, Cyan, and Grey wrap strings in ANSI escape
// codes for styled terminal output without requiring a full curses library.
//
// # Output sanitization
//
// SanitizeOutput strips ANSI escape sequences and non-printable control
// characters (including C1 controls 0x80-0x9f) from strings. Use it to
// prevent terminal injection when rendering output from external sources
// such as tool calls or subprocesses.
//
// # Prompt
//
// FPrompt reads a single trimmed, non-empty line from a bufio.Scanner after
// printing a formatted prompt string to an io.Writer. It abstracts the common
// read-loop boilerplate found in interactive REPL and chat interfaces.
// Returns (line, error) where io.EOF signals clean end-of-input and
// scanner.Err() is surfaced as a non-nil error.
//
// # Styled messages
//
// The F-prefixed functions (FInfo, FWarn, FError, FInfof, FWarnf, FErrorf)
// print prefixed, colored messages for common CLI output patterns and accept
// an io.Writer for testability. The Infof and Warnf convenience wrappers
// write to stdout.
//
// # REPL
//
// REPL provides a generic command-dispatch read-eval-print loop. Register
// commands with AddCommand (panics on empty Name or nil Run), set a banner
// with SetBanner, and call Run to start. Built-in quit/exit/q commands exit
// the loop, and an auto-generated help listing is provided when no explicit
// help command is registered. Input and output streams are configurable for
// testing and embedding.
//
// # Agent harness
//
// AgentREPL extends REPL into a full agent-harness terminal UI. It provides
// built-in tools and chat commands, lazy inference engine loading via a
// caller-supplied EngineLoader func, and an Engine(ctx) getter so custom
// commands registered via AddCommand can access the loaded engine. Configure
// via AgentConfig with title, banner, system prompt, engine loader, tool
// registry, and local directory. Chat uses orchestrate.AgentLoop and tools
// from the tool package.
//
// # Rendering
//
// TerminalRenderer returns an *inference.StreamHandler that renders streaming
// content, reasoning, and tool-call events to an io.Writer with ANSI colors.
// FPrintUsage displays token usage statistics. UsageStats is a
// dependency-free value type for passing usage data across package
// boundaries without importing inference engine packages.
package tui
