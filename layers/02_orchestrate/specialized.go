package orchestrate

import (
	"context"
	"fmt"
	"time"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
	observe "github.com/miroslav-matejovsky/pagantic/layers/10_observe"
)

const phase2Prompt = "Produce your structured output now."

// SpecializedConfig configures stateless specialized loop.
type SpecializedConfig struct {
	Engine            inference.Engine // required
	Tools             *tool.Registry
	SystemPrompt      string
	Schema            core.Schema
	Grammar           string // GBNF grammar for decoder-level constraint; empty means none
	MaxTokens         int    // required; must be > 0
	MaxToolIterations int    // required; must be > 0
	Stream            *inference.StreamHandler
	Observer          observe.EventLog
	ContextProvider   ContextProvider // optional; retrieves context using original prompt
}

// SpecializedLoop is stateless wrapper around fresh AgentLoop per call.
type SpecializedLoop struct {
	cfg SpecializedConfig
}

// NewSpecializedLoop builds SpecializedLoop. Returns error if Engine is nil,
// MaxTokens <= 0, or MaxToolIterations <= 0.
func NewSpecializedLoop(cfg SpecializedConfig) (*SpecializedLoop, error) {
	if cfg.Engine == nil {
		return nil, fmt.Errorf("specialized loop: Engine required")
	}
	if cfg.MaxTokens <= 0 {
		return nil, fmt.Errorf("specialized loop: MaxTokens must be > 0")
	}
	if cfg.MaxToolIterations <= 0 {
		return nil, fmt.Errorf("specialized loop: MaxToolIterations must be > 0")
	}
	return &SpecializedLoop{cfg: cfg}, nil
}

// Call runs optional context retrieval, optional tool phase, then structured phase.
// Context is retrieved once using original prompt, avoiding poor retrieval
// with phase2Prompt during the structured output step.
func (sl *SpecializedLoop) Call(ctx context.Context, prompt string) (*inference.Result, error) {
	if sl == nil || sl.cfg.Engine == nil {
		return nil, fmt.Errorf("specialized loop: nil engine")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	inner, err := NewAgentLoop(LoopConfig{
		Engine:            sl.cfg.Engine,
		Tools:             sl.cfg.Tools,
		SystemPrompt:      sl.cfg.SystemPrompt,
		Grammar:           sl.cfg.Grammar,
		MaxTokens:         sl.cfg.MaxTokens,
		MaxToolIterations: sl.cfg.MaxToolIterations,
		Stream:            sl.cfg.Stream,
		Observer:          sl.cfg.Observer,
	})
	if err != nil {
		return nil, fmt.Errorf("specialized loop: %w", err)
	}

	if sl.cfg.ContextProvider != nil {
		started := time.Now()
		msgs, err := sl.cfg.ContextProvider.Build(ctx, prompt)
		inner.recordEvent(started, "context", map[string]any{"query": prompt, "chunks": len(msgs)}, err)
		if err == nil {
			inner.injectContext(msgs)
		}
	}

	if sl.cfg.Tools != nil {
		if _, err := inner.Chat(ctx, prompt); err != nil {
			return nil, fmt.Errorf("specialized loop tool phase: %w", err)
		}
		return inner.ChatStructured(ctx, phase2Prompt, sl.cfg.Schema)
	}

	return inner.ChatStructured(ctx, prompt, sl.cfg.Schema)
}
