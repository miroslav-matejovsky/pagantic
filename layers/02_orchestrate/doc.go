// Package orchestrate holds layer 2 control loop.
//
// Layer 2 drives work across many inference steps. It breaks requests into
// steps, routes work across subsystems, enforces order and state moves, and
// manages retries, branching, and tool loops. System is not one inference
// call. System is controlled multi-step loop.
//
// Depends on constraint (layer 5) for structured output: ChatStructured uses
// RepairJSON and SchemaValidator to enforce valid JSON from model output.
// Grammar support passes GBNF grammar strings through to the inference engine
// for decoder-level output constraints.
//
// # System Contracts
//
// SystemRequest and SystemResponse define the stable boundary API. Adapters
// construct SystemRequest from external input and receive SystemResponse.
// SystemError carries a FailureCategory from the canonical failure taxonomy.
// Mode selects which orchestration pattern handles the request.
//
// # Execution Lifecycle
//
// Every request follows a state machine: INIT -> PLAN -> PREPARE -> EXECUTE ->
// VALIDATE -> COMPLETE (or ERROR / CANCELLED). ExecutionState tracks position.
// ExecutionContext carries per-request data through the lifecycle.
//
// # Planning
//
// Planner creates ExecutionPlan from SystemRequest. PlanExecutor executes
// plans. This separation is an explicit boundary. PlanPolicy constrains
// plan construction. PlanTrace records creation metadata.
//
// # Execution IR (Planned)
//
// StepInput, StepOutput, CandidateIR, and ContextIR are planned IR types.
// Current PlanExecutor chains steps using raw any (Step.Input/Output fields).
// These typed wrappers exist for optional use by callers that want typed
// step boundaries; PlanExecutor itself does not enforce them.
//
// Key types:
//   - AgentLoop
//   - SpecializedLoop
//   - ContextProvider
//   - PromptProvider
//   - ExecutionPlan
//   - PlanExecutor
//   - Planner
//   - StepHandler
//   - Step
//   - RoutingStrategy
//   - CandidateReranker
//   - RedundantLoop
//   - VotingStrategy
//   - SystemRequest
//   - SystemResponse
//   - SystemError
//   - Mode
//   - FailureCategory
//   - LifecycleState
//   - ExecutionState
//   - ExecutionContext
//   - StepInput
//   - StepOutput
//   - CandidateIR
//
// ContextProvider is an optional interface that retrieves relevant context
// messages before inference. The pagantic layers/03_context.ContextBuilder
// (typically imported with alias pctx) satisfies it via Go structural typing -
// no explicit interface declaration is needed in that package. Context is
// injected ephemerally per user message in AgentLoop.Chat (not stored in the
// conversation buffer, preventing accumulation across turns). In SpecializedLoop,
// context is retrieved once per Call using the original prompt and injected into
// a fresh inner loop before the tool/structured phases.
//
// CandidateReranker follows the same structural typing pattern. The
// rerank.Reranker type satisfies it without importing the rerank package.
//
// PromptProvider follows the same structural typing pattern. Prompt layer
// implementations must expose BuildSystemPrompt() (core.Message, error).
// The existing prompt.SystemPrompt.Build() and prompt.Template.Render()
// methods do not satisfy this interface; adapter code bridges between them.
package orchestrate
