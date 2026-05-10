// Package prompt defines Layer 8, Feedforward Control, for pagantic.
//
// Layer 8 shapes model behavior before execution through structured prompt
// construction. It keeps prompt assembly explicit, testable, and reusable so
// callers do not build ad hoc prompt strings.
//
// All model interactions must go through structured prompts. Key types are
// Template, SystemPrompt, InstructionSet, and ContextPolicy.
//
// # PromptProvider Pattern
//
// Orchestrate defines a PromptProvider interface that prompt layer types can
// satisfy via Go structural typing (same pattern as ContextProvider and
// CandidateReranker). This allows orchestration to consume structured prompts
// without the prompt layer importing orchestrate. Raw SystemPrompt strings
// remain a convenience default; PromptProvider is the primary mechanism
// for production use.
package prompt
