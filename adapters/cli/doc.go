// Package cli implements a single-shot command-line adapter for Pagantic.
//
// # ARCHITECTURAL ROLE
//
// This package is an Interface (Shell) adapter. It is a thin boundary between
// command-line invocation and the internal execution core. It must remain
// stateless, thin (no business logic), deterministic, and replaceable without
// affecting core behavior.
//
// INTERACTION FLOW
//
//	CLI Arguments / stdin
//	    -> cli Adapter (this package)
//	    -> ExecutionService (orchestrate layer)
//	    -> Response Formatting
//	    -> stdout
//
// Unlike the TUI adapter which provides an interactive REPL loop, the CLI
// adapter executes a single inference request and exits. Input comes from
// command-line arguments or stdin. Output goes to stdout.
//
// # PROHIBITED RESPONSIBILITIES
//
// This package must NOT perform orchestration logic, call inference directly,
// execute tools, enforce output schemas, implement validation beyond input
// contract checks, or construct prompts.
//
// # Runner
//
// Runner is the main entry point. Configure with RunConfig (engine, system
// prompt, optional tools, stream handler, timeout). Call Run with a prompt
// string to execute a single inference request and write the result to the
// configured output writer.
//
// # Input
//
// ReadPrompt reads a prompt from command-line args (joined with spaces) or
// falls back to reading all of stdin when no args are provided.
package cli
