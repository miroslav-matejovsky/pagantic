package constraint

import (
	"fmt"
	"strings"
)

// GrammarDefinition holds a GBNF grammar for decoder-level output constraints.
// GBNF (GGML BNF) is the grammar format used by llama.cpp to constrain token
// generation. When passed to the inference engine, the model can only produce
// tokens that match the grammar rules.
type GrammarDefinition struct {
	// Name identifies this grammar for logging and debugging.
	Name string
	// Grammar is the GBNF grammar string. Must contain a root rule.
	Grammar string
}

// GrammarString returns the raw GBNF text for decoder enforcement.
func (g GrammarDefinition) GrammarString() string {
	return g.Grammar
}

// Validate checks grammar for basic structural issues.
// Returns error if grammar is empty or missing root rule.
func (g GrammarDefinition) Validate() error {
	return ValidateGrammar(g.Grammar)
}

// ValidateGrammar checks GBNF grammar string for basic structural issues.
// Not a full parser - checks for non-empty content and root rule presence.
func ValidateGrammar(grammar string) error {
	trimmed := strings.TrimSpace(grammar)
	if trimmed == "" {
		return fmt.Errorf("constraint: grammar is empty")
	}

	if !hasRootRule(trimmed) {
		return fmt.Errorf("constraint: grammar missing root rule (must contain 'root ::=')")
	}

	return nil
}

// DecoderConstraint represents a constraint enforceable at the decoder level.
// Implementations return a grammar string that the inference engine uses to
// restrict token generation. This is distinct from post-hoc validation -
// decoder constraints prevent invalid output from being generated at all.
type DecoderConstraint interface {
	// GrammarString returns GBNF grammar for decoder enforcement.
	GrammarString() string
}

// GrammarConstraint wraps GrammarDefinition as a DecoderConstraint.
type GrammarConstraint struct {
	Definition GrammarDefinition
}

// GrammarString returns the GBNF grammar from the wrapped definition.
func (gc GrammarConstraint) GrammarString() string {
	return gc.Definition.GrammarString()
}

// hasRootRule checks if grammar text contains an exact root rule definition.
// The LHS of the ::= must be exactly "root" (trimmed), not just a prefix.
func hasRootRule(grammar string) bool {
	for _, line := range strings.Split(grammar, "\n") {
		trimmed := strings.TrimSpace(line)
		idx := strings.Index(trimmed, "::=")
		if idx < 0 {
			continue
		}
		lhs := strings.TrimSpace(trimmed[:idx])
		if lhs == "root" {
			return true
		}
	}
	return false
}
