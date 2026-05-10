package orchestrate

import core "github.com/miroslav-matejovsky/pagantic/layers/00_core"

// StepInput is typed wrapper for data entering a Step.
// Replaces raw any values with structured fields at step boundaries.
type StepInput struct {
	Messages   []core.Message
	Candidates []CandidateIR
	Context    *ContextIR
	ToolCalls  []core.ToolCall
	Raw        any
}

// StepOutput is typed wrapper for data leaving a Step.
// Carries results forward to next Step in plan.
type StepOutput struct {
	Content    string
	Candidates []CandidateIR
	ToolCalls  []core.ToolCall
	Raw        any
}

// CandidateIR is the canonical cross-step representation of a scored item.
// Unifies context.Chunk and rerank.Candidate at step boundaries.
// Conversion between layer-native types and CandidateIR happens at
// integration sites in user code, not inside layers.
type CandidateIR struct {
	Content  string
	Source   string
	Score    float64
	Metadata map[string]any
}

// ContextIR represents retrieved context with provenance.
// Built by context layer, consumed by orchestrate during prompt assembly.
type ContextIR struct {
	Chunks []ContextChunk
}

// ContextChunk is one piece of retrieved knowledge with source tracking.
type ContextChunk struct {
	Content string
	Source  string
	Score   float64
}
