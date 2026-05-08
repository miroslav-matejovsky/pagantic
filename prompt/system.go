package prompt

import (
	"strings"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
)

// SystemPrompt builds a system message from a base prompt, optional
// instruction sets, and an optional context policy.
type SystemPrompt struct {
	Base         string
	Instructions []InstructionSet
	Policy       *ContextPolicy
}

// Build assembles system prompt into a core.Message.
func (sp *SystemPrompt) Build() core.Message {
	parts := make([]string, 0, len(sp.Instructions)+1)
	if sp.Base != "" {
		parts = append(parts, sp.Base)
	}

	for _, instructionSet := range sp.Instructions {
		parts = append(parts, instructionSet.Format())
	}

	return core.NewSystemMessage(strings.Join(parts, "\n\n"))
}
