# pagantic

Generic agent framework for experimenting with agent architectures and LLM interactions.
Uses [kronk](https://github.com/ardanlabs/kronk) for LLM engine access and model infrastructure management.
Not intended for production use - learning tool and prototyping base.

## Packages

- **agent** - Multi-agent framework with tool loop, structured output, and tool registry
- **llm** - Streaming response parsing and LLM interaction utilities
- **tui** - Terminal UI primitives, REPL, and agent-harness for CLI apps
- **kronk** - Kronk SDK lifecycle wrapper (install, load, init)

## Examples

- **examples/simple-chat** - Minimal interactive chat REPL
