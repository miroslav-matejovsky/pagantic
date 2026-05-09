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
//   - ContextProvider
//   - ExecutionPlan
//   - Step
//   - StepExecutor
//   - RoutingStrategy
//
// ContextProvider is an optional interface that retrieves relevant context
// messages before inference. The pagantic layers/03_context.ContextBuilder
// (typically imported with alias pctx) satisfies it via Go structural typing -
// no explicit interface declaration is needed in that package. Context is
// injected ephemerally per user message in AgentLoop.Chat (not stored in the
// conversation buffer, preventing accumulation across turns). In SpecializedLoop,
// context is retrieved once per Call using the original prompt and injected into
// a fresh inner loop before the tool/structured phases.
package orchestrate
