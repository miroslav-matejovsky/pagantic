// Package api defines interface contracts for exposing pagantic as a service.
//
// # ARCHITECTURAL ROLE
//
// This package is an Interface (Shell) adapter. It is a thin boundary between
// external API consumers and the internal execution core. It must remain
// stateless, thin (no business logic), deterministic, and replaceable without
// affecting core behavior.
//
// INTERACTION FLOW
//
//	External HTTP/gRPC Request
//	    -> api Adapter (this package)
//	    -> ExecutionService (orchestrate layer)
//	    -> Response Mapping
//	    -> Structured API Output
//
// # PROHIBITED RESPONSIBILITIES
//
// This package must NOT perform orchestration logic, call inference directly,
// execute tools, enforce output schemas, implement validation beyond input
// contract checks, or construct prompts.
//
// Key types:
//   - Request, RequestValidator
//   - Response
//   - ErrorModel, ErrorCode
//   - StreamChunk, StreamingInterface
//
// TODO: HTTP server implementation.
package api
