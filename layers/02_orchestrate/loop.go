package orchestrate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
	constraint "github.com/miroslav-matejovsky/pagantic/layers/05_constraint"
	memory "github.com/miroslav-matejovsky/pagantic/layers/09_memory"
	observe "github.com/miroslav-matejovsky/pagantic/layers/10_observe"
)

const defaultMaxTokens = 2048
const defaultMaxToolIterations = 20

// ContextProvider retrieves context messages for a query.
// The pagantic layers/03_context.ContextBuilder (typically imported with alias
// pctx) satisfies this interface via Go structural typing - no explicit
// interface declaration needed in that package.
type ContextProvider interface {
	Build(ctx context.Context, query string) ([]core.Message, error)
}

// LoopConfig configures agent loop.
type LoopConfig struct {
	Engine            inference.Engine
	Tools             *tool.Registry
	SystemPrompt      string
	Grammar           string // GBNF grammar for decoder-level constraint; only applied in ChatStructured, not Chat
	MaxTokens         int
	MaxToolIterations int // max tool-call loop rounds, 0 uses default (20)
	Stream            *inference.StreamHandler
	OnToolResult      func(name, output string)
	Observer          observe.EventLog
	ContextProvider   ContextProvider // optional; retrieves context for Chat turns (multi-turn path only)
}

// AgentLoop is stateful multi-turn agent with tool loop.
type AgentLoop struct {
	cfg    LoopConfig
	memory *memory.ConversationBuffer
}

// NewAgentLoop builds AgentLoop.
func NewAgentLoop(cfg LoopConfig) *AgentLoop {
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = defaultMaxTokens
	}
	if cfg.MaxToolIterations == 0 {
		cfg.MaxToolIterations = defaultMaxToolIterations
	}

	buf := memory.NewConversationBuffer(0)
	if cfg.SystemPrompt != "" {
		buf.Append(core.NewSystemMessage(cfg.SystemPrompt))
	}

	return &AgentLoop{cfg: cfg, memory: buf}
}

// Chat sends user message, resolves tool calls, and returns final answer.
func (al *AgentLoop) Chat(ctx context.Context, userMessage string) (*inference.Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if al == nil || al.cfg.Engine == nil {
		return nil, fmt.Errorf("agent loop chat: nil engine")
	}

	// Retrieve context ephemerally - injected into first iteration request only,
	// never stored in the buffer. This prevents context from accumulating across
	// turns when the same AgentLoop is used for multi-turn conversation.
	ctxMsgs := al.fetchContext(ctx, userMessage)
	al.memory.Append(core.NewUserMessage(userMessage))

	for iteration := 0; ; iteration++ {
		if iteration >= al.cfg.MaxToolIterations {
			return nil, fmt.Errorf("agent loop chat: exceeded max tool iterations (%d)", al.cfg.MaxToolIterations)
		}

		var requestMsgs []core.Message
		if iteration == 0 {
			// First iteration: inject context ephemerally before user message.
			// Subsequent iterations (tool loop) use memory directly.
			requestMsgs = al.memoryWithEphemeralContext(ctxMsgs)
		} else {
			requestMsgs = al.memory.Messages()
		}

		req := inference.Request{
			Messages:  requestMsgs,
			MaxTokens: al.cfg.MaxTokens,
		}
		if al.cfg.Tools != nil {
			req.Tools = al.cfg.Tools.Definitions()
		}

		// Snapshot memory before inference. Used by storeResult to rebuild
		// history without ephemeral context from req.Messages.
		memBase := al.memory.Messages()

		result, err := al.infer(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("agent loop chat: %w", err)
		}

		al.storeResult(memBase, result)
		if len(result.ToolCalls) == 0 {
			return result, nil
		}
		if al.cfg.Tools == nil {
			return nil, fmt.Errorf("agent loop chat: tool call requested but no tools configured")
		}

		for _, tc := range result.ToolCalls {
			output, execErr := al.executeTool(tc)
			if execErr != nil {
				output = fmt.Sprintf("Error: %v", execErr)
			}
			if al.cfg.OnToolResult != nil {
				al.cfg.OnToolResult(tc.Name, output)
			}
			al.memory.Append(core.NewToolResultMessage(tc.ID, tc.Name, output))
		}
	}
}

