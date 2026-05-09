# pagantic

**Probabilistic Agentic Control System**

LLM harness system with deterministic control architecture around probabilistic inference.
Uses [kronk](https://github.com/ardanlabs/kronk) for LLM engine access.
Inspired by [Harness engineering for coding agent users](https://martinfowler.com/articles/harness-engineering.html).

## Architecture

10-layer system with explicit architectural boundaries:

- **core** - Shared domain types (Message, ToolCall, Schema, TokenUsage)
- **inference** - Layer 1: Execution substrate, Engine interface
- **orchestrate** - Layer 2: Control loop, AgentLoop, SpecializedLoop
- **context** - Layer 3: Knowledge retrieval, ContextBuilder (stub)
- **tool** - Layer 4: Tool registry and execution
- **constraint** - Layer 5: Output enforcement, JSON validation and repair
- **rerank** - Layer 6: Candidate evaluation and reranking (stub)
- **validate** - Layer 7: Guardrails, rule validation, retry policy
- **prompt** - Layer 8: Prompt construction, templates, instruction sets
- **memory** - Layer 9: State management, conversation buffer
- **observe** - Layer 10: Tracing, metrics, event logging
- **api** - Interface contracts, request/response types
- **kronk** - Kronk SDK lifecycle wrapper and inference adapter
- **tui** - Terminal UI, REPL, streaming renderer

## Examples

- **examples/simple-chat** - Minimal interactive chat REPL
- **examples/tool-use** - Chat with a custom Go tool (dice roller)
- **examples/structured-output** - SpecializedLoop with JSON schema output
