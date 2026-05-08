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
//
// TODO:
//   - GrammarDefinition for GBNF constraints.
//   - DecoderConstraint for decoder-level enforcement.
package constraint
