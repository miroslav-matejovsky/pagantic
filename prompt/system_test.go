package prompt

import (
	"testing"

	"github.com/miroslav-matejovsky/pagantic/core"
	"github.com/stretchr/testify/require"
)

func TestSystemPromptBuild(t *testing.T) {
	t.Run("base only", func(t *testing.T) {
		systemPrompt := SystemPrompt{Base: "base prompt"}

		message := systemPrompt.Build()

		require.Equal(t, core.RoleSystem, message.Role)
		require.Equal(t, "base prompt", message.Content)
	})

	t.Run("base plus one instruction set", func(t *testing.T) {
		systemPrompt := SystemPrompt{
			Base: "base prompt",
			Instructions: []InstructionSet{
				{
					Name:  "Rules",
					Rules: []string{"rule 1", "rule 2"},
				},
			},
		}

		message := systemPrompt.Build()

		require.Equal(t, core.RoleSystem, message.Role)
		require.Equal(t, "base prompt\n\n## Rules\n- rule 1\n- rule 2", message.Content)
	})

	t.Run("base plus multiple instruction sets", func(t *testing.T) {
		systemPrompt := SystemPrompt{
			Base: "base prompt",
			Instructions: []InstructionSet{
				{
					Name:  "Rules",
					Rules: []string{"rule 1"},
				},
				{
					Name:  "More",
					Rules: []string{"rule 2"},
				},
			},
		}

		message := systemPrompt.Build()

		require.Equal(t, core.RoleSystem, message.Role)
		require.Equal(t, "base prompt\n\n## Rules\n- rule 1\n\n## More\n- rule 2", message.Content)
	})

	t.Run("returned message has system role", func(t *testing.T) {
		systemPrompt := SystemPrompt{}

		message := systemPrompt.Build()

		require.Equal(t, core.RoleSystem, message.Role)
	})
}
