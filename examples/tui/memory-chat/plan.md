# Memory Chat TUI Example Plan

## Goal

Demonstrate `SessionState`, `WorkingMemory`, and `ConversationBuffer.Len` from
`layers/09_memory`. Multi-turn interactive chat with persistent session state
and transient working memory.

## Deadcode Covered

- `NewSessionState` (session.go:16)
- `SessionState.Set` (session.go:21)
- `SessionState.Get` (session.go:29)
- `SessionState.Delete` (session.go:38)
- `SessionState.Keys` (session.go:46)
- `NewWorkingMemory` (working.go:13)
- `WorkingMemory.Reset` (working.go:18)
- `WorkingMemory.SetResult` (working.go:24)
- `WorkingMemory.GetResult` (working.go:33)
- `ConversationBuffer.Len` (conversation.go:47)

## Why TUI (not CLI)

Memory management is meaningful in multi-turn conversations. Single-shot CLI
has no state to persist. TUI REPL provides the interactive loop needed to
demonstrate:

- State accumulation across turns (SessionState)
- Per-step transient context (WorkingMemory)
- Conversation length tracking (ConversationBuffer.Len)

## Design

### Custom Commands

- `remember <key> <value>` - Store in SessionState via Set
- `recall <key>` - Retrieve from SessionState via Get
- `forget <key>` - Remove from SessionState via Delete
- `keys` - List all session keys via Keys
- `status` - Show conversation length via Len, working memory state

### Working Memory Usage

Each chat turn:
1. Before inference: store turn number in WorkingMemory via SetResult
2. After inference: store result summary in WorkingMemory via SetResult
3. Read back via GetResult for status display
4. Reset between turns via Reset

### Conversation Buffer

AgentREPL internally uses ConversationBuffer. To exercise Len(), wrap the
chat command to report message count after each turn.

## Implementation Sketch

```go
session := memory.NewSessionState()
working := memory.NewWorkingMemory()

repl := tui.NewAgentREPL(tui.AgentConfig{
    Title:        "memory-chat",
    Banner:       "Chat with memory. Commands: remember, recall, forget, keys, status",
    SystemPrompt: "You are a helpful assistant.",
    EngineLoader: engineLoader,
    Registry:     tool.NewRegistry(),
})

repl.AddCommand(tui.Command{
    Name: "remember",
    Run: func(ctx context.Context, args []string) error {
        if len(args) < 2 {
            return fmt.Errorf("usage: remember <key> <value...>")
        }
        session.Set(args[0], strings.Join(args[1:], " "))
        return nil
    },
})

repl.AddCommand(tui.Command{
    Name: "recall",
    Run: func(ctx context.Context, args []string) error {
        if len(args) == 0 {
            return fmt.Errorf("usage: recall <key>")
        }
        val, ok := session.Get(args[0])
        if !ok {
            tui.Warnf("key %q not found", args[0])
            return nil
        }
        tui.Infof("%s = %v", args[0], val)
        return nil
    },
})

// ... forget, keys, status commands
```

## Dependencies

- `layers/09_memory` (SessionState, WorkingMemory, ConversationBuffer)
- `adapters/tui` (AgentREPL, Command)
- `kronk` (engine loading)

## Notes

- ConversationBuffer is internal to AgentLoop. To call Len() from example,
  either create a standalone ConversationBuffer for tracking, or access it
  through orchestrate if exposed.
- SessionState is thread-safe (sync.RWMutex). Good for concurrent access
  but this example is single-threaded.
- WorkingMemory is not thread-safe. Fine for sequential step execution.
