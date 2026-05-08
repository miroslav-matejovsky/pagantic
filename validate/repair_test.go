package validate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONRepairStrategy_ValidJSON(t *testing.T) {
	s := &JSONRepairStrategy{}

	result, err := s.Repair(context.Background(), `{"key":"value"}`, nil)

	require.NoError(t, err)
	require.Equal(t, `{"key":"value"}`, result)
}

func TestJSONRepairStrategy_RepairedJSON(t *testing.T) {
	s := &JSONRepairStrategy{}

	// Missing closing brace - RepairJSON fixes this
	result, err := s.Repair(context.Background(), `{"key":"value"`, nil)

	require.NoError(t, err)
	require.NotEmpty(t, result)
}

func TestJSONRepairStrategy_UnrepairableReturnsError(t *testing.T) {
	s := &JSONRepairStrategy{}

	// Pure garbage that cannot be repaired to valid JSON
	result, err := s.Repair(context.Background(), "not json at all !!!", nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "repair did not produce valid JSON")
	require.NotEmpty(t, result)
}
