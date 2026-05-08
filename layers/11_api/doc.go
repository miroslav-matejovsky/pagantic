// Package api holds layer 11 interface boundary.
//
// Layer 11 exposes system as reliable service. It validates incoming requests,
// enforces response schema, and provides stable API surface. Interface must be
// deterministic even when inner system is probabilistic.
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
