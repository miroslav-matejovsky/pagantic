// Package tool holds layer 4 deterministic capability.
//
// Layer 4 runs operations outside model. All side effects live here, never in
// model. Package centers on Tool, ToolInfo, ToolType, Registry, and
// ToolExecutor.
//
// Tool defines metadata, schema, execution, and availability checks for one
// capability. Registry groups tools and dispatches execution. ToolExecutor adds
// observability around registry calls.
//
// # Tool Safety
//
// ToolSafety declares side effect level, idempotency, and required capabilities
// for each tool. Orchestration uses these fields for retry decisions:
// idempotent tools can be retried automatically, non-idempotent failures are
// terminal. SideEffectLevel classifies what external state a tool touches.
// ToolTimeoutPolicy configures per-tool deadline behavior.
package tool
