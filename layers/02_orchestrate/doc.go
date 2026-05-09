// Package orchestrate holds layer 2 control loop.
//
// Layer 2 drives work across many inference steps. It breaks requests into
// steps, routes work across subsystems, enforces order and state moves, and
// manages retries, branching, and tool loops. System is not one inference
// call. System is controlled multi-step loop.
//
// Depends on constraint (layer 5) for structured output: ChatStructured uses
// RepairJSON and SchemaValidator to enforce valid JSON from model output.
//
// Key types:
//   - AgentLoop
//   - SpecializedLoop
//   - ExecutionPlan
//   - Step
//   - StepExecutor
//   - RoutingStrategy
package orchestrate
