# Pipeline Query Example Plan

## Goal

Demonstrate `InferHandler`, `RerankHandler`, and `ValidateHandler` constructors from
`layers/02_orchestrate/handlers.go`. These are the built-in `StepHandler` factories
for `ExecutionPlan` pipelines.

## Deadcode Covered

- `InferHandler` (handlers.go:13)
- `RerankHandler` (handlers.go:38)
- `ValidateHandler` (handlers.go:88)
- `resolveMessages` (loop.go:305) - unexported, becomes reachable via InferHandler execution
- `SystemError.Error` (contracts.go:100) - becomes reachable via error paths
- `SystemError.Unwrap` (contracts.go:108) - becomes reachable via error paths

## Difference from rerank-query

The existing `rerank-query` example builds custom `StepHandler` functions inline.
This example uses the built-in handler constructors (`InferHandler()`, `RerankHandler()`,
`ValidateHandler()`), showing the intended API for composing pipelines.

## Pipeline Design

```
Retrieve -> Rerank -> Infer -> Validate
```

### Type Flow (important)

Each handler has specific input/output types:

1. **RetrieveHandler** - Input: `string` (query), Output: `[]core.Message`
2. **RerankHandler** - Input: `RerankInput`, Output: `[]RerankCandidate`
3. **InferHandler** - Input: `inference.Request`, Output: `*inference.Result`
4. **ValidateHandler** - Input: `string`, Output: `string`

Steps 2-4 need bridge functions to convert the previous step's output into the
next step's expected input type. These bridges are custom `StepHandler` functions
that wrap the built-in handlers.

### Bridge Functions Needed

- **messages-to-rerank**: Convert `[]core.Message` to `RerankInput`
- **candidates-to-request**: Convert `[]RerankCandidate` to `inference.Request`
- **result-to-string**: Extract `Content` from `*inference.Result` for validation

## Implementation Sketch

```go
// Build handlers from constructors.
inferH := orchestrate.InferHandler(engine)
rerankH := orchestrate.RerankHandler(rerankAdapter)
validateH := orchestrate.ValidateHandler(func(output string) error {
    if !json.Valid([]byte(output)) {
        return fmt.Errorf("invalid JSON: %s", output)
    }
    return nil
})

plan := orchestrate.ExecutionPlan{
    Steps: []orchestrate.Step{
        {Name: "retrieve", Type: orchestrate.StepRetrieve, Input: prompt},
        {Name: "rerank", Type: orchestrate.StepRerank},
        {Name: "infer", Type: orchestrate.StepInfer},
        {Name: "validate", Type: orchestrate.StepValidate},
    },
}

// Custom handlers wrap built-in ones with type conversion.
executor := orchestrate.NewPlanExecutor(map[orchestrate.StepType]orchestrate.StepHandler{
    orchestrate.StepRetrieve: orchestrate.RetrieveHandler(contextProvider),
    orchestrate.StepRerank:   bridgeToRerank(rerankH, prompt),
    orchestrate.StepInfer:    bridgeToInfer(inferH, prompt),
    orchestrate.StepValidate: bridgeToValidate(validateH),
})
```

## Dependencies

- `layers/01_inference` (Engine)
- `layers/02_orchestrate` (ExecutionPlan, handlers)
- `layers/03_context` (InMemoryRetriever, ContextBuilder)
- `layers/06_rerank` (Reranker, SimpleScorer)
- `adapters/cli` (ReadPrompt)
- `kronk` (engine loading)

## Notes

- `resolveMessages` is unexported. It cannot be called from example code directly.
  It becomes reachable if production code paths that call it are themselves exercised.
  Check if `AgentLoop.Chat` or other exported functions call it.
- `SystemError` is returned by orchestration error paths. To exercise it, trigger
  an error (e.g., pass nil engine to InferHandler) and inspect the error chain.
