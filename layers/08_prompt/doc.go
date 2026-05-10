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
// Orchestrate defines a PromptProvider interface (BuildSystemPrompt() returning
// core.Message). Prompt layer implementations can satisfy this interface via
// Go structural typing without importing orchestrate - same pattern as
// ContextProvider and CandidateReranker. The current SystemPrompt.Build()
// and Template.Render() methods use different signatures; production adapters
// bridge between them or implement PromptProvider directly.
package prompt
