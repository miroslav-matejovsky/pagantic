// Package api defines interface contracts for exposing pagantic as a service.
//
// It validates incoming requests, enforces response schema, and provides
// stable API surface. Interface must be deterministic even when inner
// system is probabilistic.
//
// Key types:
//   - Request
//   - Response
//   - ErrorModel
//   - StreamingInterface
//   - RequestValidator
//
// TODO: HTTP server implementation.
package api
