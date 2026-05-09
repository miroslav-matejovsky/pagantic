package cli

/*
Package cli implements an external interaction adapter for the Pagantic system.

This package represents a Shell layer responsible for translating a specific
interaction modality into the internal execution model.

GENERAL RESPONSIBILITY

The cli adapter is a thin boundary layer between external user interaction
and the internal system (execution service / orchestration core).

It must:
- Accept input from its interaction channel (CLI, TUI, or API)
- Normalize input into a structured request model
- Invoke the core execution service
- Return structured responses to the caller


ARCHITECTURAL ROLE

This package belongs to the Interface (Shell) layer.

It is NOT part of:
- orchestration
- retrieval / embeddings
- tool execution
- validation / constraints
- prompt construction


CORE ABSTRACTIONS TO IMPLEMENT

The package should define:

1. Transport Handling
   - Input parsing (arguments, commands, HTTP payload, UI state)
   - Output formatting (stdout, UI rendering, JSON responses)

2. Request Mapping
   - Conversion of external input into internal request structures
   - Validation of basic input contract (syntax-level only)

3. Execution Dispatch
   - Interaction with a central ExecutionService / RuntimeFacade

4. Response Mapping
   - Conversion of internal responses into user-facing format


CONSTRAINTS

This package must remain:

- Stateless (except minimal interaction context)
- Thin (no business logic)
- Deterministic (no hidden transformations)
- Replaceable (can be swapped without affecting core system)


PROHIBITED RESPONSIBILITIES

This package must NOT:

- perform orchestration logic
- call inference engine directly
- perform retrieval or embedding operations
- execute tools
- enforce output schemas or grammar
- implement validation or correction logic
- construct prompts


INTERACTION FLOW

External Request
    ↓
cli Adapter (this package)
    ↓
ExecutionService (core system)
    ↓
Response Mapping
    ↓
External Output


GOAL

Ensure that all external interaction channels are:

- consistent
- replaceable
- decoupled from system core
- aligned with defined contracts

This package serves as a strict boundary enforcing separation between
user interaction and system behavior.
*/
