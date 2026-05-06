// Package agent provides a reusable multi-agent framework built on top of
// the llm package. It is not specific to any application and can be used
// in any project that needs LLM agents with tool support.
//
// # Agent types
//
// Agent is the stateful base agent. It maintains conversation history across
// Chat calls. Chat handles the tool-call loop: when the LLM requests tools,
// Agent executes them and feeds results back until the LLM produces a content
// response. ChatStructured makes a one-shot structured output call using the
// agent's accumulated history as context, with the output constrained to match
// a JSON Schema. Thinking is disabled so grammar enforcement starts from the
// first token.
//
// SpecializedAgent is a stateless wrapper around Agent for deterministic
// structured output where the JSON Schema (grammar) is fixed at construction
// time. Each Call creates a fresh internal Agent and runs two phases:
//   - Phase 1 (when Tools configured): Agent.Chat drives the tool loop.
//   - Phase 2: Agent.ChatStructured uses accumulated context + fixed schema.
//
// When no Tools are configured, only Phase 2 runs (single structured call).
// This delegation means ChatStructured is called from production via
// SpecializedAgent, keeping the responsibilities cleanly separated.
//
// Use SpecializedAgent when you know the output schema at construction time
// and want a clean single-call interface with no history management.
// Use Agent directly when you need stateful conversation or per-call schemas.
//
// # Tool system
//
// ToolProvider is the interface agents consume: Definitions() returns the
// OpenAI-style tool list passed to the LLM; Execute dispatches a tool call
// by name and returns the result.
//
// Tool is the interface for individual tools. Each tool provides metadata
// (Info), an OpenAI-style definition (Definition), execution logic (Execute),
// and availability reporting (Available). ToolInfo carries the tool name,
// ToolType (TypeGo or TypeCLI), and a short description.
//
// Registry aggregates Tool instances and implements ToolProvider directly.
// Definitions() returns only available tools (Available() == true).
// AllDefinitions() returns all tools regardless of availability.
// CheckAvailability() reports the status of every registered tool.
//
// # Usage
//
//	// Stateful chat with tools
//	a := agent.New(agent.Config{
//	    SystemPrompt: "You analyze production incidents.",
//	    Engine:       kronkEngine,
//	    Tools:        agent.NewRegistry(myTool1, myTool2),
//	})
//	result, err := a.Chat(ctx, "What caused the spike at 14:32?")
//
//	// Deterministic structured output with optional tool loop
//	a := agent.NewSpecialized(agent.SpecializedConfig{
//	    SystemPrompt: "Classify incidents by severity.",
//	    Engine:       kronkEngine,
//	    Schema: model.D{
//	        "type": "object",
//	        "properties": model.D{
//	            "severity": model.D{"type": "string", "enum": []string{"low", "medium", "high"}},
//	        },
//	        "required": []string{"severity"},
//	    },
//	    Tools: agent.NewRegistry(dataFetcher), // optional: runs tool loop before structured output
//	})
//	result, err := a.Call(ctx, "Classify this incident...")
//	// result.Content is valid JSON matching the schema
package agent
