// Package validate catches bad output and system mistakes.
//
// Layer 7 guards final output. It checks output against hard constraints,
// rules, and repair paths. It can also trigger retry loops when output stays
// bad.
//
// Two kinds live here:
//   - deterministic validation, like schema checks and local rules
//   - inferential validation, like LLM checks for hallucination or nonsense
//
// Main types:
//   - RuleValidator for deterministic rule checks
//   - SemanticValidator for inferential checks
//   - RepairStrategy for fixing bad output
//   - RetryPolicy for retry loops
//
// TODO: add LLM-backed inferential validation.
package validate
