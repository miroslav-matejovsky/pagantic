package agent

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepairJSON_AlreadyComplete(t *testing.T) {
	in := `{"sentiment": "positive", "confidence": 0.9}`
	out := repairJSON(in)
	require.Equal(t, in, out)
	require.True(t, isValidJSON(out))
}

func TestRepairJSON_MissingClosingBrace(t *testing.T) {
	// Simulates the grammar-constrained output that stops before emitting `}`.
	in := "{\n  \"confidence\": 0.9,\n  \"sentiment\": \"positive\""
	out := repairJSON(in)
	require.True(t, isValidJSON(out), "repaired JSON must unmarshal cleanly: %s", out)
	require.True(t, strings.HasSuffix(out, "\n}"), "closing brace must be on new line: %s", out)
}

func TestRepairJSON_WeirdCommaPlacement(t *testing.T) {
	// Reproduces actual kronk grammar output where `,` appears on its own line.
	in := "{\n  \"confidence\": 0.9,\n  \"explanation\": \"great\"\n  ,\n  \"sentiment\": \"positive\""
	out := repairJSON(in)
	require.True(t, isValidJSON(out), "repaired JSON must unmarshal cleanly: %s", out)
}

func TestRepairJSON_MissingClosingBracket(t *testing.T) {
	in := `{"items": [1, 2, 3`
	out := repairJSON(in)
	require.True(t, isValidJSON(out), "repaired JSON must unmarshal cleanly: %s", out)
}

func TestRepairJSON_EmptyInput(t *testing.T) {
	require.Equal(t, "", repairJSON(""))
}

func TestRepairJSON_TrailingWhitespaceTrimmed(t *testing.T) {
	in := `{"ok": true}` + "\n  "
	out := repairJSON(in)
	require.True(t, isValidJSON(out))
}

// isValidJSON reports whether s is valid JSON.
func isValidJSON(s string) bool {
	var v any
	return json.Unmarshal([]byte(s), &v) == nil
}
