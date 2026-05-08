// Package prompt defines Layer 8, Feedforward Control, for pagantic.
//
// Layer 8 shapes model behavior before execution through structured prompt
// construction. It keeps prompt assembly explicit, testable, and reusable so
// callers do not build ad hoc prompt strings.
//
// All model interactions must go through structured prompts. Key types are
// Template, SystemPrompt, InstructionSet, and ContextPolicy.
package prompt
