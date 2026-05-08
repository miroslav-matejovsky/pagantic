package prompt

import "strings"

// InstructionSet is a named group of rules injected into system prompt.
type InstructionSet struct {
	Name  string
	Rules []string
}

// Format renders instruction set as text block with header and rules.
func (is InstructionSet) Format() string {
	var builder strings.Builder

	builder.WriteString("## ")
	builder.WriteString(is.Name)

	for _, rule := range is.Rules {
		builder.WriteString("\n- ")
		builder.WriteString(rule)
	}

	return builder.String()
}
