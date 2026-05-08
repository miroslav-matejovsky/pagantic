// Package core defines shared domain types used across all layers of the LLM
// harness system. It serves as a shared vocabulary, not an architectural layer.
//
// Types here are intentionally simple data structures with no behavior beyond
// construction helpers. They represent the fundamental concepts that cross
// layer boundaries: messages, tool interactions, schemas, and usage metrics.
//
// No package in pagantic may depend on core for logic -- only for type
// definitions. Conversion to/from external formats (e.g., kronk model.D)
// happens at the inference boundary, not here.
package core
