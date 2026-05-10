// Package constraint enforces deterministic output at system boundary.
//
// Layer 5 keeps model output structured and valid. No unstructured text
// crosses boundary. Validators check raw output, repair common truncation,
// and return clear failures.
//
// Key abstractions:
//   - OutputValidator validates raw model output.
//   - ValidationResult reports validity, errors, and repaired output.
//   - RepairJSON closes common truncated JSON from LLM output.
//   - GrammarDefinition holds GBNF grammar for decoder-level constraints.
//   - DecoderConstraint interface for decoder-level enforcement.
//   - GrammarConstraint wraps GrammarDefinition as DecoderConstraint.
package constraint
