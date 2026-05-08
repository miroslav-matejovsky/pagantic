// Package inference provides layer 1 execution substrate.
//
// Layer 1 accepts structured prompt input and produces raw model output as
// text, tool calls, or token stream callbacks. It knows nothing about tools,
// workflows, schemas beyond transport, validation, or business rules.
//
// Engine hides concrete inference backend details behind typed Go interfaces.
// KronkAdapter converts between core types and kronk's model.D format.
package inference
