package prompt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstructionSetFormat(t *testing.T) {
	t.Run("formats with header and rules", func(t *testing.T) {
		instructionSet := InstructionSet{
			Name:  "Rules",
			Rules: []string{"rule 1", "rule 2"},
		}

		require.Equal(t, "## Rules\n- rule 1\n- rule 2", instructionSet.Format())
	})

	t.Run("empty rules returns just header", func(t *testing.T) {
		instructionSet := InstructionSet{Name: "Rules"}

		require.Equal(t, "## Rules", instructionSet.Format())
	})

	t.Run("empty name still works", func(t *testing.T) {
		instructionSet := InstructionSet{Rules: []string{"rule 1"}}

		require.Equal(t, "## \n- rule 1", instructionSet.Format())
	})
}
