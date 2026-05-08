// Package tool holds layer 4 deterministic capability.
//
// Layer 4 runs operations outside model. All side effects live here, never in
// model. Package centers on Tool, ToolInfo, ToolType, Registry, and
// ToolExecutor.
//
// Tool defines metadata, schema, execution, and availability checks for one
// capability. Registry groups tools and dispatches execution. ToolExecutor adds
// observability around registry calls.
package tool
