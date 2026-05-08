package api

import (
	"fmt"
	"unicode/utf8"

	"github.com/miroslav-matejovsky/pagantic/core"
)

// Request represents an incoming API request.
type Request struct {
	ID          string
	Messages    []core.Message
	Tools       []core.ToolDefinition
	Schema      *core.Schema
	MaxTokens   int
	Temperature *float64
	Stream      bool
}

// RequestValidator checks incoming requests for validity.
type RequestValidator struct {
	MaxMessageCount  int         // 0 means unlimited
	MaxContentLength int         // per message, 0 means unlimited
	RequiredRoles    []core.Role // messages must start with these roles
}

// Validate checks a request. Returns nil if valid.
func (rv *RequestValidator) Validate(req Request) error {
	if len(req.Messages) == 0 {
		return newValidationError("request must contain at least one message", map[string]any{
			"field": "messages",
		})
	}

	if max := rv.maxMessageCount(); max > 0 && len(req.Messages) > max {
		return newValidationError(fmt.Sprintf("request has %d messages; max is %d", len(req.Messages), max), map[string]any{
			"field": "messages",
			"count": len(req.Messages),
			"max":   max,
		})
	}

	requiredRoles := rv.requiredRoles()
	if len(requiredRoles) > 0 {
		if len(req.Messages) < len(requiredRoles) {
			return newValidationError("request does not satisfy required message roles", map[string]any{
				"field":          "messages",
				"required_roles": requiredRoles,
			})
		}

		for i, role := range requiredRoles {
			if req.Messages[i].Role != role {
				return newValidationError(fmt.Sprintf("message %d must have role %q", i, role), map[string]any{
					"field":    "messages",
					"index":    i,
					"expected": role,
					"actual":   req.Messages[i].Role,
				})
			}
		}
	}

	maxContentLength := rv.maxContentLength()
	for i, msg := range req.Messages {
		if !isValidRole(msg.Role) {
			return newValidationError(fmt.Sprintf("message %d has invalid role %q", i, msg.Role), map[string]any{
				"field": "messages",
				"index": i,
				"role":  msg.Role,
			})
		}

		if maxContentLength > 0 && utf8.RuneCountInString(msg.Content) > maxContentLength {
			return newValidationError(fmt.Sprintf("message %d content exceeds max length %d", i, maxContentLength), map[string]any{
				"field":  "messages",
				"index":  i,
				"length": utf8.RuneCountInString(msg.Content),
				"max":    maxContentLength,
			})
		}
	}

	return nil
}

// maxMessageCount gets message cap. Nil validator means no cap.
func (rv *RequestValidator) maxMessageCount() int {
	if rv == nil {
		return 0
	}

	return rv.MaxMessageCount
}

// maxContentLength gets content cap. Nil validator means no cap.
func (rv *RequestValidator) maxContentLength() int {
	if rv == nil {
		return 0
	}

	return rv.MaxContentLength
}

// requiredRoles gets required start roles. Nil validator means no rule.
func (rv *RequestValidator) requiredRoles() []core.Role {
	if rv == nil {
		return nil
	}

	return rv.RequiredRoles
}

// isValidRole checks known message roles.
func isValidRole(role core.Role) bool {
	switch role {
	case core.RoleSystem, core.RoleUser, core.RoleAssistant, core.RoleTool:
		return true
	default:
		return false
	}
}

// newValidationError wraps validation failure in ErrorModel.
func newValidationError(message string, details map[string]any) error {
	return &ErrorModel{
		Code:    ErrValidation,
		Message: message,
		Details: details,
	}
}
