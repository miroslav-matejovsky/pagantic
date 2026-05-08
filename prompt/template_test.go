package prompt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplateRender(t *testing.T) {
	t.Run("basic variable substitution works", func(t *testing.T) {
		template := Template{
			Raw: "Hello, {{.Name}}!",
			Variables: map[string]string{
				"Name": "Grog",
			},
		}

		rendered, err := template.Render()

		require.NoError(t, err)
		require.Equal(t, "Hello, Grog!", rendered)
	})

	t.Run("missing variable returns error", func(t *testing.T) {
		template := Template{
			Raw:       "Hello, {{.Name}}!",
			Variables: map[string]string{},
		}

		rendered, err := template.Render()

		require.Error(t, err)
		require.Empty(t, rendered)
	})

	t.Run("empty template renders empty string", func(t *testing.T) {
		template := Template{}

		rendered, err := template.Render()

		require.NoError(t, err)
		require.Empty(t, rendered)
	})

	t.Run("template with no variables renders raw text unchanged", func(t *testing.T) {
		template := Template{
			Raw: "plain text",
		}

		rendered, err := template.Render()

		require.NoError(t, err)
		require.Equal(t, "plain text", rendered)
	})
}
