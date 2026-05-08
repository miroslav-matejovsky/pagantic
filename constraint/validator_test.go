package constraint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONValidator_ValidJSONPasses(t *testing.T) {
	validator := NewJSONValidator(false)

	result := validator.Validate(`{"ok": true}`)

	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
	require.Equal(t, `{"ok": true}`, result.Output)
}

func TestJSONValidator_InvalidJSONFails(t *testing.T) {
	validator := NewJSONValidator(false)

	result := validator.Validate(`{"ok":`)

	require.False(t, result.Valid)
	require.Equal(t, []string{"output is not valid JSON"}, result.Errors)
}

func TestJSONValidator_RepairFixesTruncatedJSON(t *testing.T) {
	validator := NewJSONValidator(true)

	result := validator.Validate(`{"ok": true`)

	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
	require.Equal(t, `{"ok": true}`, result.Output)
}

func TestJSONValidator_RepairStillFailsBrokenInput(t *testing.T) {
	validator := NewJSONValidator(true)

	result := validator.Validate(`{"ok":,}`)

	require.False(t, result.Valid)
	require.Equal(t, []string{"output is not valid JSON after repair"}, result.Errors)
}
