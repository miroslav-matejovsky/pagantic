package constraint

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepairJSON_AlreadyComplete(t *testing.T) {
	in := `{"sentiment": "positive", "confidence": 0.9}`
	out := RepairJSON(in)
	require.Equal(t, in, out)
	require.True(t, isValidJSON(out))
}

func TestRepairJSON_MissingClosingBrace(t *testing.T) {
	in := "{\n  \"confidence\": 0.9,\n  \"sentiment\": \"positive\""
	out := RepairJSON(in)
	require.True(t, isValidJSON(out), "repaired JSON must unmarshal cleanly: %s", out)
	require.True(t, strings.HasSuffix(out, "\n}"), "closing brace must be on new line: %s", out)
}

func TestRepairJSON_WeirdCommaPlacement(t *testing.T) {
	in := "{\n  \"confidence\": 0.9,\n  \"explanation\": \"great\"\n  ,\n  \"sentiment\": \"positive\""
	out := RepairJSON(in)
	require.True(t, isValidJSON(out), "repaired JSON must unmarshal cleanly: %s", out)
}

func TestRepairJSON_MissingClosingBracket(t *testing.T) {
	in := `{"items": [1, 2, 3`
	out := RepairJSON(in)
	require.True(t, isValidJSON(out), "repaired JSON must unmarshal cleanly: %s", out)
}

func TestRepairJSON_EmptyInput(t *testing.T) {
	require.Equal(t, "", RepairJSON(""))
}

func TestRepairJSON_TrailingWhitespaceTrimmed(t *testing.T) {
	in := `{"ok": true}` + "\n  "
	out := RepairJSON(in)
	require.True(t, isValidJSON(out))
	require.Equal(t, `{"ok": true}`, out)
}

// isValidJSON says if JSON parse works.
func isValidJSON(s string) bool {
	var value any
	return json.Unmarshal([]byte(s), &value) == nil
}
