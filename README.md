# pagantic

**Probabilistic Agentic Control System**

LLM harness system with deterministic control architecture around probabilistic inference.
Uses [kronk](https://github.com/ardanlabs/kronk) for LLM engine access.
Inspired by [Harness engineering for coding agent users](https://martinfowler.com/articles/harness-engineering.html).

## Architecture

10-layer system with explicit architectural boundaries, adapter-based I/O, and a dedicated engine wrapper.

```
                    +---------------------------+
                    |         Adapters           |
                    |  cli  |  tui  |  api       |
                    +-------+-------+------------+
                            |
                    +-------v-------+
                    |   orchestrate |  Control loop, agent loops
                    +-------+-------+
                            |
        +-------+-------+---+---+-------+-------+
        |       |       |       |       |       |
     context  tool  constraint rerank validate prompt
        |       |       |       |       |       |
        +-------+-------+---+---+-------+-------+
                            |
                    +-------v-------+
                    |   inference   |  Engine interface
                    +-------+-------+
                            |
                    +-------v-------+
                    |     kronk     |  SDK adapter
                    +---------------+
                            |
                    +-------v-------+
                    |     core      |  Shared domain types
                    +---------------+
```

### Layers

Each layer lives under `layers/` with a numeric prefix enforcing dependency direction.

| Layer | Package | Purpose |
|-------|---------|---------|
| 0 | **core** | Shared domain types - Message, ToolCall, Schema, TokenUsage. No dependencies. |
| 1 | **inference** | Execution substrate. Defines Engine interface for model inference. |
| 2 | **orchestrate** | Control loop. AgentLoop for multi-turn chat with tool resolution. SpecializedLoop for schema-constrained single-shot calls. |
| 3 | **context** | Knowledge retrieval and context building. Retriever interface, InMemoryRetriever, ContextBuilder. Integrated with orchestrate via ContextProvider. |
| 4 | **tool** | Tool registry, execution, and availability checking. |
| 5 | **constraint** | Output enforcement. JSON validation, repair, schema validation, enum normalization. |
| 6 | **rerank** | Candidate evaluation and reranking (stub). |
| 7 | **validate** | Guardrails, rule validation, retry policy. |
| 8 | **prompt** | Prompt construction. SystemPrompt builder with composable InstructionSets. |
| 9 | **memory** | State management. ConversationBuffer for message history. |
| 10 | **observe** | Tracing, metrics, event logging via EventLog interface. |

### Adapters

Adapters live under `adapters/` and serve as thin boundary layers between external interaction channels and the internal execution core. Each adapter is stateless, deterministic, and replaceable.

| Adapter | Purpose |
|---------|---------|
| **cli** | Single-shot command-line execution. Reads prompt from args or stdin, runs inference, writes result to stdout, and exits. |
| **tui** | Interactive terminal UI. REPL with command dispatch, agent harness with streaming chat, ANSI colored output, and tool status display. |
| **api** | Service interface contracts. Request/Response types, validation, streaming interface, structured error model. |

### Engine

| Package | Purpose |
|---------|---------|
| **kronk** | Kronk SDK lifecycle wrapper. Handles library installation, model downloading, and inference adapter implementing Engine interface. |

## Examples

- **examples/cli/simple-query** - Single-shot CLI query
- **examples/cli/context-query** - Single-shot CLI query with context retrieval (RAG)
- **examples/cli/grammar-query** - GBNF grammar-constrained output
- **examples/cli/rerank-query** - Query with document reranking via ExecutionPlan
- **examples/cli/redundant-query** - Redundant inference with majority voting (TMR)
- **examples/tui/simple-chat** - Minimal interactive chat REPL
- **examples/tui/tool-use** - Chat with a custom Go tool (dice roller)
- **examples/tui/structured-output** - SpecializedLoop with JSON schema output
- **examples/tui/context-chat** - Interactive chat with per-turn context retrieval (RAG)

## Documentation

Full HTML documentation is available at `docs/index.html`.
