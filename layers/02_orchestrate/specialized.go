package orchestrate

import (
	"context"
	"fmt"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
	"github.com/miroslav-matejovsky/pagantic/observe"
)

const phase2Prompt = "Produce your structured output now."

// SpecializedConfig configures stateless specialized loop.
type SpecializedConfig struct {
	Engine       inference.Engine
	Tools        *tool.Registry
	SystemPrompt string
	Schema       core.Schema
	MaxTokens    int
	Stream       *inference.StreamHandler
	Observer     observe.EventLog
}

// SpecializedLoop is stateless wrapper around fresh AgentLoop per call.
type SpecializedLoop struct {
	cfg SpecializedConfig
}

// NewSpecializedLoop builds SpecializedLoop.
func NewSpecializedLoop(cfg SpecializedConfig) *SpecializedLoop {
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = defaultMaxTokens
	}
	return &SpecializedLoop{cfg: cfg}
}

// Call runs optional tool phase, then structured phase.
func (sl *SpecializedLoop) Call(ctx context.Context, prompt string) (*inference.Result, error) {
	inner := NewAgentLoop(LoopConfig{
		Engine:       sl.cfg.Engine,
		Tools:        sl.cfg.Tools,
		SystemPrompt: sl.cfg.SystemPrompt,
		MaxTokens:    sl.cfg.MaxTokens,
		Stream:       sl.cfg.Stream,
		Observer:     sl.cfg.Observer,
	})

	if sl.cfg.Tools != nil {
		if _, err := inner.Chat(ctx, prompt); err != nil {
			return nil, fmt.Errorf("specialized loop tool phase: %w", err)
		}
		return inner.ChatStructured(ctx, phase2Prompt, sl.cfg.Schema)
	}

	return inner.ChatStructured(ctx, prompt, sl.cfg.Schema)
}
