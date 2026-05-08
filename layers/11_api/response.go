package api

import core "github.com/miroslav-matejovsky/pagantic/layers/00_core"

// Response represents API output.
type Response struct {
	ID        string
	Content   string
	ToolCalls []core.ToolCall
	Usage     core.TokenUsage
	Error     *ErrorModel
}

// IsError returns true if response contains error.
func (r *Response) IsError() bool {
	return r != nil && r.Error != nil
}
