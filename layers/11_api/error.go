package api

import "fmt"

// ErrorCode identifies type of API error.
type ErrorCode string

const (
	ErrValidation    ErrorCode = "validation_error"
	ErrInference     ErrorCode = "inference_error"
	ErrToolExecution ErrorCode = "tool_execution_error"
	ErrTimeout       ErrorCode = "timeout"
	ErrInternal      ErrorCode = "internal_error"
)

// ErrorModel represents structured API error.
type ErrorModel struct {
	Code    ErrorCode
	Message string
	Details map[string]any
}

// Error formats code and message for error interface.
func (e *ErrorModel) Error() string {
	if e == nil {
		return ""
	}

	if e.Code == "" {
		return e.Message
	}

	if e.Message == "" {
		return string(e.Code)
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