// ChatStructured sends user message and asks for schema-shaped JSON.
func (al *AgentLoop) ChatStructured(ctx context.Context, userMessage string, schema core.Schema) (*inference.Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if al == nil || al.cfg.Engine == nil {
		return nil, fmt.Errorf("agent loop structured chat: nil engine")
	}

	messages := append(al.Messages(), core.NewUserMessage(userMessage))
	temperature := 0.3
	result, err := al.infer(ctx, inference.Request{
		Messages:    messages,
		Schema:      &schema,
		Grammar:     al.cfg.Grammar,
		MaxTokens:   al.cfg.MaxTokens,
		Temperature: &temperature,
		Options: map[string]any{
			"enable_thinking": false,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("agent loop structured chat: %w", err)
	}

	if json.Valid([]byte(result.Content)) {
		result.Content = constraint.NormalizeEnumValues(result.Content, schema)
		return al.validateSchema(result, schema)
	}

	repaired := constraint.RepairJSON(result.Content)
	if !json.Valid([]byte(repaired)) {
		return nil, fmt.Errorf("agent loop structured output: invalid JSON after repair: %s", repaired)
	}

	result.Content = constraint.NormalizeEnumValues(repaired, schema)
	return al.validateSchema(result, schema)
}

// validateSchema runs SchemaValidator if schema has constraints.
func (al *AgentLoop) validateSchema(result *inference.Result, schema core.Schema) (*inference.Result, error) {
	sv := constraint.NewSchemaValidator(schema)
	vr := sv.Validate(result.Content)
	if vr.Valid {
		return result, nil
	}
	return nil, fmt.Errorf("agent loop structured output: schema validation failed: %v", vr.Errors)
}

// Messages returns copy of conversation state.
func (al *AgentLoop) Messages() []core.Message {
	if al == nil || al.memory == nil {
		return nil
	}
	return al.memory.Messages()
}

func (al *AgentLoop) infer(ctx context.Context, req inference.Request) (*inference.Result, error) {
	engine := al.cfg.Engine
	if withStream, ok := engine.(interface {
		WithStreamHandler(*inference.StreamHandler) inference.Engine
	}); ok {
		engine = withStream.WithStreamHandler(al.cfg.Stream)
	}

	started := time.Now()
	result, err := engine.Infer(ctx, req)
	al.recordEvent(started, "infer", map[string]any{
		"messages": len(req.Messages),
		"tools":    len(req.Tools),
	}, err)
	return result, err
}

func (al *AgentLoop) executeTool(call core.ToolCall) (string, error) {
	started := time.Now()
	output, err := al.cfg.Tools.Execute(call.Name, call.Arguments)
	al.recordEvent(started, "tool", map[string]any{
		"name": call.Name,
	}, err)
	return output, err
}

func (al *AgentLoop) replaceMessages(messages []core.Message) {
	al.memory.Clear()
	for _, msg := range messages {
		al.memory.Append(msg)
	}
}

func (al *AgentLoop) recordEvent(started time.Time, action string, data map[string]any, err error) {
	if al == nil || al.cfg.Observer == nil {
		return
	}

	al.cfg.Observer.Record(observe.Event{
		Timestamp: started,
		Layer:     "orchestrate",
		Action:    action,
		Data:      data,
		Duration:  time.Since(started),
		Error:     err,
	})
}

// fetchContext retrieves context messages without storing them in the buffer.
// Returns nil on error (graceful degradation) and records an observer event.
func (al *AgentLoop) fetchContext(ctx context.Context, query string) []core.Message {
	if al.cfg.ContextProvider == nil {
		return nil
	}

	started := time.Now()
	msgs, err := al.cfg.ContextProvider.Build(ctx, query)
	al.recordEvent(started, "context", map[string]any{"query": query, "chunks": len(msgs)}, err)
	if err != nil {
		return nil
	}

	return msgs
}

// memoryWithEphemeralContext builds a request message list by inserting ctxMsgs
// before the last message (user message) in the buffer. The buffer is not modified.
func (al *AgentLoop) memoryWithEphemeralContext(ctxMsgs []core.Message) []core.Message {
	memMsgs := al.memory.Messages()
	if len(ctxMsgs) == 0 {
		return memMsgs
	}

	n := len(memMsgs)
	if n == 0 {
		return cloneMessages(ctxMsgs)
	}

	result := make([]core.Message, 0, n+len(ctxMsgs))
	result = append(result, memMsgs[:n-1]...)
	result = append(result, ctxMsgs...)
	result = append(result, memMsgs[n-1])
	return result
}

// storeResult rebuilds memory from base messages plus the assistant response in
// result. It ignores result.Messages (which may include ephemeral context from
// the request) and uses only result.Content and result.ToolCalls.
func (al *AgentLoop) storeResult(base []core.Message, result *inference.Result) {
	msgs := cloneMessages(base)
	if result == nil {
		al.replaceMessages(msgs)
		return
	}

	assistant := core.Message{
		Role:      core.RoleAssistant,
		Content:   result.Content,
		ToolCalls: cloneToolCalls(result.ToolCalls),
	}
	if assistant.Content != "" || len(assistant.ToolCalls) > 0 {
		msgs = append(msgs, assistant)
	}
	al.replaceMessages(msgs)
}

// injectContext adds messages directly into conversation buffer.
// Used by SpecializedLoop to pre-load context into a fresh inner AgentLoop
// before running Chat or ChatStructured. Safe because SpecializedLoop creates
// a fresh inner loop per call, so there is no multi-turn accumulation.
func (al *AgentLoop) injectContext(messages []core.Message) {
	for _, msg := range messages {
		al.memory.Append(msg)
	}
}

func cloneMessages(messages []core.Message) []core.Message {
	if len(messages) == 0 {
		return nil
	}

	cloned := make([]core.Message, len(messages))
	for i, msg := range messages {
		cloned[i] = msg
		cloned[i].ToolCalls = cloneToolCalls(msg.ToolCalls)
	}
	return cloned
}

func cloneToolCalls(calls []core.ToolCall) []core.ToolCall {
	if len(calls) == 0 {
		return nil
	}

	cloned := make([]core.ToolCall, len(calls))
	for i, call := range calls {
		cloned[i] = call
		cloned[i].Arguments = cloneArgs(call.Arguments)
	}
	return cloned
}

func cloneArgs(args map[string]any) map[string]any {
	if len(args) == 0 {
		return nil
	}

	cloned := make(map[string]any, len(args))
	for key, value := range args {
		cloned[key] = value
	}
	return cloned
}
